package eseriesmel

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

const (
	eventsMetricName    = "events"
	defaultBatchSize    = 500
	defaultClientTime   = "30s"
	availableSuffix     = "/available"
	fieldSequenceNumber = "sequenceNumber"
	fieldTimeStamp      = "timeStamp"
	fieldDescription    = "description"
	fieldEventType      = "eventType"
	fieldComponentType  = "componentType"
	fieldPriority       = "priority"
	fieldName           = "name"
	labelMessage        = "message"

	// keySeparator is a non-printable control character: unlike any printable
	// choice, it can't collide with user-assigned volume/pool names in location.
	keySeparator = "\x1f"

	// configFieldEventType is the template's per-entry key, not the API field name.
	configFieldEventType = "event_type"

	// undefinedEnumValue is the sentinel the API uses across its enums.
	undefinedEnumValue = "__UNDEFINED"
	unknownDisplayName = "unknown"

	// "relative" means the real component type is nested one level deeper.
	componentTypeRelative             = "relative"
	componentRelativeLocationTypePath = "componentLocation.componentRelativeLocation.componentType"
)

var priorityDisplayNames = map[string]string{
	"priorityDefault":   "default",
	"priorityCritical":  "critical",
	"priorityInfo":      "info",
	"priorityEmergency": "emergency",
	"priorityAlert":     "alert",
	"priorityError":     "error",
	"priorityWarning":   "warning",
	"priorityNotice":    "notice",
	"priorityDebug":     "debug",
	undefinedEnumValue:  unknownDisplayName,
}

func normalizePriority(raw string) string {
	if display, ok := priorityDisplayNames[raw]; ok {
		return display
	}
	return raw
}

// normalizeComponentType hides internal placeholders (__UNDEFINED, unresolved
// "relative") behind "unknown".
func normalizeComponentType(raw string) string {
	if raw == undefinedEnumValue || raw == componentTypeRelative {
		return unknownDisplayName
	}
	return raw
}

// EseriesMel polls E-Series MEL events using a sequenceNumber cursor.
type EseriesMel struct {
	*collector.AbstractCollector
	Client    *rest.Client
	Prop      *Prop
	arrayID   string
	arrayName string
	// nextSeqNum is the next fetch's startSequenceNumber, moved by advanceCursor.
	nextSeqNum int64
	// lastTip is the endingSeqNum last observed from fetchExtent, used to detect
	// a genuine counter reset independent of where the cursor happens to be.
	lastTip int64
	seeded  bool
}

// Prop holds the parsed template configuration for this collector.
type Prop struct {
	Object         string
	Query          string
	TemplatePath   string
	Filter         []string          // server-side query fragments, appended verbatim
	EventCatalog   map[string]string // eventType -> catalog message name; required, defines the allow-list
	BatchSize      int
	InstanceKeys   []string          // API field names that make up the instance key, in order
	InstanceLabels map[string]string // API field name -> display label name
}

func init() {
	plugin.RegisterModule(&EseriesMel{})
}

func (e *EseriesMel) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.eseriesmel",
		New: func() plugin.Module { return new(EseriesMel) },
	}
}

func (e *EseriesMel) Init(a *collector.AbstractCollector) error {
	var err error

	e.AbstractCollector = a

	e.InitProp()

	if e.Prop.TemplatePath, err = e.LoadTemplate(); err != nil {
		return err
	}

	if err := e.ParseTemplate(); err != nil {
		return err
	}

	if err := e.InitClient(); err != nil {
		return err
	}

	if e.Options.IsTest {
		mx := matrix.New(e.Name, e.Object, e.Object)
		if exportOptions := e.Params.GetChildS("export_options"); exportOptions != nil {
			mx.SetExportOptions(exportOptions)
		}
		e.Matrix = make(map[string]*matrix.Matrix)
		e.Matrix[e.Object] = mx
	} else {
		if err := collector.Init(e); err != nil {
			return err
		}
	}

	if err := e.InitMatrix(); err != nil {
		return err
	}

	e.Logger.Debug(
		"initialized",
		slog.String("object", e.Prop.Object),
		slog.Int("batchSize", e.Prop.BatchSize),
		slog.String("timeout", e.Client.Timeout.String()),
	)

	return nil
}

