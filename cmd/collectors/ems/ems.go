package ems

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/ems/metrictransformer"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slice"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

const defaultDataPollDuration = 3 * time.Minute
const maxURLSize = 8_000 // bytes
const severityFilterPrefix = "message.severity="
const defaultSeverityFilter = "alert|emergency|error|informational|notice"

// defaultMaxLookback caps how far back the collector asks for events.
// lastFilterTime only advances on a successful poll, so without a cap the
// window widens by one poll interval after every failure.
const defaultMaxLookback = 15 * time.Minute

// emsVisibilityLag holds the end of the window back from cluster time. An event
// becomes queryable a moment after its own timestamp, so closing the window at
// exactly cluster time would drop events stamped just before the boundary - the
// next poll starts at that boundary and would never ask for them again.
const emsVisibilityLag = 5 * time.Second

// defaultEmsBatchSize is larger than collectors.DefaultBatchSize because the
// cost of this endpoint is per request, not per record.
const defaultEmsBatchSize = "10000"

const defaultMaxRecords = 20_000
const MaxBookendInstances = 1000
const DefaultBookendResolutionDuration = 28 * 24 * time.Hour // 28 days == 672 hours
const Hyphen = "-"
const AutoResolved = "autoresolved"

type Ems struct {
	*rest2.Rest    // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	Query          string
	TemplatePath   string
	emsProp        map[string][]*emsProp // array is used here to handle same ems written with different ops, matches or exports. Example: arw.volume.state ems with op as disabled or dry-run
	Filter         []string
	Fields         []string
	ReturnTimeOut  *int
	lastFilterTime int64
	maxURLSize     int
	DefaultLabels  []string
	severityFilter string
	eventNames     []string                 // consist of all ems events supported
	bookendEmsMap  map[string]*set.Set      // This is reverse bookend ems map, [Resolving ems]:[Set of Issuing ems]. Using Set here to ensure that it has slice of unique issuing ems
	resolveAfter   map[string]time.Duration // This is resolve after map, [Issuing ems]:[Duration]. After this duration, ems got auto resolved.
	batchSize      string
	maxLookback    time.Duration
	maxRecords     int
	walkBudget     time.Duration
}

type Metric struct {
	Label      string
	Name       string
	MetricType string
	Exportable bool
}

type Matches struct {
	Name  string
	value string
}

// missingLabelKey identifies a template label that an ONTAP release does not
// send for a given event.
type missingLabelKey struct {
	ems   string
	label string
}

type missingLabelCount struct {
	count   int
	example string
}

type emsProp struct {
	Name           string
	InstanceKeys   []string
	InstanceLabels map[string]string
	Metrics        map[string]*Metric
	Plugins        []plugin.Plugin // built-in or custom plugins
	Matches        []*Matches
	Labels         map[string]string
}

func init() {
	plugin.RegisterModule(&Ems{})
}

func (e *Ems) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.ems",
		New: func() plugin.Module { return new(Ems) },
	}
}

func (e *Ems) InitEmsProp() {
	e.emsProp = make(map[string][]*emsProp)
}

func (e *Ems) Init(a *collector.AbstractCollector) error {

	var err error

	e.Rest = &rest2.Rest{AbstractCollector: a}
	e.Fields = []string{"*"}
	e.maxURLSize = maxURLSize
	e.severityFilter = severityFilterPrefix + defaultSeverityFilter

	// init Rest props
	e.InitProp()
	// init ems props
	e.InitEmsProp()

	e.bookendEmsMap = make(map[string]*set.Set)
	e.resolveAfter = make(map[string]time.Duration)

	if err := e.InitClient(); err != nil {
		return err
	}

	if e.TemplatePath, err = e.LoadTemplate(); err != nil {
		return err
	}

	if err := collector.Init(e); err != nil {
		return err
	}

	if err := e.InitCache(); err != nil {
		return err
	}

	return e.InitMatrix()
}

func (e *Ems) InitMatrix() error {
	mat := e.Matrix[e.Object]
	// overwrite from abstract collector
	mat.Object = e.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", e.Remote.Name)
	mat.SetGlobalLabel("cluster_uuid", e.Remote.UUID)

	if e.Params.HasChildS("labels") {
		for _, l := range e.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}
	return nil
}

func (e *Ems) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "MetricTransformer":
		return metrictransformer.New(abc)
	default:
		e.Logger.Warn("no ems plugin found", slog.String("kind", kind))
	}
	return nil
}

