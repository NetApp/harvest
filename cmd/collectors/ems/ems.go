package ems

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

const defaultDataPollDuration = 3 * time.Minute
const maxURLSize = 8_000 //bytes
const severityFilterPrefix = "message.severity="
const defaultSeverityFilter = "alert|emergency|error|informational|notice"
const MaxBookendInstances = 1000
const DefaultBookendResolutionDuration = 28 * 24 * time.Hour // 28 days == 672 hours
const Hyphen = "-"

type Ems struct {
	*rest2.Rest    // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	Query          string
	TemplatePath   string
	emsProp        map[string][]*emsProp // array is used here to handle same ems written with different ops, matches or exports. Example: arw.volume.state ems with op as disabled or dry-run
	Filter         []string
	Fields         []string
	ReturnTimeOut  string
	lastFilterTime int64
	maxURLSize     int
	DefaultLabels  []string
	severityFilter string
	eventNames     []string            // consist of all ems events supported
	bookendEmsMap  map[string][]string // This is reverse bookend ems map, [Resolving ems]:[Issuing ems slice]
	resolveAfter   time.Duration
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

	e.bookendEmsMap = make(map[string][]string)

	if err = e.InitClient(); err != nil {
		return err
	}

	if e.TemplatePath, err = e.LoadTemplate(); err != nil {
		return err
	}

	if err = collector.Init(e); err != nil {
		return err
	}

	if err = e.InitCache(); err != nil {
		return err
	}

	if err = e.InitMatrix(); err != nil {
		return err
	}

	return nil
}

func (e *Ems) InitMatrix() error {
	mat := e.Matrix[e.Object]
	// overwrite from abstract collector
	mat.Object = e.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", e.Client.Cluster().Name)
	mat.SetGlobalLabel("cluster_uuid", e.Client.Cluster().UUID)

	if e.Params.HasChildS("labels") {
		for _, l := range e.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}
	return nil
}

func (e *Ems) LoadPlugin(kind string, _ *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	default:
		e.Logger.Warn().Str("kind", kind).Msg("no ems plugin found ")
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
	e.Logger.Debug().Int("max_url_size", e.maxURLSize).Msgf("")

	if s := e.Params.GetChildContentS("severity"); s != "" {
		e.severityFilter = severityFilterPrefix + s
	}
	e.Logger.Debug().Str("severityFilter", e.severityFilter).Msgf("")

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
		e.ReturnTimeOut = returnTimeout
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
			e.Logger.Warn().Msg("Missing event name")
			continue
		}

		//populate prop counter for asup
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
					e.Logger.Error().Stack().Err(err).Msg("Failed to load plugin")
				}
			}
			if line1.GetNameS() == "resolve_when_ems" {
				e.ParseResolveEms(line1, prop)
			}
		}
		e.emsProp[prop.Name] = append(e.emsProp[prop.Name], &prop)
	}
	//add severity filter
	e.Filter = append(e.Filter, e.severityFilter)
	return nil
}

func (e *Ems) getClusterTime() (time.Time, error) {
	var (
		err         error
		records     []gjson.Result
		clusterTime time.Time
	)

	query := "private/cli/cluster/date"
	fields := []string{"date"}

	href := rest.BuildHref(query, strings.Join(fields, ","), nil, "", "", "1", e.ReturnTimeOut, "")

	if records, err = e.GetRestData(href); err != nil {
		return clusterTime, err
	}
	if len(records) == 0 {
		return clusterTime, errs.New(errs.ErrConfig, e.Object+" date not found on cluster")
	}

	for _, instanceData := range records {
		currentClusterDate := instanceData.Get("date")
		if currentClusterDate.Exists() {
			t, err := time.Parse(time.RFC3339, currentClusterDate.String())
			if err != nil {
				e.Logger.Error().Str("date", currentClusterDate.String()).Err(err).Msg("Failed to load cluster date")
				continue
			}
			clusterTime = t
			break
		}
	}

	e.Logger.Debug().Str("cluster time", clusterTime.String()).Msg("")
	return clusterTime, nil
}

