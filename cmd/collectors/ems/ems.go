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
const severityFilter = "message.severity=alert|emergency|error|informational"
const emsEventMatrixPrefix = "ems_" //Used to clean up old instances excluding the parent
var emsBatchSize = 10

type Ems struct {
	*rest2.Rest     // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	Query           string
	TemplatePath    string
	emsProp         map[string]*emsProp
	Filter          []string
	Fields          []string
	ReturnTimeOut   string
	clusterTimezone *time.Location
	lastFilterTime  string
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
	Counters       map[string]string
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

func (e *Ems) Init(a *collector.AbstractCollector) error {

	var err error

	e.Rest = &rest2.Rest{AbstractCollector: a}
	e.emsProp = make(map[string]*emsProp)
	e.Fields = []string{"*"}

	e.InitProp()

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

	if e.Params.HasChildS("labels") {
		for _, l := range e.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}
	return nil
}

func (e *Ems) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	//handle custom plugins
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
			emsBatchSize = s
		}
	}
	e.Logger.Trace().Int("batch_size", emsBatchSize).Msgf("")

	if export := e.Params.GetChildS("export_options"); export != nil {
		e.Matrix[e.Object].SetExportOptions(export)
	}

	if e.Query = e.Params.GetChildContentS("query"); e.Query == "" {
		return errors.New(errors.MissingParam, "query")
	}

	// create metric cache
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
		prop.Counters = make(map[string]string)
		prop.Metrics = make(map[string]*Metric)

		for _, line1 := range line.GetChildren() {
			if line1.GetNameS() == "name" {
				prop.Name = line1.GetContentS()
				if prop.Name == "" {
					e.Logger.Warn().Msg("Missing event name for ems")
					continue
				}
			}
			if line1.GetNameS() == "exports" {
				if len(line1.GetAllChildContentS()) == 0 {
					e.Logger.Warn().Str("name", prop.Name).Msg("Missing exports for ems")
					continue
				}
				e.ParseRestCounters(line1, &prop)
			}
			if line1.GetNameS() == "matches" {
				e.ParseMatches(line1, &prop)
			}
			if line1.GetNameS() == "labels" {
				e.ParseLabels(line1, &prop)
			}
			if line1.GetNameS() == "plugins" {
				if prop.Plugins, err = e.LoadEmsPlugins(line1); err != nil {
					e.Logger.Error().Stack().Err(err).Msg("Failed to load plugin")
				}
				//set plugin at collector level
				e.Plugins[prop.Name] = prop.Plugins
			}
		}
		e.emsProp[prop.Name] = &prop
	}
	e.Filter = append(e.Filter, severityFilter)
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

// returns timefilter (clustertime - polldata duration)
func (e *Ems) getTimeStampFilter() string {
	dataDuration, err := GetDataInterval(e.GetParams(), DefaultDataPollDuration)
	if err != nil {
		e.Logger.Error().Stack().Err(err).Str("DefaultDataPollDuration", DefaultDataPollDuration.String()).Msg("Failed to parse duration. using default")
	}
	fromTime := time.Now().In(e.clusterTimezone).Add(-dataDuration).Format(time.RFC3339)
	if e.lastFilterTime != "" {
		fromTime = e.lastFilterTime
	}
	return "time=>=" + fromTime
}

func (e *Ems) fetchEMSData(names []string, filter []string) ([]interface{}, error) {
	var (
		records []interface{}
		err     error
	)
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
	for k := range e.Matrix {
		if strings.HasPrefix(k, emsEventMatrixPrefix) {
			delete(e.Matrix, k)
		}
	}

	startTime = time.Now()

	toTime := time.Now().In(e.clusterTimezone).Format(time.RFC3339)
	timeFilter := e.getTimeStampFilter()
	filter := append(e.Filter, timeFilter)
	//build filter
	var names []string
	for key := range e.emsProp {
		names = append(names, key)
	}

	// Split names into batches
	batch := emsBatchSize

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

	_, count = e.HandleResults(results[1], e.emsProp)

	e.Logger.Info().
		Uint64("instances", numRecords.Uint()).
		Uint64("dataPoints", count).
		Str("apiTime", apiD.String()).
		Str("parseTime", parseD.String()).
		Msg("Collected")

	_ = e.Metadata.LazySetValueInt64("count", "data", numRecords.Int())
	_ = e.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = e.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = e.Metadata.LazySetValueUint64("datapoint_count", "data", count)
	e.AddCollectCount(count)

	e.lastFilterTime = toTime
	return e.Matrix, nil
}

// fetch pollData interval
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
		value := gjson.Get(instanceData.String(), property)
		return value
	} else {
		property = strings.Replace(property, "parameters.", "", -1)
	}
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
func (e *Ems) HandleResults(result gjson.Result, prop map[string]*emsProp) (map[string]*matrix.Matrix, uint64) {
	var (
		err   error
		count uint64
		mx    *matrix.Matrix
	)

	var m = e.Matrix

	result.ForEach(func(key, instanceData gjson.Result) bool {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			e.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
			return true
		}
		messageName := instanceData.Get("message.name")
		if !messageName.Exists() {
			e.Logger.Warn().Msg("skip instance, missing message name")
			return true
		} else {
			k := emsEventMatrixPrefix + messageName.String()
			if _, ok := m[k]; !ok {
				mx = matrix.New(messageName.String(), e.Prop.Object, messageName.String())
				mx.SetGlobalLabels(e.Matrix[e.Object].GetGlobalLabels())
				m[k] = mx
			} else {
				mx = m[k]
			}
		}

		for _, p := range prop {

			if p.Name != messageName.String() {
				continue
			}

			//TODO matches implementation

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

			instance = mx.GetInstance(instanceKey)

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
					count++
				} else {
					// spams a lot currently due to missing label mappings. Moved to debug for now till rest gaps are filled
					e.Logger.Debug().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
				}
			}

			//set labels
			for k, v := range p.Labels {
				instance.SetLabel(k, v)
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
					// this code may not execute as ems only support events metric
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
		return true
	})
	return m, count
}

// Interface guards
var (
	_ collector.Collector = (*Ems)(nil)
)