func (e *Ems) InitCache() error {

	var (
		events *node.Node
	)

	if x := e.Params.GetChildContentS("object"); x != "" {
		e.Prop.Object = x
	} else {
		e.Prop.Object = strings.ToLower(e.Object)
	}

	e.maxURLSize = e.LoadParam("max_url_size", e.maxURLSize)
	e.maxRecords = e.LoadParam("max_records", defaultMaxRecords)
	if e.maxRecords <= 0 {
		e.Logger.Warn("max_records must be positive, using default",
			slog.Int("max_records", e.maxRecords),
			slog.Int("default", defaultMaxRecords),
		)
		e.maxRecords = defaultMaxRecords
	}

	// Ems overrides Rest.InitCache, so the batch_size handling there does not
	// run for this collector and has to be repeated here.
	e.batchSize = defaultEmsBatchSize
	if b := e.Params.GetChildContentS("batch_size"); b != "" {
		// Atoi alone accepts 0 and negatives, which would ask ONTAP for an
		// empty or nonsensical page.
		if n, err := strconv.Atoi(b); err == nil && n > 0 {
			e.batchSize = b
		} else {
			e.Logger.Warn("Invalid value of batch_size, using default",
				slog.String("batch_size", b),
				slog.String("default", defaultEmsBatchSize),
			)
		}
	}

	e.maxLookback = defaultMaxLookback
	if m := e.Params.GetChildContentS("max_lookback"); m != "" {
		// Must exceed emsVisibilityLag. At or below it the clamp pulls
		// fromTime to at least toTime, every window comes out empty, and the
		// collector silently stops collecting anything at all.
		if d, err := time.ParseDuration(m); err == nil && d > emsVisibilityLag {
			e.maxLookback = d
		} else {
			e.Logger.Warn("Invalid value of max_lookback, using default",
				slogx.Err(err),
				slog.String("max_lookback", m),
				slog.String("minimum", emsVisibilityLag.String()),
				slog.String("default", defaultMaxLookback.String()),
			)
		}
	}

	// One collection must fit inside one poll interval. Overrunning it does not
	// buy fresher data - it just backs polls up behind each other.
	budget, err := collectors.GetDataInterval(e.GetParams(), defaultDataPollDuration)
	if err != nil {
		e.Logger.Warn("Failed to parse duration. using default",
			slogx.Err(err),
			slog.String("defaultDataPollDuration", defaultDataPollDuration.String()),
		)
		budget = defaultDataPollDuration
	}
	e.walkBudget = budget

	if s := e.Params.GetChildContentS("severity"); s != "" {
		e.severityFilter = severityFilterPrefix + s
	}

	if export := e.Params.GetChildS("export_options"); export != nil {
		e.Matrix[e.Object].SetExportOptions(export)
	}

	if e.Query = e.Params.GetChildContentS("query"); e.Query == "" {
		return errs.New(errs.ErrMissingParam, "query")
	}

	// Used for autosupport
	e.Prop.Query = e.Query

	if exports := e.Params.GetChildS("exports"); exports != nil {
		for _, line := range exports.GetChildren() {
			if line != nil {
				e.DefaultLabels = append(e.DefaultLabels, line.GetContentS())
			}
		}
	}

	if events = e.Params.GetChildS("events"); events == nil || len(events.GetChildren()) == 0 {
		return errs.New(errs.ErrMissingParam, "events")
	}

	// default value for ONTAP is 15 sec
	if returnTimeout := e.Params.GetChildContentS("return_timeout"); returnTimeout != "" {
		iReturnTimeout, err := strconv.Atoi(returnTimeout)
		if err != nil {
			e.Logger.Warn("Invalid value of returnTimeout", slog.String("returnTimeout", returnTimeout))
		} else {
			e.ReturnTimeOut = &iReturnTimeout
		}
	}

	// init plugins
	if e.Plugins == nil {
		e.Plugins = make(map[string][]plugin.Plugin)
	}

	for _, line := range events.GetChildren() {
		prop := emsProp{
			InstanceKeys:   make([]string, 0),
			InstanceLabels: make(map[string]string),
			Metrics:        make(map[string]*Metric),
		}

		// check if name is present in template
		if line.GetChildContentS("name") == "" {
			e.Logger.Error("Missing event name")
			continue
		}

		// populate prop counter for asup
		eventName := line.GetChildContentS("name")
		e.Prop.Counters[eventName] = eventName

		e.ParseDefaults(&prop)

		for _, line1 := range line.GetChildren() {
			if line1.GetNameS() == "name" {
				prop.Name = line1.GetContentS()
			}
			if line1.GetNameS() == "exports" {
				e.ParseExports(line1, &prop)
			}
			if line1.GetNameS() == "matches" {
				e.ParseMatches(line1, &prop)
			}
			if line1.GetNameS() == "labels" {
				e.ParseLabels(line1, &prop)
			}
			if line1.GetNameS() == "plugins" {
				if err := e.LoadPlugins(line1, e, prop.Name); err != nil {
					e.Logger.Error("Failed to load plugin", slogx.Err(err))
				}
			}
			if line1.GetNameS() == "resolve_when_ems" {
				e.ParseResolveEms(line1, prop)
			}
		}
		e.emsProp[prop.Name] = append(e.emsProp[prop.Name], &prop)
	}
	// add severity filter
	e.Filter = append(e.Filter, e.severityFilter)

	// Logged once at startup rather than per poll. These are the settings that
	// determine how much work a collection does, so they belong in any log
	// shipped with a report of EMS collection being slow.
	e.Logger.Info(
		"EMS collector settings",
		slog.String("batchSize", e.batchSize),
		slog.Int("maxRecords", e.maxRecords),
		slog.Duration("maxLookback", e.maxLookback),
		slog.Duration("walkBudget", e.walkBudget),
		slog.Int("eventsInTemplate", len(e.emsProp)),
		slog.String("severityFilter", e.severityFilter),
	)

	return nil
}

