package main

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"
	//zapi_collector "goharvest2/collectors/zapi/collector"

	"goharvest2/poller/collector"
	"goharvest2/share/dict"
	"goharvest2/share/errors"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"goharvest2/share/set"
	"goharvest2/share/tree/node"
	"goharvest2/share/util"

	client "goharvest2/apis/zapi"
)

// default parameter values
const (
	INSTANCE_KEY    = "uuid"
	BATCH_SIZE      = 500
	LATENCY_IO_REQD = 0 //10
)

type ZapiPerf struct {
	*collector.AbstractCollector
	//*zapi_collector.Zapi  // provides: Connection, System, Object, Query, TemplateFn, TemplateType
	Connection      *client.Client
	System          *client.System
	object          string
	Query           string
	TemplateFn      string
	TemplateType    string
	batch_size      int
	latency_io_reqd int
	instance_key    string
	array_labels    map[string][]string
}

func New(a *collector.AbstractCollector) collector.Collector {
	return &ZapiPerf{AbstractCollector: a}
}

/*
func New(a *collector.AbstractCollector) collector.Collector {
	z := zapi_collector.NewZapi(a)
    return &ZapiPerf{Zapi: z}
}*/

/* Copied from Zapi
Don't change => merge!
*/

func (c *ZapiPerf) Init() error {

	// @TODO check if cert/key files exist
	if c.Params.GetChildContentS("auth_style") == "certificate_auth" {
		if c.Params.GetChildS("ssl_cert") == nil {
			cert_path := path.Join(c.Options.ConfPath, "cert", c.Options.Poller+".pem")
			c.Params.NewChildS("ssl_cert", cert_path)
			logger.Debug(c.Prefix, "added ssl_cert path [%s]", cert_path)
		}

		if c.Params.GetChildS("ssl_key") == nil {
			key_path := path.Join(c.Options.ConfPath, "cert", c.Options.Poller+".key")
			c.Params.NewChildS("ssl_key", key_path)
			logger.Debug(c.Prefix, "added ssl_key path [%s]", key_path)
		}
	}

	var err error
	if c.Connection, err = client.New(c.Params); err != nil {
		return err
	}

	// @TODO handle connectivity-related errors (retry a few times)
	if c.System, err = c.Connection.GetSystem(); err != nil {
		//logger.Error(c.Prefix, "system info: %v", err)
		return err
	}
	logger.Debug(c.Prefix, "Connected to: %s", c.System.String())

	c.TemplateFn = c.Params.GetChildS("objects").GetChildContentS(c.Object) // @TODO err handling

	model := "cdot"
	if !c.System.Clustered {
		model = "7mode"
	}
	template, err := c.ImportSubTemplate(model, "default", c.TemplateFn, c.System.Version)
	if err != nil {
		logger.Error(c.Prefix, "Error importing subtemplate: %s", err)
		return err
	}
	c.Params.Union(template)

	// object name from subtemplate
	if c.object = c.Params.GetChildContentS("object"); c.object == "" {
		return errors.New(errors.MISSING_PARAM, "object")
	}

	// api query literal
	if c.Query = c.Params.GetChildContentS("query"); c.Query == "" {
		return errors.New(errors.MISSING_PARAM, "query")
	}

	// Invoke generic initializer
	// this will load Schedule, initialize Data and Metadata
	if err := collector.Init(c); err != nil {
		return err
	}

	// overwrite from abstract collector
	c.Data.Object = c.object

	// Add system (cluster) name
	c.Data.SetGlobalLabel("cluster", c.System.Name)
	if !c.System.Clustered {
		c.Data.SetGlobalLabel("node", c.System.Name)
	}

	// Initialize counter cache
	counters := c.Params.GetChildS("counters")
	if counters == nil {
		return errors.New(errors.MISSING_PARAM, "counters")
	}

	if err = c.InitCache(); err != nil {
		return err
	}

	logger.Debug(c.Prefix, "Successfully initialized")
	return nil
}

