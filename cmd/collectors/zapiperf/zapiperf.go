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
	"context"
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
	collector2 "github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"maps"
	"slices"
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
	timestampMetricName     = "timestamp"
)

var workloadDetailMetrics = []string{"resource_latency"}

type ZapiPerf struct {
	*zapi.Zapi              // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	object                  string
	filter                  string
	batchSize               int
	latencyIoReqd           int
	instanceKeys            []string
	instanceLabels          map[string]string
	histogramLabels         map[string][]string
	scalarCounters          []string
	qosLabels               map[string]string
	isCacheEmpty            bool
	keyName                 string
	keyNameIndex            int
	testFilePath            string // Used only from unit test
	recordsToSave           int    // Number of records to save when using the recorder
	pollDataCalls           int
	pollInstanceCalls       int
	allowPartialAggregation bool // allow partial aggregation for this collector
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

	z.InitQOS()

	z.recordsToSave = collector.RecordKeepLast(z.Params, z.Logger)

	z.Logger.Debug("initialized")
	return nil
}

func (z *ZapiPerf) InitQOS() {
	counters := z.Params.GetChildS("counters")
	if counters != nil {
		refine := counters.GetChildS("refine")
		if refine != nil {
			withServiceLatency := refine.GetChildContentS("with_service_latency")
			if withServiceLatency != "false" {
				workloadDetailMetrics = append(workloadDetailMetrics, "service_time_latency")
			}
		}
	}
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
		z.Logger.Info("no zapiPerf plugin found for %s", slog.String("kind", kind))
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
	allowPartialAggregation := z.loadParamStr("allow_partial_aggregation", "false")
	if allowPartialAggregation == "true" {
		z.allowPartialAggregation = true
	}
	z.keyName, z.keyNameIndex = z.initKeyName()
	// hack to override from AbstractCollector
	// @TODO need cleaner solution
	if z.object == "" {
		z.object = z.Object
	}
	z.Matrix[z.Object].Object = z.object
	z.Logger.Debug("->", slog.String("z.Object", z.object), slog.String("z.object", z.object))

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
		z.Logger.Debug("using", slog.String(name, x))
		return x
	}
	z.Logger.Debug("using", slog.String(name, defaultValue))
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
		z.Logger.Debug("", slog.String("name", name), slog.String("value", v))
		return []string{v}
	}

	p := z.Params.GetChildS(name)
	if p != nil {
		if v := p.GetAllChildContentS(); v != nil {
			z.Logger.Debug("", slog.String("name", name), slog.Any("values", v))
			return v
		}
	}

	z.Logger.Debug("", slog.String("name", name), slog.String("defaultValue", defaultValue))
	return []string{defaultValue}
}