// timeWindow returns the time filter for this poll plus the window it covers.
//
// Only the lower bound goes to ONTAP. This endpoint rejects a two-sided filter
// on `time` outright - 400, code 262188, target `time`, "Field \"time\" was
// specified twice".
// fetchEMSData enforces the end of the window client-side instead.
//
// The window still needs an end. Without one the watermark could only advance
// to the newest record that happened to be seen, which is not the same as the
// window being complete. Discarding the records past toTime costs
// little: they are the few that arrive during the request itself.
func (e *Ems) timeWindow(clusterTime time.Time) ([]string, int64, int64) {
	toTime := clusterTime.Add(-emsVisibilityLag).Unix()
	fromTime := e.lastFilterTime

	// check if this is the first request
	if fromTime == 0 {
		// if first request fetch cluster time
		dataDuration, err := collectors.GetDataInterval(e.GetParams(), defaultDataPollDuration)
		if err != nil {
			e.Logger.Warn(
				"Failed to parse duration. using default",
				slogx.Err(err),
				slog.String("defaultDataPollDuration", defaultDataPollDuration.String()),
			)
		}
		fromTime = clusterTime.Add(-dataDuration).Unix()
	}

	if oldest := clusterTime.Add(-e.maxLookback).Unix(); fromTime < oldest {
		e.Logger.Warn(
			"EMS watermark is older than max_lookback, skipping ahead. Events in the gap are not collected",
			slog.Time("watermark", time.Unix(fromTime, 0)),
			slog.Duration("gap", clusterTime.Sub(time.Unix(fromTime, 0))),
			slog.Duration("maxLookback", e.maxLookback),
		)
		fromTime = oldest
	}

	return []string{fmt.Sprintf("time=>=%d", fromTime)}, fromTime, toTime
}

// sortRecords orders events chronologically, so that when an issuing ems and
// the ems that resolves it arrive in the same poll, the resolving one is
// processed last and the issue is cleared. HandleResults depends on that order.
//
// This was previously done with order_by=index asc in the query. Two reasons it
// moved here, measured on a 32-node cluster against the collector's own query
// shape over three interleaved rounds:
//
//   - ONTAP caps a sorted page at 5000 records while an unsorted one returns up
//     10000.
//   - The sort itself added ~17% to each request (10.3/17.7/3.1s with it,
//     7.5/16.3/2.7s without).
//
// Sorting here costs nothing and preserves the ordering the bookend logic
// needs.
//
// time is the primary key rather than index. index is a per-node sequence
// number, so it does not order events against each other across nodes; on a
// 32-node cluster sorting by it alone is close to arbitrary. index remains the
// tiebreaker for events sharing a timestamp.
func sortRecords(records []gjson.Result) {
	slices.SortStableFunc(records, func(a, b gjson.Result) int {
		return cmp.Or(
			cmp.Compare(eventTime(a), eventTime(b)),
			cmp.Compare(a.Get("index").Int(), b.Get("index").Int()),
		)
	})
}

// inWindow reports whether a record falls inside the window ending at toTime.
// A zero timestamp means the record's time could not be read; it is kept
// rather than silently dropped over an unparseable field.
func inWindow(record gjson.Result, toTime int64) bool {
	t := eventTime(record)
	return t == 0 || t <= toTime
}

// nextWatermark returns the lower bound for the poll following one that covered
// a window ending at toTime. inWindow keeps records at exactly toTime and the
// query is inclusive (time=>=), so the next window must open one second later
// or that second's events are collected and handled twice. These two functions
// are the whole window boundary and have to stay consistent with each other.
func nextWatermark(toTime int64) int64 {
	return toTime + 1
}

// eventTime returns an event's timestamp as epoch seconds. ONTAP renders `time`
// as an RFC3339 string on this endpoint, but accepts epoch seconds in filters,
// so both forms are handled. Returns 0 when the timestamp cannot be read, which
// callers treat as "keep the record" rather than silently discarding it.
func eventTime(record gjson.Result) int64 {
	t := record.Get("time")
	if !t.Exists() {
		return 0
	}
	if secs := t.Int(); secs > 0 {
		return secs
	}
	return int64(collectors.HandleTimestamp(t.ClonedString()))
}

