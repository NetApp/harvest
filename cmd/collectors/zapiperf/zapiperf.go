/*
	Copyright NetApp Inc, 2021 All rights reserved

	ZapiPerf collects and processes metrics from the "perf" APIs of the
	ZAPI protocol. This collector inherits some methods and fields of
	the Zapi collector (as they use the same protocol). However,
	ZapiPerf calculates final metric values from the deltas of two
	consecutive polls.

	The exact formula of doing these calculations, depends on the property
	of each counter and some counters require a "base-counter" additionally.

	Counter properties and other metadata are fetched from the target system
	during PollCounter() making the collector ONTAP-version independent.

	The collector maintains a cache of instances, updated periodically as well,
	during PollInstance().

	The source code prioritizes performance over simplicity/readability.
	Additionally, some objects (e.g. workloads) come with twists that
	force the collector to do acrobatics. Don't expect to easily
	comprehend what comes below.
*/
package zapiperf

import (
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/fcp"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/headroom"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/nic"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/vscan"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errors"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strconv"
	"strings"
	"time"

	zapi "github.com/netapp/harvest/v2/cmd/collectors/zapi/collector"
)

const (
	// default parameter values
	instanceKey   = "uuid"
	batchSize     = 500
	latencyIoReqd = 10
	// objects that need special handling
	objWorkload             = "workload"
	objWorkloadDetail       = "workload_detail"
	objWorkloadVolume       = "workload_volume"
	objWorkloadDetailVolume = "workload_detail_volume"
)

const BILLION = 1000000000

type ZapiPerf struct {
	*zapi.Zapi      // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	object          string
	batchSize       int
	latencyIoReqd   int
	instanceKey     string
	instanceLabels  map[string]string
	histogramLabels map[string][]string
	scalarCounters  []string
	qosLabels       map[string]string
	isCacheEmpty    bool
}

func init() {
	plugin.RegisterModule(ZapiPerf{})
}

func (ZapiPerf) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.zapiperf",
		New: func() plugin.Module { return new(ZapiPerf) },
	}
}

func (me *ZapiPerf) Init(a *collector.AbstractCollector) error {
	me.Zapi = &zapi.Zapi{AbstractCollector: a}

	if err := me.InitVars(); err != nil {
		return err
	}
	// Invoke generic initializer
	// this will load Schedule, initialize data and metadata Matrices
	if err := collector.Init(me); err != nil {
		return err
	}

	if err := me.InitMatrix(); err != nil {
		return err
	}

	if err := me.InitCache(); err != nil {
		return err
	}

	me.Logger.Debug().Msg("initialized")
	return nil
}

func (me *ZapiPerf) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Nic":
		return nic.New(abc)
	case "Fcp":
		return fcp.New(abc)
	case "Headroom":
		return headroom.New(abc)
	case "Volume":
		return volume.New(abc)
	case "Vscan":
		return vscan.New(abc)
	default:
		me.Logger.Info().Msgf("no zapiPerf plugin found for %s", kind)
	}
	return nil
}

func (me *ZapiPerf) InitCache() error {
	me.histogramLabels = make(map[string][]string)
	me.instanceLabels = make(map[string]string)
	me.instanceKey = me.loadParamStr("instance_key", instanceKey)
	me.batchSize = me.loadParamInt("batch_size", batchSize)
	me.latencyIoReqd = me.loadParamInt("latency_io_reqd", latencyIoReqd)
	me.isCacheEmpty = true
	me.object = me.loadParamStr("object", "")
	// hack to override from AbstractCollector
	// @TODO need cleaner solution
	if me.object == "" {
		me.object = me.Object
	}
	me.Matrix[me.Object].Object = me.object
	me.Logger.Debug().Msgf("object= %s --> %s", me.Object, me.object)
	return nil
}

// load a string parameter or use defaultValue
func (me *ZapiPerf) loadParamStr(name, defaultValue string) string {

	var x string

	if x = me.Params.GetChildContentS(name); x != "" {
		me.Logger.Debug().Msgf("using %s = [%s]", name, x)
		return x
	}
	me.Logger.Debug().Msgf("using %s = [%s] (default)", name, defaultValue)
	return defaultValue
}

// load an int parameter or use defaultValue
func (me *ZapiPerf) loadParamInt(name string, defaultValue int) int {

	var (
		x string
		n int
		e error
	)

	if x = me.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			me.Logger.Debug().Msgf("using %s = [%d]", name, n)
			return n
		}
		me.Logger.Warn().Msgf("invalid parameter %s = [%s] (expected integer)", name, x)
	}

	me.Logger.Debug().Msgf("using %s = [%d] (default)", name, defaultValue)
	return defaultValue
}