// load workload_class or use defaultValue
func (z *ZapiPerf) loadWorkloadClassQuery(defaultValue string) string {

	var x *node.Node

	name := "workload_class"

	if x = z.Params.GetChildS(name); x != nil {
		v := x.GetAllChildContentS()
		if len(v) == 0 {
			z.Logger.Debug(
				"",
				slog.String("name", name),
				slog.String("defaultValue", defaultValue),
			)
			return defaultValue
		}
		s := strings.Join(v, "|")
		z.Logger.Debug("", slog.String("name", name), slog.String("value", s))
		return s
	}
	z.Logger.Debug("", slog.String("name", name), slog.String("defaultValue", defaultValue))
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
			z.Logger.Debug("using", slog.String("name", name), slog.Int("value", n))
			return n
		}
		z.Logger.Warn("invalid parameter (expected integer)", slog.String("name", name), slog.String("value", x))
	}

	z.Logger.Debug("using", slog.String("name", name), slog.Int("value", defaultValue))
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

	prevMat := z.Matrix[z.Object]
	z.Client.Metadata.Reset()

	// clone matrix without numeric data and non-exportable all instances
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: false})
	curMat.Reset()

	timestamp := curMat.GetMetric(timestampMetricName)
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

	// Sort the counters and instanceKeys so they are deterministic

	for _, key := range z.scalarCounters {
		requestCounters.NewChildS("counter", key)
	}

	// load histograms
	sortedHistogramKeys := slices.Sorted(maps.Keys(z.histogramLabels))
	for _, key := range sortedHistogramKeys {
		requestCounters.NewChildS("counter", key)
	}

	// load instance labels
	sortedLabels := slices.Sorted(maps.Keys(z.instanceLabels))
	for _, key := range sortedLabels {
		requestCounters.NewChildS("counter", key)
	}

	slices.Sort(instanceKeys)

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
				// Avoid adding duplicate keys. It can happen for flex-cache case
				if !addedKeys[key] {
					requestInstances.NewChildS(z.keyName, key)
					addedKeys[key] = true
				}
			}
		}

		startIndex = endIndex

		if err = z.Client.BuildRequest(request); err != nil {
			z.Logger.Error("Build request", slogx.Err(err), slog.String("objectname", z.Query))
			return nil, err
		}

		z.pollDataCalls++
		if z.pollDataCalls >= z.recordsToSave {
			z.pollDataCalls = 0
		}

		var headers map[string]string

		poller, err := conf.PollerNamed(z.Options.Poller)
		if err != nil {
			slog.Error("failed to find poller", slogx.Err(err), slog.String("poller", z.Options.Poller))
		}

		if poller.IsRecording() {
			headers = map[string]string{
				"From": strconv.Itoa(z.pollDataCalls),
			}
		}

		response, rd, pd, err := z.Client.InvokeWithTimers(z.testFilePath, headers)

		if err != nil {
			errMsg := strings.ToLower(err.Error())
			// if ONTAP complains about batch size, use a smaller batch size
			if strings.Contains(errMsg, "resource limit exceeded") && z.batchSize > 100 {
				z.Logger.Error(
					"Changed batch_size",
					slogx.Err(err),
					slog.Int("oldBatchSize", z.batchSize),
					slog.Int("newBatchSize", z.batchSize-100),
				)
				z.batchSize -= 100
				return nil, nil
			} else if strings.Contains(errMsg, "timeout: operation") && z.batchSize > 100 {
				z.Logger.Error(
					"ONTAP timeout, reducing batch size",
					slogx.Err(err),
					slog.Int("oldBatchSize", z.batchSize),
					slog.Int("newBatchSize", z.batchSize-100),
				)
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

		// timestamp for batch instances
		// ignore timestamp from ZAPI which is always integer
		// we want float, since our poll interval can be a float
		ts := float64(time.Now().UnixNano()) / collector2.BILLION

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
					z.Logger.Warn("Instance key has unexpected format", slog.String("key", key))
					continue
				}

				for _, wm := range workloadDetailMetrics {
					mLayer := layer + wm
					if l := curMat.GetMetric(mLayer); l == nil {
						z.Logger.Warn("metric missing in cache", slog.String("layer", mLayer))
						continue
					}
				}
			}

			if key == "" {
				if z.Logger.Enabled(context.Background(), slog.LevelDebug) {
					z.Logger.Debug(
						"Skip instance, key is empty",
						slog.Any("instanceKey", z.instanceKeys),
						slog.String("name", i.GetChildContentS("name")),
						slog.String("uuid", i.GetChildContentS("uuid")),
					)
				}
				continue
			}

			instance := curMat.GetInstance(key)
			if instance == nil {
				z.Logger.Debug("Skip instance key, not found in cache", slog.String("key", key))
				continue
			}

			if z.isPartialAggregation(i) {
				if z.allowPartialAggregation {
					// Partial aggregation detected but allowed processing - mark as complete and exportable
					instance.SetPartial(false)
					instance.SetExportable(true)
				} else {
					// Partial aggregation detected and not allowed processing - mark instance as partial and non-exportable
					instance.SetPartial(true)
					instance.SetExportable(false)
					numPartials++
				}
			} else {
				// Aggregation is complete - mark as complete and exportable
				instance.SetPartial(false)
				instance.SetExportable(true)
			}

			counters := i.GetChildS("counters")
			if counters == nil {
				z.Logger.Debug("Skip instance key, no data counters", slog.String("key", key))
				continue
			}

			// add batch timestamp as custom counter
			timestamp.SetValueFloat64(instance, ts)

			for _, cnt := range counters.GetChildren() {

				name := cnt.GetChildContentS("name")
				value := cnt.GetChildContentS("value")

				// validation
				if name == "" || value == "" {
					// skip counters with empty value or name
					continue
				}

				// ZAPI counter for us is either instance label (string)
				// or numeric metric (scalar or histogram)

				// store as instance label
				if display, has := z.instanceLabels[name]; has {
					instance.SetLabel(display, value)
					continue
				}

				// store as array counter / histogram
				if labels, has := z.histogramLabels[name]; has {

					values := strings.Split(value, ",")

					if len(labels) != len(values) {
						// warn & skip
						z.Logger.Error(
							"Histogram labels don't match parsed values",
							slog.String("labels", name),
							slog.String("value", value),
							slog.Int("instIndex", instIndex),
						)
						continue
					}

					for i, label := range labels {
						if metric := curMat.GetMetric(name + "." + label); metric != nil {
							if err = metric.SetValueString(instance, values[i]); err != nil {
								z.Logger.Error(
									"Set histogram value failed",
									slogx.Err(err),
									slog.String("name", name),
									slog.String("label", label),
									slog.String("value", values[i]),
									slog.Int("instIndex", instIndex),
								)
							} else {
								count++
							}
						} else {
							z.Logger.Warn(
								"Histogram name. Label not in cache",
								slog.String("name", name),
								slog.String("label", label),
								slog.String("value", value),
								slog.Int("instIndex", instIndex),
							)
						}
					}
					continue
				}

				// special case for workload_detail
				if z.Query == objWorkloadDetail || z.Query == objWorkloadDetailVolume {
					for _, wm := range workloadDetailMetrics {
						wMetric := curMat.GetMetric(layer + wm)

						switch {
						case wm == "resource_latency" && (name == "wait_time" || name == "service_time"):
							if err := wMetric.AddValueString(instance, value); err != nil {
								z.Logger.Error(
									"Add resource_latency failed",
									slogx.Err(err),
									slog.String("name", name),
									slog.String("value", value),
									slog.Int("instIndex", instIndex),
								)
							} else {
								count++
							}
							continue
						case wm == "service_time_latency" && name == "service_time":
							if err = wMetric.SetValueString(instance, value); err != nil {
								z.Logger.Error(
									"Add service_time_latency failed",
									slogx.Err(err),
									slog.String("name", name),
									slog.String("value", value),
									slog.Int("instIndex", instIndex),
								)
							} else {
								count++
							}
						case wm == "wait_time_latency" && name == "wait_time":
							if err = wMetric.SetValueString(instance, value); err != nil {
								z.Logger.Error(
									"Add wait_time_latency failed",
									slogx.Err(err),
									slog.String("name", name),
									slog.String("value", value),
									slog.Int("instIndex", instIndex),
								)
							} else {
								count++
							}
						}
					}
					continue
				}

				// store as scalar metric
				if metric := curMat.GetMetric(name); metric != nil {
					if err = metric.SetValueString(instance, value); err != nil {
						z.Logger.Error(
							"Set metric failed",
							slogx.Err(err),
							slog.String("name", name),
							slog.String("value", value),
							slog.Int("instIndex", instIndex),
						)
					} else {
						count++
					}
					continue
				}

				z.Logger.Warn(
					"Counter not in cache",
					slog.Int("instIndex", instIndex),
					slog.String("name", name),
					slog.String("value", value),
				)
			} // end loop over counters
		} // end loop over instances
	} // end batch request

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
		z.Logger.Debug("skip postprocessing until next poll (previous cache empty)")
		z.Matrix[z.Object] = curMat
		z.isCacheEmpty = false
		return nil, nil
	}

	calcStart := time.Now()

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
	if _, err = curMat.Delta(timestampMetricName, prevMat, cachedData, z.allowPartialAggregation, z.Logger); err != nil {
		z.Logger.Error("(timestamp) calculate delta:", slogx.Err(err))
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
		if skips, err = curMat.Delta(key, prevMat, cachedData, z.allowPartialAggregation, z.Logger); err != nil {
			z.Logger.Error("Calculate delta", slogx.Err(err), slog.String("key", key))
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
			if z.Query == objWorkloadDetail || z.Query == objWorkloadDetailVolume {
				// The workload detail generates metrics at the resource level. The 'service_time' and 'wait_time' metrics are used as raw values for these resource-level metrics. Their denominator, 'visits', is not collected; therefore, a check is added here to prevent warnings.
				// There is no need to cook these metrics further.
				if key == "service_time" || key == "wait_time" {
					continue
				}
			}
			z.Logger.Warn(
				"Base counter missing",
				slog.String("key", key),
				slog.String("property", property),
				slog.String("comment", metric.GetComment()),
			)
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
				skips, err = curMat.DivideWithThreshold(key, metric.GetComment(), z.latencyIoReqd, cachedData, prevMat, timestampMetricName, z.Logger)
			} else {
				skips, err = curMat.Divide(key, metric.GetComment())
			}

			if err != nil {
				z.Logger.Error("Division by base", slogx.Err(err), slog.String("key", key))
			}
			totalSkips += skips

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if skips, err = curMat.MultiplyByScalar(key, 100); err != nil {
				z.Logger.Error("Multiply by scalar", slogx.Err(err), slog.String("key", key))
			} else {
				totalSkips += skips
			}
			continue
		}
		z.Logger.Error(
			"Unknown property",
			slogx.Err(err),
			slog.String("key", key),
			slog.String("property", property),
		)
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		if metric.GetProperty() == "rate" {
			if skips, err = curMat.Divide(orderedKeys[i], timestampMetricName); err != nil {
				z.Logger.Error(
					"Calculate rate",
					slogx.Err(err),
					slog.Int("i", i),
					slog.String("key", orderedKeys[i]),
				)
				continue
			}
			totalSkips += skips
		}
	}

	calcD := time.Since(calcStart)

	_ = z.Metadata.LazySetValueInt64("calc_time", "data", calcD.Microseconds())
	_ = z.Metadata.LazySetValueUint64("skips", "data", uint64(totalSkips)) //nolint:gosec

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
func (z *ZapiPerf) getParentOpsCounters(data *matrix.Matrix, keyAttr string) (time.Duration, time.Duration, error) {

	var (
		ops          *matrix.Metric
		object       string
		instanceKeys []string
		apiT, parseT time.Duration
	)

	if z.Query == objWorkloadDetail {
		object = objWorkload
	} else {
		object = objWorkloadVolume
	}

	z.Logger.Debug(
		"starting redundancy poll for ops from parent object",
		slog.String("query", z.Query),
		slog.String("object", object),
	)

	apiT = 0 * time.Second
	parseT = 0 * time.Second

	if ops = data.GetMetric("ops"); ops == nil {
		z.Logger.Error("ops counter not found in cache")
		return apiT, parseT, errs.New(errs.ErrMissingParam, "counter ops")
	}

	instanceKeys = data.GetInstanceKeys()
	slices.Sort(instanceKeys)

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

		z.Logger.Debug(
			"starting batch poll for instances",
			slog.Int("startIndex", startIndex),
			slog.Int("endIndex", endIndex),
		)

		request.PopChildS(keyAttr + "s")
		requestInstances := request.NewChildS(keyAttr+"s", "")
		for _, key := range instanceKeys[startIndex:endIndex] {
			requestInstances.NewChildS(keyAttr, key)
		}

		startIndex = endIndex

		if err := z.Client.BuildRequest(request); err != nil {
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
				if z.Logger.Enabled(context.Background(), slog.LevelDebug) {
					z.Logger.Debug(
						"skip instance",
						slog.Any("key", z.instanceKeys),
						slog.String("name", i.GetChildContentS("name")),
						slog.String("uuid", i.GetChildContentS("uuid")),
					)
				}
				continue
			}

			instance := data.GetInstance(key)
			if instance == nil {
				z.Logger.Warn("skip instance, not found in cache", slog.String("key", key))
				continue
			}

			counters := i.GetChildS("counters")
			if counters == nil {
				z.Logger.Debug("skip instance, no data counters", slog.String("key", key))
				continue
			}

			for _, cnt := range counters.GetChildren() {

				name := cnt.GetChildContentS("name")
				value := cnt.GetChildContentS("value")

				if name == "ops" {
					if err = ops.SetValueString(instance, value); err != nil {
						z.Logger.Error(
							"set metric value",
							slogx.Err(err),
							slog.String("name", name),
							slog.String("value", value),
						)
					} else {
						count++
					}
				} else {
					z.Logger.Error(
						"unrequested metric",
						slog.String("name", name),
					)
				}
			}
		}
	}
	z.Logger.Debug(
		"completed redundant ops poll",
		slog.String("query", z.Query),
		slog.String("object", object),
		slog.Int("count", count),
	)
	return apiT, parseT, nil
}