func (e *EseriesMel) InitProp() {
	e.Prop = &Prop{
		InstanceKeys:   make([]string, 0),
		InstanceLabels: make(map[string]string),
		EventCatalog:   make(map[string]string),
		BatchSize:      defaultBatchSize,
	}
}

func (e *EseriesMel) InitClient() error {
	var err error

	clientTimeout := e.Params.GetChildContentS("client_timeout")
	if clientTimeout == "" {
		clientTimeout = defaultClientTime
	}

	duration, err := time.ParseDuration(clientTimeout)
	if err != nil {
		e.Logger.Info("Using default timeout", slog.String("timeout", defaultClientTime))
		duration, _ = time.ParseDuration(defaultClientTime)
	}

	poller, err := conf.PollerNamed(e.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, e.Logger)

	// MEL events are cursor-based, never shared/cached across objects.
	if e.Client, err = rest.New(poller, duration, credentials, ""); err != nil {
		return err
	}

	if e.Options.IsTest {
		return nil
	}

	if err := e.Client.Init(1, e.Remote); err != nil {
		return err
	}

	e.Remote = e.Client.Remote()

	return nil
}

func (e *EseriesMel) InitMatrix() error {
	mat := e.Matrix[e.Object]
	mat.Object = e.Prop.Object

	if exportOptions := e.Params.GetChildS("export_options"); exportOptions != nil {
		mat.SetExportOptions(exportOptions)
	}

	if e.Params.HasChildS("labels") {
		for _, l := range e.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	// Default until discoverArray's first successful fetch.
	mat.SetGlobalLabel("management_version", unknownDisplayName)

	if _, err := mat.NewMetricFloat64(eventsMetricName, eventsMetricName); err != nil {
		return err
	}

	e.Logger.Debug(
		"initialized cache",
		slog.Any("instanceKeys", e.Prop.InstanceKeys),
		slog.Int("numLabels", len(e.Prop.InstanceLabels)),
	)
	return nil
}

// PollCounter refreshes the array identity and its array-scoped global labels.
func (e *EseriesMel) PollCounter() (map[string]*matrix.Matrix, error) {
	return nil, e.discoverArray()
}

func (e *EseriesMel) discoverArray() error {
	array, err := e.Client.DiscoverArray(e.Logger)
	if err != nil {
		return err
	}

	e.arrayID = array.ID
	e.arrayName = array.Name

	mat := e.Matrix[e.Object]
	mat.SetGlobalLabel("array", e.arrayName)
	mat.SetGlobalLabel("chassis_serial", array.System.Get("chassisSerialNumber").ClonedString())

	if version, err := e.Client.GetManagementVersion(e.arrayID); err != nil {
		e.Logger.Warn("failed to fetch management version, keeping previous value", slogx.Err(err))
	} else {
		mat.SetGlobalLabel("management_version", normalizeManagementVersion(version))
	}

	e.Logger.Debug("discovered array", slog.String("id", e.arrayID), slog.String("name", e.arrayName))
	return nil
}

// normalizeManagementVersion reduces a version string to major.minor (e.g. "12.10").
func normalizeManagementVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return version
	}
	return parts[0] + "." + parts[1]
}

// melExtent is the array's MEL sequence-number range. endingSeqNum is
// exclusive (one past the newest real event)
type melExtent struct {
	startingSeqNum int64
	endingSeqNum   int64
}

func (e *EseriesMel) fetchExtent() (melExtent, error) {
	query := rest.NewURLBuilder().
		APIPath(e.Prop.Query + availableSuffix).
		ArrayID(e.arrayID).
		Build()

	results, err := e.Client.Fetch(e.Client.APIPath+"/"+query, nil)
	if err != nil {
		return melExtent{}, err
	}
	if len(results) == 0 {
		return melExtent{}, errs.New(errs.ErrNoInstance, "mel-events/available returned no data")
	}

	obj := results[0]
	startingSeqNum := obj.Get("startingSeqNum")
	endingSeqNum := obj.Get("endingSeqNum")
	if !startingSeqNum.Exists() || !endingSeqNum.Exists() {
		return melExtent{}, errs.New(errs.ErrAPIResponse, "mel-events/available missing startingSeqNum/endingSeqNum")
	}

	return melExtent{
		startingSeqNum: startingSeqNum.Int(),
		endingSeqNum:   endingSeqNum.Int(),
	}, nil
}

