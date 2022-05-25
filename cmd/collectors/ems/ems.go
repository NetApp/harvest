package ems

import (
	"encoding/json"
	"fmt"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/errors"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

const DefaultDataPollDuration = 3 * time.Minute
const DefaultBatchSize = 25
const DefaultSeverityFilter = "message.severity=alert|emergency|error|informational|notice"
const emsEventMatrixPrefix = "ems#" //Used to clean up old instances excluding the parent

type Ems struct {
	*rest2.Rest     // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	Query           string
	TemplatePath    string
	emsProp         map[string][]*emsProp
	Filter          []string
	Fields          []string
	ReturnTimeOut   string
	clusterTimezone *time.Location
	lastFilterTime  string
	batchSize       int
	DefaultLabels   []string
	severityFilter  string
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
	plugin.RegisterModule(Ems{})
}

func (Ems) HarvestModule() plugin.ModuleInfo {
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
	e.batchSize = DefaultBatchSize
	e.severityFilter = DefaultSeverityFilter

	// init Rest props
	e.InitProp()
	// init ems props
	e.InitEmsProp()

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

	if err = e.getClusterTimeZone(); err != nil {
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

func (e *Ems) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	default:
		e.Logger.Warn().Str("kind", kind).Msg("no ems plugin found ")
	}
	return nil
}

func (e *Ems) InitCache() error {

	var (
		events *node.Node
		err    error
	)

	if x := e.Params.GetChildContentS("object"); x != "" {
		e.Prop.Object = x
	} else {
		e.Prop.Object = strings.ToLower(e.Object)
	}

	if b := e.Params.GetChildContentS("batch_size"); b != "" {
		if s, err := strconv.Atoi(b); err == nil {
			e.batchSize = s
		}
	}
	e.Logger.Debug().Int("batch_size", e.batchSize).Msgf("")

	if s := e.Params.GetChildContentS("severity"); s != "" {
		e.severityFilter = s
	}
	e.Logger.Debug().Str("severityFilter", e.severityFilter).Msgf("")

	if export := e.Params.GetChildS("export_options"); export != nil {
		e.Matrix[e.Object].SetExportOptions(export)
	}

	if e.Query = e.Params.GetChildContentS("query"); e.Query == "" {
		return errors.New(errors.MissingParam, "query")
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

	if events = e.Params.GetChildS("events"); events == nil {
		return errors.New(errors.MissingParam, "events")
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
		if line.GetChildS("name") == nil || line.GetChildS("name").GetContentS() == "" {
			e.Logger.Warn().Msg("Missing event name")
			continue
		}

		//populate prop counter for asup
		eventName := line.GetChildS("name").GetContentS()
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
				if err = e.LoadPlugins(line1, e, prop.Name); err != nil {
					e.Logger.Error().Stack().Err(err).Msg("Failed to load plugin")
				}
			}
		}
		e.emsProp[prop.Name] = append(e.emsProp[prop.Name], &prop)
	}
	//add severity filter
	e.Filter = append(e.Filter, e.severityFilter)
	return nil
}

func (e *Ems) getClusterTimeZone() error {
	var (
		content []byte
		err     error
		records []interface{}
	)

	// /api/cluster?fields=timezone is supported from 9.7. Using private CLI to support 9.6
	query := "private/cli/cluster/date"
	fields := []string{"timezone"}

	href := rest.BuildHref(query, strings.Join(fields, ","), nil, "", "", "1", e.ReturnTimeOut, "")

	if records, err = e.GetRestData(href); err != nil {
		e.Logger.Error().Stack().Err(err).Msg("Failed to fetch data")
		return err
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}

	content, err = json.Marshal(all)
	if err != nil {
		e.Logger.Error().Err(err).Str("ApiPath", e.Query).Msg("Unable to marshal rest pagination")
	}

	if !gjson.ValidBytes(content) {
		return fmt.Errorf("json is not valid for: %s", e.Query)
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return errors.New(errors.ErrNoInstance, e.Object+" timezone not found on cluster")
	}

	results[1].ForEach(func(key, instanceData gjson.Result) bool {
		timezone := instanceData.Get("timezone")
		if timezone.Exists() {
			loc, err := time.LoadLocation(timezone.String())
			if err != nil {
				e.Logger.Error().Stack().Err(err).Msg("Failed to set cluster time")
				return true
			} else {
				e.clusterTimezone = loc
			}
		}
		return true
	})
	if e.clusterTimezone == nil {
		return errors.New(errors.ErrNoInstance, e.Object+" timezone not found on cluster")
	} else {
		e.Logger.Info().Str("cluster time zone", e.clusterTimezone.String()).Msg("")
		return nil
	}
}