func (z *ZapiPerf) PollCounter() (map[string]*matrix.Matrix, error) {
	var (
		err                                      error
		request, response, counterList           *node.Node
		oldMetrics, oldLabels, replaced, missing *set.Set
		wanted                                   map[string]string
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
	for key := range z.instanceLabels {
		oldLabels.Add(key)
	}

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

	// build request
	request = node.NewXMLS("perf-object-counter-list-info")
	request.NewChildS("objectname", z.Query)

	if err := z.Client.BuildRequest(request); err != nil {
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
			counter.SetChildContentS("properties", p)
		}

		display, ok := wanted[key]
		// counter not requested
		if !ok {
			continue
		}

		// deprecated and possibly replaced counter
		// if there is no replacement continue instead of skipping
		if counter.GetChildContentS("is-deprecated") == "true" {
			if r := counter.GetChildContentS("replaced-by"); r != "" {
				z.Logger.Info(
					"Replaced deprecated counter",
					slog.String("key", key),
					slog.String("replacement", r),
				)
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
		} else {
			// add counter as numeric metric
			oldMetrics.Remove(key)
			if r := z.addCounter(counter, key, display, true, counters); r != "" {
				_, ok = wanted[r]
				if !ok {
					if z.Query == objWorkloadDetail || z.Query == objWorkloadDetailVolume {
						// It is not needed because 'ops' is used as the denominator in latency calculations.
						if r == "visits" {
							continue
						}
					}
					missing.Add(r) // required base counter, missing in template
				}
			}
		}
	}

	// identify missing counters which do not exist in counter-schema
	for k := range wanted {
		_, ok := counters[k]
		if !ok {
			z.Logger.Warn("Metric not found in counterSchema", slog.String("key", k))
		}

	}

	// second loop for replaced counters
	if replaced.Size() > 0 {
		z.Logger.Debug("attempting to retrieve metadata of replaced counters", slog.Int("size", replaced.Size()))
		for name, counter := range counters {
			if replaced.Has(name) {
				oldMetrics.Remove(name)
				z.Logger.Debug("adding replaced counter", slog.String("name", name))
				if r := z.addCounter(counter, name, name, true, counters); r != "" {
					_, ok := wanted[r]
					if !ok {
						missing.Add(r) // required base counter, missing in template
						z.Logger.Debug(
							"marking as required base counter",
							slog.String("name", name),
							slog.String("r", r),
						)
					}
				}
			}
		}
	}

	// third loop for required base counters, not in template
	if missing.Size() > 0 {
		z.Logger.Debug(
			"attempting to retrieve metadata of missing base counters",
			slog.Int("missing", missing.Size()),
		)
		for name, counter := range counters {
			if missing.Has(name) {
				oldMetrics.Remove(name)
				z.Logger.Debug("adding missing base counter", slog.String("name", name))
				z.addCounter(counter, name, "", false, counters)
			}
		}
	}

	// @TODO check dtype!!!!
	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	if !oldMetrics.Has(timestampMetricName) {
		m, err := mat.NewMetricFloat64(timestampMetricName)
		if err != nil {
			z.Logger.Error("add timestamp metric", slogx.Err(err))
		}
		m.SetProperty("raw")
		m.SetExportable(false)
	}

	// hack for workload objects, @TODO replace with a plugin
	if z.Query == objWorkload || z.Query == objWorkloadDetail || z.Query == objWorkloadVolume || z.Query == objWorkloadDetailVolume {

		// for these two objects, we need to create latency/ops counters for each of the workload layers
		// there original counters will be discarded
		if z.Query == objWorkloadDetail || z.Query == objWorkloadDetailVolume {

			var service, wait, ops *matrix.Metric
			oldMetrics.Remove("service_time")
			oldMetrics.Remove("wait_time")
			oldMetrics.Remove("ops")

			if service = mat.GetMetric("service_time"); service == nil {
				z.Logger.Error("metric [service_time] required to calculate workload missing")
			}

			if wait = mat.GetMetric("wait_time"); wait == nil {
				z.Logger.Error("metric [wait_time] required to calculate workload missing")
			}

			if service == nil || wait == nil {
				return nil, errs.New(errs.ErrMissingParam, "workload metrics")
			}

			if ops = mat.GetMetric("ops"); ops == nil {
				if ops, err = mat.NewMetricFloat64("ops"); err != nil {
					return nil, err
				}
				ops.SetProperty("rate")
			}

			service.SetExportable(false)
			wait.SetExportable(false)

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
		if key != timestampMetricName && !strings.Contains(key, ".") {
			mat.RemoveMetric(key)
			z.Logger.Debug("removed metric", slog.String("key", key))
		}
	}

	for key := range oldLabels.Iter() {
		delete(z.instanceLabels, key)
		z.Logger.Debug("removed label", slog.String("key", key))
	}

	numMetrics := len(mat.GetMetrics())

	// update metadata for collector logs
	_ = z.Metadata.LazySetValueInt64("api_time", "counter", apiD.Microseconds())
	_ = z.Metadata.LazySetValueInt64("parse_time", "counter", time.Since(parseT).Microseconds())
	_ = z.Metadata.LazySetValueUint64("metrics", "counter", uint64(numMetrics))
	_ = z.Metadata.LazySetValueUint64("bytesRx", "counter", z.Client.Metadata.BytesRx)
	_ = z.Metadata.LazySetValueUint64("numCalls", "counter", z.Client.Metadata.NumCalls)

	if numMetrics == 0 {
		return nil, errs.New(errs.ErrNoMetric, "")
	}

	slices.Sort(z.scalarCounters)
	return nil, nil
}

// create new or update existing metric based on Zapi counter metadata
func (z *ZapiPerf) addCounter(counter *node.Node, name, display string, enabled bool, cache map[string]*node.Node) string {

	var (
		property, baseCounter string
		err                   error
	)

	mat := z.Matrix[z.Object]

	p := counter.GetChildContentS("properties")
	switch {
	case strings.Contains(p, "raw"):
		property = "raw"
	case strings.Contains(p, "delta"):
		property = "delta"
	case strings.Contains(p, "rate"):
		property = "rate"
	case strings.Contains(p, "average"):
		property = "average"
	case strings.Contains(p, "percent"):
		property = "percent"
	default:
		z.Logger.Warn(
			"skip counter with unknown property",
			slog.String("name", name),
			slog.String("property", p),
		)
		return ""
	}

	baseCounter = counter.GetChildContentS("base-counter")

	if display == "" {
		display = strings.ReplaceAll(name, "-", "_") // redundant for zapiperf
	}

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
			z.Logger.Warn(
				"skipping of type array",
				slog.String("name", name),
				slog.String("e", e),
			)
			return ""
		}

		if baseCounter != "" {
			if base, ok := cache[baseCounter]; ok {
				if base.GetChildContentS("type") == "array" {
					baseLabels, e = parseHistogramLabels(base)
					if e != "" {
						z.Logger.Warn("skipping",
							slog.String("name", name),
							slog.String("baseCounter", baseCounter),
							slog.String("e", e),
						)
						return ""
					} else if len(baseLabels) != len(labels) {
						z.Logger.Warn(
							"skipping",
							slog.String("name", name),
							slog.String("baseCounter", baseCounter),
						)
						return ""
					}
				}
			} else {
				z.Logger.Warn("skipping", slog.String("name", name), slog.String("baseCounter", baseCounter))
				return ""
			}
		}

		// ONTAP does not have a `type` for histogram. Harvest tests the `desc` field to determine
		// if a counter is a histogram
		isHistogram = false
		if len(labels) > 0 && strings.Contains(description, "histogram") {
			key := name + ".bucket"
			histogramMetric = mat.GetMetric(key)
			if histogramMetric == nil {
				histogramMetric, err = mat.NewMetricFloat64(key, display)
				if err != nil {
					z.Logger.Error(
						"unable to create histogram metric",
						slogx.Err(err),
						slog.String("key", key),
					)
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
			m = mat.GetMetric(key)
			if m == nil {
				m, err = mat.NewMetricFloat64(key, display)
				if err != nil {
					z.Logger.Error(
						"unable to create array metric",
						slogx.Err(err),
						slog.String("key", key),
					)
					return ""
				}
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
		m := mat.GetMetric(name)
		if m == nil {
			m, err = mat.NewMetricFloat64(name, display)
			if err != nil {
				z.Logger.Error(
					"unable to create scalar metric",
					slogx.Err(err),
					slog.String("name", name),
				)
				return ""
			}
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
		newSize                           int
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

	nameAttr = "name"
	uuidAttr = "uuid"
	keyAttrs = z.instanceKeys

	// hack for workload objects: get instances from Zapi
	switch {
	case z.Query == objWorkload || z.Query == objWorkloadDetail || z.Query == objWorkloadVolume || z.Query == objWorkloadDetailVolume:
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
	case z.Client.IsClustered():
		request = node.NewXMLS("perf-object-instance-list-info-iter")
		request.NewChildS("objectname", z.Query)
		if z.filter != "" {
			request.NewChildS("filter-data", z.filter)
		}
		instancesAttr = "attributes-list"
	default:
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

		z.pollInstanceCalls++
		if z.pollInstanceCalls >= z.recordsToSave/3 {
			z.pollInstanceCalls = 0
		}

		var headers map[string]string

		poller, err := conf.PollerNamed(z.Options.Poller)
		if err != nil {
			slog.Error("failed to find poller", slogx.Err(err), slog.String("poller", z.Options.Poller))
		}

		if poller.IsRecording() {
			headers = map[string]string{
				"From": strconv.Itoa(z.pollInstanceCalls),
			}
		}

		responseData, err := z.Client.InvokeBatchRequest(request, batchTag, z.testFilePath, headers)

		if err != nil {
			if errors.Is(err, errs.ErrAPIRequestRejected) {
				z.Logger.Info(
					err.Error(),
					slog.String("request", request.GetNameS()),
					slog.String("batchTag", batchTag),
				)
			} else {
				z.Logger.Error(
					"InvokeBatchRequest failed",
					slogx.Err(err),
					slog.String("request", request.GetNameS()),
					slog.String("batchTag", batchTag),
				)
			}
			apiD += time.Since(apiT)
			break
		}

		results = responseData.Result
		batchTag = responseData.Tag
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
				if z.Logger.Enabled(context.Background(), slog.LevelDebug) {
					z.Logger.Debug(
						"skip instance",
						slog.Any("key", z.instanceKeys),
						slog.String("name", name),
						slog.String("uuid", uuid),
					)
				}
			} else if oldInstances.Has(key) {
				// instance already in cache
				oldInstances.Remove(key)
				instance := mat.GetInstance(key)
				z.updateQosLabels(i, instance)
				continue
			} else if instance, err := mat.NewInstance(key); err != nil {
				z.Logger.Error("add instance", slogx.Err(err))
			} else {
				z.updateQosLabels(i, instance)
			}
		}
		parseD += time.Since(parseT)
	}

	for key := range oldInstances.Iter() {
		mat.RemoveInstance(key)
		z.Logger.Debug("removed instance", slog.String("key", key))
	}

	newSize = len(mat.GetInstances())

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

func (z *ZapiPerf) updateQosLabels(qos *node.Node, instance *matrix.Instance) {
	if z.Query == objWorkload || z.Query == objWorkloadDetail || z.Query == objWorkloadVolume || z.Query == objWorkloadDetailVolume {
		for label, display := range z.qosLabels {
			if value := qos.GetChildContentS(label); value != "" {
				instance.SetLabel(display, value)
			}
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
			z.Logger.Warn("skip instance, missing key", slog.String("key", k))
		}
	}
	return strings.Join(values, keyToken)
}

// Interface guards
var (
	_ collector.Collector = (*ZapiPerf)(nil)
)
