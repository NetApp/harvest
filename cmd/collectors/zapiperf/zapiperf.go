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
	"errors"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/disk"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/externalserviceoperation"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/fabricpool"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/fcp"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/fcvi"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/flexcache"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/headroom"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/nic"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/volumetag"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/vscan"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/rs/zerolog"
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
	keyToken      = "?#"
	// objects that need special handling
	objWorkload             = "workload"
	objWorkloadDetail       = "workload_detail"
	objWorkloadVolume       = "workload_volume"
	objWorkloadDetailVolume = "workload_detail_volume"
	objWorkloadClass        = "user_defined|system_defined"
	objWorkloadVolumeClass  = "autovolume"
	BILLION                 = 1_000_000_000
)

var workloadDetailMetrics = []string{"resource_latency", "service_time_latency"}

type ZapiPerf struct {
	*zapi.Zapi      // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	object          string
	filter          string
	batchSize       int
	latencyIoReqd   int
	instanceKeys    []string
	instanceLabels  map[string]string
	histogramLabels map[string][]string
	scalarCounters  []string
	qosLabels       map[string]string
	isCacheEmpty    bool
	keyName         string
	keyNameIndex    int
	testFilePath    string // Used only from unit test
}

func init() {
	plugin.RegisterModule(&ZapiPerf{})
}

func (z *ZapiPerf) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.zapiperf",
		New: func() plugin.Module { return new(ZapiPerf) },
	}
}

func (z *ZapiPerf) Init(a *collector.AbstractCollector) error {
	z.Zapi = &zapi.Zapi{AbstractCollector: a}

	if err := z.InitVars(); err != nil {
		return err
	}
	// Invoke generic initializer
	// this will load Schedule, initialize data and metadata Matrices
	if err := collector.Init(z); err != nil {
		return err
	}

	if err := z.InitMatrix(); err != nil {
		return err
	}

	if err := z.InitCache(); err != nil {
		return err
	}

	z.Logger.Debug().Msg("initialized")
	return nil
}

func (z *ZapiPerf) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Nic":
		return nic.New(abc)
	case "Fcp":
		return fcp.New(abc)
	case "FabricPool":
		return fabricpool.New(abc)
	case "Headroom":
		return headroom.New(abc)
	case "Volume":
		return volume.New(abc)
	case "VolumeTag":
		return volumetag.New(abc)
	case "Vscan":
		return vscan.New(abc)
	case "Disk":
		return disk.New(abc)
	case "ExternalServiceOperation":
		return externalserviceoperation.New(abc)
	case "FCVI":
		return fcvi.New(abc)
	case "FlexCache":
		return flexcache.New(abc)
	default:
		z.Logger.Info().Msgf("no zapiPerf plugin found for %s", kind)
	}
	return nil
}

func (z *ZapiPerf) InitCache() error {
	z.histogramLabels = make(map[string][]string)
	z.instanceLabels = make(map[string]string)
	z.instanceKeys = z.loadParamArray("instance_key", instanceKey)
	z.filter = z.loadFilter()
	z.batchSize = z.loadParamInt("batch_size", batchSize)
	z.latencyIoReqd = z.loadParamInt("latency_io_reqd", latencyIoReqd)
	z.isCacheEmpty = true
	z.object = z.loadParamStr("object", "")
	z.keyName, z.keyNameIndex = z.initKeyName()
	// hack to override from AbstractCollector
	// @TODO need cleaner solution
	if z.object == "" {
		z.object = z.Object
	}
	z.Matrix[z.Object].Object = z.object
	z.Logger.Debug().Msgf("object= %s --> %s", z.Object, z.object)

	// Add metadata metric for skips/numPartials
	_, _ = z.Metadata.NewMetricUint64("skips")
	_, _ = z.Metadata.NewMetricUint64("numPartials")

	return nil
}

func (z *ZapiPerf) initKeyName() (string, int) {
	// determine what will serve as instance key (either "uuid" or "instance")
	keyName := "instance-uuid"
	keyNameIndex := 0
	// either instance-uuid or instance can be passed as key not both
	for i, k := range z.instanceKeys {
		if k == "uuid" {
			keyName = "instance-uuid"
			keyNameIndex = i
			break
		} else if k == "name" {
			keyName = "instance"
			keyNameIndex = i
			break
		}
	}
	return keyName, keyNameIndex
}

// load a string parameter or use defaultValue
func (z *ZapiPerf) loadParamStr(name, defaultValue string) string {

	var x string

	if x = z.Params.GetChildContentS(name); x != "" {
		z.Logger.Debug().Msgf("using %s = [%s]", name, x)
		return x
	}
	z.Logger.Debug().Msgf("using %s = [%s] (default)", name, defaultValue)
	return defaultValue
}

func (z *ZapiPerf) loadFilter() string {

	counters := z.Params.GetChildS("counters")
	if counters != nil {
		if x := counters.GetChildS("filter"); x != nil {
			filter := strings.Join(x.GetAllChildContentS(), ",")
			return filter
		}
	}
	return ""
}

// load a string parameter or use defaultValue
func (z *ZapiPerf) loadParamArray(name, defaultValue string) []string {

	if v := z.Params.GetChildContentS(name); v != "" {
		z.Logger.Debug().
			Str("name", name).
			Str("value", v).
			Send()
		return []string{v}
	}

	p := z.Params.GetChildS(name)
	if p != nil {
		if v := p.GetAllChildContentS(); v != nil {
			z.Logger.Debug().
				Str("name", name).
				Strs("values", v).
				Send()
			return v
		}
	}

	z.Logger.Debug().
		Str("name", name).
		Str("defaultValue", defaultValue).
		Send()
	return []string{defaultValue}
}

// load workload_class or use defaultValue
func (z *ZapiPerf) loadWorkloadClassQuery(defaultValue string) string {

	var x *node.Node

	name := "workload_class"

	if x = z.Params.GetChildS(name); x != nil {
		v := x.GetAllChildContentS()
		if len(v) == 0 {
			z.Logger.Debug().
				Str("name", name).
				Str("defaultValue", defaultValue).
				Send()
			return defaultValue
		}
		s := strings.Join(v, "|")
		z.Logger.Debug().
			Str("name", name).
			Str("value", s).
			Send()
		return s
	}
	z.Logger.Debug().
		Str("name", name).
		Str("defaultValue", defaultValue).
		Send()
	return defaultValue
}