// returns time filter (clustertime - polldata duration)
func (e *Ems) getTimeStampFilter(clusterTime time.Time) string {
	fromTime := e.lastFilterTime
	// check if this is the first request
	if e.lastFilterTime == 0 {
		// if first request fetch cluster time
		dataDuration, err := GetDataInterval(e.GetParams(), defaultDataPollDuration)
		if err != nil {
			e.Logger.Warn().Err(err).
				Str("defaultDataPollDuration", defaultDataPollDuration.String()).
				Msg("Failed to parse duration. using default")
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
		ok               bool
		metr             matrix.Metric
		bookendCacheSize int
		metricTimestamp  float64
	)

	query := "api/support/ems/messages"
	fields := []string{"name"}

	href := rest.BuildHref(query, strings.Join(fields, ","), nil, "", "", "", e.ReturnTimeOut, query)

	if records, err = e.GetRestData(href); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, errs.New(errs.ErrNoInstance, e.Object+" no ems message found on cluster")
	}

	var emsEventCatalogue []string
	for _, instanceData := range records {
		name := instanceData.Get("name")
		if name.Exists() {
			emsEventCatalogue = append(emsEventCatalogue, name.String())
		}
	}

	// collect all event names
	var names []string
	for key := range e.emsProp {
		names = append(names, key)
	}

	//filter out names which exists on the cluster. ONTAP rest ems throws error for a message.name filter if that event is not supported by that cluster
	filteredNames, _ := util.Intersection(names, emsEventCatalogue)
	e.Logger.Debug().Strs("querying for events", filteredNames).Msg("")
	_, missingNames := util.Intersection(filteredNames, names)
	e.Logger.Debug().Strs("skipped events", missingNames).Msg("")
	e.eventNames = filteredNames

	// check instance timestamp and remove it after given resolve_after duration and warning when total instance in cache > 1000 instance
	for _, issuingEmsList := range e.bookendEmsMap {
		for _, issuingEms := range issuingEmsList {
			if mx := e.Matrix[issuingEms]; mx != nil {
				for instanceKey, instance := range mx.GetInstances() {
					if metr, ok = mx.GetMetrics()["timestamp"]; !ok {
						e.Logger.Error().
							Str("name", "timestamp").
							Msg("failed to get metric")
					}
					// check instance timestamp and remove it after given resolve_after value
					if metricTimestamp, ok, _ = metr.GetValueFloat64(instance); ok {
						if collectors.IsTimestampOlderThanDuration(metricTimestamp, e.resolveAfter) {
							mx.RemoveInstance(instanceKey)
						}
					}
				}

				if instances := mx.GetInstances(); len(instances) == 0 {
					delete(e.Matrix, issuingEms)
					continue
				}

				bookendCacheSize += len(mx.GetInstances())
			}
		}
	}

	e.Logger.Info().Int("total instances", bookendCacheSize).Msg("")
	// warning when total instance in cache > 1000 instance
	if bookendCacheSize > MaxBookendInstances {
		e.Logger.Warn().Int("total instances", bookendCacheSize).Msg("cache has more than 1000 instances")
	}
	return nil, nil
}

func (e *Ems) PollData() (map[string]*matrix.Matrix, error) {

	var (
		count        uint64
		apiD, parseD time.Duration
		startTime    time.Time
		err          error
		records      []gjson.Result
	)

	e.Logger.Debug().Msg("starting data poll")

	// Update cache for bookend ems
	e.updateMatrix()

	startTime = time.Now()

	// add time filter
	clusterTime, err := e.getClusterTime()
	if err != nil {
		return nil, err
	}
	toTime := clusterTime.Unix()
	timeFilter := e.getTimeStampFilter(clusterTime)
	filter := append(e.Filter, timeFilter)

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
			end = end - 1
			h = e.getHref(e.eventNames[start:end], filter)
			hrefs = append(hrefs, h)
			start = end
		} else {
			if end == len(e.eventNames)-1 {
				end = len(e.eventNames)
				h = e.getHref(e.eventNames[start:end], filter)
				hrefs = append(hrefs, h)
			}
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

	if len(records) == 0 {
		e.Logger.Info().
			Int("queried", len(e.eventNames)).
			Msg("No EMS events returned")
		e.lastFilterTime = toTime
		_ = e.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
		_ = e.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
		_ = e.Metadata.LazySetValueUint64("metrics", "data", 0)
		_ = e.Metadata.LazySetValueUint64("instances", "data", 0)
		return nil, nil
	}

	startTime = time.Now()
	_, count = e.HandleResults(records, e.emsProp)

	parseD = time.Since(startTime)

	var instanceCount int
	for _, v := range e.Matrix {
		instanceCount += len(v.GetInstances())
	}

	e.Logger.Info().
		Int("instances", instanceCount).
		Uint64("dataPoints", count).
		Str("apiTime", apiD.String()).
		Str("parseTime", parseD.String()).
		Msg("Collected")

	_ = e.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = e.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = e.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = e.Metadata.LazySetValueUint64("instances", "data", uint64(instanceCount))

	e.AddCollectCount(count)

	// update lastFilterTime to current cluster time
	e.lastFilterTime = toTime
	return e.Matrix, nil
}