func (c *ZapiPerf) InitCache() error {
	c.array_labels = make(map[string][]string)
	c.instance_key = c.loadParamStr("instance_key", INSTANCE_KEY)
	c.batch_size = c.loadParamInt("batch_size", BATCH_SIZE)
	c.latency_io_reqd = c.loadParamInt("latency_io_reqd", LATENCY_IO_REQD)
	return nil
}

func (c *ZapiPerf) loadParamStr(name, default_value string) string {

	var x string

	if x = c.Params.GetChildContentS(name); x != "" {
		logger.Debug(c.Prefix, "using %s = [%s]", name, x)
		return x
	}
	logger.Debug(c.Prefix, "using %s = [%s] (default)", name, default_value)
	return default_value
}

func (c *ZapiPerf) loadParamInt(name string, default_value int) int {

	var x string
	var n int
	var e error

	if x = c.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			logger.Debug(c.Prefix, "using %s = [%d]", name, n)
			return n
		}
		logger.Warn(c.Prefix, "invalid parameter %s = [%s] (expected integer)", name, x)
	}

	logger.Debug(c.Prefix, "using %s = [%d] (default)", name, default_value)
	return default_value
}

func (c *ZapiPerf) PollData() (*matrix.Matrix, error) {

	var err error
	logger.Debug(c.Prefix, "Updating data cache")

	NewData := c.Data.Clone(false)
	if err = NewData.InitData(); err != nil {
		return nil, err
	}

	timestamp := NewData.GetMetric("timestamp")
	if timestamp == nil {
		return nil, errors.New(errors.ERR_CONFIG, "missing timestamp metric")
	}

	// for updating metadata
	count := 0
	batch_count := 0
	api_d := time.Duration(0 * time.Second)
	parse_d := time.Duration(0 * time.Second)

	// determine what will serve as instance key (either "uuid" or "instance")
	key_name := "instance-uuid"
	if c.instance_key == "name" {
		key_name = "instance"
	}

	// list of instance keys (instance names or uuids) for which
	// we will request counter data
	instance_keys := make([]string, 0, len(NewData.Instances))
	for key := range NewData.GetInstances() {
		instance_keys = append(instance_keys, key)
	}

	// build ZAPI request
	request := node.NewXmlS("perf-object-get-instances")
	request.NewChildS("objectname", c.Query)

	// load requested counters (metrics + labels)
	request_counters := request.NewChildS("counters", "")
	for key := range NewData.GetMetrics() {
		request_counters.NewChildS("counter", key)
	}
	for key := range NewData.GetLabels() {
		request_counters.NewChildS("counter", key)
	}

	// batch indices
	start_index := 0
	end_index := 0

	for end_index < len(instance_keys) {

		// update batch indices
		end_index += c.batch_size
		if end_index > len(instance_keys) {
			end_index = len(instance_keys)
		}

		logger.Debug(c.Prefix, "starting batch poll for instances [%d:%d]", start_index, end_index)

		request.PopChildS(key_name + "s")
		request_instances := request.NewChildS(key_name+"s", "")
		for _, key := range instance_keys[start_index:end_index] {
			request_instances.NewChildS(key_name, key)
		}

		start_index = end_index

		if err = c.Connection.BuildRequest(request); err != nil {
			logger.Error(c.Prefix, "build request: %v", err)
			//break
			return nil, err
		}

		response, rd, pd, err := c.Connection.InvokeWithTimers()
		if err != nil {
			//logger.Error(c.Prefix, "data request: %v", err)
			//@TODO handle "resource limit exceeded"
			//break
			return nil, err
		}

		api_d += rd
		parse_d += pd
		batch_count += 1

		// fetch instances
		instances := response.GetChildS("instances")
		if instances == nil || len(instances.GetChildren()) == 0 {
			err = errors.New(errors.ERR_NO_INSTANCE, "")
			break
		}

		logger.Debug(c.Prefix, "fetched batch with %d instances", len(instances.GetChildren()))

		// timestamp for batch instances
		// ignore timestamp from ZAPI which is always integer
		// we want float, since our poll interval can be float
		ts := float64(time.Now().UnixNano()) / 1000000000

		for _, i := range instances.GetChildren() {

			key := i.GetChildContentS(c.instance_key)
			if key == "" {
				logger.Debug(c.Prefix, "skip instance, no key [%s] (name=%s, uuid=%s)", c.instance_key, i.GetChildContentS("name"), i.GetChildContentS("uuid"))
				continue
			}

			instance := NewData.GetInstance(key)
			if instance == nil {
				logger.Debug(c.Prefix, "skip instance [%s], not found in cache", key)
				continue
			}

			counters := i.GetChildS("counters")
			if counters == nil {
				logger.Debug(c.Prefix, "skip instance [%s], no data counters", key)
				continue
			}

			logger.Debug(c.Prefix, "fetching data of instance [%s]", key)

			// add batch timestamp as custom counter
			NewData.SetValue(timestamp, instance, ts)

			for _, cnt := range counters.GetChildren() {

				name := cnt.GetChildContentS("name")
				value := cnt.GetChildContentS("value")

				// sanity check
				if name == "" || value == "" {
					logger.Debug(c.Prefix, "skipping raw counter [%s] with value [%s]", name, value)
					continue
				}

				// ZAPI counter for us is either instance label (string)
				// or numeric metric (scalar or string)

				if _, has := NewData.GetLabel(name); has { // @TODO implement
					NewData.SetInstanceLabel(instance, name, value)
					logger.Debug(c.Prefix, "+ label data [%s= %s%s%s]", name, util.Yellow, value, util.End)
					continue
				}

				// process array counter
				if strings.Contains(value, ",") {
					labels, ok := c.array_labels[name]
					if ok {
						// warn & skip
						logger.Error(c.Prefix, "metric [%s] array labels not in cache, skip", name, value)
						continue
					}
					values := strings.Split(string(value), ",")
					if len(labels) != len(values) {
						// warn & skip
						logger.Error(c.Prefix, "metric [%s] array labels don't match with values (%d), skip", name, len(values))
						continue
					}

					for i, label := range labels {
						key := name + "." + label
						if m := NewData.GetMetric(key); m != nil {
							if e := NewData.SetValueString(m, instance, values[i]); e != nil {
								logger.Error(c.Prefix, "set metric [%s] with value [%s]: %v", key, values[i], e)
							} else {
								logger.Debug(c.Prefix, "+ data [%s] = [%s%s%s]", key, util.Pink, values[i], util.End)
								count += 1
							}
						} else {
							logger.Error(c.Prefix, "metric [%s] not in cache, skip", key, value)
						}
					}
					// process scalar counter
				} else {
					if m := NewData.GetMetric(name); m != nil {
						if e := NewData.SetValueString(m, instance, value); e != nil {
							logger.Error(c.Prefix, "set metric [%s] with value [%s]: %v", name, value, e)
						} else {
							logger.Debug(c.Prefix, "+ data [%s] = [%s%s%s]", name, util.Cyan, value, util.End)
							count += 1
						}
					} else {
						logger.Error(c.Prefix, "metric [%s] not in cache, skip", name, value)
					}
				}
			} // end loop over counters
		} // end loop over instances
	} // end batch request

	// terminate if serious errors
	// @TODO handle...

	if err != nil {
		return nil, err
	}

	// update metadata
	c.Metadata.SetValueSS("api_time", "data", float64(api_d.Microseconds()))
	c.Metadata.SetValueSS("parse_time", "data", float64(parse_d.Microseconds()))
	c.Metadata.SetValueSS("count", "data", float64(count))
	c.AddCount(count)

	logger.Debug(c.Prefix, "collected data: %d batch polls, %d data points", batch_count, count)

	// skip calculating from delta if no data from previous poll
	if c.Data.IsEmpty() {
		logger.Debug(c.Prefix, "no postprocessing until next poll (new data empty: %v)", NewData.IsEmpty())
		c.Data = NewData
		return nil, nil
	}

	logger.Debug(c.Prefix, "starting delta calculations from previous poll")
	logger.Debug(c.Prefix, "data has dimensions (%d x %d)", len(NewData.Data), len(NewData.Data[0]))

	calc_start := time.Now()

	// cache data, to store after calculations
	CachedData := NewData.Clone(true)

	// order metrics, such that those requiring base counters are processed last
	ordered_metrics := make([]*matrix.Metric, 0, len(NewData.Metrics))
	for _, m := range NewData.Metrics {
		if m.BaseCounter == "" { // does not require base counter
			ordered_metrics = append(ordered_metrics, m)
		}
	}
	for _, m := range NewData.Metrics {
		if m.BaseCounter != "" { // requires base counter
			ordered_metrics = append(ordered_metrics, m)
		}
	}

	// calculate timestamp delta first since many counters require it for postprocessing
	// timestamp has "raw" property, so won't be postprocessed automatically
	// fmt.Printf("\npostprocessing %s%s%s - %s%v%s\n", util.Red, timestamp.Name, util.End, util.Bold, timestamp.Properties, util.End)
	logger.Debug(c.Prefix, "cooking [%s] (%s)", timestamp.Name, timestamp.Properties)
	print_vector("current", NewData.Data[timestamp.Index])
	print_vector("previous", c.Data.Data[timestamp.Index])
	NewData.Delta(c.Data, timestamp.Index)
	print_vector(util.Green+"delta"+util.End, NewData.Data[timestamp.Index])

	for _, m := range ordered_metrics {

		// raw counters don't require postprocessing
		if strings.Contains(m.Properties, "raw") {
			continue
		}

		//logger.Debug(c.Prefix, "Postprocessing %s%s%s (%s%v%s)", util.Red, m.Name, util.End, util.Bold, m.Properties, util.End)
		// fmt.Printf("\npostprocessing %s%s%s - %s%v%s\n", util.Red, m.Name, util.End, util.Bold, m.Properties, util.End)
		// scalar not depending on base counter

		if m.BaseCounter == "" {
			// fmt.Printf("scalar - no basecounter\n")
			logger.Debug(c.Prefix, "cooking [%d] [%s%s%s] (%s)", m.Index, util.Cyan, m.Name, util.End, m.Properties)
			c.calculate_from_delta(NewData, m.Name, m.Index, -1, m.Properties) // -1 indicates no base counter
		} else if b := NewData.GetMetric(m.BaseCounter); b != nil {
			// fmt.Printf("scalar - with basecounter %s%s%s (%s)\n", util.Red, m.BaseCounter, util.End, b.Properties)
			logger.Debug(c.Prefix, "cooking [%d] [%s%s%s] (%s) using base counter [%s] (%s)", m.Index, util.Cyan, m.Name, util.End, m.Properties, b.Name, b.Properties)
			c.calculate_from_delta(NewData, m.Name, m.Index, b.Index, m.Properties)
		} else {
			logger.Error(c.Prefix, "required base [%s] for scalar [%s] missing", m.BaseCounter, m.Name)
		}
	}

	c.Metadata.SetValueSS("calc_time", "data", float64(time.Since(calc_start).Microseconds()))
	// store cache for next poll
	c.Data = CachedData
	//c.Data.IsEmpty = false // @redundant

	return NewData, nil
}