func (z *ZapiPerf) updateWorkloadQuery(query *node.Node) {
	// filter -> workload-class takes precedence over workload_class param at root level
	// filter -> is-constituent takes precedence over refine -> with_constituents
	workloadClass := ""
	isConstituent := ""
	counters := z.Params.GetChildS("counters")
	if counters != nil {
		filter := counters.GetChildS("filter")
		if filter != nil {
			for _, n := range filter.GetChildren() {
				name := n.GetNameS()
				content := n.GetContentS()
				query.NewChildS(name, content)
				if name == "workload-class" {
					workloadClass = content
				}
				if name == "is-constituent" {
					isConstituent = content
				}
			}
		}
	}
	if workloadClass == "" {
		var workloadClassQuery string
		if z.Query == objWorkloadVolume || z.Query == objWorkloadDetailVolume {
			workloadClassQuery = z.loadWorkloadClassQuery(objWorkloadVolumeClass)
		} else {
			workloadClassQuery = z.loadWorkloadClassQuery(objWorkloadClass)
		}
		query.NewChildS("workload-class", workloadClassQuery)
	}
	if isConstituent == "" {
		if counters != nil {
			refine := counters.GetChildS("refine")
			if refine != nil {
				isConstituent = refine.GetChildContentS("with_constituents")
				if isConstituent == "false" {
					query.NewChildS("is-constituent", isConstituent)
				}
			}
		}
	}
}

// load an int parameter or use defaultValue
func (z *ZapiPerf) loadParamInt(name string, defaultValue int) int {

	var (
		x string
		n int
		e error
	)

	if x = z.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			z.Logger.Debug().Msgf("using %s = [%d]", name, n)
			return n
		}
		z.Logger.Warn().Msgf("invalid parameter %s = [%s] (expected integer)", name, x)
	}

	z.Logger.Debug().Msgf("using %s = [%d] (default)", name, defaultValue)
	return defaultValue
}

func (z *ZapiPerf) isPartialAggregation(instance *node.Node) bool {
	aggregation := instance.GetChildS("aggregation")
	if aggregation != nil {
		aggregationData := aggregation.GetChildS("aggregation-data")
		if aggregationData != nil {
			result := aggregationData.GetChildS("result")
			if result != nil {
				r := result.GetContentS()
				return r == "partial_aggregation"
			}
		}
	}
	return false
}