func (e *EseriesMel) PollData() (map[string]*matrix.Matrix, error) {
	var (
		apiTime   time.Duration
		parseTime time.Duration
		count     uint64
	)

	if e.arrayID == "" {
		if err := e.discoverArray(); err != nil {
			return nil, err
		}
	}

	mat := e.Matrix[e.Object]
	// Fresh instances every poll: only currently-firing events are exported.
	e.Matrix[e.Object] = mat.CloneMetricTemplate()
	mat = e.Matrix[e.Object]

	e.Client.Metadata.Reset()

	apiStart := time.Now()

	extent, err := e.fetchExtent()
	if err != nil {
		return nil, err
	}

	if !e.seeded {
		// First poll: start at the tip so historical events aren't replayed.
		e.nextSeqNum = extent.endingSeqNum
		e.lastTip = extent.endingSeqNum
		e.seeded = true
	}

	var events []gjson.Result

	if rolledBack(e.lastTip, extent) {
		e.Logger.Warn(
			"mel sequence counter rolled back, resetting cursor to range start",
			slog.Int64("previousTip", e.lastTip),
			slog.Int64("newStartingSeqNum", extent.startingSeqNum),
			slog.Int64("newEndingSeqNum", extent.endingSeqNum),
		)
		e.nextSeqNum = extent.startingSeqNum
	} else {
		e.detectGap(extent)
	}

	if shouldSkipFetch(extent, e.nextSeqNum) {
		e.Logger.Debug("no new mel events since last poll, skipping fetch")
	} else {
		var run pageRun
		if run, err = e.fetchNewEvents(extent); err != nil {
			return nil, err
		}
		events = run.events
		e.logStopReason(extent, run)
		e.nextSeqNum = advanceCursor(extent, run)
	}
	e.lastTip = extent.endingSeqNum

	apiTime = time.Since(apiStart)

	parseStart := time.Now()
	count = e.pollData(mat, events)
	parseTime = time.Since(parseStart)

	dataInst := e.Metadata.MustGetInstance("data")
	e.Metadata.MustSetValueInt64("api_time", dataInst, apiTime.Microseconds())
	e.Metadata.MustSetValueInt64("parse_time", dataInst, parseTime.Microseconds())
	e.Metadata.MustSetValueUint64("metrics", dataInst, count)
	e.Metadata.MustSetValueUint64("instances", dataInst, uint64(len(mat.GetInstances())))
	e.Metadata.MustSetValueUint64("bytesRx", dataInst, e.Client.Metadata.BytesRx.Load())
	e.Metadata.MustSetValueUint64("numCalls", dataInst, e.Client.Metadata.NumCalls.Load())
	e.AddCollectCount(count)

	// Surface the collapse ratio since "Collected"'s metrics/instances fields don't.
	if count > 0 {
		uniqueInstances := len(mat.GetInstances())
		e.Logger.Info(
			"mel events deduped",
			slog.Uint64("processed", count),
			slog.Int("uniqueInstances", uniqueInstances),
			slog.Uint64("deduped", dedupedCount(count, uniqueInstances)),
		)
	}

	return e.Matrix, nil
}

// logStopReason surfaces pagination outcomes where this poll didn't read everything.
func (e *EseriesMel) logStopReason(extent melExtent, run pageRun) {
	switch run.reason {
	case stopStalled:
		e.Logger.Warn(
			"mel page carried no sequence number beyond the cursor, skipping range",
			slog.Int64("cursor", e.nextSeqNum),
			slog.Int64("nextSeq", run.nextSeq),
			slog.Int64("endingSeqNum", extent.endingSeqNum),
		)
	case stopCovered, stopShortPage:
	}
}

// dedupedCount returns how many occurrences collapsed into an existing instance.
func dedupedCount(processed uint64, uniqueInstances int) uint64 {
	u := uint64(uniqueInstances) // #nosec G115 -- len() is always non-negative
	if processed < u {
		return 0
	}
	return processed - u
}