// PollData updates the data cache of the collector. During first poll, no data will
// be emitted. Afterwards, final metric values will be calculated from previous poll.
func (me *ZapiPerf) PollData() (map[string]*matrix.Matrix, error) {

	var (
		instanceKeys    []string
		resourceLatency matrix.Metric // for workload* objects
		err             error
	)

	me.Logger.Debug().Msg("updating data cache")
	m := me.Matrix[me.Object]
	// clone matrix without numeric data
	newData := m.Clone(false, true, true)
	newData.Reset()

	timestamp := newData.GetMetric("timestamp")
	if timestamp == nil {
		return nil, errors.New(errors.ErrConfig, "missing timestamp metric") // @TODO errconfig??
	}

	// for updating metadata
	count := uint64(0)
	batchCount := 0
	apiT := 0 * time.Second
	parseT := 0 * time.Second

	// determine what will serve as instance key (either "uuid" or "instance")
	keyName := "instance-uuid"
	if me.instanceKey == "name" {
		keyName = "instance"
	}

	// list of instance keys (instance names or uuids) for which
	// we will request counter data
	if me.Query == objWorkloadDetail || me.Query == objWorkloadDetailVolume {
		if resourceMap := me.Params.GetChildS("resource_map"); resourceMap == nil {
			return nil, errors.New(errors.MissingParam, "resource_map")
		} else {
			instanceKeys = make([]string, 0)
			for _, layer := range resourceMap.GetAllChildNamesS() {
				for key := range m.GetInstances() {
					instanceKeys = append(instanceKeys, key+"."+layer)
				}
			}
		}
	} else {
		instanceKeys = newData.GetInstanceKeys()
	}

	// build ZAPI request
	request := node.NewXMLS("perf-object-get-instances")
	request.NewChildS("objectname", me.Query)

	// load requested counters (metrics + labels)
	requestCounters := request.NewChildS("counters", "")
	// load scalar metrics
	for _, key := range me.scalarCounters {
		requestCounters.NewChildS("counter", key)
	}
	// load histograms
	for key := range me.histogramLabels {
		requestCounters.NewChildS("counter", key)
	}
	// load instance labels
	for key := range me.instanceLabels {
		requestCounters.NewChildS("counter", key)
	}

	// batch indices
	startIndex := 0
	endIndex := 0

	for endIndex < len(instanceKeys) {

		// update batch indices
		endIndex += me.batchSize
		if endIndex > len(instanceKeys) {
			endIndex = len(instanceKeys)
		}

		me.Logger.Debug().
			Int("startIndex", startIndex).
			Int("endIndex", endIndex).
			Msg("Starting batch poll for instances")

		request.PopChildS(keyName + "s")
		requestInstances := request.NewChildS(keyName+"s", "")
		for _, key := range instanceKeys[startIndex:endIndex] {
			requestInstances.NewChildS(keyName, key)
		}

		startIndex = endIndex

		if err = me.Client.BuildRequest(request); err != nil {
			me.Logger.Error().Stack().Err(err).
				Str("objectname", me.Query).
				Msg("Build request")
			return nil, err
		}

		response, rd, pd, err := me.Client.InvokeWithTimers()
		if err != nil {
			// if ONTAP complains about batch size, use a smaller batch size
			if strings.Contains(err.Error(), "resource limit exceeded") && me.batchSize > 100 {
				me.Logger.Error().Stack().Err(err)
				me.Logger.Info().
					Int("oldBatchSize", me.batchSize).
					Int("newBatchSize", me.batchSize-100).
					Msg("Changed batch_size")
				me.batchSize -= 100
				return nil, nil
			}
			return nil, err
		}

		apiT += rd
		parseT += pd
		batchCount++

		// fetch instances
		instances := response.GetChildS("instances")
		if instances == nil || len(instances.GetChildren()) == 0 {
			err = errors.New(errors.ErrNoInstance, "")
			break
		}

		me.Logger.Debug().
			Int("instances", len(instances.GetChildren())).
			Msg("Fetched batch with instances")

		// timestamp for batch instances
		// ignore timestamp from ZAPI which is always integer
		// we want float, since our poll interval can be float
		ts := float64(time.Now().UnixNano()) / BILLION

		for _, i := range instances.GetChildren() {

			key := i.GetChildContentS(me.instanceKey)

			// special case for these two objects
			// we need to process each latency layer for each instance/counter
			if me.Query == objWorkloadDetail || me.Query == objWorkloadDetailVolume {

				layer := "" // latency layer (resource) for workloads

				if x := strings.Split(key, "."); len(x) == 2 {
					key = x[0]
					layer = x[1]
				} else {
					me.Logger.Warn().
						Str("key", key).
						Msg("Instance key has unexpected format")
					continue
				}

				if resourceLatency = newData.GetMetric(layer); resourceLatency == nil {
					me.Logger.Warn().
						Str("layer", layer).
						Msg("Resource-latency metric missing in cache")
					continue
				}
			}

			if key == "" {
				me.Logger.Debug().
					Str("instanceKey", me.instanceKey).
					Str("name", i.GetChildContentS("name")).
					Str("uuid", i.GetChildContentS("uuid")).
					Msg("Skip instance, key is empty")
				continue
			}

			instance := newData.GetInstance(key)
			if instance == nil {
				me.Logger.Debug().
					Str("key", key).
					Msg("Skip instance key, not found in cache")
				continue
			}

			counters := i.GetChildS("counters")
			if counters == nil {
				me.Logger.Debug().
					Str("key", key).
					Msg("Skip instance key, no data counters")
				continue
			}

			me.Logger.Debug().
				Str("key", key).
				Msg("Fetching data of instance")

			// add batch timestamp as custom counter
			if err := timestamp.SetValueFloat64(instance, ts); err != nil {
				me.Logger.Error().Stack().Err(err).Msg("set timestamp value: ")
			}

			for _, cnt := range counters.GetChildren() {

				name := cnt.GetChildContentS("name")
				value := cnt.GetChildContentS("value")

				// sanity check
				if name == "" || value == "" {
					me.Logger.Debug().
						Str("counter", name).
						Str("value", value).
						Msg("Skipping incomplete counter")
					continue
				}

				me.Logger.Trace().
					Str("key", key).
					Str("counter", name).
					Str("value", value).
					Msg("Parsing counter")

				// ZAPI counter for us is either instance label (string)
				// or numeric metric (scalar or histogram)

				// store as instance label
				if display, has := me.instanceLabels[name]; has {
					instance.SetLabel(display, value)
					me.Logger.Trace().
						Str("display", display).
						Str("value", value).
						Msg("SetLabel")
					continue
				}

				// store as array counter / histogram
				if labels, has := me.histogramLabels[name]; has {

					values := strings.Split(value, ",")

					if len(labels) != len(values) {
						// warn & skip
						me.Logger.Error().
							Stack().
							Str("labels", name).
							Str("value", value).
							Msg("Histogram labels don't match parsed values")
						continue
					}

					for i, label := range labels {
						if metric := newData.GetMetric(name + "." + label); metric != nil {
							if err = metric.SetValueString(instance, values[i]); err != nil {
								me.Logger.Error().
									Stack().
									Err(err).
									Str("name", name).
									Str("label", label).
									Str("value", values[i]).
									Msg("Set histogram value failed")
							} else {
								me.Logger.Trace().
									Str("name", name).
									Str("label", label).
									Str("value", values[i]).
									Msg("Set histogram name.label = value")
								count++
							}
						} else {
							me.Logger.Warn().
								Str("name", name).
								Str("label", label).
								Str("value", value).
								Msg("Histogram name.label not in cache")
						}
					}
					continue
				}

				// special case for workload_detail
				if me.Query == objWorkloadDetail || me.Query == objWorkloadDetailVolume {
					if name == "wait_time" || name == "service_time" {
						if err := resourceLatency.AddValueString(instance, value); err != nil {
							me.Logger.Error().
								Stack().
								Err(err).
								Str("name", name).
								Str("value", value).
								Msg("Add resource-latency failed")
						} else {
							me.Logger.Trace().
								Str("name", name).
								Str("value", value).
								Msg("Add resource-latency")
							count++
						}
						continue
					}
					// "visits" are ignored
					if name == "visits" {
						continue
					}
				}

				// store as scalar metric
				if metric := newData.GetMetric(name); metric != nil {
					if err = metric.SetValueString(instance, value); err != nil {
						me.Logger.Error().
							Stack().
							Err(err).
							Str("name", name).
							Str("value", value).
							Msg("Set metric failed")
					} else {
						me.Logger.Trace().
							Str("name", name).
							Str("value", value).
							Msg("Set metric")
						count++
					}
					continue
				}

				me.Logger.Warn().Str("counter", name).Str("value", value).Msg("Counter not found in cache")

			} // end loop over counters

		} // end loop over instances
	} // end batch request

	me.Logger.Debug().
		Uint64("count", count).
		Int("batchCount", batchCount).
		Msg("Collected data points in batch polls")

	if me.Query == objWorkloadDetail || me.Query == objWorkloadDetailVolume {
		if rd, pd, err := me.getParentOpsCounters(newData, keyName); err == nil {
			apiT += rd
			parseT += pd
		} else {
			// no point to continue as we can't calculate the other counters
			return nil, err
		}
	}

	// update metadata
	_ = me.Metadata.LazySetValueInt64("api_time", "data", apiT.Microseconds())
	_ = me.Metadata.LazySetValueInt64("parse_time", "data", parseT.Microseconds())
	_ = me.Metadata.LazySetValueUint64("count", "data", count)
	me.AddCollectCount(count)

	// skip calculating from delta if no data from previous poll
	if me.isCacheEmpty {
		me.Logger.Debug().Msg("skip postprocessing until next poll (previous cache empty)")
		me.Matrix[me.Object] = newData
		me.isCacheEmpty = false
		return nil, nil
	}

	calcStart := time.Now()

	me.Logger.Debug().Msg("starting delta calculations from previous cache")

	// cache raw data for next poll
	cachedData := newData.Clone(true, true, true) // @TODO implement copy data

	// order metrics, such that those requiring base counters are processed last
	orderedMetrics := make([]matrix.Metric, 0, len(newData.GetMetrics()))
	orderedKeys := make([]string, 0, len(orderedMetrics))

	for key, metric := range newData.GetMetrics() {
		if metric.GetComment() == "" { // does not require base counter
			orderedMetrics = append(orderedMetrics, metric)
			orderedKeys = append(orderedKeys, key)
		}
	}
	for key, metric := range newData.GetMetrics() {
		if metric.GetComment() != "" { // requires base counter
			orderedMetrics = append(orderedMetrics, metric)
			orderedKeys = append(orderedKeys, key)
		}
	}

	// calculate timestamp delta first since many counters require it for postprocessing.
	// Timestamp has "raw" property, so it isn't post-processed automatically
	if err = timestamp.Delta(m.GetMetric("timestamp")); err != nil {
		me.Logger.Error().Stack().Err(err).Msg("(timestamp) calculate delta:")
		// @TODO terminate since other counters will be incorrect
	}

	var base matrix.Metric

	for i, metric := range orderedMetrics {

		property := metric.GetProperty()
		key := orderedKeys[i]

		// RAW - submit without post-processing
		if property == "raw" {
			continue
		}

		// all other properties - first calculate delta
		if err = metric.Delta(m.GetMetric(key)); err != nil {
			me.Logger.Error().Stack().Err(err).Str("key", key).Msg("Calculate delta")
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
		if base = newData.GetMetric(metric.GetComment()); base == nil {
			me.Logger.Warn().
				Str("key", key).
				Str("property", property).
				Str("comment", metric.GetComment()).
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
				err = metric.DivideWithThreshold(base, me.latencyIoReqd)
			} else {
				err = metric.Divide(base)
			}

			if err != nil {
				me.Logger.Error().Stack().Err(err).Str("key", key).Msg("Division by base")
			}

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if err = metric.MultiplyByScalar(100); err != nil {
				me.Logger.Error().Stack().Err(err).Str("key", key).Msg("Multiply by scalar")
			}
			continue
		}
		me.Logger.Error().Stack().Err(err).
			Str("key", key).
			Str("property", property).
			Msg("Unknown property")
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		if metric.GetProperty() == "rate" {
			if err = metric.Divide(timestamp); err != nil {
				me.Logger.Error().Stack().Err(err).
					Int("i", i).
					Str("key", orderedKeys[i]).
					Msg("Calculate rate")
			}
		}
	}

	_ = me.Metadata.LazySetValueInt64("calc_time", "data", time.Since(calcStart).Microseconds())

	// store cache for next poll
	me.Matrix[me.Object] = cachedData

	newDataMap := make(map[string]*matrix.Matrix)
	newDataMap[me.Object] = newData
	return newDataMap, nil
}