// errWalkBudget stops the pagination walk once the record cap or the deadline
// is reached. It never escapes fetchEMSData.
var errWalkBudget = errors.New("ems walk budget exhausted")

// fetchEMSData pages through href, dropping records past the end of the window
// and stopping early once maxRecords or the deadline is reached.
//
// Partial results are returned rather than discarded.
func (e *Ems) fetchEMSData(href string, deadline time.Time, toTime int64, quota int) ([]gjson.Result, bool, error) {
	var (
		records   []gjson.Result
		truncated bool
	)

	err := rest.FetchAllStream(e.Client, &e.RequestMetadata, href, func(batch []gjson.Result, _ int64) error {
		for _, record := range batch {
			// Client-side end bound. This endpoint rejects a two-sided filter
			// on `time`, so the end of the window is enforced here - see
			// timeWindow.
			if !inWindow(record, toTime) {
				continue
			}
			records = append(records, record)
		}

		// quota is what is left of the poll-wide record cap, not a per-href
		// allowance, so several hrefs cannot together collect a multiple of
		// max_records.
		//
		// The time check refuses to *begin* a page unless a whole
		// client_timeout still fits inside the budget. The REST client takes
		// no context, so a request already in flight cannot be cut short;
		// without this a page starting just before the deadline could overrun
		// the poll interval by the full timeout, which is the pile-up the
		// budget exists to prevent.
		if len(records) >= quota || time.Until(deadline) < e.Client.GetTimeout() {
			truncated = true
			return errWalkBudget
		}
		return nil
	})

	switch {
	case errors.Is(err, errWalkBudget):
		return records, truncated, nil
	case errors.Is(err, errs.ErrNoInstance):
		// no events in the window is normal, not a failure
		return nil, false, nil
	case err != nil:
		return nil, false, err
	}

	return records, truncated, nil
}

// buildHrefs splits the event-name list into as many queries as maxURLSize allows.
func (e *Ems) buildHrefs(filter []string) ([]string, error) {
	var hrefs []string
	start := 0
	for end := 0; end < len(e.eventNames); end++ {
		h := e.getHref(e.eventNames[start:end], filter)
		if len(h) > e.maxURLSize {
			if end == 0 {
				return nil, fmt.Errorf("maxURLSize=%d is too small to form queries. Increase it to at least %d",
					e.maxURLSize, len(h))
			}
			end--
			h = e.getHref(e.eventNames[start:end], filter)
			hrefs = append(hrefs, h)
			start = end
		} else if end == len(e.eventNames)-1 {
			end = len(e.eventNames)
			h = e.getHref(e.eventNames[start:end], filter)
			hrefs = append(hrefs, h)
		}
	}
	return hrefs, nil
}

// hrefQuota returns how many records the href at index i may collect, given
// what the poll has already taken.
//
// The record cap is divided between the remaining hrefs rather than offered to
// each in turn. The event names are split across hrefs in a stable order, so a
// first-come allowance lets one flooding event name consume the whole cap and
// leave the trailing names unqueried - permanently, since the split does not
// change between polls. A rare LUN.offline in the last chunk would go
// uncollected because wafl.vol.autoSize.done is noisy in the first.
//
// Unused allowance rolls forward: an href that returns less than its share
// leaves the remainder to those after it, so dividing the cap costs nothing
// when no href is flooding.
func hrefQuota(maxRecords, taken, i, hrefs int) int {
	remaining := maxRecords - taken
	if remaining <= 0 {
		return 0
	}
	return max(1, remaining/(hrefs-i))
}

// collect walks every href within a shared record cap and time budget.
func (e *Ems) collect(hrefs []string, deadline time.Time, toTime int64) ([]gjson.Result, bool, error) {
	var (
		records   []gjson.Result
		truncated bool
		skipped   int
	)

	for i, h := range hrefs {
		quota := hrefQuota(e.maxRecords, len(records), i, len(hrefs))
		if quota <= 0 {
			// Cap reached. Remaining hrefs are not queried at all, so say how
			// many rather than reporting only that something was dropped.
			skipped = len(hrefs) - i
			truncated = true
			break
		}

		r, t, err := e.fetchEMSData(h, deadline, toTime, quota)
		if err != nil {
			return nil, false, err
		}
		records = append(records, r...)

		if t && !time.Now().Before(deadline.Add(-e.Client.GetTimeout())) {
			// Out of time, not just out of quota: there is no room to start a
			// request for the hrefs after this one either.
			skipped = len(hrefs) - i - 1
			truncated = true
			break
		}
		if t {
			// This href hit its share of the cap. Keep going: the others are
			// entitled to theirs, and time remains to ask for it.
			truncated = true
		}
	}

	if skipped > 0 {
		e.Logger.Warn(
			"EMS budget spent before every event-name query ran",
			slog.Int("queriesSkipped", skipped),
			slog.Int("queries", len(hrefs)),
		)
	}

	return records, truncated, nil
}

