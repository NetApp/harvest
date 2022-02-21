package restperf

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	rest2 "goharvest2/cmd/collectors/rest"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"strings"
	"time"
)

type RestPerf struct {
	*rest2.Rest  // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	prop         *prop
	counterInfo  map[string]*counter
	isCacheEmpty bool
}

type counter struct {
	name        string
	description string
	counterType string
	units       string
	denominator string
}

type prop struct {
	object         string
	query          string
	instanceKeys   []string
	templatePath   string
	instanceLabels map[string]string
	metrics        map[string]metric
	counters       map[string]string
	returnTimeOut  string
	fields         []string
	apiType        string // public, private
}

type metric struct {
	label      string
	name       string
	metricType string
}

func init() {
	plugin.RegisterModule(RestPerf{})
}

func (RestPerf) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.restperf",
		New: func() plugin.Module { return new(RestPerf) },
	}
}

func (r *RestPerf) Init(a *collector.AbstractCollector) error {

	var err error

	r.Rest = &rest2.Rest{AbstractCollector: a}

	r.prop = &prop{}
	r.counterInfo = make(map[string]*counter)

	if err = r.InitClient(); err != nil {
		return err
	}

	if err, r.prop.templatePath = r.LoadTemplate(); err != nil {
		return err
	}

	if err = collector.Init(r); err != nil {
		return err
	}

	if err = r.InitCache(); err != nil {
		return err
	}

	if err = r.InitMatrix(); err != nil {
		return err
	}

	r.Logger.Info().Msgf("initialized cache with %d metrics", len(r.Matrix.GetMetrics()))

	return nil
}