func print_vector(tag string, x []float64) {
	vector := []string{}
	for _, n := range x {
		vector = append(vector, strconv.FormatFloat(float64(n), 'f', 5, 64))
	}
	logger.Debug("--------", "%-35s => %v", tag, vector)
}

func (c *ZapiPerf) calculate_from_delta(NewData *matrix.Matrix, metricName string, metricIndex, baseIndex int, properties string) {

	PrevData := c.Data // for convenience

	print_vector("current", NewData.Data[metricIndex])
	print_vector("previous", PrevData.Data[metricIndex])

	// calculate metric delta for all instances from previous cache
	NewData.Delta(PrevData, metricIndex)

	print_vector("delta", NewData.Data[metricIndex])

	// fmt.Println()

	if strings.Contains(properties, "delta") {
		print_vector(fmt.Sprintf("%s delta%s", util.Green, util.End), NewData.Data[metricIndex])
		//B.Data[metric.Index] = delta
		return
	}

	if strings.Contains(properties, "rate") {
		if ts := NewData.GetMetric("timestamp"); ts != nil {
			NewData.Divide(metricIndex, ts.Index, float64(0))
			print_vector(fmt.Sprintf("%s rate%s", util.Green, util.End), NewData.Data[metricIndex])
		} else {
			logger.Error(c.Prefix, "timestamp counter not found")
		}
		return
	}

	// For the next two properties we need base counters
	// We assume that delta of base counters is already calculated

	if baseIndex < 0 {
		logger.Error(c.Prefix, "no base counter index for") // should never happen
		return
	}

	// @TODO minimum ops threshold
	if strings.Contains(properties, "average") {

		if strings.Contains(metricName, "latency") {
			NewData.Divide(metricIndex, baseIndex, float64(c.latency_io_reqd))
		} else {
			NewData.Divide(metricIndex, baseIndex, float64(0))
		}
		print_vector(fmt.Sprintf("%s average%s", util.Green, util.End), NewData.Data[metricIndex])
		return
	}

	if strings.Contains(properties, "percent") {
		NewData.Divide(metricIndex, baseIndex, float64(0))
		NewData.MultByScalar(metricIndex, float64(100))
		print_vector(fmt.Sprintf("%s percent%s", util.Green, util.End), NewData.Data[metricIndex])
		return
	}

	logger.Error(c.Prefix, "unexpected properties: %s", properties)
	return
}