// PollInstance queries the cluster's EMS catalog and intersects that catalog with the EMS template.
// This is required because ONTAP EMS Rest endpoint fails when queried for an EMS message that does not exist.
func (e *Ems) PollInstance() (map[string]*matrix.Matrix, error) {
	var (
		err              error
		records          []gjson.Result
		bookendCacheSize int
	)

	query := "api/support/ems/messages"
	fields := []string{"name"}

	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		ReturnTimeout(e.ReturnTimeOut).
		Build()

	apiT := time.Now()
	if records, err = e.GetRestData(href); err != nil {
		return nil, err
	}
	apiD := time.Since(apiT)

	parseT := time.Now()
	if len(records) == 0 {
		return nil, errs.New(errs.ErrNoInstance, e.Object+" no ems message found on cluster")
	}

	var emsEventCatalogue []string
	for _, instanceData := range records {
		name := instanceData.Get("name")
		if name.Exists() {
			emsEventCatalogue = append(emsEventCatalogue, name.ClonedString())
		}
	}

	// collect all event names
	names := make([]string, 0, len(e.emsProp))
	for key := range e.emsProp {
		names = append(names, key)
	}

	// Filter out names which exist on the cluster.
	// ONTAP rest ems throws error for a message.name filter if that event is not supported by that cluster
	filteredNames, _ := slice.Intersection(names, emsEventCatalogue)
	_, missingNames := slice.Intersection(filteredNames, names)
	e.Logger.Debug("filtered ems events", slog.Any("skipped events", missingNames))
	e.eventNames = filteredNames

	// warning when total instance in cache > 1000 instance
	for _, issuingEmsList := range e.bookendEmsMap {
		for _, issuingEms := range issuingEmsList.Slice() {
			if mx := e.Matrix[issuingEms]; mx != nil {
				bookendCacheSize += len(mx.GetInstances())
			}
		}
	}

	e.Logger.Info("", slog.Int("total instances", bookendCacheSize))
	// warning when total instance in cache > 1000 instance
	if bookendCacheSize > MaxBookendInstances {
		e.Logger.Warn("cache has more than 1000 instances", slog.Int("total instances", bookendCacheSize))
	}

	// update metadata for collector logs
	instanceInst := e.Metadata.MustGetInstance("instance")
	e.Metadata.MustSetValueInt64("api_time", instanceInst, apiD.Microseconds())
	e.Metadata.MustSetValueInt64("parse_time", instanceInst, time.Since(parseT).Microseconds())
	e.Metadata.MustSetValueUint64("instances", instanceInst, uint64(bookendCacheSize))

	return nil, nil
}