// returns time filter (clustertime - polldata duration)
func (e *Ems) getTimeStampFilter() string {
	var fromTime string
	// check if not first request
	if e.lastFilterTime != "" {
		fromTime = e.lastFilterTime
	} else {
		// if first request fetch cluster time
		dataDuration, err := GetDataInterval(e.GetParams(), DefaultDataPollDuration)
		if err != nil {
			e.Logger.Warn().Err(err).Str("DefaultDataPollDuration", DefaultDataPollDuration.String()).Msg("Failed to parse duration. using default")
		}
		fromTime = time.Now().In(e.clusterTimezone).Add(-dataDuration).Format(time.RFC3339)
	}
	return "time=>=" + fromTime
}

func (e *Ems) fetchEMSData(names []string, filter []string) ([]interface{}, error) {
	var (
		records []interface{}
		err     error
	)
	//event name filter
	nameFilter := "message.name=" + strings.Join(names[:], ",")
	filter = append(filter, nameFilter)

	href := rest.BuildHref(e.Query, strings.Join(e.Fields, ","), filter, "", "", "", e.ReturnTimeOut, e.Query)

	e.Logger.Debug().Str("href", href).Msg("")
	if records, err = e.GetRestData(href); err != nil {
		e.Logger.Error().Stack().Err(err).Msg("Failed to fetch data")
		return nil, err
	}
	return records, nil
}

func (e *Ems) PollData() (map[string]*matrix.Matrix, error) {

	var (
		content      []byte
		count        uint64
		apiD, parseD time.Duration
		startTime    time.Time
		err          error
		records      []interface{}
	)

	e.Logger.Debug().Msg("starting data poll")
	// remove instances from last poll except the parent object
	for k := range e.Matrix {
		if strings.HasPrefix(k, emsEventMatrixPrefix) {
			delete(e.Matrix, k)
		}
	}

	startTime = time.Now()

	// add time filter
	toTime := time.Now().In(e.clusterTimezone).Format(time.RFC3339)
	timeFilter := e.getTimeStampFilter()
	filter := append(e.Filter, timeFilter)

	// collect all event names
	var names []string
	for key := range e.emsProp {
		names = append(names, key)
	}

	// Split names into batches
	batch := e.batchSize

	for i := 0; i < len(names); i += batch {
		j := i + batch
		if j > len(names) {
			j = len(names)
		}
		r, err := e.fetchEMSData(names[i:j], filter)
		if err != nil {
			e.Logger.Error().Stack().Err(err).Msg("Failed to fetch data")
			return nil, err
		} else {
			records = append(records, r...)
		}
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}
	apiD = time.Since(startTime)

	content, err = json.Marshal(all)
	if err != nil {
		e.Logger.Error().Err(err).Str("ApiPath", e.Query).Msg("Unable to marshal rest pagination")
	}

	if !gjson.ValidBytes(content) {
		return nil, fmt.Errorf("json is not valid for: %s", e.Query)
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, errors.New(errors.ErrNoInstance, "no "+e.Object+" instances on cluster")
	}

	e.Logger.Debug().Str("object", e.Object).Str("number of records extracted", numRecords.String()).Msg("")

	startTime = time.Now()
	_, count = e.HandleResults(results[1], e.emsProp)
	parseD = time.Since(startTime)

	var instanceCount uint64 = 0
	for _, v := range e.Matrix {
		instanceCount += uint64(len(v.GetInstances()))
	}

	e.Logger.Info().
		Uint64("instances", instanceCount).
		Uint64("dataPoints", count).
		Str("apiTime", apiD.String()).
		Str("parseTime", parseD.String()).
		Msg("Collected")

	_ = e.Metadata.LazySetValueInt64("count", "data", numRecords.Int())
	_ = e.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = e.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = e.Metadata.LazySetValueUint64("datapoint_count", "data", count)
	e.AddCollectCount(count)

	// update lastfiltertime to current cluster time
	e.lastFilterTime = toTime
	return e.Matrix, nil
}