func (c *ZapiPerf) PollCounter() (*matrix.Matrix, error) {

	var (
		err                                        error
		request, response, counter_list            *node.Node
		old_metrics, old_labels, replaced, missing *set.Set
		wanted                                     *dict.Dict
		old_metrics_size, old_labels_size          int
		counters                                   map[string]*node.Node
	)

	counters = make(map[string]*node.Node)
	old_metrics = set.New() // current set of metrics, so we can remove from matrix if not updated
	old_labels = set.New()  // current set of labels
	wanted = dict.New()     // counters listed in template, maps raw name to display name
	missing = set.New()     // required base counters, missing in template
	replaced = set.New()    // deprecated and replaced counters

	for key := range c.Data.GetMetrics() {
		old_metrics.Add(key)
	}
	old_metrics_size = old_metrics.Size()

	for key := range c.Data.GetLabels() {
		old_labels.Add(key)
	}
	old_labels_size = old_labels.Size()

	// parse list of counters defined in template
	if counter_list = c.Params.GetChildS("counters"); counter_list != nil {
		for _, cnt := range counter_list.GetAllChildContentS() {
			if renamed := strings.Split(cnt, "=>"); len(renamed) == 2 {
				wanted.Set(strings.TrimSpace(renamed[0]), strings.TrimSpace(renamed[1]))
			} else {
				display := strings.ReplaceAll(cnt, "-", "_")
				if strings.HasPrefix(display, c.object) {
					display = strings.TrimPrefix(display, c.object)
					display = strings.TrimPrefix(display, "_")
				}
				wanted.Set(cnt, display)
			}
		}
	} else {
		return nil, errors.New(errors.MISSING_PARAM, "counters")
	}

	logger.Debug(c.Prefix, "updating metric cache (old cache has %d metrics and %d labels", old_metrics.Size(), old_labels.Size())

	// build request
	request = node.NewXmlS("perf-object-counter-list-info")
	request.NewChildS("objectname", c.Query)

	if err = c.Connection.BuildRequest(request); err != nil {
		return nil, err
	}

	if response, err = c.Connection.Invoke(); err != nil {
		return nil, err
	}

	// fetch counter elements
	if elems := response.GetChildS("counters"); elems != nil && len(elems.GetChildren()) != 0 {
		for _, counter := range elems.GetChildren() {
			if name := counter.GetChildContentS("name"); name != "" {
				counters[name] = counter
			}
		}
	} else {
		return nil, errors.New(errors.ERR_NO_METRIC, "no counters in response")
	}

	for key, counter := range counters {
		display, ok := wanted.GetHas(key)
		// counter not requested
		if !ok {
			logger.Trace(c.Prefix, "%sskip [%s], not requested%s", util.Grey, key, util.End)
			continue
		}

		// deprecated and possibly replaced counter
		if counter.GetChildContentS("is-deprecated") == "true" {
			if r := counter.GetChildContentS("replaced-by"); r != "" {
				logger.Info(c.Prefix, "replaced deprecated counter [%s] with [%s]", key, r)
				if !wanted.Has(r) {
					replaced.Add(r)
				}
			} else {
				logger.Info(c.Prefix, "skip [%s], deprecated", key)
				continue
			}
		}
		// override counter properties from template
		if p := c.GetOverride(key); p != "" {
			counter.SetChildContentS("properties", p)
		}

		// string metric, add as instance label
		if strings.Contains(counter.GetChildContentS("properties"), "string") {
			old_labels.Delete(key)
			if display == "instance_name" {
				display = c.object
			}
			c.Data.AddLabel(key, display)
			logger.Debug(c.Prefix, "%s+[%s] added as label name (%s)%s", util.Yellow, key, display, util.End)
		} else {
			// add counter as numeric metric
			old_metrics.Delete(key)
			if r := c.add_counter(counter, key, display, true, counters); r != "" && !wanted.Has(r) {
				missing.Add(r) // required base counter, missing in template
				logger.Debug(c.Prefix, "%smarking [%s] as required base counter for [%s]%s", util.Red, r, key, util.End)
			}
		}
	}

	// second loop for replaced counters
	if replaced.Size() > 0 {
		logger.Debug(c.Prefix, "attempting to retrieve metadata of %d replaced counters", replaced.Size())
		for name, counter := range counters {
			if replaced.Has(name) {
				old_metrics.Delete(name)
				logger.Debug(c.Prefix, "adding [%s] (replacment for deprecated counter)", name)
				if r := c.add_counter(counter, name, name, true, counters); r != "" && !wanted.Has(r) {
					missing.Add(r) // required base counter, missing in template
					logger.Debug(c.Prefix, "%smarking [%s] as required base counter for [%s]%s", util.Red, r, name, util.End)
				}
			}
		}
	}

	// third loop for required base counters, not in template
	if missing.Size() > 0 {
		logger.Debug(c.Prefix, "attempting to retrieve metadata of %d missing base counters", missing.Size())
		for name, counter := range counters {
			//logger.Debug(c.Prefix, "%shas??? [%s]%s", util.Grey, name, util.End)
			if missing.Has(name) {
				old_metrics.Delete(name)
				logger.Debug(c.Prefix, "adding [%s] (missing base counter)", name)
				c.add_counter(counter, name, "", false, counters)
			}
		}
	}

	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	if !old_metrics.Has("timestamp") {
		_, err := c.Data.AddMetricExtended("timestamp", "timestamp", "", "raw", false)
		if err != nil {
			logger.Error(c.Prefix, "add timestamp metric: %v", err)
		}
	}

	for key := range old_metrics.Iter() {
		if !(key == "timestamp") {
			c.Data.RemoveMetric(key)
			logger.Debug(c.Prefix, "removed metric [%s]", key)
		}
	}

	for key := range old_labels.Iter() {
		c.Data.RemoveLabel(key)
		logger.Debug(c.Prefix, "removed label [%s]", key)
	}

	metrics_added := c.Data.SizeMetrics() - (old_metrics_size - old_metrics.Size())
	labels_added := c.Data.SizeLabels() - (old_labels_size - old_labels.Size())

	logger.Debug(c.Prefix, "added %d new, removed %d metrics (total: %d)", metrics_added, old_metrics.Size(), c.Data.SizeMetrics())
	logger.Debug(c.Prefix, "added %d new, removed %d labels (total: %d)", labels_added, old_labels.Size(), c.Data.SizeLabels())

	if c.Data.SizeMetrics() == 0 {
		return nil, errors.New(errors.ERR_NO_METRIC, "")
	}

	return nil, nil
}