// rolledBack reports whether the array's MEL counter was reset (not just purged).
func rolledBack(prevTip int64, extent melExtent) bool {
	return extent.endingSeqNum < prevTip
}

func shouldSkipFetch(extent melExtent, nextSeqNum int64) bool {
	return extent.endingSeqNum == nextSeqNum
}

// detectGap logs when events were purged before this collector could read them.
func (e *EseriesMel) detectGap(extent melExtent) {
	if extent.startingSeqNum <= e.nextSeqNum {
		return
	}
	missed := extent.startingSeqNum - e.nextSeqNum
	e.Logger.Warn(
		"mel events purged before being read",
		slog.Int64("missedCount", missed),
		slog.Int64("nextSeqNum", e.nextSeqNum),
		slog.Int64("actualStart", extent.startingSeqNum),
	)
	// Align the cursor: the missed range no longer exists and can't be recovered.
	e.nextSeqNum = extent.startingSeqNum
}

func (e *EseriesMel) fetchNewEvents(extent melExtent) (pageRun, error) {
	return paginate(e.nextSeqNum, extent.endingSeqNum, e.Prop.BatchSize, e.fetchPage)
}

func (e *EseriesMel) fetchPage(start int64) ([]gjson.Result, error) {
	filters := append(append([]string{}, e.Prop.Filter...),
		fmt.Sprintf("startSequenceNumber=%d", start),
		fmt.Sprintf("count=%d", e.Prop.BatchSize),
	)

	query := rest.NewURLBuilder().
		APIPath(e.Prop.Query).
		ArrayID(e.arrayID).
		Filter(filters).
		Build()

	e.Logger.Debug("fetching mel events page", slog.String("query", query))
	return e.Client.Fetch(e.Client.APIPath+"/"+query, nil)
}

// stopReason explains why a pagination run ended; it's diagnostic only and
// doesn't affect how advanceCursor computes the next cursor.
type stopReason int

const (
	// stopCovered means paging reached the extent's tip.
	stopCovered stopReason = iota
	// stopShortPage means the array returned fewer than count records (empty
	// included). count caps records returned, not records scanned, so nothing
	// more matches at or above the requested sequence number.
	stopShortPage
	// stopStalled means the array ignored startSequenceNumber; breaking avoids
	// re-issuing the same request forever.
	stopStalled
)

// pageRun is the outcome of one pagination run.
type pageRun struct {
	events []gjson.Result
	// nextSeq is the next sequence number to fetch; becomes the next poll's cursor.
	nextSeq int64
	reason  stopReason
}

// paginate fetches pages of batchSize starting at nextSeq until it reaches
// endSeqExclusive or one of the other stopReason conditions is met. There's
// no artificial page cap: detectGap keeps endSeqExclusive-nextSeq bounded by
// the array's ring buffer size on every poll, which bounds the iteration count.
func paginate(nextSeq, endSeqExclusive int64, batchSize int, fetchPage func(start int64) ([]gjson.Result, error)) (pageRun, error) {
	run := pageRun{nextSeq: nextSeq, reason: stopCovered}

	for run.nextSeq < endSeqExclusive {
		page, err := fetchPage(run.nextSeq)
		if err != nil {
			return pageRun{}, err
		}

		pageNextSeq := run.nextSeq
		for _, ev := range page {
			if seq := ev.Get(fieldSequenceNumber).Int(); seq+1 > pageNextSeq {
				pageNextSeq = seq + 1
			}
		}
		// Empty pages have no sequence number to measure progress; the
		// short-page check below ends the run.
		if len(page) > 0 && pageNextSeq <= run.nextSeq {
			run.reason = stopStalled
			break
		}

		run.events = append(run.events, page...)
		run.nextSeq = pageNextSeq

		if len(page) < batchSize {
			run.reason = stopShortPage
			break
		}
	}

	return run, nil
}

// advanceCursor returns the next poll's startSequenceNumber: the tip, or
// beyond it if nextSeq led (more records arrived while paging was in flight).
// If nextSeq trailed instead (stopShortPage/stopStalled), this jumps the
// cursor forward to the tip, deliberately skipping sequence numbers that were
// never read.
func advanceCursor(extent melExtent, run pageRun) int64 {
	return max(extent.endingSeqNum, run.nextSeq)
}