func (e *Ems) PollData() (map[string]*matrix.Matrix, error) {

	var (
		count, instanceCount uint64
		apiD, parseD         time.Duration
		startTime            time.Time
		err                  error
		records              []gjson.Result
	)

	// Update cache for bookend ems
	e.updateMatrix(time.Now())

	e.RequestMetadata.Reset()

	startTime = time.Now()

	// add time filter
	clusterTime, err := collectors.GetClusterTime(e.Client, e.ReturnTimeOut, e.Logger)
	if err != nil {
		return nil, err
	}
	timeFilters, fromTime, toTime := e.timeWindow(clusterTime)
	if toTime <= fromTime {
		// Nothing to ask for yet. Happens when the poll interval is shorter
		// than emsVisibilityLag, or when cluster time moved backwards.
		e.Logger.Debug("EMS window is empty, skipping poll",
			slog.Time("from", time.Unix(fromTime, 0)),
			slog.Time("to", time.Unix(toTime, 0)),
		)
		return e.Matrix, nil
	}

	// Copy rather than append to e.Filter directly: append would write through
	// to the shared backing array whenever it has spare capacity.
	filter := make([]string, 0, len(e.Filter)+len(timeFilters))
	filter = append(filter, e.Filter...)
	filter = append(filter, timeFilters...)

	// build hrefs up to maxURLSize
	hrefs, err := e.buildHrefs(filter)
	if err != nil {
		return nil, err
	}

	deadline := startTime.Add(e.walkBudget)
	records, truncated, err := e.collect(hrefs, deadline, toTime)
	if err != nil {
		return nil, err
	}

	if truncated {
		// recordsPerSec is the throughput ONTAP actually delivered. The
		// endpoint is served under a throughput budget, so this is the number
		// that says whether a window is collectable at all: a window holding
		// more records than budget x throughput can never be finished,
		// regardless of how the request is shaped.
		var recordsPerSec float64
		if elapsed := time.Since(startTime).Seconds(); elapsed > 0 {
			recordsPerSec = float64(len(records)) / elapsed
		}
		e.Logger.Warn(
			"EMS collection hit its budget, some events in this window were not collected",
			slog.Int("records", len(records)),
			slog.Int("maxRecords", e.maxRecords),
			slog.Duration("budget", e.walkBudget),
			slog.Float64("recordsPerSec", recordsPerSec),
			slog.Time("from", time.Unix(fromTime, 0)),
			slog.Time("to", time.Unix(toTime, 0)),
		)
	}

	apiD = time.Since(startTime)

	startTime = time.Now()
	sortRecords(records)
	_, count, instanceCount = e.HandleResults(records, e.emsProp)

	parseD = time.Since(startTime)

	dataInst := e.Metadata.MustGetInstance("data")
	e.Metadata.MustSetValueInt64("api_time", dataInst, apiD.Microseconds())
	e.Metadata.MustSetValueInt64("parse_time", dataInst, parseD.Microseconds())
	e.Metadata.MustSetValueUint64("metrics", dataInst, count)
	e.Metadata.MustSetValueUint64("instances", dataInst, instanceCount)
	// numCalls is the page count for this poll, which is the number that shows
	// whether batch_size is actually reducing round trips.
	e.Metadata.MustSetValueUint64("bytesRx", dataInst, e.RequestMetadata.BytesRx.Load())
	e.Metadata.MustSetValueUint64("numCalls", dataInst, e.RequestMetadata.NumCalls.Load())

	e.AddCollectCount(count)

	// The watermark advances even when the walk was truncated: holding it back
	// would only widen the next window and make the next poll slower.
	e.lastFilterTime = nextWatermark(toTime)
	return e.Matrix, nil
}

func (e *Ems) getHref(names []string, filter []string) string {
	// Copy: getHref is called repeatedly with the same filter slice while
	// hrefs are being sized. Appending in place would let one call's
	// name filter leak into the next.
	f := make([]string, 0, len(filter)+2)
	f = append(f, filter...)

	nameFilter := "message.name=" + strings.Join(names, ",")
	f = append(f, nameFilter)
	// Deliberately no order_by. The ordering the bookend logic needs is applied
	// by sortRecords after collection instead - see there for why.

	href := rest.NewHrefBuilder().
		APIPath(e.Query).
		Fields(e.Fields).
		Filter(f).
		MaxRecords(e.batchSize).
		ReturnTimeout(e.ReturnTimeOut).
		Build()
	return href
}

func parseProperties(instanceData gjson.Result, property string) gjson.Result {

	if !strings.HasPrefix(property, "parameters.") {
		// if prefix is not parameters.
		value := gjson.Get(instanceData.ClonedString(), property)
		return value
	}
	// strip parameters. from property name
	_, after, found := strings.Cut(property, "parameters.")
	if found {
		property = after
	}

	// process parameter search
	t := gjson.Get(instanceData.ClonedString(), "parameters.#.name")

	for _, name := range t.Array() {
		if name.ClonedString() == property {
			value := gjson.Get(instanceData.ClonedString(), "parameters.#(name="+property+").value")
			return value
		}
	}
	return gjson.Result{}
}