// Poll counter "ops" of the related/parent object, required for objects
// workload_detail and workload_detail_volume. This counter is already
// collected by the other ZapiPerf collectors, so this poll is redundant
// (until we implement some sort of inter-collector communication).
func (me *ZapiPerf) getParentOpsCounters(data *matrix.Matrix, KeyAttr string) (time.Duration, time.Duration, error) {

	var (
		ops          matrix.Metric
		object       string
		instanceKeys []string
		apiT, parseT time.Duration
		err          error
	)

	if me.Query == objWorkloadDetail {
		object = objWorkload
	} else {
		object = objWorkloadVolume
	}

	me.Logger.Debug().Msgf("(%s) starting redundancy poll for ops from parent object (%s)", me.Query, object)

	apiT = 0 * time.Second
	parseT = 0 * time.Second

	if ops = data.GetMetric("ops"); ops == nil {
		me.Logger.Error().Stack().Err(nil).Msgf("ops counter not found in cache")
		return apiT, parseT, errors.New(errors.MissingParam, "counter ops")
	}

	instanceKeys = data.GetInstanceKeys()

	// build ZAPI request
	request := node.NewXMLS("perf-object-get-instances")
	request.NewChildS("objectname", object)

	requestCounters := request.NewChildS("counters", "")
	requestCounters.NewChildS("counter", "ops")

	// batch indices
	startIndex := 0
	endIndex := 0

	count := 0

	for endIndex < len(instanceKeys) {

		// update batch indices
		endIndex += me.batchSize
		if endIndex > len(instanceKeys) {
			endIndex = len(instanceKeys)
		}

		me.Logger.Debug().Msgf("starting batch poll for instances [%d:%d]", startIndex, endIndex)

		request.PopChildS(KeyAttr + "s")
		requestInstances := request.NewChildS(KeyAttr+"s", "")
		for _, key := range instanceKeys[startIndex:endIndex] {
			requestInstances.NewChildS(KeyAttr, key)
		}

		startIndex = endIndex

		if err = me.Client.BuildRequest(request); err != nil {
			return apiT, parseT, err
		}

		response, rt, pt, err := me.Client.InvokeWithTimers()
		if err != nil {
			return apiT, parseT, err
		}

		apiT += rt
		parseT += pt

		// fetch instances
		instances := response.GetChildS("instances")
		if instances == nil || len(instances.GetChildren()) == 0 {
			return apiT, parseT, err
		}

		for _, i := range instances.GetChildren() {

			key := i.GetChildContentS(me.instanceKey)

			if key == "" {
				me.Logger.Debug().Msgf("skip instance, no key [%s] (name=%s, uuid=%s)", me.instanceKey, i.GetChildContentS("name"), i.GetChildContentS("uuid"))
				continue
			}

			instance := data.GetInstance(key)
			if instance == nil {
				me.Logger.Warn().Msgf("skip instance [%s], not found in cache", key)
				continue
			}

			counters := i.GetChildS("counters")
			if counters == nil {
				me.Logger.Debug().Msgf("skip instance [%s], no data counters", key)
				continue
			}

			for _, cnt := range counters.GetChildren() {

				name := cnt.GetChildContentS("name")
				value := cnt.GetChildContentS("value")

				me.Logger.Trace().Msgf("(%s%s%s%s) parsing counter = %v", color.Grey, key, color.End, name, value)

				if name == "ops" {
					if err = ops.SetValueString(instance, value); err != nil {
						me.Logger.Error().Stack().Err(err).Msgf("set metric (%s) value [%s]", name, value)
					} else {
						me.Logger.Trace().Msgf("+ metric (%s) = [%s%s%s]", name, color.Cyan, value, color.End)
						count++
					}
				} else {
					me.Logger.Error().Stack().Err(nil).Msgf("unrequested metric (%s)", name)
				}
			}
		}
	}
	me.Logger.Debug().Msgf("(%s) completed redundant ops poll (%s): collected %d", me.Query, object, count)
	return apiT, parseT, nil
}