func (e *EseriesMel) pollData(mat *matrix.Matrix, events []gjson.Result) uint64 {
	var count uint64

	metr, ok := mat.GetMetrics()[eventsMetricName]
	if !ok {
		e.Logger.Error("events metric missing from matrix", slog.String("metric", eventsMetricName))
		return 0
	}

	for _, eventData := range events {
		if !eventData.IsObject() {
			continue
		}

		message, allowed := e.eventCatalogEntry(eventData)
		if !allowed {
			continue
		}

		instKey, missing := e.instanceKey(eventData)
		if len(missing) == len(e.Prop.InstanceKeys) {
			e.Logger.Debug("no instance key fields present, skipping",
				slog.String("object", e.Object),
				slog.Any("instanceKeys", e.Prop.InstanceKeys))
			continue
		}
		if len(missing) > 0 {
			e.Logger.Debug("instance key field missing, using partial key",
				slog.Any("missing", missing),
				slog.String("key", instKey))
		}

		instance, _ := mat.GetOrCreateInstance(instKey)
		instance.SetExportable(true)

		// Labels must come from the same occurrence as the metric value, so only
		// (re)write them when this occurrence wins the max-timestamp comparison.
		ts := eventData.Get(fieldTimeStamp).Float()
		if current, ok := metr.GetValueFloat64(instance); !ok || ts > current {
			metr.SetValueFloat64(instance, ts)

			instance.ClearLabels()
			for apiField, display := range e.Prop.InstanceLabels {
				value := eventData.Get(apiField)
				if apiField == fieldComponentType {
					value = resolvedComponentType(eventData)
				}
				if !value.Exists() {
					continue
				}
				labelValue := value.ClonedString()
				switch apiField {
				case fieldPriority:
					labelValue = normalizePriority(labelValue)
				case fieldComponentType:
					labelValue = normalizeComponentType(labelValue)
				}
				instance.SetLabel(display, labelValue)
			}
			instance.SetLabel(labelMessage, message)
		}
		count++

		if e.Logger.Enabled(context.Background(), slog.LevelDebug) {
			e.logOccurrence(eventData, message)
		}
	}

	return count
}

func (e *EseriesMel) eventCatalogEntry(eventData gjson.Result) (string, bool) {
	name, ok := e.Prop.EventCatalog[eventData.Get(fieldEventType).ClonedString()]
	return name, ok
}

func (e *EseriesMel) instanceKey(eventData gjson.Result) (string, []string) {
	var key strings.Builder
	var missing []string
	for i, apiField := range e.Prop.InstanceKeys {
		if i > 0 {
			key.WriteString(keySeparator)
		}
		if value := eventData.Get(apiField); value.Exists() {
			key.WriteString(value.ClonedString())
		} else {
			missing = append(missing, apiField)
		}
	}
	return key.String(), missing
}

func (e *EseriesMel) logOccurrence(eventData gjson.Result, message string) {
	e.Logger.Debug(
		"mel event",
		slog.String("array", e.arrayName),
		slog.Int64("sequence_number", eventData.Get(fieldSequenceNumber).Int()),
		slog.String("timestamp", eventData.Get(fieldTimeStamp).ClonedString()),
		slog.String("event_type", eventData.Get(fieldEventType).ClonedString()),
		slog.String("message", message),
		slog.String("location", eventData.Get("location").ClonedString()),
		slog.String("component_type", normalizeComponentType(resolvedComponentType(eventData).ClonedString())),
		slog.String("severity", normalizePriority(eventData.Get(fieldPriority).ClonedString())),
		slog.Bool("critical", eventData.Get("critical").Bool()),
		slog.String("description", eventData.Get(fieldDescription).ClonedString()),
	)
}

func resolvedComponentType(eventData gjson.Result) gjson.Result {
	flat := eventData.Get(fieldComponentType)
	if flat.String() != componentTypeRelative {
		return flat
	}
	nested := eventData.Get(componentRelativeLocationTypePath)
	if nested.Exists() && nested.String() != "" {
		return nested
	}
	return flat
}

var (
	_ collector.Collector = (*EseriesMel)(nil)
)