// HandleResults function is used for handling the rest response for parent as well as endpoints calls,
func (e *Ems) HandleResults(result []gjson.Result, prop map[string][]*emsProp) (map[string]*matrix.Matrix, uint64, uint64) {
	var (
		err                  error
		count, instanceCount uint64
		mx                   *matrix.Matrix
	)

	var m = e.Matrix

	// Resolving ems events that matched no cached issuing ems, counted by name.
	unresolved := make(map[string]int)
	unresolvedIssuers := make(map[string]string)

	// Template labels the ONTAP response did not carry, counted per event+label.
	missingLabels := make(map[missingLabelKey]*missingLabelCount)

	for _, instanceData := range result {
		var (
			instanceKey string
		)

		if !instanceData.IsObject() {
			e.Logger.Warn("Instance data is not object, skipping", slog.String("type", instanceData.Type.String()))
			continue
		}
		messageName := instanceData.Get("message.name")

		// verify if message name exists in ONTAP response
		if !messageName.Exists() {
			e.Logger.Error("skip instance, missing message name")
			continue
		}
		msgName := messageName.ClonedString()
		if issuingEmsList, ok := e.bookendEmsMap[msgName]; ok {
			props := prop[msgName]
			if len(props) == 0 {
				e.Logger.Warn("Ems properties not found", slog.String("resolving ems", msgName))
				continue
			}
			// resolving ems would only have 1 prop record
			p := props[0]
			bookendKey := e.getInstanceKeys(p, instanceData)

			emsResolved := false
			/* Below logic will evaluate this way:
			   case 1: For Bookend ems (one to one): [LUN.offline - LUN.offline]
			     - loop would iterate once
			     - if issuing ems exist then resolve else log warning as unable to find matching ems in cache
			   case 2: For Bookend with same resoling ems (many to one): [monitor.fan.critical, monitor.fan.failed, monitor.fan.warning - monitor.fan.ok]
			     - loop would iterate for all possible issuing ems
			     - if one or more issuing ems exist then resolve all matching ems else log warning as unable to find matching ems in cache
			*/
			for _, issuingEms := range issuingEmsList.Slice() {
				if mx = m[issuingEms]; mx != nil {
					metr, exist := mx.GetMetrics()["events"]
					if !exist {
						e.Logger.Warn("failed to get metric", slog.String("name", "events"))
						continue
					}

					// get all active instances by issuingems-bookendkey
					for key, instance := range mx.GetInstances() {
						if !strings.HasSuffix(key, issuingEms+bookendKey) {
							continue
						}
						metr.SetValueFloat64(instance, 0)
						instance.SetExportable(true)
						emsResolved = true
					}
				}
			}

			if !emsResolved {
				// Counted rather than logged per record. A resolving ems with
				// no cached issuing ems is normal - the issuing event may
				// predate this poller, or have fallen outside the window. These will be
				// summarized after the loop.
				unresolved[msgName]++
				if _, ok := unresolvedIssuers[msgName]; !ok {
					issuers := issuingEmsList.Slice()
					// Set iteration order is random, so sort to keep the
					// summary stable from poll to poll.
					slices.Sort(issuers)
					unresolvedIssuers[msgName] = strings.Join(issuers, ",")
				}
			}
		} else {
			existingEms := false
			if _, ok := m[msgName]; !ok {
				// create matrix if not exists for the ems event
				mx = matrix.New(msgName, e.Prop.Object, msgName)
				mx.UpdateGlobalLabels(e.Matrix[e.Object].GetGlobalLabels())
				m[msgName] = mx
			} else {
				existingEms = true
				mx = m[msgName]
			}

			// Check matches at all same name ems
			var isMatch bool
			// Check matches at each ems
			var isMatchPs bool
			// Check instance count at all same name ems
			instanceLabelCount := uint64(0)
			// Check instance count at each ems
			var instanceLabelCountPs uint64

			// parse ems properties for the instance
			if ps, ok := prop[msgName]; ok {
				for _, p := range ps {
					isMatchPs = false
					instanceLabelCountPs = 0
					instanceKey = e.getInstanceKeys(p, instanceData)
					instance := mx.GetInstance(instanceKey)

					if instance == nil {
						if instance, err = mx.NewInstance(instanceKey); err != nil {
							e.Logger.Error("", slogx.Err(err), slog.String("instanceKey", instanceKey))
							continue
						}
					}

					// explicitly set to true. for bookend case, it was set as false for the same instance[when multiple ems received]
					instance.SetExportable(true)

					for label, display := range p.InstanceLabels {
						if label == AutoResolved {
							instance.SetLabel(display, "")
							continue
						}
						value := parseProperties(instanceData, label)
						if value.Exists() {
							if value.IsArray() {
								labelArray := make([]string, 0, len(value.Array()))
								for _, r := range value.Array() {
									labelString := r.ClonedString()
									labelArray = append(labelArray, labelString)
								}
								instance.SetLabel(display, strings.Join(labelArray, ","))
							} else {
								instance.SetLabel(display, value.ClonedString())
							}
							instanceLabelCountPs++
						} else {
							// Counted rather than logged per record. A label the
							// template asks for but the ONTAP release does not
							// send is missing from every instance of that event,
							// so one line per record says nothing extra.
							k := missingLabelKey{ems: msgName, label: label}
							if seen, ok := missingLabels[k]; ok {
								seen.count++
							} else {
								missingLabels[k] = &missingLabelCount{count: 1, example: instanceKey}
							}
						}
					}

					// set labels
					for k, v := range p.Labels {
						instance.SetLabel(k, v)
					}

					// matches filtering
					if len(p.Matches) == 0 {
						isMatchPs = true
					} else {
						for _, match := range p.Matches {
							if value := instance.GetLabel(match.Name); value != "" {
								if value == match.value {
									isMatchPs = true
									break
								}
							} else {
								// value not found
								e.Logger.Warn(
									"label is not found",
									slog.String("instanceKey", instanceKey),
									slog.String("name", match.Name),
									slog.String("value", match.value),
								)
							}
						}
					}
					if !isMatchPs {
						continue
					}

					for _, metric := range p.Metrics {
						metr, ok := mx.GetMetrics()[metric.Name]
						if !ok {
							if metr, err = mx.NewMetricFloat64(metric.Name); err != nil {
								e.Logger.Error("failed to get metric",
									slogx.Err(err),
									slog.String("name", metric.Name),
								)
								continue
							}
							metr.SetExportable(metric.Exportable)
						}
						switch metric.Name {
						case "events":
							metr.SetValueFloat64(instance, 1)
						case "timestamp":
							metr.SetValueFloat64(instance, float64(time.Now().UnixMicro()))
						default:
							e.Logger.Warn("Unable to find metric",
								slog.String("key", metric.Name),
								slog.String("metric", metric.Label),
							)
						}
					}
					instanceLabelCount += instanceLabelCountPs
					isMatch = true
				}
			}
			if !isMatch {
				mx.RemoveInstance(instanceKey)
				if !existingEms {
					delete(m, msgName)
				}
				continue
			}
			count += instanceLabelCount
		}
	}

	missingKeys := slices.SortedFunc(maps.Keys(missingLabels), func(a, b missingLabelKey) int {
		return cmp.Or(cmp.Compare(a.ems, b.ems), cmp.Compare(a.label, b.label))
	})
	for _, k := range missingKeys {
		m := missingLabels[k]
		e.Logger.Warn(
			"Missing label value",
			slog.String("ems", k.ems),
			slog.String("label", k.label),
			slog.Int("count", m.count),
			slog.String("exampleInstanceKey", m.example),
		)
	}

	for _, name := range slices.Sorted(maps.Keys(unresolved)) {
		e.Logger.Warn(
			"Unable to find matching issue ems in cache",
			slog.String("resolving ems", name),
			slog.String("issuing ems", unresolvedIssuers[name]),
			slog.Int("count", unresolved[name]),
		)
	}

	for _, v := range e.Matrix {
		for _, i := range v.GetInstances() {
			if i.IsExportable() {
				instanceCount++
			}
		}
	}

	return m, count, instanceCount
}

