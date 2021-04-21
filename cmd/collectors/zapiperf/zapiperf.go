//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package main

import (
	"goharvest2/cmd/poller/collector"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/set"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"strconv"
	"strings"
	"time"

	zapi "goharvest2/cmd/collectors/zapi/collector"
)

// default parameter values
const (
	INSTANCE_KEY    = "uuid"
	BATCH_SIZE      = 500
	LATENCY_IO_REQD = 10
)

const BILLION = 1000000000

type ZapiPerf struct {
	//*collector.AbstractCollector
	*zapi.Zapi // provides: Connection, System, Object, Query, TemplateFn, TemplateType
	//Connection      *client.Client
	//System          *client.System
	object string
	//Query           string
	//TemplateFn      string
	//TemplateType    string
	batch_size      int
	latency_io_reqd int
	instance_key    string
	instance_labels map[string]string
	array_labels    map[string][]string
	status_label    string
	status_ok_value string
	cache_empty     bool
}

func New(a *collector.AbstractCollector) collector.Collector {
	//return &ZapiPerf{AbstractCollector: a}
	return &ZapiPerf{Zapi: zapi.NewZapi(a)}
}

func (me *ZapiPerf) Init() error {

	if err := me.InitVars(); err != nil {
		return err
	}
	// Invoke generic initializer
	// this will load Schedule, initialize Data and Metadata
	if err := collector.Init(me); err != nil {
		return err
	}

	if err := me.InitMatrix(); err != nil {
		return err
	}

	if err := me.InitCache(); err != nil {
		return err
	}

	logger.Debug(me.Prefix, "initialized")
	return nil
}

func (me *ZapiPerf) InitCache() error {
	me.array_labels = make(map[string][]string)
	me.instance_labels = make(map[string]string)
	me.instance_key = me.loadParamStr("instance_key", INSTANCE_KEY)
	me.batch_size = me.loadParamInt("batch_size", BATCH_SIZE)
	me.latency_io_reqd = me.loadParamInt("latency_io_reqd", LATENCY_IO_REQD)
	me.cache_empty = true
	return nil
}

func (me *ZapiPerf) loadParamStr(name, default_value string) string {

	var x string

	if x = me.Params.GetChildContentS(name); x != "" {
		logger.Debug(me.Prefix, "using %s = [%s]", name, x)
		return x
	}
	logger.Debug(me.Prefix, "using %s = [%s] (default)", name, default_value)
	return default_value
}

func (me *ZapiPerf) loadParamInt(name string, default_value int) int {

	var x string
	var n int
	var e error

	if x = me.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			logger.Debug(me.Prefix, "using %s = [%d]", name, n)
			return n
		}
		logger.Warn(me.Prefix, "invalid parameter %s = [%s] (expected integer)", name, x)
	}

	logger.Debug(me.Prefix, "using %s = [%d] (default)", name, default_value)
	return default_value
}