// PollData updates the data cache of the collector. During first poll, no data will
// be emitted. Afterward, final metric values will be calculated from previous poll.
func (z *ZapiPerf) PollData() (map[string]*matrix.Matrix, error) {

	var (
		instanceKeys []string
		err          error
		skips        int
		numPartials  uint64
		apiT         time.Duration
		parseT       time.Duration
	)

	z.Logger.Trace().Msg("updating data cache")
	prevMat := z.Matrix[z.Object]
	z.Client.Metadata.Reset()

	// clone matrix without numeric data and non-exportable all instances
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: false})
	curMat.Reset()

	timestamp := curMat.GetMetric("timestamp")
	if timestamp == nil {
		return nil, errs.New(errs.ErrConfig, "missing timestamp metric") // @TODO errconfig??
	}

	// for updating metadata
	count := uint64(0)
	batchCount := 0

	// list of instance keys (instance names or uuids) for which
	// we will request counter data
	if z.Query == objWorkloadDetail || z.Query == objWorkloadDetailVolume {
		resourceMap := z.Params.GetChildS("resource_map")
		if resourceMap == nil {
			return nil, errs.New(errs.ErrMissingParam, "resource_map")
		}
		instanceKeys = make([]string, 0)
		for _, layer := range resourceMap.GetAllChildNamesS() {
			for key := range prevMat.GetInstances() {
				instanceKeys = append(instanceKeys, key+"."+layer)
			}
		}
	} else {
		instanceKeys = curMat.GetInstanceKeys()
	}

	// build ZAPI request
	request := node.NewXMLS("perf-object-get-instances")
	request.NewChildS("objectname", z.Query)

	// load requested counters (metrics + labels)
	requestCounters := request.NewChildS("counters", "")
	// load scalar metrics
	for _, key := range z.scalarCounters {
		requestCounters.NewChildS("counter", key)
	}
	// load histograms
	for key := range z.histogramLabels {
		requestCounters.NewChildS("counter", key)
	}
	// load instance labels
	for key := range z.instanceLabels {
		requestCounters.NewChildS("counter", key)
	}

	// batch indices
	startIndex := 0
	endIndex := 0

	for endIndex < len(instanceKeys) {

		// update batch indices
		endIndex += z.batchSize
		// In case of unit test, for loop should run once
		if z.testFilePath != "" {
			endIndex = len(instanceKeys)
		}
		if endIndex > len(instanceKeys) {
			endIndex = len(instanceKeys)
		}

		z.Logger.Trace().
			Int("startIndex", startIndex).
			Int("endIndex", endIndex).
			Msg("Starting batch poll for instances")

		request.PopChildS(z.keyName + "s")
		requestInstances := request.NewChildS(z.keyName+"s", "")
		addedKeys := make(map[string]bool)
		for _, key := range instanceKeys[startIndex:endIndex] {
			if len(z.instanceKeys) == 1 {
				requestInstances.NewChildS(z.keyName, key)
			} else {
				if strings.Contains(key, keyToken) {
					v := strings.Split(key, keyToken)
					if z.keyNameIndex < len(v) {
						key = v[z.keyNameIndex]
					}
				}
				// avoid adding duplicate keys. It can happen for flex-cache case
				if !addedKeys[key] {
					requestInstances.NewChildS(z.keyName, key)
					addedKeys[key] = true
				}
			}
		}

		startIndex = endIndex

		if err = z.Client.BuildRequest(request); err != nil {
			z.Logger.Error().Err(err).
				Str("objectname", z.Query).
				Msg("Build request")
			return nil, err
		}

		response, rd, pd, err := z.Client.InvokeWithTimers(z.testFilePath)
		if err != nil {
			// if ONTAP complains about batch size, use a smaller batch size
			if strings.Contains(err.Error(), "resource limit exceeded") && z.batchSize > 100 {
				z.Logger.Error().Err(err).
					Int("oldBatchSize", z.batchSize).
					Int("newBatchSize", z.batchSize-100).
					Msg("Changed batch_size")
				z.batchSize -= 100
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
			break
		}

		z.Logger.Trace().
			Int("instances", len(instances.GetChildren())).
			Msg("Fetched batch with instances")

		// timestamp for batch instances
		// ignore timestamp from ZAPI which is always integer
		// we want float, since our poll interval can be a float
		ts := float64(time.Now().UnixNano()) / BILLION

		for instIndex, i := range instances.GetChildren() {

			key := z.buildKeyValue(i, z.instanceKeys)

			var layer = "" // latency layer (resource) for workloads

			// special case for these two objects
			// we need to process each latency layer for each instance/counter
			if z.Query == objWorkloadDetail || z.Query == objWorkloadDetailVolume {

				if x := strings.Split(key, "."); len(x) == 2 {
					key = x[0]
					layer = x[1]
				} else {
					z.Logger.Warn().
						Str("key", key).
						Msg("Instance key has unexpected format")
					continue
				}

				for _, wm := range workloadDetailMetrics {
					mLayer := layer + wm
					if l := curMat.GetMetric(mLayer); l == nil {
						z.Logger.Warn().
							Str("layer", mLayer).
							Msg("metric missing in cache")
						continue
					}
				}
			}

			if key == "" {
				z.Logger.Debug().
					Strs("instanceKey", z.instanceKeys).
					Str("name", i.GetChildContentS("name")).
					Str("uuid", i.GetChildContentS("uuid")).
					Msg("Skip instance, key is empty")
				continue
			}

			instance := curMat.GetInstance(key)
			if instance == nil {
				z.Logger.Debug().
					Str("key", key).
					Msg("Skip instance key, not found in cache")
				continue
			}

			if z.isPartialAggregation(i) {
				instance.SetPartial(true)
				numPartials++
			}

			instance.SetExportable(true)
			counters := i.GetChildS("counters")
			if counters == nil {
				z.Logger.Debug().
					Str("key", key).
					Msg("Skip instance key, no data counters")
				continue
			}

			z.Logger.Trace().
				Str("key", key).
				Msg("Fetching data of instance")

			// add batch timestamp as custom counter
			if err := timestamp.SetValueFloat64(instance, ts); err != nil {
				z.Logger.Error().Err(err).Msg("set timestamp value: ")
			}

			for _, cnt := range counters.GetChildren() {

				name := cnt.GetChildContentS("name")
				value := cnt.GetChildContentS("value")

				// sanity check
				if name == "" || value == "" {
					// skip counters with empty value or name
					z.Logger.Trace().
						Str("counter", name).
						Str("value", value).
						Msg("Skipping incomplete counter")
					continue
				}

				// ZAPI counter for us is either instance label (string)
				// or numeric metric (scalar or histogram)

				// store as instance label
				if display, has := z.instanceLabels[name]; has {
					instance.SetLabel(display, value)
					z.Logger.Trace().
						Str("display", display).
						Int("instIndex", instIndex).
						Str("value", value).
						Msg("SetLabel")
					continue
				}

				// store as array counter / histogram
				if labels, has := z.histogramLabels[name]; has {

					values := strings.Split(value, ",")

					if len(labels) != len(values) {
						// warn & skip
						z.Logger.Error().
							Stack().
							Str("labels", name).
							Str("value", value).
							Int("instIndex", instIndex).
							Msg("Histogram labels don't match parsed values")
						continue
					}

					for i, label := range labels {
						if metric := curMat.GetMetric(name + "." + label); metric != nil {
							if err = metric.SetValueString(instance, values[i]); err != nil {
								z.Logger.Error().
									Stack().
									Err(err).
									Str("name", name).
									Str("label", label).
									Str("value", values[i]).
									Int("instIndex", instIndex).
									Msg("Set histogram value failed")
							} else {
								z.Logger.Trace().
									Str("name", name).
									Str("label", label).
									Str("value", values[i]).
									Int("instIndex", instIndex).
									Msg("Set histogram name.label = value")
								count++
							}
						} else {
							z.Logger.Warn().
								Str("name", name).
								Str("label", label).
								Str("value", value).
								Int("instIndex", instIndex).
								Msg("Histogram name. Label not in cache")
						}
					}
					continue
				}

				// special case for workload_detail
				if z.Query == objWorkloadDetail || z.Query == objWorkloadDetailVolume {
					for _, wm := range workloadDetailMetrics {
						// "visits" are ignored
						if name == "visits" {
							continue
						}

						wMetric := curMat.GetMetric(layer + wm)

						if wm == "resource_latency" && (name == "wait_time" || name == "service_time") {
							if err := wMetric.AddValueString(instance, value); err != nil {
								z.Logger.Error().
									Err(err).
									Str("name", name).
									Str("value", value).
									Int("instIndex", instIndex).
									Msg("Add resource_latency failed")
							} else {
								z.Logger.Trace().
									Str("name", name).
									Str("value", value).
									Int("instIndex", instIndex).
									Msg("Add resource_latency")
								count++
							}
							continue
						} else if wm == "service_time_latency" && name == "service_time" {
							if err = wMetric.SetValueString(instance, value); err != nil {
								z.Logger.Error().
									Err(err).
									Str("name", name).
									Str("value", value).
									Int("instIndex", instIndex).
									Msg("Add service_time_latency failed")
							} else {
								z.Logger.Trace().
									Int("instIndex", instIndex).
									Str("name", name).
									Str("value", value).
									Int("instIndex", instIndex).
									Msg("Add service_time_latency")
								count++
							}
						} else if wm == "wait_time_latency" && name == "wait_time" {
							if err = wMetric.SetValueString(instance, value); err != nil {
								z.Logger.Error().
									Err(err).
									Str("name", name).
									Str("value", value).
									Int("instIndex", instIndex).
									Msg("Add wait_time_latency failed")
							} else {
								z.Logger.Trace().
									Int("instIndex", instIndex).
									Str("name", name).
									Str("value", value).
									Int("instIndex", instIndex).
									Msg("Add wait_time_latency")
								count++
							}
						}
					}
					continue
				}

				// store as scalar metric
				if metric := curMat.GetMetric(name); metric != nil {
					if err = metric.SetValueString(instance, value); err != nil {
						z.Logger.Error().
							Err(err).
							Str("name", name).
							Str("value", value).
							Int("instIndex", instIndex).
							Msg("Set metric failed")
					} else {
						z.Logger.Trace().
							Int("instIndex", instIndex).
							Str("key", key).
							Str("counter", name).
							Str("value", value).
							Msg("Set metric")
						count++
					}
					continue
				}

				z.Logger.Warn().
					Int("instIndex", instIndex).
					Str("counter", name).
					Str("value", value).
					Msg("Counter not found in cache")
			} // end loop over counters
		} // end loop over instances
	} // end batch request

	z.Logger.Trace().
		Uint64("count", count).
		Int("batchCount", batchCount).
		Msg("Collected data points in batch polls")

	if z.Query == objWorkloadDetail || z.Query == objWorkloadDetailVolume {
		if rd, pd, err := z.getParentOpsCounters(curMat, z.keyName); err == nil {
			apiT += rd
			parseT += pd
		} else {
			// no point to continue as we can't calculate the other counters
			return nil, err
		}
	}

	// update metadata
	_ = z.Metadata.LazySetValueInt64("api_time", "data", apiT.Microseconds())
	_ = z.Metadata.LazySetValueInt64("parse_time", "data", parseT.Microseconds())
	_ = z.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = z.Metadata.LazySetValueUint64("instances", "data", uint64(len(instanceKeys)))
	_ = z.Metadata.LazySetValueUint64("bytesRx", "data", z.Client.Metadata.BytesRx)
	_ = z.Metadata.LazySetValueUint64("numCalls", "data", z.Client.Metadata.NumCalls)
	_ = z.Metadata.LazySetValueUint64("numPartials", "data", numPartials)

	z.AddCollectCount(count)

	// skip calculating from delta if no data from previous poll
	if z.isCacheEmpty {
		z.Logger.Debug().Msg("skip postprocessing until next poll (previous cache empty)")
		z.Matrix[z.Object] = curMat
		z.isCacheEmpty = false
		return nil, nil
	}

	calcStart := time.Now()

	z.Logger.Trace().Msg("starting delta calculations from previous cache")

	// cache raw data for next poll
	cachedData := curMat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true, PartialInstances: true}) // @TODO implement copy data

	// order metrics, such that those requiring base counters are processed last
	orderedMetrics := make([]*matrix.Metric, 0, len(curMat.GetMetrics()))
	orderedKeys := make([]string, 0, len(orderedMetrics))

	for key, metric := range curMat.GetMetrics() {
		if metric.GetComment() == "" && metric.Buckets() == nil { // does not require base counter
			orderedMetrics = append(orderedMetrics, metric)
			orderedKeys = append(orderedKeys, key)
		}
	}
	for key, metric := range curMat.GetMetrics() {
		if metric.GetComment() != "" && metric.Buckets() == nil { // requires base counter
			orderedMetrics = append(orderedMetrics, metric)
			orderedKeys = append(orderedKeys, key)
		}
	}

	// calculate timestamp delta first since many counters require it for postprocessing.
	// Timestamp has "raw" property, so it isn't post-processed automatically
	if _, err = curMat.Delta("timestamp", prevMat, z.Logger); err != nil {
		z.Logger.Error().Err(err).Msg("(timestamp) calculate delta:")
		// @TODO terminate since other counters will be incorrect
	}

	var base *matrix.Metric
	var totalSkips int

	for i, metric := range orderedMetrics {

		property := metric.GetProperty()
		key := orderedKeys[i]

		// RAW - submit without post-processing
		if property == "raw" {
			continue
		}

		// all other properties - first calculate delta
		if skips, err = curMat.Delta(key, prevMat, z.Logger); err != nil {
			z.Logger.Error().Err(err).Str("key", key).Msg("Calculate delta")
			continue
		}
		totalSkips += skips

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
		if base = curMat.GetMetric(metric.GetComment()); base == nil {
			z.Logger.Warn().
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
				skips, err = curMat.DivideWithThreshold(key, metric.GetComment(), z.latencyIoReqd, cachedData, prevMat, z.Logger)
			} else {
				skips, err = curMat.Divide(key, metric.GetComment(), z.Logger)
			}

			if err != nil {
				z.Logger.Error().Err(err).Str("key", key).Msg("Division by base")
			}
			totalSkips += skips

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if skips, err = curMat.MultiplyByScalar(key, 100, z.Logger); err != nil {
				z.Logger.Error().Err(err).Str("key", key).Msg("Multiply by scalar")
			} else {
				totalSkips += skips
			}
			continue
		}
		z.Logger.Error().Err(err).
			Str("key", key).
			Str("property", property).
			Msg("Unknown property")
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		if metric.GetProperty() == "rate" {
			if skips, err = curMat.Divide(orderedKeys[i], "timestamp", z.Logger); err != nil {
				z.Logger.Error().Err(err).
					Int("i", i).
					Str("key", orderedKeys[i]).
					Msg("Calculate rate")
				continue
			}
			totalSkips += skips
		}
	}

	calcD := time.Since(calcStart)

	_ = z.Metadata.LazySetValueInt64("calc_time", "data", calcD.Microseconds())
	_ = z.Metadata.LazySetValueUint64("skips", "data", uint64(totalSkips))

	// store cache for next poll
	z.Matrix[z.Object] = cachedData

	newDataMap := make(map[string]*matrix.Matrix)
	newDataMap[z.Object] = curMat
	return newDataMap, nil
}