func (r *RestPerf) InitMatrix() error {
	// overwrite from abstract collector
	r.Matrix.Object = r.prop.object
	// Add system (cluster) name
	r.Matrix.SetGlobalLabel("cluster", r.Client.Cluster().Name)
	if r.Params.HasChildS("labels") {
		for _, l := range r.Params.GetChildS("labels").GetChildren() {
			r.Matrix.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}
	return nil
}

func (r *RestPerf) InitCache() error {

	var (
		counters                        *node.Node
		display, name, kind, metricType string
	)

	if x := r.Params.GetChildContentS("object"); x != "" {
		r.prop.object = x
	} else {
		r.prop.object = strings.ToLower(r.Object)
	}

	if e := r.Params.GetChildS("export_options"); e != nil {
		r.Matrix.SetExportOptions(e)
	}

	if r.prop.query = r.Params.GetChildContentS("query"); r.prop.query == "" {
		return errors.New(errors.MISSING_PARAM, "query")
	}

	// create metric cache
	if counters = r.Params.GetChildS("counters"); counters == nil {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	// default value for ONTAP is 15 sec
	if returnTimeout := r.Params.GetChildContentS("return_timeout"); returnTimeout != "" {
		r.prop.returnTimeOut = returnTimeout
	}

	r.prop.instanceKeys = make([]string, 0)
	r.prop.instanceLabels = make(map[string]string)
	r.prop.counters = make(map[string]string)
	r.prop.metrics = make(map[string]metric)

	for _, c := range counters.GetAllChildContentS() {
		if c != "" {
			name, display, kind, metricType = util.ParseMetric(c)
			r.Logger.Debug().
				Str("kind", kind).
				Str("name", name).
				Str("display", display).
				Msg("Collected")

			r.prop.counters[name] = display
			switch kind {
			case "key":
				r.prop.instanceLabels[name] = display
				r.prop.instanceKeys = append(r.prop.instanceKeys, name)
			case "label":
				r.prop.instanceLabels[name] = display
			case "float":
				m := metric{label: display, name: name, metricType: metricType}
				r.prop.metrics[name] = m
			}
		}
	}

	r.prop.fields = []string{"*"}
	r.Logger.Info().Strs("extracted Instance Keys", r.prop.instanceKeys).Msg("")
	r.Logger.Info().Int("count metrics", len(r.prop.metrics)).Int("count labels", len(r.prop.instanceLabels)).Msg("initialized metric cache")

	return nil
}

func (r *RestPerf) PollCounter() (*matrix.Matrix, error) {
	var (
		content []byte
		err     error
		records []interface{}
	)

	counterQuery := r.prop.query[:strings.LastIndex(r.prop.query, "/")]

	href := rest.BuildHref(counterQuery, "", nil, "", "", "", r.prop.returnTimeOut, counterQuery)
	r.Logger.Debug().Str("href", href).Msg("")
	if href == "" {
		return nil, errors.New(errors.ERR_CONFIG, "empty url")
	}

	err = rest.FetchData(r.Client, href, &records)
	if err != nil {
		r.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, err
	}

	all := rest.Pagination{
		Records: records,
	}

	content, err = json.Marshal(all)
	if err != nil {
		r.Logger.Error().Err(err).Str("ApiPath", r.prop.query).Msg("Unable to marshal rest pagination")
	}

	if !gjson.ValidBytes(content) {
		return nil, fmt.Errorf("json is not valid for: %s", r.prop.query)
	}

	results := gjson.GetManyBytes(content, "records.0.counter_schemas", "records.0.name", "records.0.description")

	results[0].ForEach(func(key, c gjson.Result) bool {

		if !c.IsObject() {
			r.Logger.Warn().Str("type", c.Type.String()).Msg("Counter is not object, skipping")
			return true
		}

		name := c.Get("name").String()
		if _, has := r.prop.metrics[name]; has {
			if _, ok := r.counterInfo[name]; !ok {
				r.counterInfo[name] = &counter{
					name:        c.Get("name").String(),
					description: c.Get("description").String(),
					counterType: c.Get("type").String(),
					units:       c.Get("units").String(),
					denominator: c.Get("denominator.name").String(),
				}
			}
		} else {
			r.Logger.Trace().
				Str("key", name).
				Msg("Skip counter not requested")
		}
		return true
	})

	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	if r.Matrix.GetMetric("timestamp") == nil {
		m, err := r.Matrix.NewMetricFloat64("timestamp")
		if err != nil {
			r.Logger.Error().Stack().Err(err).Msg("add timestamp metric")
		}
		m.SetProperty("raw")
		m.SetExportable(false)
	}

	return r.Matrix, nil
}

func parseProperties(instanceData gjson.Result, property string) gjson.Result {
	t := gjson.Get(instanceData.String(), "properties.#.name")

	for _, name := range t.Array() {
		if name.String() == property {
			value := gjson.Get(instanceData.String(), "properties.#(name="+property+").value")
			return value
		}
	}
	return gjson.Result{}
}

func parseMetric(instanceData gjson.Result, metric string) gjson.Result {
	t := gjson.Get(instanceData.String(), "counters.#.name")

	for _, name := range t.Array() {
		if name.String() == metric {
			value := gjson.Get(instanceData.String(), "counters.#(name="+metric+").value")
			return value
		}
	}
	return gjson.Result{}
}

func (r *RestPerf) PollData() (*matrix.Matrix, error) {

	var (
		content      []byte
		count        uint64
		apiD, parseD time.Duration
		startTime    time.Time
		err          error
		records      []interface{}
	)

	r.Logger.Debug().Msg("updating data cache")

	// clone matrix without numeric data
	newData := r.Matrix.Clone(false, true, true)
	newData.Reset()

	timestamp := newData.GetMetric("timestamp")
	if timestamp == nil {
		return nil, errors.New(errors.ERR_CONFIG, "missing timestamp metric")
	}

	r.Matrix.Reset()

	startTime = time.Now()

	href := rest.BuildHref(r.prop.query, strings.Join(r.prop.fields, ","), nil, "", "", "", r.prop.returnTimeOut, r.prop.query)

	r.Logger.Debug().Str("href", href).Msg("")
	if href == "" {
		return nil, errors.New(errors.ERR_CONFIG, "empty url")
	}

	err = rest.FetchData(r.Client, href, &records)
	if err != nil {
		r.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, err
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}
	apiD = time.Since(startTime)

	content, err = json.Marshal(all)
	if err != nil {
		r.Logger.Error().Err(err).Str("ApiPath", r.prop.query).Msg("Unable to marshal rest pagination")
	}

	startTime = time.Now()
	if !gjson.ValidBytes(content) {
		return nil, fmt.Errorf("json is not valid for: %s", r.prop.query)
	}
	parseD = time.Since(startTime)

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no "+r.Object+" instances on cluster")
	}

	r.Logger.Debug().Str("object", r.Object).Str("number of records extracted", numRecords.String()).Msg("")

	results[1].ForEach(func(key, instanceData gjson.Result) bool {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			r.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
			return true
		}

		// extract instance key(s)
		for _, k := range r.prop.instanceKeys {
			value := parseProperties(instanceData, k)
			if value.Exists() {
				instanceKey += value.String()
			} else {
				r.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
				break
			}
		}

		if r.Params.GetChildContentS("only_cluster_instance") != "true" {
			if instanceKey == "" {
				return true
			}
		}

		if instance = r.Matrix.GetInstance(instanceKey); instance == nil {
			if instance, err = r.Matrix.NewInstance(instanceKey); err != nil {
				r.Logger.Error().Err(err).Str("Instance key", instanceKey).Msg("")
				return true
			}
		}

		for label, display := range r.prop.instanceLabels {
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
				r.Logger.Debug().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
			}
		}

		for name, metric := range r.prop.metrics {
			metr, ok := r.Matrix.GetMetrics()[name]
			if !ok {
				if metr, err = r.Matrix.NewMetricFloat64(name); err != nil {
					r.Logger.Error().Err(err).
						Str("name", name).
						Msg("NewMetricFloat64")
				}
			}
			f := parseMetric(instanceData, name)
			if f.Exists() {
				metr.SetName(metric.label)
				if err = metr.SetValueFloat64(instance, f.Float()); err != nil {
					r.Logger.Error().Err(err).Str("key", metric.name).Str("metric", metric.label).
						Msg("Unable to set float key on metric")
				}
				count++
			}
		}
		return true
	})

	if err != nil {
		r.Logger.Error().Err(err).Msg("Error while processing end points")
	}

	r.Logger.Info().
		Uint64("dataPoints", count).
		Str("apiTime", apiD.String()).
		Str("parseTime", parseD.String()).
		Msg("Collected")

	_ = r.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = r.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = r.Metadata.LazySetValueUint64("count", "data", count)
	r.AddCollectCount(count)

	return r.Matrix, nil
}

func (r *RestPerf) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	default:
		r.Logger.Warn().Str("kind", kind).Msg("no rest plugin found ")
	}
	return nil
}

// Interface guards
var (
	_ collector.Collector = (*RestPerf)(nil)
)