func (me *ZapiPerf) PollCounter() (map[string]*matrix.Matrix, error) {

	var (
		err                                      error
		request, response, counterList           *node.Node
		oldMetrics, oldLabels, replaced, missing *set.Set
		wanted                                   *dict.Dict
		oldMetricsSize, oldLabelsSize            int
		counters                                 map[string]*node.Node
	)

	me.scalarCounters = make([]string, 0)
	counters = make(map[string]*node.Node)
	oldMetrics = set.New() // current set of metrics, so we can remove from matrix if not updated
	oldLabels = set.New()  // current set of labels
	wanted = dict.New()    // counters listed in template, maps raw name to display name
	missing = set.New()    // required base counters, missing in template
	replaced = set.New()   // deprecated and replaced counters

	mat := me.Matrix[me.Object]
	for key := range mat.GetMetrics() {
		oldMetrics.Add(key)
	}
	oldMetricsSize = oldMetrics.Size()

	for key := range me.instanceLabels {
		oldLabels.Add(key)
	}
	oldLabelsSize = oldLabels.Size()

	// parse list of counters defined in template
	if counterList = me.Params.GetChildS("counters"); counterList != nil {
		for _, cnt := range counterList.GetAllChildContentS() {
			if renamed := strings.Split(cnt, "=>"); len(renamed) == 2 {
				wanted.Set(strings.TrimSpace(renamed[0]), strings.TrimSpace(renamed[1]))
			} else if cnt == "instance_name" {
				wanted.Set(cnt, me.object)
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
		return nil, errors.New(errors.MissingParam, "counters")
	}

	me.Logger.Debug().
		Int("oldMetrics", oldMetricsSize).
		Int("oldLabels", oldLabelsSize).
		Msg("Updating metric cache")

	// build request
	request = node.NewXMLS("perf-object-counter-list-info")
	request.NewChildS("objectname", me.Query)

	if err = me.Client.BuildRequest(request); err != nil {
		return nil, err
	}

	if response, err = me.Client.Invoke(); err != nil {
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
		return nil, errors.New(errors.ErrNoMetric, "no counters in response")
	}

	for key, counter := range counters {

		// override counter properties from template
		if p := me.GetOverride(key); p != "" {
			me.Logger.Debug().Msgf("%soverride counter [%s] property [%s] => [%s]%s", color.Red, key, counter.GetChildContentS("properties"), p, color.End)
			counter.SetChildContentS("properties", p)
		}

		display, ok := wanted.GetHas(key)
		// counter not requested
		if !ok {
			me.Logger.Trace().
				Str("key", key).
				Msg("Skip counter not requested")
			continue
		}

		// deprecated and possibly replaced counter
		// if there is no replacement continue instead of skipping
		if counter.GetChildContentS("is-deprecated") == "true" {
			if r := counter.GetChildContentS("replaced-by"); r != "" {
				me.Logger.Info().
					Str("key", key).
					Str("replacement", r).
					Msg("Replaced deprecated counter")
				if !wanted.Has(r) {
					replaced.Add(r)
				}
			}
		}

		// string metric, add as instance label
		if strings.Contains(counter.GetChildContentS("properties"), "string") {
			oldLabels.Delete(key)
			me.instanceLabels[key] = display
			me.Logger.Debug().Msgf("%s+[%s] added as label name (%s)%s", color.Yellow, key, display, color.End)
		} else {
			// add counter as numeric metric
			oldMetrics.Delete(key)
			if r := me.addCounter(counter, key, display, true, counters); r != "" && !wanted.Has(r) {
				missing.Add(r) // required base counter, missing in template
				me.Logger.Debug().Msgf("%smarking [%s] as required base counter for [%s]%s", color.Red, r, key, color.End)
			}
		}
	}

	// second loop for replaced counters
	if replaced.Size() > 0 {
		me.Logger.Debug().Msgf("attempting to retrieve metadata of %d replaced counters", replaced.Size())
		for name, counter := range counters {
			if replaced.Has(name) {
				oldMetrics.Delete(name)
				me.Logger.Debug().Msgf("adding [%s] (replacement for deprecated counter)", name)
				if r := me.addCounter(counter, name, name, true, counters); r != "" && !wanted.Has(r) {
					missing.Add(r) // required base counter, missing in template
					me.Logger.Debug().Msgf("%smarking [%s] as required base counter for [%s]%s", color.Red, r, name, color.End)
				}
			}
		}
	}

	// third loop for required base counters, not in template
	if missing.Size() > 0 {
		me.Logger.Debug().
			Int("missing", missing.Size()).
			Msg("Attempting to retrieve metadata of missing base counters")
		for name, counter := range counters {
			if missing.Has(name) {
				oldMetrics.Delete(name)
				me.Logger.Debug().Str("name", name).Msg("Adding missing base counter")
				me.addCounter(counter, name, "", false, counters)
			}
		}
	}

	// @TODO check dtype!!!!
	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	if !oldMetrics.Has("timestamp") {
		m, err := mat.NewMetricFloat64("timestamp")
		if err != nil {
			me.Logger.Error().Stack().Err(err).Msg("add timestamp metric")
		}
		m.SetProperty("raw")
		m.SetExportable(false)
	}

	// hack for workload objects, @TODO replace with a plugin
	if me.Query == objWorkload || me.Query == objWorkloadDetail || me.Query == objWorkloadVolume || me.Query == objWorkloadDetailVolume {

		// for these two objects, we need to create latency/ops counters for each of the workload layers
		// there original counters will be discarded
		if me.Query == objWorkloadDetail || me.Query == objWorkloadDetailVolume {

			var service, wait, visits, ops matrix.Metric
			oldMetrics.Delete("service_time")
			oldMetrics.Delete("wait_time")
			oldMetrics.Delete("visits")
			oldMetrics.Delete("ops")

			if service = mat.GetMetric("service_time"); service == nil {
				me.Logger.Error().Stack().Err(nil).Msg("metric [service_time] required to calculate workload missing")
			}

			if wait = mat.GetMetric("wait_time"); wait == nil {
				me.Logger.Error().Stack().Err(nil).Msg("metric [wait-time] required to calculate workload missing")
			}

			if visits = mat.GetMetric("visits"); visits == nil {
				me.Logger.Error().Stack().Err(nil).Msg("metric [visits] required to calculate workload missing")
			}

			if service == nil || wait == nil || visits == nil {
				return nil, errors.New(errors.MissingParam, "workload metrics")
			}

			if ops = mat.GetMetric("ops"); ops == nil {
				if ops, err = mat.NewMetricFloat64("ops"); err != nil {
					return nil, err
				}
				ops.SetProperty(visits.GetProperty())
				me.Logger.Debug().Msgf("+ [resource_ops] [%s] added workload ops metric with property (%s)", ops.GetName(), ops.GetProperty())
			}

			service.SetExportable(false)
			wait.SetExportable(false)
			visits.SetExportable(false)

			if resourceMap := me.Params.GetChildS("resource_map"); resourceMap == nil {
				return nil, errors.New(errors.MissingParam, "resource_map")
			} else {
				for _, x := range resourceMap.GetChildren() {
					name := x.GetNameS()
					resource := x.GetContentS()

					if m := mat.GetMetric(name); m != nil {
						oldMetrics.Delete(name)
						continue
					}
					if m, err := mat.NewMetricFloat64(name); err != nil {
						return nil, err
					} else {
						m.SetName("resource_latency")
						m.SetLabel("resource", resource)
						m.SetProperty(service.GetProperty())
						// base counter is the ops of the same resource
						m.SetComment("ops")

						oldMetrics.Delete(name)
						me.Logger.Debug().Msgf("+ [%s] (=> %s) added workload latency metric", name, resource)
					}
				}
			}
		}

		if qosLabels := me.Params.GetChildS("qos_labels"); qosLabels == nil {
			return nil, errors.New(errors.MissingParam, "qos_labels")
		} else {
			me.qosLabels = make(map[string]string)
			for _, label := range qosLabels.GetAllChildContentS() {

				display := strings.ReplaceAll(label, "-", "_")
				if x := strings.Split(label, "=>"); len(x) == 2 {
					label = strings.TrimSpace(x[0])
					display = strings.TrimSpace(x[1])
				}
				me.qosLabels[label] = display
				//me.instanceLabels[label] = display
				//oldLabels.Delete(label)
			}
		}
	}

	for key := range oldMetrics.Iter() {
		// temporary fix: prevent removing array counters
		// @TODO
		if key != "timestamp" && !strings.Contains(key, ".") {
			mat.RemoveMetric(key)
			me.Logger.Debug().Msgf("removed metric [%s]", key)
		}
	}

	for key := range oldLabels.Iter() {
		delete(me.instanceLabels, key)
		me.Logger.Debug().Msgf("removed label [%s]", key)
	}

	metricsAdded := len(mat.GetMetrics()) - (oldMetricsSize - oldMetrics.Size())
	labelsAdded := len(me.instanceLabels) - (oldLabelsSize - oldLabels.Size())

	me.Logger.Debug().Msgf("added %d new, removed %d metrics (total: %d)", metricsAdded, oldMetrics.Size(), len(mat.GetMetrics()))
	me.Logger.Debug().Msgf("added %d new, removed %d labels (total: %d)", labelsAdded, oldLabels.Size(), len(me.instanceLabels))

	if len(mat.GetMetrics()) == 0 {
		return nil, errors.New(errors.ErrNoMetric, "")
	}

	return nil, nil
}

// create new or update existing metric based on Zapi counter metadata
func (me *ZapiPerf) addCounter(counter *node.Node, name, display string, enabled bool, cache map[string]*node.Node) string {

	var (
		property, baseCounter, unit string
		err                         error
	)

	mat := me.Matrix[me.Object]

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
		me.Logger.Warn().Msgf("skip counter [%s] with unknown property [%s]", name, p)
		return ""
	}

	baseCounter = counter.GetChildContentS("base-counter")
	unit = counter.GetChildContentS("unit")

	if display == "" {
		display = strings.ReplaceAll(name, "-", "_") // redundant for zapiperf
	}

	me.Logger.Debug().Msgf("handling counter [%s] with property [%s] and unit [%s]", name, property, unit)

	// counter type is array, each element will be converted to a metric instance
	if counter.GetChildContentS("type") == "array" {

		var labels, baseLabels []string
		var e string

		if labels, e = parseHistogramLabels(counter); e != "" {
			me.Logger.Warn().Msgf("skipping [%s] of type array: %s", name, e)
			return ""
		}

		if baseCounter != "" {
			if base, ok := cache[baseCounter]; ok {
				if base.GetChildContentS("type") == "array" {
					baseLabels, e = parseHistogramLabels(base)
					if e != "" {
						me.Logger.Warn().Msgf("skipping [%s], base counter [%s] is array, but %s", name, baseCounter, e)
						return ""
					} else if len(baseLabels) != len(labels) {
						me.Logger.Warn().Msgf("skipping [%s], array labels don't match with base counter labels [%s]", name, baseCounter)
						return ""
					}
				}
			} else {
				me.Logger.Warn().Msgf("skipping [%s], base counter [%s] not found", name, baseCounter)
				return ""
			}
		}

		for _, label := range labels {

			var m matrix.Metric

			key := name + "." + label
			baseKey := baseCounter
			if baseCounter != "" && len(baseLabels) != 0 {
				baseKey += "." + baseLabels[0]
			}

			if m = mat.GetMetric(key); m != nil {
				me.Logger.Debug().Msgf("updating array metric [%s] attributes", key)
			} else if m, err = mat.NewMetricFloat64(key); err == nil {
				me.Logger.Debug().Msgf("%s+[%s] added array metric (%s), element with label (%s)%s", color.Pink, name, display, label, color.End)
			} else {
				me.Logger.Error().Stack().Err(err).Msgf("add array metric element [%s]: ", key)
				return ""
			}

			m.SetName(display)
			m.SetProperty(property)
			m.SetComment(baseKey)
			m.SetExportable(enabled)

			if x := strings.Split(label, "."); len(x) == 2 {
				m.SetLabel("metric", x[0])
				m.SetLabel("submetric", x[1])
			} else {
				m.SetLabel("metric", label)
			}
		}
		// cache labels only when parsing counter was success
		me.histogramLabels[name] = labels

		// counter type is scalar
	} else {
		var m matrix.Metric
		if m = mat.GetMetric(name); m != nil {
			me.Logger.Debug().Msgf("updating scalar metric [%s] attributes", name)
		} else if m, err = mat.NewMetricFloat64(name); err == nil {
			me.Logger.Debug().Msgf("%s+[%s] added scalar metric (%s)%s", color.Cyan, name, display, color.End)
		} else {
			me.Logger.Error().Stack().Err(err).Msgf("add scalar metric [%s]", name)
			return ""
		}

		me.scalarCounters = append(me.scalarCounters, name)
		m.SetName(display)
		m.SetProperty(property)
		m.SetComment(baseCounter)
		m.SetExportable(enabled)

	}
	return baseCounter
}

// override counter property
func (me *ZapiPerf) GetOverride(counter string) string {
	if o := me.Params.GetChildS("override"); o != nil {
		return o.GetChildContentS(counter)
	}
	return ""
}

// parse ZAPI array counter (histogram), so we can store it
// as multiple flat metrics
func parseHistogramLabels(elem *node.Node) ([]string, string) {
	var (
		labels []string
		msg    string
	)

	if x := elem.GetChildS("labels"); x == nil {
		msg = "array labels missing"
	} else if d := len(x.GetChildren()); d == 1 {
		labels = strings.Split(node.DecodeHTML(x.GetChildren()[0].GetContentS()), ",")
	} else if d == 2 {
		labelsA := strings.Split(node.DecodeHTML(x.GetChildren()[0].GetContentS()), ",")
		labelsB := strings.Split(node.DecodeHTML(x.GetChildren()[1].GetContentS()), ",")
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

// PollInstance updates instance cache
func (me *ZapiPerf) PollInstance() (map[string]*matrix.Matrix, error) {

	var (
		err                                        error
		request, results                           *node.Node
		oldInstances                               *set.Set
		oldSize, newSize, removed, added           int
		keyAttr, instancesAttr, nameAttr, uuidAttr string
	)

	oldInstances = set.New()
	mat := me.Matrix[me.Object]
	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}
	oldSize = oldInstances.Size()

	me.Logger.Debug().Msgf("updating instance cache (old cache has: %d)", oldInstances.Size())

	nameAttr = "name"
	uuidAttr = "uuid"
	keyAttr = me.instanceKey

	// hack for workload objects: get instances from Zapi
	if me.Query == objWorkload || me.Query == objWorkloadDetail || me.Query == objWorkloadVolume || me.Query == objWorkloadDetailVolume {
		request = node.NewXMLS("qos-workload-get-iter")
		queryElem := request.NewChildS("query", "")
		infoElem := queryElem.NewChildS("qos-workload-info", "")
		if me.Query == objWorkloadVolume || me.Query == objWorkloadDetailVolume {
			infoElem.NewChildS("workload-class", "autovolume")
		} else {
			infoElem.NewChildS("workload-class", "user-defined")
		}

		instancesAttr = "attributes-list"
		nameAttr = "workload-name"
		uuidAttr = "workload-uuid"
		if me.instanceKey == "instance_name" || me.instanceKey == "name" {
			keyAttr = "workload-name"
		} else {
			keyAttr = "workload-uuid"
		}
		// syntax for cdot/perf
	} else if me.Client.IsClustered() {
		request = node.NewXMLS("perf-object-instance-list-info-iter")
		request.NewChildS("objectname", me.Query)
		instancesAttr = "attributes-list"
		// syntax for 7mode/perf
	} else {
		request = node.NewXMLS("perf-object-instance-list-info")
		request.NewChildS("objectname", me.Query)
		instancesAttr = "instances"
	}

	if me.Client.IsClustered() {
		request.NewChildS("max-records", strconv.Itoa(me.batchSize))
	}

	batchTag := "initial"

	for {

		if results, batchTag, err = me.Client.InvokeBatchRequest(request, batchTag); err != nil {
			me.Logger.Error().Err(err).
				Str("request", request.GetNameS()).
				Str("batchTag", batchTag).
				Msg("InvokeBatchRequest failed")
			break
		}

		if results == nil {
			break
		}

		// fetch instances
		instances := results.GetChildS(instancesAttr)
		if instances == nil || len(instances.GetChildren()) == 0 {
			break
		}

		for _, i := range instances.GetChildren() {

			if key := i.GetChildContentS(keyAttr); key == "" {
				// instance key missing
				name := i.GetChildContentS(nameAttr)
				uuid := i.GetChildContentS(uuidAttr)
				me.Logger.Debug().Msgf("skip instance, missing key [%s] (name=%s, uuid=%s)", me.instanceKey, name, uuid)
			} else if oldInstances.Delete(key) {
				// instance already in cache
				me.Logger.Debug().Msgf("updated instance [%s%s%s%s]", color.Bold, color.Yellow, key, color.End)
				continue
			} else if instance, err := mat.NewInstance(key); err != nil {
				me.Logger.Error().Err(err).Msg("add instance")
			} else {
				me.Logger.Debug().
					Str("key", key).
					Msg("Added new instance")
				if me.Query == objWorkload || me.Query == objWorkloadDetail || me.Query == objWorkloadVolume || me.Query == objWorkloadDetailVolume {
					for label, display := range me.qosLabels {
						if value := i.GetChildContentS(label); value != "" {
							instance.SetLabel(display, value)
						}
					}
					me.Logger.Debug().Msgf("(%s) [%s] added QOS labels: %s", me.Query, key, instance.GetLabels().String())
				}
			}
		}
	}

	for key := range oldInstances.Iter() {
		mat.RemoveInstance(key)
		me.Logger.Debug().Msgf("removed instance [%s]", key)
	}

	removed = oldInstances.Size()
	newSize = len(mat.GetInstances())
	added = newSize - (oldSize - removed)

	me.Logger.Debug().Msgf("added %d new, removed %d (total instances %d)", added, removed, newSize)

	if newSize == 0 {
		return nil, errors.New(errors.ErrNoInstance, "")
	}

	return nil, err
}

// Interface guards
var (
	_ collector.Collector = (*ZapiPerf)(nil)
)