func (c *ZapiPerf) add_counter(counter *node.Node, name, display string, enabled bool, cache map[string]*node.Node) string {

	var properties, base_counter, unit string
	var err error

	properties = counter.GetChildContentS("properties")
	base_counter = counter.GetChildContentS("base-counter")
	unit = counter.GetChildContentS("unit")

	if display == "" {
		display = strings.ReplaceAll(name, "-", "_") // redundant for zapiperf
	}

	logger.Debug(c.Prefix, "handling counter [%s] with properties [%s] and unit [%s]", name, properties, unit)

	// counter type is array, each element will be converted to a metric instance
	if counter.GetChildContentS("type") == "array" {

		var labels, base_labels []string
		var e string

		if labels, e = parse_array_labels(counter); e != "" {
			logger.Warn(c.Prefix, "skipping [%s] of type array: %s", name, e)
			return ""
		}

		if base_counter != "" {
			if base, ok := cache[base_counter]; ok {
				if base.GetChildContentS("type") == "array" {
					base_labels, e = parse_array_labels(base)
					if e != "" {
						logger.Warn(c.Prefix, "skipping [%s], base counter [%s] is array, but %s", name, base_counter, e)
						return ""
					} else if len(base_labels) != len(labels) {
						logger.Warn(c.Prefix, "skipping [%s], array labels does not match with base counter (%s)", name, base_counter)
						return ""
					}
				}
			} else {
				logger.Warn(c.Prefix, "skipping [%s], base counter [%s] not found", name, base_counter)
				return ""
			}
		}

		for i, label := range labels {

			var m *matrix.Metric

			key := name + "." + label

			if m = c.Data.GetMetric(key); m != nil {
				logger.Debug("updating [%s] array metric element", key)
				m.Name = display
				m.Properties = properties
				m.Enabled = enabled
				// no base counter or base counter is scalar
			} else if base_counter == "" || len(base_labels) == 0 {
				m, err = c.Data.AddMetricExtended(key, display, base_counter, properties, enabled)
			} else {
				m, err = c.Data.AddMetricExtended(key, display, base_counter+"."+base_labels[i], properties, enabled)
			}

			if err != nil {
				logger.Error(c.Prefix, "add array metric element [%s]: %v", key, err)
			} else {
				if x := strings.Split(label, "."); len(x) == 2 {
					m.Labels.Set("metric", x[0])
					m.Labels.Set("submetric", x[1])
				} else {
					m.Labels.Set("metric", label)
				}
				logger.Debug(c.Prefix, "%s+[%s] added array metric (%s), element with label (%s)", util.Pink, name, display, label, util.End)
			}
		}
		// cache labels only when parsing counter was success
		c.array_labels[name] = labels

		// counter type is scalar
	} else {
		var m *matrix.Metric
		if m = c.Data.GetMetric(name); m != nil {
			logger.Debug(c.Prefix, "%s+[%s] updated scalar metric (%s)%s", util.Cyan, name, display, util.End)
			m.Name = display
			m.BaseCounter = base_counter
			m.Properties = properties
			m.Enabled = enabled
		} else if _, err = c.Data.AddMetricExtended(name, display, base_counter, properties, enabled); err != nil {
			logger.Error(c.Prefix, "add scalar metric [%s]: %v", name, err)
		} else {
			logger.Debug(c.Prefix, "%s+[%s] added as scalar metric (%s)%s", util.Cyan, name, display, util.End)
		}
	}
	return base_counter
}