func (e *Ems) getHref(names []string, filter []string) string {
	nameFilter := "message.name=" + strings.Join(names, ",")
	filter = append(filter, nameFilter)
	// If both issuing ems and resolving ems would come together in same poll, This index ordering would make sure that latest ems would process last. So, if resolving ems would be latest, it will resolve the issue.
	// add filter as order by index in descending order
	orderByIndexFilter := "order_by=" + "index%20desc"
	filter = append(filter, orderByIndexFilter)

	href := rest.BuildHref(e.Query, strings.Join(e.Fields, ","), filter, "", "", "", e.ReturnTimeOut, e.Query)
	return href
}

// GetDataInterval fetch pollData interval
func GetDataInterval(param *node.Node, defaultInterval time.Duration) (time.Duration, error) {
	var dataIntervalStr string
	var durationVal time.Duration
	var err error
	schedule := param.GetChildS("schedule")
	if schedule != nil {
		dataInterval := schedule.GetChildS("data")
		if dataInterval != nil {
			dataIntervalStr = dataInterval.GetContentS()
			if durationVal, err = time.ParseDuration(dataIntervalStr); err == nil {
				return durationVal, nil
			}
			return defaultInterval, err
		}
	}
	return defaultInterval, nil
}

func parseProperties(instanceData gjson.Result, property string) gjson.Result {

	if !strings.HasPrefix(property, "parameters.") {
		// if prefix is not parameters.
		value := gjson.Get(instanceData.String(), property)
		return value
	}
	//strip parameters. from property name
	_, after, found := strings.Cut(property, "parameters.")
	if found {
		property = after
	}

	//process parameter search
	t := gjson.Get(instanceData.String(), "parameters.#.name")

	for _, name := range t.Array() {
		if name.String() == property {
			value := gjson.Get(instanceData.String(), "parameters.#(name="+property+").value")
			return value
		}
	}
	return gjson.Result{}
}