// GetDataInterval fetch pollData interval
func GetDataInterval(param *node.Node, defaultInterval time.Duration) (time.Duration, error) {
	var dataIntervalStr = ""
	schedule := param.GetChildS("schedule")
	if schedule != nil {
		dataInterval := schedule.GetChildS("data")
		if dataInterval != nil {
			dataIntervalStr = dataInterval.GetContentS()
			if durationVal, err := time.ParseDuration(dataIntervalStr); err == nil {
				return durationVal, nil
			} else {
				return defaultInterval, err
			}
		}
	}
	return defaultInterval, nil
}

func parseProperties(instanceData gjson.Result, property string) gjson.Result {

	if !strings.HasPrefix(property, "parameters.") {
		// if prefix is not parameters.
		value := gjson.Get(instanceData.String(), property)
		return value
	} else {
		//strip parameters. from property name
		property = strings.Replace(property, "parameters.", "", -1)
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
func (e *Ems) HandleResults(result gjson.Result, prop map[string][]*emsProp) (map[string]*matrix.Matrix, uint64) {
	var (
		err   error
		count uint64
		mx    *matrix.Matrix
	)

	var m = e.Matrix

	result.ForEach(func(key, instanceData gjson.Result) bool {
		var (
			instanceKey string
		)

		var instanceLabelCount uint64 = 0

		if !instanceData.IsObject() {
			e.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
			return true
		}
		messageName := instanceData.Get("message.name")
		// verify if message name exists in ontap response
		if !messageName.Exists() {
			e.Logger.Warn().Msg("skip instance, missing message name")
			return true
		} else {
			k := emsEventMatrixPrefix + messageName.String()
			if _, ok := m[k]; !ok {
				//create matrix if not exists for the ems event
				mx = matrix.New(messageName.String(), e.Prop.Object, messageName.String())
				mx.SetGlobalLabels(e.Matrix[e.Object].GetGlobalLabels())
				m[k] = mx
			} else {
				mx = m[k]
			}
		}

		//parse ems properties for the instance
		isMatch := false
		if ps, ok := prop[messageName.String()]; ok {
			for _, p := range ps {
				instanceKey = ""
				instanceLabelCount = 0
				// extract instance key(s)
				for _, k := range p.InstanceKeys {
					value := parseProperties(instanceData, k)
					if value.Exists() {
						instanceKey += value.String()
					} else {
						e.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
						break
					}
				}

				instance := mx.GetInstance(instanceKey)

				if instance == nil {
					if instance, err = mx.NewInstance(instanceKey); err != nil {
						e.Logger.Error().Err(err).Str("Instance key", instanceKey).Msg("")
						return true
					}
				}

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
					for _, v := range p.Matches {
						if value := instance.GetLabel(v.Name); value != "" {
							if value == v.value {
								isMatch = true
								break
							}
						} else {
							//value not found
							e.Logger.Warn().Str("Instance key", instanceKey).Str("name", v.Name).Str("value", v.value).Msg("label is not found")
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
								Msg("NewMetricFloat64")
						}
					}
					if metric.Name == "events" {
						if err = metr.SetValueFloat64(instance, 1); err != nil {
							e.Logger.Error().Err(err).Str("key", metric.Name).Str("metric", metric.Label).
								Msg("Unable to set float key on metric")
						}
					} else {
						// this code will not execute as ems only support events metric
						f := instanceData.Get(metric.Name)
						if f.Exists() {
							if err = metr.SetValueFloat64(instance, f.Float()); err != nil {
								e.Logger.Error().Err(err).Str("key", metric.Name).Str("metric", metric.Label).
									Msg("Unable to set float key on metric")
							}
						}
					}
				}
			}
		}
		if !isMatch {
			mx.RemoveInstance(instanceKey)
			return true
		} else {
			count += instanceLabelCount
		}
		return true
	})
	return m, count
}

// Interface guards
var (
	_ collector.Collector = (*Ems)(nil)
)