func (e *Ems) getInstanceKeys(p *emsProp, instanceData gjson.Result) string {
	var instanceKey strings.Builder
	// extract instance key(s)
	for _, k := range p.InstanceKeys {
		value := parseProperties(instanceData, k)
		if value.Exists() {
			instanceKey.WriteString(Hyphen + value.ClonedString())
		} else {
			e.Logger.Error("skip instance, missing key", slog.String("key", k))
			break
		}
	}
	return instanceKey.String()
}

func (e *Ems) updateMatrix(begin time.Time) {
	tempMap := make(map[string]*matrix.Matrix)
	// store the bookend ems metric in tempMap
	for _, issuingEmsList := range e.bookendEmsMap {
		for _, issuingEms := range issuingEmsList.Slice() {
			if mx, exist := e.Matrix[issuingEms]; exist {
				tempMap[issuingEms] = mx
			}
		}
	}

	// We want to ensure that the existing matrix is an empty clone so that it gets updated in the Prometheus cache.
	// This prevents older instances from appearing in the previous poll.
	for k, v := range e.Matrix {
		e.Matrix[k] = v.CloneMetricTemplate()
	}

	for issuingEms, mx := range tempMap {
		eventMetric, ok := mx.GetMetrics()["events"]
		if !ok {
			e.Logger.Error(
				"failed to get metric",
				slog.String("issuingEms", issuingEms),
				slog.String("name", "events"),
			)
			continue
		}

		timestampMetric, ok := mx.GetMetrics()["timestamp"]
		if !ok {
			e.Logger.Error(
				"failed to get metric",
				slog.String("issuingEms", issuingEms),
				slog.String("name", "timestamp"),
			)
			continue
		}
		for instanceKey, instance := range mx.GetInstances() {
			// set export to false
			instance.SetExportable(false)

			if val, exist := eventMetric.GetValueFloat64(instance); exist && val == 0 {
				mx.RemoveInstance(instanceKey)
				continue
			}

			// check instance timestamp and remove it after given resolve_after duration
			if metricTimestamp, ok := timestampMetric.GetValueFloat64(instance); ok {
				if collectors.IsTimestampOlderThanDuration(begin, metricTimestamp, e.resolveAfter[issuingEms]) {
					// Set events metric value as 0 and export instance to true with label autoresolved as true.
					eventMetric.SetValueFloat64(instance, 0)
					instance.SetExportable(true)
					instance.SetLabel(AutoResolved, "true")
				}
			}
		}
		if instances := mx.GetInstances(); len(instances) == 0 {
			// We want to ensure that the existing matrix is an empty clone so that it gets updated in the Prometheus cache.
			// This prevents older instances from appearing in the previous poll.
			e.Matrix[issuingEms] = mx.CloneMetricTemplate()
			continue
		}
		e.Matrix[issuingEms] = mx
	}
}

// Interface guards
var (
	_ collector.Collector = (*Ems)(nil)
)