// Poll counter "ops" of the related/parent object, required for objects
// workload_detail and workload_detail_volume. This counter is already
// collected by the other ZapiPerf collectors, so this poll is redundant
// (until we implement some sort of inter-collector communication).
func (z *ZapiPerf) getParentOpsCounters(data *matrix.Matrix, KeyAttr string) (time.Duration, time.Duration, error) {

	var (
		ops          *matrix.Metric
		object       string
		instanceKeys []string
		apiT, parseT time.Duration
		err          error
	)

	if z.Query == objWorkloadDetail {
		object = objWorkload
	} else {
		object = objWorkloadVolume
	}

	z.Logger.Debug().Msgf("(%s) starting redundancy poll for ops from parent object (%s)", z.Query, object)

	apiT = 0 * time.Second
	parseT = 0 * time.Second

	if ops = data.GetMetric("ops"); ops == nil {
		z.Logger.Error().Err(nil).Msgf("ops counter not found in cache")
		return apiT, parseT, errs.New(errs.ErrMissingParam, "counter ops")
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
		endIndex += z.batchSize
		if endIndex > len(instanceKeys) {
			endIndex = len(instanceKeys)
		}

		z.Logger.Debug().Msgf("starting batch poll for instances [%d:%d]", startIndex, endIndex)

		request.PopChildS(KeyAttr + "s")
		requestInstances := request.NewChildS(KeyAttr+"s", "")
		for _, key := range instanceKeys[startIndex:endIndex] {
			requestInstances.NewChildS(KeyAttr, key)
		}

		startIndex = endIndex

		if err = z.Client.BuildRequest(request); err != nil {
			return apiT, parseT, err
		}

		response, rt, pt, err := z.Client.InvokeWithTimers("")
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

			key := z.buildKeyValue(i, z.instanceKeys)

			if key == "" {
				z.Logger.Debug().
					Strs("key", z.instanceKeys).
					Str("name", i.GetChildContentS("name")).
					Str("uuid", i.GetChildContentS("uuid")).
					Msg("skip instance")
				continue
			}

			instance := data.GetInstance(key)
			if instance == nil {
				z.Logger.Warn().Msgf("skip instance [%s], not found in cache", key)
				continue
			}

			counters := i.GetChildS("counters")
			if counters == nil {
				z.Logger.Debug().Msgf("skip instance [%s], no data counters", key)
				continue
			}

			for _, cnt := range counters.GetChildren() {

				name := cnt.GetChildContentS("name")
				value := cnt.GetChildContentS("value")

				z.Logger.Trace().Msgf("(%s%s%s%s) parsing counter = %v", color.Grey, key, color.End, name, value)

				if name == "ops" {
					if err = ops.SetValueString(instance, value); err != nil {
						z.Logger.Error().Err(err).Msgf("set metric (%s) value [%s]", name, value)
					} else {
						z.Logger.Trace().Msgf("+ metric (%s) = [%s%s%s]", name, color.Cyan, value, color.End)
						count++
					}
				} else {
					z.Logger.Error().Err(nil).Msgf("unrequested metric (%s)", name)
				}
			}
		}
	}
	z.Logger.Debug().Msgf("(%s) completed redundant ops poll (%s): collected %d", z.Query, object, count)
	return apiT, parseT, nil
}

