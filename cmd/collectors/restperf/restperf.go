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
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	latencyIoReqd = 10
	BILLION       = 1000000000
)

type RestPerf struct {
	*rest2.Rest // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	prop        *prop
}

type counter struct {
	name        string
	description string
	counterType string
	unit        string
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
	isCacheEmpty   bool
	counterInfo    map[string]*counter
	latencyIoReqd  int
}

type metric struct {
	label      string
	name       string
	metricType string
}

type metricOntapResponse struct {
	label   string
	value   string
	isArray bool
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
	r.prop.counterInfo = make(map[string]*counter)

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

	r.prop.latencyIoReqd = r.loadParamInt("latency_io_reqd", latencyIoReqd)
	r.prop.isCacheEmpty = true

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

// load a string parameter or use defaultValue
func (r *RestPerf) loadParamStr(name, defaultValue string) string {

	var x string

	if x = r.Params.GetChildContentS(name); x != "" {
		r.Logger.Debug().Msgf("using %s = [%s]", name, x)
		return x
	}
	r.Logger.Debug().Msgf("using %s = [%s] (default)", name, defaultValue)
	return defaultValue
}

// load an int parameter or use defaultValue
func (r *RestPerf) loadParamInt(name string, defaultValue int) int {

	var (
		x string
		n int
		e error
	)

	if x = r.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			r.Logger.Debug().Msgf("using %s = [%d]", name, n)
			return n
		}
		r.Logger.Warn().Msgf("invalid parameter %s = [%s] (expected integer)", name, x)
	}

	r.Logger.Debug().Msgf("using %s = [%d] (default)", name, defaultValue)
	return defaultValue
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
			if _, ok := r.prop.counterInfo[name]; !ok {
				r.prop.counterInfo[name] = &counter{
					name:        c.Get("name").String(),
					description: c.Get("description").String(),
					counterType: c.Get("type").String(),
					unit:        c.Get("unit").String(),
					denominator: c.Get("denominator.name").String(),
				}
				if p := r.GetOverride(name); p != "" {
					r.prop.counterInfo[name].counterType = p
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

func arrayMetricToString(value string) string {
	metricTypeRegex := regexp.MustCompile(`[(.*?)]`)
	r := strings.NewReplacer("\n", "", " ", "", "\"", "")
	s := r.Replace(value)

	match := metricTypeRegex.FindAllStringSubmatch(s, -1)
	if match != nil {
		name := match[0][1]
		return name
	}
	return value
}

func parseMetricOntapResponse(instanceData gjson.Result, metric string) *metricOntapResponse {
	t := gjson.Get(instanceData.String(), "counters.#.name")

	for _, name := range t.Array() {
		if name.String() == metric {
			value := gjson.Get(instanceData.String(), "counters.#(name="+metric+").value")
			if value.String() != "" {
				return &metricOntapResponse{value: value.String(), label: "", isArray: false}
			}
			values := gjson.Get(instanceData.String(), "counters.#(name="+metric+").values")
			if values.String() != "" {
				label := gjson.Get(instanceData.String(), "counters.#(name="+metric+").labels")
				return &metricOntapResponse{value: arrayMetricToString(values.String()), label: arrayMetricToString(label.String()), isArray: true}
			}
		}
	}
	return &metricOntapResponse{}
}

// override counter property
func (r *RestPerf) GetOverride(counter string) string {
	if o := r.Params.GetChildS("override"); o != nil {
		return o.GetChildContentS(counter)
	}
	return ""
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

	startTime = time.Now()

	href := rest.BuildHref(r.prop.query, strings.Join(r.prop.fields, ","), nil, "", "", "", r.prop.returnTimeOut, r.prop.query)

	r.Logger.Debug().Str("href", href).Msg("")
	if href == "" {
		return nil, errors.New(errors.ERR_CONFIG, "empty url")
	}

	ts := float64(time.Now().UnixNano()) / BILLION
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

		if instance = newData.GetInstance(instanceKey); instance == nil {
			if instance, err = newData.NewInstance(instanceKey); err != nil {
				r.Logger.Error().Err(err).Str("Instance key", instanceKey).Msg("")
				return true
			}
		}

		//// add batch timestamp as custom counter

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
			f := parseMetricOntapResponse(instanceData, name)
			if f.value != "" {
				if f.isArray {
					labels := strings.Split(f.label, ",")
					values := strings.Split(f.value, ",")

					if len(labels) != len(values) {
						// warn & skip
						r.Logger.Error().
							Stack().
							Str("labels", f.label).
							Str("value", f.value).
							Msg("labels don't match parsed values")
						continue
					}

					for i, label := range labels {
						metr, ok := newData.GetMetrics()[name+"."+label]
						if !ok {
							if metr, err = newData.NewMetricFloat64(name + "." + label); err != nil {
								r.Logger.Error().Err(err).
									Str("name", name+"."+label).
									Msg("NewMetricFloat64")
								continue
							}
							metr.SetName(name)
							metr.SetLabel("metric", label)
							// differentiate between array and normal counter
							metr.SetComment(name)
						}
						if err = metr.SetValueString(instance, values[i]); err != nil {
							r.Logger.Error().
								Stack().
								Err(err).
								Str("name", name).
								Str("label", label).
								Str("value", values[i]).
								Msg("Set value failed")
							continue
						} else {
							r.Logger.Trace().
								Str("name", name).
								Str("label", label).
								Str("value", values[i]).
								Msg("Set name.label = value")
							count++
						}
					}
				} else {
					metr, ok := newData.GetMetrics()[name]
					if !ok {
						if metr, err = newData.NewMetricFloat64(name); err != nil {
							r.Logger.Error().Err(err).
								Str("name", name).
								Msg("NewMetricFloat64")
						}
					}
					metr.SetName(metric.label)
					if c, err := strconv.ParseFloat(f.value, 64); err == nil {
						if err = metr.SetValueFloat64(instance, c); err != nil {
							r.Logger.Error().Err(err).Str("key", metric.name).Str("metric", metric.label).
								Msg("Unable to set float key on metric")
						}
					} else {
						r.Logger.Error().Err(err).Str("key", metric.name).Str("metric", metric.label).
							Msg("Unable to set float key on metric")
					}
					count++
				}
			}
		}
		if err = newData.GetMetric("timestamp").SetValueFloat64(instance, ts); err != nil {
			r.Logger.Error().Stack().Err(err).Msg("set timestamp value: ")
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

	// skip calculating from delta if no data from previous poll
	if r.prop.isCacheEmpty {
		r.Logger.Debug().Msg("skip postprocessing until next poll (previous cache empty)")
		r.Matrix = newData
		r.prop.isCacheEmpty = false
		return nil, nil
	}

	calcStart := time.Now()

	r.Logger.Debug().Msg("starting delta calculations from previous cache")

	// cache raw data for next poll
	cachedData := newData.Clone(true, true, true)
	// order metrics, such that those requiring base counters are processed last
	orderedMetrics := make([]matrix.Metric, 0, len(newData.GetMetrics()))
	orderedKeys := make([]string, 0, len(orderedMetrics))

	for key, metric := range newData.GetMetrics() {
		if metric.GetName() != "timestamp" {
			counter := r.prop.counterInfo[metric.GetName()]
			if counter == nil {
				// check if it is an array counter
				counter = r.prop.counterInfo[metric.GetComment()]
			}
			if counter != nil {
				if counter.denominator == "" {
					// does not require base counter
					orderedMetrics = append(orderedMetrics, metric)
					orderedKeys = append(orderedKeys, key)
				}
			} else {
				r.Logger.Warn().Str("counter", metric.GetName()).Msg("Misconfiguration in template counter ")
			}
		}
	}

	for key, metric := range newData.GetMetrics() {
		if metric.GetName() != "timestamp" {
			counter := r.prop.counterInfo[metric.GetName()]
			if counter == nil {
				// check if it is an array counter
				counter = r.prop.counterInfo[metric.GetComment()]
			}
			if counter != nil {
				if counter.denominator != "" {
					// does not require base counter
					orderedMetrics = append(orderedMetrics, metric)
					orderedKeys = append(orderedKeys, key)
				}
			} else {
				r.Logger.Warn().Str("counter", metric.GetName()).Msg("Misconfiguration in template counter ")
			}
		}
	}

	// calculate timestamp delta first since many counters require it for postprocessing.
	// Timestamp has "raw" property, so it isn't post-processed automatically
	if err = timestamp.Delta(r.Matrix.GetMetric("timestamp")); err != nil {
		r.Logger.Error().Stack().Err(err).Msg("(timestamp) calculate delta:")
	}

	var base matrix.Metric

	for i, metric := range orderedMetrics {
		counter := r.prop.counterInfo[metric.GetName()]
		if counter == nil {
			// check if it is an array counter
			counter = r.prop.counterInfo[metric.GetComment()]
		}
		if counter == nil {
			r.Logger.Error().Stack().Err(err).Str("counter", metric.GetName()).Msg("Missing counter:")
		}
		property := counter.counterType
		key := orderedKeys[i]

		// RAW - submit without post-processing
		if property == "raw" {
			continue
		}

		// all other properties - first calculate delta
		if err = metric.Delta(r.Matrix.GetMetric(key)); err != nil {
			r.Logger.Error().Stack().Err(err).Str("key", key).Msg("Calculate delta")
			continue
		}

		// DELTA - subtract previous value from current
		if property == "delta" {
			// already done
			continue
		}

		// RATE - delta, normalized by elapsed time
		if property == "rate" {
			// defer calculation, so we can first calculate averages/percents
			// Note: calculating rate before averages are averages/percentages are calculated
			// used to be a bug in Harvest 2.0 (Alpha, RC1, RC2) resulting in very high latency values
			continue
		}

		// For the next two properties we need base counters
		// We assume that delta of base counters is already calculated
		// (name of base counter is stored as Comment)
		if base = newData.GetMetric(counter.denominator); base == nil {
			r.Logger.Warn().
				Str("key", key).
				Str("property", property).
				Str("denominator", counter.denominator).
				Msg("Base counter missing")
			continue
		}

		// remaining properties: average and percent
		//
		// AVERAGE - delta, divided by base-counter delta
		//
		// PERCENT - average * 100
		// special case for latency counter: apply minimum number of iops as threshold
		if property == "average" || property == "percent" {

			if strings.HasSuffix(metric.GetName(), "latency") {
				err = metric.DivideWithThreshold(base, r.prop.latencyIoReqd)
			} else {
				err = metric.Divide(base)
			}

			if err != nil {
				r.Logger.Error().Stack().Err(err).Str("key", key).Msg("Division by base")
			}

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if err = metric.MultiplyByScalar(100); err != nil {
				r.Logger.Error().Stack().Err(err).Str("key", key).Msg("Multiply by scalar")
			}
			continue
		}
		r.Logger.Error().Stack().Err(err).
			Str("key", key).
			Str("property", property).
			Msg("Unknown property")
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		counter := r.prop.counterInfo[metric.GetName()]
		if counter == nil {
			// check if it is an array counter
			counter = r.prop.counterInfo[metric.GetComment()]
		}
		if counter == nil {
			r.Logger.Error().Stack().Err(err).Str("counter", metric.GetName()).Msg("Missing counter:")
		}
		property := counter.counterType
		if property == "rate" {
			if err = metric.Divide(timestamp); err != nil {
				r.Logger.Error().Stack().Err(err).
					Int("i", i).
					Str("key", orderedKeys[i]).
					Msg("Calculate rate")
			}
		}
	}

	_ = r.Metadata.LazySetValueInt64("calc_time", "data", time.Since(calcStart).Microseconds())
	// store cache for next poll
	r.Matrix = cachedData

	return newData, nil
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