func (me *ZapiPerf) PollData() (*matrix.Matrix, error) {

	var err error
	logger.Debug(me.Prefix, "updating data cache")

	// clone matrix without numeric data
	NewData := me.Matrix.Clone(false, true, true)
	NewData.Reset()

	timestamp := NewData.GetMetric("timestamp")
	if timestamp == nil {
		return nil, errors.New(errors.ERR_CONFIG, "missing timestamp metric") // @TODO errconfig??
	}

	// for updating metadata
	count := uint64(0)
	batch_count := 0
	api_d := time.Duration(0 * time.Second)
	parse_d := time.Duration(0 * time.Second)

	// determine what will serve as instance key (either "uuid" or "instance")
	key_name := "instance-uuid"
	if me.instance_key == "name" {
		key_name = "instance"
	}

	// list of instance keys (instance names or uuids) for which
	// we will request counter data
	instance_keys := NewData.GetInstanceKeys()

	// build ZAPI request
	request := node.NewXmlS("perf-object-get-instances")
	request.NewChildS("objectname", me.Query)

	// load requested counters (metrics + labels)
	request_counters := request.NewChildS("counters", "")
	// load scalar metrics
	for key, m := range NewData.GetMetrics() {
		// no histograms
		if !m.HasLabels() {
			request_counters.NewChildS("counter", key)
		}
	}
	// load histograms
	for key := range me.array_labels {
		request_counters.NewChildS("counter", key)
	}
	// load instance labels
	for key := range me.instance_labels {
		request_counters.NewChildS("counter", key)
	}

	// batch indices
	start_index := 0
	end_index := 0

	for end_index < len(instance_keys) {

		// update batch indices
		end_index += me.batch_size
		if end_index > len(instance_keys) {
			end_index = len(instance_keys)
		}

		logger.Debug(me.Prefix, "starting batch poll for instances [%d:%d]", start_index, end_index)

		request.PopChildS(key_name + "s")
		request_instances := request.NewChildS(key_name+"s", "")
		for _, key := range instance_keys[start_index:end_index] {
			request_instances.NewChildS(key_name, key)
		}

		start_index = end_index

		if err = me.Connection.BuildRequest(request); err != nil {
			logger.Error(me.Prefix, "build request: %v", err)
			//break?
			return nil, err
		}

		response, rd, pd, err := me.Connection.InvokeWithTimers()
		if err != nil {
			//logger.Error(me.Prefix, "data request: %v", err)
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

		logger.Debug(me.Prefix, "fetched batch with %d instances", len(instances.GetChildren()))

		// timestamp for batch instances
		// ignore timestamp from ZAPI which is always integer
		// we want float, since our poll interval can be float
		ts := float64(time.Now().UnixNano()) / BILLION

		for _, i := range instances.GetChildren() {

			key := i.GetChildContentS(me.instance_key)
			if key == "" {
				logger.Debug(me.Prefix, "skip instance, no key [%s] (name=%s, uuid=%s)", me.instance_key, i.GetChildContentS("name"), i.GetChildContentS("uuid"))
				continue
			}

			instance := NewData.GetInstance(key)
			if instance == nil {
				logger.Debug(me.Prefix, "skip instance [%s], not found in cache", key)
				continue
			}

			counters := i.GetChildS("counters")
			if counters == nil {
				logger.Debug(me.Prefix, "skip instance [%s], no data counters", key)
				continue
			}

			logger.Debug(me.Prefix, "fetching data of instance [%s]", key)

			// add batch timestamp as custom counter
			timestamp.SetValueFloat64(instance, ts)

			for _, cnt := range counters.GetChildren() {

				name := cnt.GetChildContentS("name")
				value := cnt.GetChildContentS("value")

				// sanity check
				// @TODO - redundant
				if name == "" || value == "" {
					logger.Debug(me.Prefix, "skipping incomplete counter [%s] with value [%s]", name, value)
					continue
				}

				logger.Trace(me.Prefix, "(%s%s%s) parsing counter (%s) = %v", util.Grey, key, util.End, name, value)

				// ZAPI counter for us is either instance label (string)
				// or numeric metric (scalar or string)

				// store as instance label
				if display, has := me.instance_labels[name]; has {
					instance.SetLabel(display, value)
					logger.Trace(me.Prefix, "+ label (%s) = [%s%s%s]", display, util.Yellow, value, util.End)
					continue
				}

				// store as array counter / histogram
				if labels, has := me.array_labels[name]; has {

					values := strings.Split(string(value), ",")

					if len(labels) != len(values) {
						// warn & skip
						logger.Error(me.Prefix, "histogram (%s) labels don't match with parsed values [%s]", name, value)
						continue
					}

					for i, label := range labels {
						if metric := NewData.GetMetric(name + "." + label); metric != nil {
							if err = metric.SetValueString(instance, values[i]); err != nil {
								logger.Error(me.Prefix, "set histogram (%s.%s) value [%s]: %v", name, label, values[i], err)
							} else {
								logger.Trace(me.Prefix, "+ histogram (%s.%s) = [%s%s%s]", name, label, util.Pink, values[i], util.End)
								count += 1
							}
						} else {
							logger.Warn(me.Prefix, "histogram (%s.%s) = [%s] not in cache", name, label, value)
						}
					}
					continue
				}

				// store as scalar metric
				if metric := NewData.GetMetric(name); metric != nil {
					if err = metric.SetValueString(instance, value); err != nil {
						logger.Error(me.Prefix, "set metric (%s) value [%s]: %v", name, value, err)
					} else {
						logger.Trace(me.Prefix, "+ metric (%s) = [%s%s%s]", name, util.Cyan, value, util.End)
						count += 1
					}
					continue
				}

				logger.Warn(me.Prefix, "counter (%s) [%s] not found in cache", name, value)

			} // end loop over counters

			// @TODO what is this?
			if metric := NewData.GetMetric("status"); metric != nil && metric.GetType() == "uint8" {
				if me.status_label != "" {
					if instance.GetLabel(me.status_label) == me.status_ok_value {
						metric.SetValueUint8(instance, 0)
						logger.Trace(me.Prefix, "(%s%s%s) status (%s= %s) = [0]", util.Grey, key, util.End, me.status_label, instance.GetLabel(me.status_label))
					} else {
						metric.SetValueUint8(instance, 1)
						logger.Trace(me.Prefix, "(%s%s%s) status (%s= %s) = [0]", util.Grey, key, util.End, me.status_label, instance.GetLabel(me.status_label))
					}
				}
			}
		} // end loop over instances
	} // end batch request

	// terminate if serious errors
	// @TODO handle...

	if err != nil {
		return nil, err
	}

	// update metadata
	me.Metadata.LazySetValueInt64("api_time", "data", api_d.Microseconds())
	me.Metadata.LazySetValueInt64("parse_time", "data", parse_d.Microseconds())
	me.Metadata.LazySetValueUint64("count", "data", count)
	me.AddCollectCount(count)

	logger.Debug(me.Prefix, "collected %d data points in %d batch polls", count, batch_count)

	// skip calculating from delta if no data from previous poll
	if me.cache_empty {
		logger.Debug(me.Prefix, "skip postprocessing until next poll (previous cache empty)")
		me.Matrix = NewData
		me.cache_empty = false
		return nil, nil
	}

	calc_start := time.Now()

	logger.Debug(me.Prefix, "starting delta calculations from previous cache")
	//logger.Debug(me.Prefix, "data has dimensions (%d x %d)", len(NewData.Data), len(NewData.Data[0]))

	// cache data, to store after calculations
	CachedData := NewData.Clone(true, true, true) // @TODO implement copy data

	// order metrics, such that those requiring base counters are processed last
	ordered_metrics := make([]matrix.Metric, 0, len(NewData.GetMetrics()))
	ordered_keys := make([]string, 0, len(ordered_metrics))

	for key, metric := range NewData.GetMetrics() {
		if metric.GetComment() == "" { // does not require base counter
			ordered_metrics = append(ordered_metrics, metric)
			ordered_keys = append(ordered_keys, key)
		}
	}
	for key, metric := range NewData.GetMetrics() {
		if metric.GetComment() != "" { // requires base counter
			ordered_metrics = append(ordered_metrics, metric)
			ordered_keys = append(ordered_keys, key)
		}
	}

	// calculate timestamp delta first since many counters require it for postprocessing
	// timestamp has "raw" property, so won't be postprocessed automatically
	// fmt.Printf("\npostprocessing %s%s%s - %s%v%s\n", util.Red, timestamp.Name, util.End, util.Bold, timestamp.Properties, util.End)
	//logger.Debug(me.Prefix, "cooking [%s] (%s)", timestamp.Name, timestamp.Properties)
	//print_vector("current", NewData.Data[timestamp.Index])
	//print_vector("previous", me.Data.Data[timestamp.Index])
	if err = timestamp.Delta(me.Matrix.GetMetric("timestamp")); err != nil {
		logger.Error(me.Prefix, "(timestamp) calculate delta: %v", err)
		// @TODO terminate since other counters will be incorrect
	}

	//NewData.Delta(me.Data, timestamp.Index)
	//print_vector(util.Green+"delta"+util.End, NewData.Data[timestamp.Index])

	var base matrix.Metric

	for i, metric := range ordered_metrics {

		property := metric.GetProperty()
		key := ordered_keys[i]

		// raw counters don't require postprocessing
		if property == "raw" || property == "" {
			continue
		}

		// for all the other properties we start with delta
		if err = metric.Delta(me.Matrix.GetMetric(key)); err != nil {
			logger.Error(me.Prefix, "(%s) calculate delta: %v", key, err)
			continue
		}

		if property == "delta" {
			// already done
			continue
		}

		// rate is delta, normalized by elapsed time
		if property == "rate" {
			if err = metric.Divide(timestamp); err != nil {
				logger.Error(me.Prefix, "(%s) calculate rate: %v", key, err)
			}
			continue
		}

		// For the next two properties we need base counters
		// We assume that delta of base counters is already calculated
		// (name of base counter is stored as Comment)
		if base = NewData.GetMetric(metric.GetComment()); base == nil {
			logger.Warn(me.Prefix, "(%s) <%s> base counter (%s) missing", key, property, metric.GetComment())
			continue
		}

		// average and percentage are calculated by dividing by the value of the base counter
		// special case for latency counter: apply minimum number of iops as threshold
		if property == "average" || property == "percent" {

			if strings.HasSuffix(metric.GetName(), "_latency") {
				err = metric.DivideWithThreshold(base, me.latency_io_reqd)
			} else {
				err = metric.Divide(base)
			}

			if err != nil {
				logger.Error(me.Prefix, "(%s) division by base: %v", key, err)
			}

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if err = metric.MultiplyByScalar(100); err != nil {
				logger.Error(me.Prefix, "(%s) multiply by scalar: %v", key, err)
			}
			continue
		}

		logger.Error(me.Prefix, "(%s) unknown property: %s", key, property)
	}

	me.Metadata.LazySetValueInt64("calc_time", "data", time.Since(calc_start).Microseconds())
	// store cache for next poll
	me.Matrix = CachedData
	//me.Data.IsEmpty = false // @redundant

	return NewData, nil
}

func (me *ZapiPerf) PollCounter() (*matrix.Matrix, error) {

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

	for key := range me.Matrix.GetMetrics() {
		old_metrics.Add(key)
	}
	old_metrics_size = old_metrics.Size()

	for key := range me.instance_labels {
		old_labels.Add(key)
	}
	old_labels_size = old_labels.Size()

	// parse list of counters defined in template
	if counter_list = me.Params.GetChildS("counters"); counter_list != nil {
		for _, cnt := range counter_list.GetAllChildContentS() {
			if renamed := strings.Split(cnt, "=>"); len(renamed) == 2 {
				wanted.Set(strings.TrimSpace(renamed[0]), strings.TrimSpace(renamed[1]))
			} else {
				display := strings.ReplaceAll(cnt, "-", "_")
				if strings.HasPrefix(display, me.object) {
					display = strings.TrimPrefix(display, me.object)
					display = strings.TrimPrefix(display, "_")
				}
				wanted.Set(cnt, display)
			}
		}
	} else {
		return nil, errors.New(errors.MISSING_PARAM, "counters")
	}

	logger.Debug(me.Prefix, "updating metric cache (old cache has %d metrics and %d labels", old_metrics.Size(), old_labels.Size())

	// build request
	request = node.NewXmlS("perf-object-counter-list-info")
	request.NewChildS("objectname", me.Query)

	if err = me.Connection.BuildRequest(request); err != nil {
		return nil, err
	}

	if response, err = me.Connection.Invoke(); err != nil {
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

		// override counter properties from template
		if p := me.GetOverride(key); p != "" {
			counter.SetChildContentS("properties", p)
		}

		display, ok := wanted.GetHas(key)
		// counter not requested
		if !ok {
			logger.Trace(me.Prefix, "%sskip [%s], not requested%s", util.Grey, key, util.End)
			continue
		}

		// deprecated and possibly replaced counter
		if counter.GetChildContentS("is-deprecated") == "true" {
			if r := counter.GetChildContentS("replaced-by"); r != "" {
				logger.Info(me.Prefix, "replaced deprecated counter [%s] with [%s]", key, r)
				if !wanted.Has(r) {
					replaced.Add(r)
				}
			} else {
				logger.Info(me.Prefix, "skip [%s], deprecated", key)
				continue
			}
		}

		// string metric, add as instance label
		if strings.Contains(counter.GetChildContentS("properties"), "string") {
			old_labels.Delete(key)
			if display == "instance_name" {
				display = me.object
			}
			me.instance_labels[key] = display
			logger.Debug(me.Prefix, "%s+[%s] added as label name (%s)%s", util.Yellow, key, display, util.End)
		} else {
			// add counter as numeric metric
			old_metrics.Delete(key)
			if r := me.add_counter(counter, key, display, true, counters); r != "" && !wanted.Has(r) {
				missing.Add(r) // required base counter, missing in template
				logger.Debug(me.Prefix, "%smarking [%s] as required base counter for [%s]%s", util.Red, r, key, util.End)
			}
		}
	}

	// second loop for replaced counters
	if replaced.Size() > 0 {
		logger.Debug(me.Prefix, "attempting to retrieve metadata of %d replaced counters", replaced.Size())
		for name, counter := range counters {
			if replaced.Has(name) {
				old_metrics.Delete(name)
				logger.Debug(me.Prefix, "adding [%s] (replacment for deprecated counter)", name)
				if r := me.add_counter(counter, name, name, true, counters); r != "" && !wanted.Has(r) {
					missing.Add(r) // required base counter, missing in template
					logger.Debug(me.Prefix, "%smarking [%s] as required base counter for [%s]%s", util.Red, r, name, util.End)
				}
			}
		}
	}

	// third loop for required base counters, not in template
	if missing.Size() > 0 {
		logger.Debug(me.Prefix, "attempting to retrieve metadata of %d missing base counters", missing.Size())
		for name, counter := range counters {
			//logger.Debug(me.Prefix, "%shas??? [%s]%s", util.Grey, name, util.End)
			if missing.Has(name) {
				old_metrics.Delete(name)
				logger.Debug(me.Prefix, "adding [%s] (missing base counter)", name)
				me.add_counter(counter, name, "", false, counters)
			}
		}
	}

	// @TODO check dtype!!!!
	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	if !old_metrics.Has("timestamp") {
		m, err := me.Matrix.NewMetricFloat64("timestamp")
		if err != nil {
			logger.Error(me.Prefix, "add timestamp metric: %v", err)
		}
		m.SetProperty("raw")
		m.SetExportable(false)
	}

	// @TODO what the hell is this?
	if x := me.Params.GetChildS("instance_status"); x != nil && !old_metrics.Has("status") {
		me.status_label = x.GetChildContentS("label")
		me.status_ok_value = x.GetChildContentS("ok_value")
		if me.status_label == "" || me.status_ok_value == "" {
			return nil, errors.New(errors.MISSING_PARAM, "label or ok_value missing")
		}
		m, err := me.Matrix.NewMetricUint8("status")
		if err != nil {
			logger.Error(me.Prefix, "add status metric: %v", err)
			return nil, err
		}
		m.SetProperty("raw")
		logger.Debug(me.Prefix, "added status metric for label [%s] (ok_value: %s)", me.status_label, me.status_ok_value)
	}

	for key := range old_metrics.Iter() {
		// temporary fix: prevent removing array counters
		// @TODO
		if key != "timestamp" && !strings.Contains(key, ".") {
			me.Matrix.RemoveMetric(key)
			logger.Debug(me.Prefix, "removed metric [%s]", key)
		}
	}

	for key := range old_labels.Iter() {
		//me.Data.RemoveLabel(key)
		delete(me.instance_labels, key)
		logger.Debug(me.Prefix, "removed label [%s]", key)
	}

	metrics_added := len(me.Matrix.GetMetrics()) - (old_metrics_size - old_metrics.Size())
	labels_added := len(me.instance_labels) - (old_labels_size - old_labels.Size())

	logger.Debug(me.Prefix, "added %d new, removed %d metrics (total: %d)", metrics_added, old_metrics.Size(), len(me.Matrix.GetMetrics()))
	logger.Debug(me.Prefix, "added %d new, removed %d labels (total: %d)", labels_added, old_labels.Size(), len(me.instance_labels))

	if len(me.Matrix.GetMetrics()) == 0 {
		return nil, errors.New(errors.ERR_NO_METRIC, "")
	}

	return nil, nil
}

func (me *ZapiPerf) add_counter(counter *node.Node, name, display string, enabled bool, cache map[string]*node.Node) string {

	var property, base_counter, unit string
	var err error

	p := counter.GetChildContentS("properties")
	if strings.Contains(p, "raw") {
		property = "raw"
	} else if strings.Contains(p, "delta") {
		property = "delta"
	} else if strings.Contains(p, "rate") {
		property = "rate"
	} else if strings.Contains(p, "average") {
		property = "average"
	} else if strings.Contains(p, "percent") {
		property = "percent"
	} else {
		logger.Warn(me.Prefix, "skip counter [%s] with unknown property [%s]", name, p)
		return ""
	}

	base_counter = counter.GetChildContentS("base-counter")
	unit = counter.GetChildContentS("unit")

	if display == "" {
		display = strings.ReplaceAll(name, "-", "_") // redundant for zapiperf
	}

	logger.Debug(me.Prefix, "handling counter [%s] with property [%s] and unit [%s]", name, property, unit)

	// counter type is array, each element will be converted to a metric instance
	if counter.GetChildContentS("type") == "array" {

		var labels, base_labels []string
		var e string

		if labels, e = parse_array_labels(counter); e != "" {
			logger.Warn(me.Prefix, "skipping [%s] of type array: %s", name, e)
			return ""
		}

		if base_counter != "" {
			if base, ok := cache[base_counter]; ok {
				if base.GetChildContentS("type") == "array" {
					base_labels, e = parse_array_labels(base)
					if e != "" {
						logger.Warn(me.Prefix, "skipping [%s], base counter [%s] is array, but %s", name, base_counter, e)
						return ""
					} else if len(base_labels) != len(labels) {
						logger.Warn(me.Prefix, "skipping [%s], array labels don't match with base counter labels [%s]", name, base_counter)
						return ""
					}
				}
			} else {
				logger.Warn(me.Prefix, "skipping [%s], base counter [%s] not found", name, base_counter)
				return ""
			}
		}

		for _, label := range labels {

			var m matrix.Metric

			key := name + "." + label
			base_key := base_counter
			if base_counter != "" && len(base_labels) != 0 {
				base_key += "." + base_labels[0]
			}

			if m = me.Matrix.GetMetric(key); m != nil {
				logger.Debug(me.Prefix, "updating array metric [%s] attributes", key)
			} else if m, err = me.Matrix.NewMetricFloat64(key); err == nil {
				logger.Debug(me.Prefix, "%s+[%s] added array metric (%s), element with label (%s)%s", util.Pink, name, display, label, util.End)
			} else {
				logger.Error(me.Prefix, "add array metric element [%s]: %v", key, err)
				return ""
			}

			m.SetName(display)
			m.SetProperty(property)
			m.SetComment(base_key)
			m.SetExportable(enabled)

			if x := strings.Split(label, "."); len(x) == 2 {
				m.SetLabel("metric", x[0])
				m.SetLabel("submetric", x[1])
			} else {
				m.SetLabel("metric", label)
			}
		}
		// cache labels only when parsing counter was success
		me.array_labels[name] = labels

		// counter type is scalar
	} else {
		var m matrix.Metric
		if m = me.Matrix.GetMetric(name); m != nil {
			logger.Debug(me.Prefix, "updating scalar metric [%s] attributes", name)
		} else if m, err = me.Matrix.NewMetricFloat64(name); err == nil {
			logger.Debug(me.Prefix, "%s+[%s] added scalar metric (%s)%s", util.Cyan, name, display, util.End)
		} else {
			logger.Error(me.Prefix, "add scalar metric [%s]: %v", name, err)
			return ""
		}

		m.SetName(display)
		m.SetProperty(property)
		m.SetComment(base_counter)
		m.SetExportable(enabled)

	}
	return base_counter
}

func (me *ZapiPerf) GetOverride(counter string) string {
	if o := me.Params.GetChildS("override"); o != nil {
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

func (me *ZapiPerf) PollInstance() (*matrix.Matrix, error) {

	var (
		err                                error
		request, results                   *node.Node
		old_instances                      *set.Set
		old_size, new_size, removed, added int
		instances_attr                     string
	)

	old_instances = set.New()
	for key := range me.Matrix.GetInstances() {
		old_instances.Add(key)
	}
	old_size = old_instances.Size()

	logger.Debug(me.Prefix, "updating instance cache (old cache has: %d)", old_instances.Size())

	if me.Connection.IsClustered() {
		request = node.NewXmlS("perf-object-instance-list-info-iter")
		instances_attr = "attributes-list"
	} else {
		request = node.NewXmlS("perf-object-instance-list-info")
		instances_attr = "instances"
	}

	request.NewChildS("objectname", me.Query)
	if me.System.Clustered {
		request.NewChildS("max-records", strconv.Itoa(me.batch_size))
	}

	batch_tag := "initial"

	for {

		results, batch_tag, err = me.Connection.InvokeBatchRequest(request, batch_tag)

		if err != nil {
			logger.Error(me.Prefix, "instance request: %v", err)
			break
		}

		if results == nil {
			break
		}

		// fetch instances
		instances := results.GetChildS(instances_attr)
		if instances == nil || len(instances.GetChildren()) == 0 {
			break
		}

		for _, i := range instances.GetChildren() {

			if key := i.GetChildContentS(me.instance_key); key == "" {
				// instance key missing
				n := i.GetChildContentS("name")
				u := i.GetChildContentS("uuid")
				logger.Debug(me.Prefix, "skip instance, missing key [%s] (name=%s, uuid=%s)", me.instance_key, n, u)
			} else if old_instances.Delete(key) {
				// instance already in cache
				continue
			} else if _, err = me.Matrix.NewInstance(key); err != nil {
				logger.Warn(me.Prefix, "add instance: %v", err)
			} else {
				logger.Debug(me.Prefix, "added new instance [%s]", key)
			}
		}
	}

	for key := range old_instances.Iter() {
		me.Matrix.RemoveInstance(key)
		logger.Debug(me.Prefix, "removed instance [%s]", key)
	}

	removed = old_instances.Size()
	new_size = len(me.Matrix.GetInstances())
	added = new_size - (old_size - removed)

	logger.Debug(me.Prefix, "added %d new, removed %d (total instances %d)", added, removed, new_size)

	if new_size == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "")
	}

	return nil, err
}