func (z *ZapiPerf) PollCounter() (map[string]*matrix.Matrix, error) {
	var (
		err                                      error
		request, response, counterList           *node.Node
		oldMetrics, oldLabels, replaced, missing *set.Set
		wanted                                   map[string]string
		oldMetricsSize, oldLabelsSize            int
		counters                                 map[string]*node.Node
		apiT, parseT                             time.Time
		apiD                                     time.Duration
	)

	z.scalarCounters = make([]string, 0)
	counters = make(map[string]*node.Node)
	oldMetrics = set.New()           // current set of metrics, so we can remove from matrix if not updated
	oldLabels = set.New()            // current set of labels
	wanted = make(map[string]string) // counters listed in template, maps raw name to display name
	missing = set.New()              // required base counters, missing in template
	replaced = set.New()             // deprecated and replaced counters

	mat := z.Matrix[z.Object]
	z.Client.Metadata.Reset()

	for key := range mat.GetMetrics() {
		oldMetrics.Add(key)
	}
	oldMetricsSize = oldMetrics.Size()

	for key := range z.instanceLabels {
		oldLabels.Add(key)
	}
	oldLabelsSize = oldLabels.Size()

	// parse list of counters defined in template
	if counterList = z.Params.GetChildS("counters"); counterList != nil {
		for _, cnt := range counterList.GetAllChildContentS() {
			if renamed := strings.Split(cnt, "=>"); len(renamed) == 2 {
				wanted[strings.TrimSpace(renamed[0])] = strings.TrimSpace(renamed[1])
			} else if cnt == "instance_name" {
				wanted["instance_name"] = z.object
			} else {
				display := strings.ReplaceAll(cnt, "-", "_")
				if strings.HasPrefix(display, z.object) {
					display = strings.TrimPrefix(display, z.object)
					display = strings.TrimPrefix(display, "_")
				}
				wanted[cnt] = display
			}
		}
	} else {
		return nil, errs.New(errs.ErrMissingParam, "counters")
	}

	z.Logger.Trace().
		Int("oldMetrics", oldMetricsSize).
		Int("oldLabels", oldLabelsSize).
		Msg("Updating metric cache")

	// build request
	request = node.NewXMLS("perf-object-counter-list-info")
	request.NewChildS("objectname", z.Query)

	if err = z.Client.BuildRequest(request); err != nil {
		return nil, err
	}

	apiT = time.Now()
	if response, err = z.Client.Invoke(z.testFilePath); err != nil {
		return nil, err
	}
	apiD = time.Since(apiT)
	parseT = time.Now()

	// fetch counter elements
	if elems := response.GetChildS("counters"); elems != nil && len(elems.GetChildren()) != 0 {
		for _, counter := range elems.GetChildren() {
			if name := counter.GetChildContentS("name"); name != "" {
				counters[name] = counter
			}
		}
	} else {
		return nil, errs.New(errs.ErrNoMetric, "no counters in response")
	}

	for key, counter := range counters {

		// override counter properties from template
		if p := z.GetOverride(key); p != "" {
			z.Logger.Trace().Msgf("%soverride counter [%s] property [%s] => [%s]%s", color.Red, key, counter.GetChildContentS("properties"), p, color.End)
			counter.SetChildContentS("properties", p)
		}

		display, ok := wanted[key]
		// counter not requested
		if !ok {
			z.Logger.Trace().
				Str("key", key).
				Msg("Skip counter not requested")
			continue
		}

		// deprecated and possibly replaced counter
		// if there is no replacement continue instead of skipping
		if counter.GetChildContentS("is-deprecated") == "true" {
			if r := counter.GetChildContentS("replaced-by"); r != "" {
				z.Logger.Info().
					Str("key", key).
					Str("replacement", r).
					Msg("Replaced deprecated counter")
				_, ok = wanted[r]
				if !ok {
					replaced.Add(r)
				}
			}
		}

		// string metric, add as instance label
		if strings.Contains(counter.GetChildContentS("properties"), "string") {
			oldLabels.Remove(key)
			z.instanceLabels[key] = display
			z.Logger.Trace().Msgf("%s+[%s] added as label name (%s)%s", color.Yellow, key, display, color.End)
		} else {
			// add counter as numeric metric
			oldMetrics.Remove(key)
			if r := z.addCounter(counter, key, display, true, counters); r != "" {
				_, ok = wanted[r]
				if !ok {
					missing.Add(r) // required base counter, missing in template
					z.Logger.Trace().Msgf("%smarking [%s] as required base counter for [%s]%s", color.Red, r, key, color.End)
				}
			}
		}
	}

	// second loop for replaced counters
	if replaced.Size() > 0 {
		z.Logger.Debug().Msgf("attempting to retrieve metadata of %d replaced counters", replaced.Size())
		for name, counter := range counters {
			if replaced.Has(name) {
				oldMetrics.Remove(name)
				z.Logger.Debug().Msgf("adding [%s] (replacement for deprecated counter)", name)
				if r := z.addCounter(counter, name, name, true, counters); r != "" {
					_, ok := wanted[r]
					if !ok {
						missing.Add(r) // required base counter, missing in template
						z.Logger.Debug().Msgf("%smarking [%s] as required base counter for [%s]%s", color.Red, r, name, color.End)
					}
				}
			}
		}
	}

	// third loop for required base counters, not in template
	if missing.Size() > 0 {
		z.Logger.Debug().
			Int("missing", missing.Size()).
			Msg("Attempting to retrieve metadata of missing base counters")
		for name, counter := range counters {
			if missing.Has(name) {
				oldMetrics.Remove(name)
				z.Logger.Debug().Str("name", name).Msg("Adding missing base counter")
				z.addCounter(counter, name, "", false, counters)
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
			z.Logger.Error().Err(err).Msg("add timestamp metric")
		}
		m.SetProperty("raw")
		m.SetExportable(false)
	}

	// hack for workload objects, @TODO replace with a plugin
	if z.Query == objWorkload || z.Query == objWorkloadDetail || z.Query == objWorkloadVolume || z.Query == objWorkloadDetailVolume {

		// for these two objects, we need to create latency/ops counters for each of the workload layers
		// there original counters will be discarded
		if z.Query == objWorkloadDetail || z.Query == objWorkloadDetailVolume {

			var service, wait, visits, ops *matrix.Metric
			oldMetrics.Remove("service_time")
			oldMetrics.Remove("wait_time")
			oldMetrics.Remove("visits")
			oldMetrics.Remove("ops")

			if service = mat.GetMetric("service_time"); service == nil {
				z.Logger.Error().Err(nil).Msg("metric [service_time] required to calculate workload missing")
			}

			if wait = mat.GetMetric("wait_time"); wait == nil {
				z.Logger.Error().Err(nil).Msg("metric [wait-time] required to calculate workload missing")
			}

			if visits = mat.GetMetric("visits"); visits == nil {
				z.Logger.Error().Err(nil).Msg("metric [visits] required to calculate workload missing")
			}

			if service == nil || wait == nil || visits == nil {
				return nil, errs.New(errs.ErrMissingParam, "workload metrics")
			}

			if ops = mat.GetMetric("ops"); ops == nil {
				if ops, err = mat.NewMetricFloat64("ops"); err != nil {
					return nil, err
				}
				ops.SetProperty(visits.GetProperty())
				z.Logger.Trace().
					Str("opsName", ops.GetName()).
					Str("opsProperty", ops.GetProperty()).
					Msg("[resource_ops] added workload ops metric with property")
			}

			service.SetExportable(false)
			wait.SetExportable(false)
			visits.SetExportable(false)

			resourceMap := z.Params.GetChildS("resource_map")
			if resourceMap == nil {
				return nil, errs.New(errs.ErrMissingParam, "resource_map")
			}
			for _, x := range resourceMap.GetChildren() {
				for _, wm := range workloadDetailMetrics {

					name := x.GetNameS() + wm
					resource := x.GetContentS()

					if m := mat.GetMetric(name); m != nil {
						oldMetrics.Remove(name)
						continue
					}
					m, err := mat.NewMetricFloat64(name, wm)
					if err != nil {
						return nil, err
					}
					m.SetLabel("resource", resource)
					m.SetProperty(service.GetProperty())
					// base counter is the ops of the same resource
					m.SetComment("ops")

					oldMetrics.Remove(name)
					z.Logger.Trace().
						Str("name", name).
						Str("resource", resource).
						Msg("added workload latency metric")
				}
			}
		}

		qosLabels := z.Params.GetChildS("qos_labels")
		if qosLabels == nil {
			return nil, errs.New(errs.ErrMissingParam, "qos_labels")
		}
		z.qosLabels = make(map[string]string)
		for _, label := range qosLabels.GetAllChildContentS() {

			display := strings.ReplaceAll(label, "-", "_")
			if x := strings.Split(label, "=>"); len(x) == 2 {
				label = strings.TrimSpace(x[0])
				display = strings.TrimSpace(x[1])
			}
			z.qosLabels[label] = display
		}
	}

	for key := range oldMetrics.Iter() {
		// temporary fix: prevent removing array counters
		// @TODO
		if key != "timestamp" && !strings.Contains(key, ".") {
			mat.RemoveMetric(key)
			z.Logger.Debug().Msgf("removed metric [%s]", key)
		}
	}

	for key := range oldLabels.Iter() {
		delete(z.instanceLabels, key)
		z.Logger.Debug().Msgf("removed label [%s]", key)
	}

	numMetrics := len(mat.GetMetrics())
	metricsAdded := numMetrics - (oldMetricsSize - oldMetrics.Size())
	labelsAdded := len(z.instanceLabels) - (oldLabelsSize - oldLabels.Size())

	z.Logger.Trace().Int("new", metricsAdded).Int("removed", oldMetrics.Size()).Int("total", numMetrics).Msg("metrics")
	z.Logger.Trace().Int("new", labelsAdded).Int("removed", oldLabels.Size()).Int("total", len(z.instanceLabels)).Msg("labels")

	// update metadata for collector logs
	_ = z.Metadata.LazySetValueInt64("api_time", "counter", apiD.Microseconds())
	_ = z.Metadata.LazySetValueInt64("parse_time", "counter", time.Since(parseT).Microseconds())
	_ = z.Metadata.LazySetValueUint64("metrics", "counter", uint64(numMetrics))
	_ = z.Metadata.LazySetValueUint64("bytesRx", "counter", z.Client.Metadata.BytesRx)
	_ = z.Metadata.LazySetValueUint64("numCalls", "counter", z.Client.Metadata.NumCalls)

	if numMetrics == 0 {
		return nil, errs.New(errs.ErrNoMetric, "")
	}

	return nil, nil
}

// create new or update existing metric based on Zapi counter metadata
func (z *ZapiPerf) addCounter(counter *node.Node, name, display string, enabled bool, cache map[string]*node.Node) string {

	var (
		property, baseCounter, unit string
		err                         error
	)

	mat := z.Matrix[z.Object]

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
		z.Logger.Warn().Msgf("skip counter [%s] with unknown property [%s]", name, p)
		return ""
	}

	baseCounter = counter.GetChildContentS("base-counter")
	unit = counter.GetChildContentS("unit")

	if display == "" {
		display = strings.ReplaceAll(name, "-", "_") // redundant for zapiperf
	}

	z.Logger.Trace().Msgf("handling counter [%s] with property [%s] and unit [%s]", name, property, unit)

	// counter type is array, each element will be converted to a metric instance
	if counter.GetChildContentS("type") == "array" {

		var (
			labels, baseLabels []string
			e                  string
			description        string
			isHistogram        bool
			histogramMetric    *matrix.Metric
		)

		description = strings.ToLower(counter.GetChildContentS("desc"))

		if labels, e = parseHistogramLabels(counter); e != "" {
			z.Logger.Warn().Msgf("skipping [%s] of type array: %s", name, e)
			return ""
		}

		if baseCounter != "" {
			if base, ok := cache[baseCounter]; ok {
				if base.GetChildContentS("type") == "array" {
					baseLabels, e = parseHistogramLabels(base)
					if e != "" {
						z.Logger.Warn().Msgf("skipping [%s], base counter [%s] is array, but %s", name, baseCounter, e)
						return ""
					} else if len(baseLabels) != len(labels) {
						z.Logger.Warn().Msgf("skipping [%s], array labels don't match with base counter labels [%s]", name, baseCounter)
						return ""
					}
				}
			} else {
				z.Logger.Warn().Msgf("skipping [%s], base counter [%s] not found", name, baseCounter)
				return ""
			}
		}

		// ONTAP does not have a `type` for histogram. Harvest tests the `desc` field to determine
		// if a counter is a histogram
		isHistogram = false
		if len(labels) > 0 && strings.Contains(description, "histogram") {
			key := name + ".bucket"
			histogramMetric = mat.GetMetric(key)
			if histogramMetric != nil {
				z.Logger.Trace().Str("metric", key).Msg("Updating array metric attributes")
			} else {
				histogramMetric, err = mat.NewMetricFloat64(key, display)
				if err != nil {
					z.Logger.Error().Err(err).Str("key", key).Msg("unable to create histogram metric")
					return ""
				}
			}
			baseKey := baseCounter
			if baseCounter != "" && len(baseLabels) != 0 {
				baseKey += "." + baseLabels[0]
			}
			histogramMetric.SetProperty(property)
			histogramMetric.SetComment(baseKey)
			histogramMetric.SetExportable(enabled)
			histogramMetric.SetBuckets(&labels)
			isHistogram = true
		}

		for i, label := range labels {

			var m *matrix.Metric

			key := name + "." + label
			baseKey := baseCounter
			if baseKey != "" && len(baseLabels) != 0 {
				baseKey += "." + baseLabels[i]
			}
			if m = mat.GetMetric(key); m != nil {
				z.Logger.Trace().Msgf("updating array metric [%s] attributes", key)
			} else if m, err = mat.NewMetricFloat64(key, display); err == nil {
				z.Logger.Trace().Msgf("%s+[%s] added array metric (%s), element with label (%s)%s", color.Pink, name, display, label, color.End)
			} else {
				z.Logger.Error().Err(err).Msgf("add array metric element [%s]: ", key)
				return ""
			}

			m.SetProperty(property)
			m.SetComment(baseKey)
			m.SetExportable(enabled)

			if x := strings.Split(label, "."); len(x) == 2 {
				m.SetLabel("metric", x[0])
				m.SetLabel("submetric", x[1])
			} else {
				m.SetLabel("metric", label)
				if isHistogram {
					// Save the index of this label so the labels can be exported in order
					m.SetLabel("comment", strconv.Itoa(i))
					// Save the bucket name so the flattened metrics can find their bucket when exported
					m.SetLabel("bucket", name+".bucket")
					m.SetHistogram(true)
				}
			}
		}
		// cache labels only when parsing counter was success
		z.histogramLabels[name] = labels

		// counter type is scalar
	} else {
		var m *matrix.Metric
		if m = mat.GetMetric(name); m != nil {
			z.Logger.Trace().Msgf("updating scalar metric [%s] attributes", name)
		} else if m, err = mat.NewMetricFloat64(name, display); err == nil {
			z.Logger.Trace().Msgf("%s+[%s] added scalar metric (%s)%s", color.Cyan, name, display, color.End)
		} else {
			z.Logger.Error().Err(err).Msgf("add scalar metric [%s]", name)
			return ""
		}

		z.scalarCounters = append(z.scalarCounters, name)
		m.SetProperty(property)
		m.SetComment(baseCounter)
		m.SetExportable(enabled)

	}
	return baseCounter
}