// HandleResults function is used for handling the rest response for parent as well as endpoints calls,
func (e *Ems) HandleResults(result []gjson.Result, prop map[string][]*emsProp) (map[string]*matrix.Matrix, uint64) {
	var (
		err   error
		count uint64
		mx    *matrix.Matrix
	)

	var m = e.Matrix

	for _, instanceData := range result {
		var (
			instanceKey string
		)

		var instanceLabelCount uint64
		if !instanceData.IsObject() {
			e.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
			continue
		}
		messageName := instanceData.Get("message.name")
		// verify if message name exists in ONTAP response
		if !messageName.Exists() {
			e.Logger.Warn().Msg("skip instance, missing message name")
			continue
		}
		msgName := messageName.String()

		if issuingEmsList, ok := e.bookendEmsMap[msgName]; ok {
			props := prop[msgName]
			if len(props) == 0 {
				e.Logger.Warn().Str("resolving ems", msgName).
					Msg("Ems properties not found")
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
			for _, issuingEms := range issuingEmsList {
				if mx = m[issuingEms]; mx != nil {
					metr, exist := mx.GetMetrics()["events"]
					if !exist {
						e.Logger.Warn().
							Str("name", "events").
							Msg("failed to get metric")
						continue
					}

					// get all active instances by issuingems-bookendkey
					if instances := mx.GetInstancesBySuffix(issuingEms + bookendKey); len(instances) != 0 {
						for _, instance := range instances {
							if err = metr.SetValueFloat64(instance, 0); err != nil {
								e.Logger.Error().Err(err).Str("key", "events").
									Msg("Unable to set float key on metric")
								continue
							}
							instance.SetExportable(true)
						}
						emsResolved = true
					}
				}
			}

			if !emsResolved {
				e.Logger.Warn().Str("resolving ems", msgName).Str("issue ems", strings.Join(issuingEmsList, ",")).
					Msg("Unable to find matching issue ems in cache")
			}
		} else {
			if _, ok := m[msgName]; !ok {
				//create matrix if not exists for the ems event
				mx = matrix.New(msgName, e.Prop.Object, msgName)
				mx.SetGlobalLabels(e.Matrix[e.Object].GetGlobalLabels())
				m[msgName] = mx
			} else {
				mx = m[msgName]
			}

			//parse ems properties for the instance
			isMatch := false
			if ps, ok := prop[msgName]; ok {
				for _, p := range ps {
					instanceKey = e.getInstanceKeys(p, instanceData)
					instance := mx.GetInstance(instanceKey)

					if instance == nil {
						if instance, err = mx.NewInstance(instanceKey); err != nil {
							e.Logger.Error().Err(err).Str("Instance key", instanceKey).Msg("")
							continue
						}
					}

					// explicitly set to true. for bookend case, it was set as false for the same instance[when multiple ems received]
					instance.SetExportable(true)

					for label, display := range p.InstanceLabels {
						value := parseProperties(instanceData, label)
						if value.Exists() {
							if value.IsArray() {
								var labelArray []string
								for _, r := range value.Array() {
									labelString := r.String()
									labelArray = append(labelArray, labelString)
								}
								instance.SetLabel(display, strings.Join(labelArray, ","))
							} else {
								instance.SetLabel(display, value.String())
							}
							instanceLabelCount++
						} else {
							e.Logger.Warn().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
						}
					}

					//set labels
					for k, v := range p.Labels {
						instance.SetLabel(k, v)
					}

					//matches filtering
					if len(p.Matches) == 0 {
						isMatch = true
					} else {
						for _, match := range p.Matches {
							if value := instance.GetLabel(match.Name); value != "" {
								if value == match.value {
									isMatch = true
									break
								}
							} else {
								//value not found
								e.Logger.Warn().
									Str("Instance key", instanceKey).
									Str("name", match.Name).
									Str("value", match.value).
									Msg("label is not found")
							}
						}
					}
					if !isMatch {
						instanceLabelCount = 0
						continue
					}

					for _, metric := range p.Metrics {
						metr, ok := mx.GetMetrics()[metric.Name]
						if !ok {
							if metr, err = mx.NewMetricFloat64(metric.Name); err != nil {
								e.Logger.Error().Err(err).
									Str("name", metric.Name).
									Msg("failed to get metric")
								continue
							}
							metr.SetExportable(metric.Exportable)
						}
						if metric.Name == "events" {
							if err = metr.SetValueFloat64(instance, 1); err != nil {
								e.Logger.Error().Err(err).Str("key", metric.Name).Str("metric", metric.Label).
									Msg("Unable to set float key on metric")
							}
						} else if metric.Name == "timestamp" {
							if err = metr.SetValueFloat64(instance, float64(time.Now().UnixMicro())); err != nil {
								e.Logger.Error().Err(err).Str("key", metric.Name).Str("metric", metric.Label).
									Msg("Unable to set timestamp on metric")
							}
						} else {
							// this code will not execute as ems only support [events, timestamp] metric
							e.Logger.Warn().Str("key", metric.Name).Str("metric", metric.Label).
								Msg("Unable to find metric")
						}
					}
				}
			}
			if !isMatch {
				mx.RemoveInstance(instanceKey)
				continue
			}
			count += instanceLabelCount
		}
	}
	return m, count
}

func (e *Ems) getInstanceKeys(p *emsProp, instanceData gjson.Result) string {
	var instanceKey string
	// extract instance key(s)
	for _, k := range p.InstanceKeys {
		value := parseProperties(instanceData, k)
		if value.Exists() {
			instanceKey += Hyphen + value.String()
		} else {
			e.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
			break
		}
	}
	return instanceKey
}

func (e *Ems) updateMatrix() {
	var (
		ok   bool
		val  float64
		metr matrix.Metric
	)

	tempMap := make(map[string]*matrix.Matrix)
	// store the bookend ems metric in tempMap
	for _, issuingEmsList := range e.bookendEmsMap {
		for _, issuingEms := range issuingEmsList {
			if mx, exist := e.Matrix[issuingEms]; exist {
				tempMap[issuingEms] = mx
			}
		}
	}

	// remove all ems matrix except parent object
	mat := e.Matrix[e.Object]
	e.Matrix = make(map[string]*matrix.Matrix)
	e.Matrix[e.Object] = mat

	for issuingEms, mx := range tempMap {
		if metr, ok = mx.GetMetrics()["events"]; !ok {
			e.Logger.Error().
				Str("name", "events").
				Msg("failed to get metric")
			continue
		}

		for instanceKey, instance := range mx.GetInstances() {
			// set export to false
			instance.SetExportable(false)

			if val, ok, _ = metr.GetValueFloat64(instance); ok && val == 0 {
				mx.RemoveInstance(instanceKey)
			}
		}
		if instances := mx.GetInstances(); len(instances) == 0 {
			delete(e.Matrix, issuingEms)
			continue
		}
		e.Matrix[issuingEms] = mx
	}
}

// Interface guards
var (
	_ collector.Collector = (*Ems)(nil)
)
