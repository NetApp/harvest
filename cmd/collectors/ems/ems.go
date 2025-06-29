package ems

import (
	"fmt"
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
	mat.SetGlobalLabel("cluster", e.Client.Remote().Name)
	mat.SetGlobalLabel("cluster_uuid", e.Client.Remote().UUID)

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

	if b := e.Params.GetChildContentS("max_url_size"); b != "" {
		if s, err := strconv.Atoi(b); err == nil {
			e.maxURLSize = s
		}
	}
	e.Logger.Debug("", slog.Int("max_url_size", e.maxURLSize))

	if s := e.Params.GetChildContentS("severity"); s != "" {
		e.severityFilter = severityFilterPrefix + s
	}
	e.Logger.Debug("", slog.String("severityFilter", e.severityFilter))

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
		prop := emsProp{}

		prop.InstanceKeys = make([]string, 0)
		prop.InstanceLabels = make(map[string]string)
		prop.Metrics = make(map[string]*Metric)

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
	return nil
}

// returns time filter (clustertime - polldata duration)
func (e *Ems) getTimeStampFilter(clusterTime time.Time) string {
	fromTime := e.lastFilterTime
	// check if this is the first request
	if e.lastFilterTime == 0 {
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
	return fmt.Sprintf("time=>=%d", fromTime)
}

func (e *Ems) fetchEMSData(href string) ([]gjson.Result, error) {
	var (
		records []gjson.Result
		err     error
	)
	if records, err = e.GetRestData(href); err != nil {
		return nil, err
	}
	return records, nil
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
	_ = e.Metadata.LazySetValueInt64("api_time", "instance", apiD.Microseconds())
	_ = e.Metadata.LazySetValueInt64("parse_time", "instance", time.Since(parseT).Microseconds())
	_ = e.Metadata.LazySetValueUint64("instances", "instance", uint64(bookendCacheSize)) //nolint:gosec

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

	startTime = time.Now()

	// add time filter
	clusterTime, err := collectors.GetClusterTime(e.Client, e.ReturnTimeOut, e.Logger)
	if err != nil {
		return nil, err
	}
	toTime := clusterTime.Unix()
	timeFilter := e.getTimeStampFilter(clusterTime)
	filter := e.Filter
	filter = append(filter, timeFilter)

	// build hrefs up to maxURLSize
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
	for _, h := range hrefs {
		r, err := e.fetchEMSData(h)
		if err != nil {
			return nil, err
		}
		records = append(records, r...)
	}

	apiD = time.Since(startTime)

	startTime = time.Now()
	_, count, instanceCount = e.HandleResults(records, e.emsProp)

	parseD = time.Since(startTime)

	_ = e.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = e.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = e.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = e.Metadata.LazySetValueUint64("instances", "data", instanceCount)

	e.AddCollectCount(count)

	// update lastFilterTime to current cluster time
	e.lastFilterTime = toTime
	return e.Matrix, nil
}

func (e *Ems) getHref(names []string, filter []string) string {
	nameFilter := "message.name=" + strings.Join(names, ",")
	filter = append(filter, nameFilter)
	// If both issuing ems and resolving ems would come together in same poll, This index ordering would make sure that latest ems would process last. So, if resolving ems would be latest, it will resolve the issue.
	// add filter as order by index in ascending order
	orderByIndexFilter := "order_by=" + "index%20asc"
	filter = append(filter, orderByIndexFilter)

	href := rest.NewHrefBuilder().
		APIPath(e.Query).
		Fields(e.Fields).
		Filter(filter).
		MaxRecords(collectors.DefaultBatchSize).
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
					if instances := mx.GetInstancesBySuffix(issuingEms + bookendKey); len(instances) != 0 {
						for _, instance := range instances {
							metr.SetValueFloat64(instance, 0)
							instance.SetExportable(true)
						}
						emsResolved = true
					}
				}
			}

			if !emsResolved {
				e.Logger.Warn(
					"Unable to find matching issue ems in cache",
					slog.String("resolving ems", msgName),
					slog.String("issuing ems", strings.Join(issuingEmsList.Slice(), ",")),
				)
			}
		} else {
			if _, ok := m[msgName]; !ok {
				// create matrix if not exists for the ems event
				mx = matrix.New(msgName, e.Prop.Object, msgName)
				mx.SetGlobalLabels(e.Matrix[e.Object].GetGlobalLabels())
				m[msgName] = mx
			} else {
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
								var labelArray []string
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
							e.Logger.Warn(
								"Missing label value",
								slog.String("instanceKey", instanceKey),
								slog.String("label", label),
							)
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
				continue
			}
			count += instanceLabelCount
		}
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
	var instanceKey string
	// extract instance key(s)
	for _, k := range p.InstanceKeys {
		value := parseProperties(instanceData, k)
		if value.Exists() {
			instanceKey += Hyphen + value.ClonedString()
		} else {
			e.Logger.Error("skip instance, missing key", slog.String("key", k))
			break
		}
	}
	return instanceKey
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
		e.Matrix[k] = v.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: false})
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
			e.Matrix[issuingEms] = mx.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: false})
			continue
		}
		e.Matrix[issuingEms] = mx
	}
}

// Interface guards
var (
	_ collector.Collector = (*Ems)(nil)
)