// GetOverride overrides a counter property
func (z *ZapiPerf) GetOverride(counter string) string {
	if o := z.Params.GetChildS("override"); o != nil {
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
func (z *ZapiPerf) PollInstance() (map[string]*matrix.Matrix, error) {

	var (
		err                               error
		request, results                  *node.Node
		oldInstances                      *set.Set
		oldSize, newSize, removed, added  int
		instancesAttr, nameAttr, uuidAttr string
		keyAttrs                          []string
		apiD, parseD                      time.Duration
		apiT, parseT                      time.Time
	)

	oldInstances = set.New()
	mat := z.Matrix[z.Object]
	z.Client.Metadata.Reset()

	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}
	oldSize = oldInstances.Size()

	z.Logger.Trace().
		Int("oldInstancesSize", oldInstances.Size()).
		Msg("updating instance cache")

	nameAttr = "name"
	uuidAttr = "uuid"
	keyAttrs = z.instanceKeys

	// hack for workload objects: get instances from Zapi
	if z.Query == objWorkload || z.Query == objWorkloadDetail || z.Query == objWorkloadVolume || z.Query == objWorkloadDetailVolume {
		request = node.NewXMLS("qos-workload-get-iter")
		queryElem := request.NewChildS("query", "")
		infoElem := queryElem.NewChildS("qos-workload-info", "")
		z.updateWorkloadQuery(infoElem)
		instancesAttr = "attributes-list"
		nameAttr = "workload-name"
		uuidAttr = "workload-uuid"
		var ikey string
		if len(z.instanceKeys) > 0 {
			ikey = z.instanceKeys[0]
		}
		if ikey == "instance_name" || ikey == "name" {
			keyAttrs = []string{"workload-name"}
		} else {
			keyAttrs = []string{"workload-uuid"}
		}
		// syntax for cdot/perf
	} else if z.Client.IsClustered() {
		request = node.NewXMLS("perf-object-instance-list-info-iter")
		request.NewChildS("objectname", z.Query)
		if z.filter != "" {
			request.NewChildS("filter-data", z.filter)
		}
		instancesAttr = "attributes-list"
	} else {
		// syntax for 7mode/perf
		request = node.NewXMLS("perf-object-instance-list-info")
		request.NewChildS("objectname", z.Query)
		instancesAttr = "instances"
	}

	if z.Client.IsClustered() {
		request.NewChildS("max-records", strconv.Itoa(z.batchSize))
	}

	batchTag := "initial"

	for {
		apiT = time.Now()
		if results, batchTag, err = z.Client.InvokeBatchRequest(request, batchTag, z.testFilePath); err != nil {
			if errors.Is(err, errs.ErrAPIRequestRejected) {
				z.Logger.Info().
					Str("request", request.GetNameS()).
					Msg(err.Error())
			} else {
				z.Logger.Error().
					Err(err).
					Str("request", request.GetNameS()).
					Str("batchTag", batchTag).
					Msg("InvokeBatchRequest failed")
			}
			apiD += time.Since(apiT)
			break
		}

		apiD += time.Since(apiT)
		parseT = time.Now()

		if results == nil {
			break
		}

		// fetch instances
		instances := results.GetChildS(instancesAttr)
		if instances == nil || len(instances.GetChildren()) == 0 {
			break
		}

		for _, i := range instances.GetChildren() {

			key := z.buildKeyValue(i, keyAttrs)
			if key == "" {
				// instance key missing
				name := i.GetChildContentS(nameAttr)
				uuid := i.GetChildContentS(uuidAttr)
				z.Logger.Debug().
					Strs("key", z.instanceKeys).
					Str("name", name).
					Str("uuid", uuid).
					Msg("skip instance")
			} else if oldInstances.Has(key) {
				// instance already in cache
				oldInstances.Remove(key)
				instance := mat.GetInstance(key)
				z.updateQosLabels(i, instance, key)
				z.Logger.Trace().Msgf("updated instance [%s%s%s%s]", color.Bold, color.Yellow, key, color.End)
				continue
			} else if instance, err := mat.NewInstance(key); err != nil {
				z.Logger.Error().Err(err).Msg("add instance")
			} else {
				z.Logger.Trace().
					Str("key", key).
					Msg("Added new instance")
				z.updateQosLabels(i, instance, key)
			}
		}
		parseD += time.Since(parseT)
	}

	for key := range oldInstances.Iter() {
		mat.RemoveInstance(key)
		z.Logger.Debug().Msgf("removed instance [%s]", key)
	}

	removed = oldInstances.Size()
	newSize = len(mat.GetInstances())
	added = newSize - (oldSize - removed)

	z.Logger.Debug().Int("new", added).Int("removed", removed).Int("total", newSize).Msg("instances")

	// update metadata for collector logs
	_ = z.Metadata.LazySetValueInt64("api_time", "instance", apiD.Microseconds())
	_ = z.Metadata.LazySetValueInt64("parse_time", "instance", parseD.Microseconds())
	_ = z.Metadata.LazySetValueUint64("instances", "instance", uint64(newSize))
	_ = z.Metadata.LazySetValueUint64("bytesRx", "instance", z.Client.Metadata.BytesRx)
	_ = z.Metadata.LazySetValueUint64("numCalls", "instance", z.Client.Metadata.NumCalls)
	if newSize == 0 {
		return nil, errs.New(errs.ErrNoInstance, "")
	}

	return nil, err
}

func (z *ZapiPerf) updateQosLabels(qos *node.Node, instance *matrix.Instance, key string) {
	if z.Query == objWorkload || z.Query == objWorkloadDetail || z.Query == objWorkloadVolume || z.Query == objWorkloadDetailVolume {
		for label, display := range z.qosLabels {
			if value := qos.GetChildContentS(label); value != "" {
				instance.SetLabel(display, value)
			}
		}
		if z.Logger.GetLevel() == zerolog.DebugLevel {
			z.Logger.Debug().
				Str("query", z.Query).
				Str("key", key).
				Str("qos labels", dict.String(instance.GetLabels())).
				Send()
		}
	}
}

func (z *ZapiPerf) buildKeyValue(i *node.Node, keys []string) string {
	if len(keys) == 1 {
		return i.GetChildContentS(keys[0])
	}
	var values []string
	for _, k := range keys {
		value := i.GetChildContentS(k)
		if value != "" {
			values = append(values, value)
		} else {
			z.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
		}
	}
	return strings.Join(values, keyToken)
}

// Interface guards
var (
	_ collector.Collector = (*ZapiPerf)(nil)
)