func (c *ZapiPerf) GetOverride(counter string) string {
	if o := c.Params.GetChildS("override"); o != nil {
		return o.GetChildContentS(counter)
	}
	return ""
}

func parse_array_labels(elem *node.Node) ([]string, string) {
	var labels []string
	var msg string

	if x := elem.GetChildS("labels"); x == nil {
		msg = "array labels missing"
	} else if d := len(x.GetChildren()); d == 1 {
		labels = strings.Split(node.DecodeHtml(x.GetChildren()[0].GetContentS()), ",")
	} else if d == 2 {
		labelsA := strings.Split(node.DecodeHtml(x.GetChildren()[0].GetContentS()), ",")
		labelsB := strings.Split(node.DecodeHtml(x.GetChildren()[1].GetContentS()), ",")
		for _, a := range labelsA {
			for _, b := range labelsB {
				labels = append(labels, a+"."+b)
			}
		}
	} else {
		msg = "unexpected dimensions"
	}

	return labels, msg
}

func (c *ZapiPerf) PollInstance() (*matrix.Matrix, error) {

	var (
		err                                error
		request                            *node.Node
		old_instances                      *set.Set
		old_size, new_size, removed, added int
		instances_attr                     string
	)

	old_instances = set.New()
	for key := range c.Data.GetInstances() {
		old_instances.Add(key)
	}
	old_size = old_instances.Size()

	logger.Debug(c.Prefix, "updating instance cache (old cache has: %d)", old_instances.Size())

	if c.System.Clustered {
		request = node.NewXmlS("perf-object-instance-list-info-iter")
		instances_attr = "attributes-list"
	} else {
		request = node.NewXmlS("perf-object-instance-list-info")
		instances_attr = "instances"
	}

	request.NewChildS("objectname", c.Query)
	if c.System.Clustered {
		request.NewChildS("max-records", strconv.Itoa(c.batch_size))
	}

	batch_tag := "initial"

	for batch_tag != "" {

		// build request
		if batch_tag != "initial" {
			request.PopChildS("tag")
			request.NewChildS("tag", batch_tag)
		}

		if err = c.Connection.BuildRequest(request); err != nil {
			logger.Error(c.Prefix, "build request: %v", err)
			break
		}

		response, err := c.Connection.Invoke()
		if err != nil {
			logger.Error(c.Prefix, "instance request: %v", err)
			break
		}

		// @TODO next-tag bug
		batch_tag = response.GetChildContentS("next-tag")

		// fetch instances
		instances := response.GetChildS(instances_attr)
		if instances == nil || len(instances.GetChildren()) == 0 {
			break
		}

		for _, i := range instances.GetChildren() {

			if key := i.GetChildContentS(c.instance_key); key == "" {
				// instance key missing
				n := i.GetChildContentS("name")
				u := i.GetChildContentS("uuid")
				logger.Debug(c.Prefix, "skip instance, missing key [%s] (name=%s, uuid=%s)", c.instance_key, n, u)
			} else if old_instances.Delete(key) {
				// instance already in cache
				continue
			} else if _, e := c.Data.AddInstance(key); e != nil {
				logger.Warn(c.Prefix, "add instance: %v", e)
			} else {
				logger.Debug(c.Prefix, "added new instance [%s]", key)
			}
		}
	}

	for key := range old_instances.Iter() {
		c.Data.RemoveInstance(key)
		logger.Debug(c.Prefix, "removed instance [%s]", key)
	}

	removed = old_instances.Size()
	new_size = c.Data.SizeInstances()
	added = new_size - (old_size - removed)

	logger.Debug(c.Prefix, "added %d new, removed %d (total instances %d)", added, removed, new_size)

	if new_size == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "")
	}

	return nil, err

}
