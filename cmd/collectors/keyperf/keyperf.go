package keyperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/keyperf/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/volumetopmetrics"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	rest2 "github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	latencyIoReqd = 10
)

type KeyPerf struct {
	*rest.Rest    // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	perfProp      *perfProp
	pollDataCalls int
	recordsToSave int // Number of records to save when using the recorder
}

type counter struct {
	name        string
	counterType string
	unit        string
	denominator string
}

type perfProp struct {
	isCacheEmpty        bool
	counterInfo         map[string]*counter
	latencyIoReqd       int
	timestampMetricName string
}

func init() {
	plugin.RegisterModule(&KeyPerf{})
}

func (kp *KeyPerf) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.keyperf",
		New: func() plugin.Module { return new(KeyPerf) },
	}
}

func (kp *KeyPerf) Init(a *collector.AbstractCollector) error {

	var err error

	kp.Rest = &rest.Rest{AbstractCollector: a}

	kp.perfProp = &perfProp{}

	kp.InitProp()

	kp.perfProp.counterInfo = make(map[string]*counter)

	if err := kp.InitClient(); err != nil {
		return err
	}

	if kp.Prop.TemplatePath, err = kp.LoadTemplate(); err != nil {
		return err
	}

	kp.InitVars(a.Params)

	if err := kp.InitEndPoints(); err != nil {
		return err
	}

	if err := collector.Init(kp); err != nil {
		return err
	}

	if err := kp.InitCache(); err != nil {
		return err
	}

	if err := kp.InitMatrix(); err != nil {
		return err
	}

	kp.buildCounters()

	kp.recordsToSave = collector.RecordKeepLast(kp.Params, kp.Logger)

	kp.Logger.Debug(
		"initialized cache",
		slog.Int("numMetrics", len(kp.Prop.Metrics)),
		slog.String("timeout", kp.Client.Timeout.String()),
	)
	return nil
}

func (kp *KeyPerf) InitMatrix() error {
	mat := kp.Matrix[kp.Object]
	// init perf properties
	kp.perfProp.latencyIoReqd = kp.loadParamInt("latency_io_reqd", latencyIoReqd)
	kp.perfProp.isCacheEmpty = true
	// overwrite from abstract collector
	mat.Object = kp.Prop.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", kp.Remote.Name)
	if kp.Params.HasChildS("labels") {
		for _, l := range kp.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	// Add metadata metric for skips/numPartials
	_, _ = kp.Metadata.NewMetricUint64("skips")
	_, _ = kp.Metadata.NewMetricUint64("numPartials")
	return nil
}

// load an int parameter or use defaultValue
func (kp *KeyPerf) loadParamInt(name string, defaultValue int) int {

	var (
		x string
		n int
		e error
	)

	if x = kp.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			kp.Logger.Debug("", slog.String("name", name), slog.Int("n", n))
			return n
		}
		kp.Logger.Warn("invalid parameter", slog.String("parameter", name), slog.String("x", x))
	}

	kp.Logger.Debug("using values", slog.String("name", name), slog.Int("defaultValue", defaultValue))
	return defaultValue
}

func (kp *KeyPerf) LoadPlugin(kind string, p *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Volume":
		return volume.New(p)
	case "VolumeTopClients":
		return volumetopmetrics.New(p)
	default:
		kp.Logger.Info("no KeyPerf plugin found", slog.String("kind", kind))
	}
	return nil
}

func findStaticCounterDefPath() string {
	prodPath := "conf/keyperf/static_counter_definitions.yaml"
	testPath := "../../../conf/keyperf/static_counter_definitions.yaml"

	if _, err := os.Stat(prodPath); err == nil {
		return prodPath
	}
	return testPath
}

func (kp *KeyPerf) buildCounters() {
	staticCounterDef, err := LoadStaticCounterDefinitions(kp.Prop.Object, findStaticCounterDefPath(), kp.Logger)
	if err != nil {
		// It's acceptable to continue even if there are errors, as the remaining counters will still be processed.
		// Any counters that require counter metadata will be skipped.
		kp.Logger.Error("Failed to load static counter definitions", slogx.Err(err))
	}

	// Check if the statistics.timestamp metric exists; if not, create it
	_, exists := kp.Prop.Metrics["statistics.timestamp"]
	if !exists {
		kp.Prop.Metrics["statistics.timestamp"] = &rest.Metric{
			Label:      "timestamp",
			Name:       "statistics.timestamp",
			Exportable: true,
		}
	}

	// handle statistics.timestamp for endpoints
	for _, endpoint := range kp.Endpoints {
		eProp := endpoint.Prop
		_, exists = eProp.Metrics["statistics.timestamp"]
		if !exists {
			eProp.Metrics["statistics.timestamp"] = &rest.Metric{
				Label:      "timestamp",
				Name:       "statistics.timestamp",
				Exportable: true,
			}
		}

		for k, v := range eProp.Metrics {
			if _, exists = kp.Prop.Metrics[k]; !exists {
				kp.Prop.Metrics[k] = v
			}
		}
	}

	for k, v := range kp.Prop.Metrics {
		if _, exists := kp.perfProp.counterInfo[k]; !exists {
			var ctr *counter

			switch {
			case strings.Contains(k, "latency"):
				ctr = &counter{
					name:        k,
					counterType: "average",
					unit:        "microsec",
					denominator: strings.Replace(k, "latency", "iops", 1),
				}
			case strings.Contains(k, "iops"):
				ctr = &counter{
					name:        k,
					counterType: "rate",
					unit:        "per_sec",
				}
			case strings.Contains(k, "throughput"):
				ctr = &counter{
					name:        k,
					counterType: "rate",
					unit:        "b_per_sec",
				}
			case strings.Contains(k, "timestamp"):
				kp.perfProp.timestampMetricName = k
				ctr = &counter{
					name:        k,
					counterType: "delta",
					unit:        "sec",
				}
			default:
				// look up metric in staticCounterDef
				if counterDef, exists := staticCounterDef.CounterDefinitions[v.Name]; exists {
					ctr = &counter{
						name:        k,
						counterType: counterDef.Type,
						denominator: counterDef.BaseCounter,
					}
					if counterDef.BaseCounter != "" {
						// Ensure denominator exists in counterInfo
						if _, denomExists := kp.perfProp.counterInfo[counterDef.BaseCounter]; !denomExists {
							var baseCounterType string
							if baseCounterDef, baseCounterExists := staticCounterDef.CounterDefinitions[counterDef.BaseCounter]; baseCounterExists {
								baseCounterType = baseCounterDef.Type
							}
							if baseCounterType != "" {
								kp.perfProp.counterInfo[counterDef.BaseCounter] = &counter{
									name:        counterDef.BaseCounter,
									counterType: staticCounterDef.CounterDefinitions[counterDef.BaseCounter].Type,
								}
								if _, dExists := kp.Prop.Metrics[counterDef.BaseCounter]; !dExists {
									m := &rest.Metric{Label: "", Name: counterDef.BaseCounter, MetricType: "", Exportable: false}
									kp.Prop.Metrics[counterDef.BaseCounter] = m
								}
							}
						}
					}
				} else {
					slog.Warn("Skipping metric due to unknown metricType", slog.String("name", k), slog.String("metricType", v.MetricType))
				}
			}

			if ctr != nil {
				kp.perfProp.counterInfo[k] = ctr
			}
		}
	}
}

func (kp *KeyPerf) PollData() (map[string]*matrix.Matrix, error) {
	var (
		apiD, parseD time.Duration
		metricCount  uint64
		numPartials  uint64
		startTime    time.Time
		prevMat      *matrix.Matrix
		curMat       *matrix.Matrix
	)

	startTime = time.Now()
	kp.Client.Metadata.Reset()

	href := kp.Prop.Href
	kp.Logger.Debug("fetching href", slog.String("href", href))
	if href == "" {
		return nil, errs.New(errs.ErrConfig, "empty url")
	}

	kp.pollDataCalls++
	if kp.pollDataCalls >= kp.recordsToSave {
		kp.pollDataCalls = 0
	}

	var headers map[string]string

	poller, err := conf.PollerNamed(kp.Options.Poller)
	if err != nil {
		slog.Error("failed to find poller", slogx.Err(err), slog.String("poller", kp.Options.Poller))
	}

	if poller.IsRecording() {
		headers = map[string]string{
			"From": strconv.Itoa(kp.pollDataCalls),
		}
	}

	// Track old instances before processing batches
	oldInstances := set.New()
	for key := range kp.Matrix[kp.Object].GetInstances() {
		oldInstances.Add(key)
	}

	prevMat = kp.Matrix[kp.Object]
	// clone matrix without numeric data
	curMat = prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	processBatch := func(perfRecords []gjson.Result, timestamp int64) error {
		if len(perfRecords) == 0 {
			return nil
		}

		// Process the current batch of records
		count, np, batchParseD := kp.processPerfRecords(perfRecords, curMat, oldInstances, timestamp)
		numPartials += np
		metricCount += count
		parseD += batchParseD
		return nil
	}

	if err := rest2.FetchAllStream(kp.Client, kp.Prop.Href, processBatch, headers); err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	apiD += time.Since(startTime)

	if err != nil {
		return nil, err
	}

	// Process endpoints after all batches have been processed
	eCount, endpointAPID := kp.ProcessEndPoints(curMat, kp.ProcessEndPoint, oldInstances)
	metricCount += eCount
	apiD += endpointAPID

	// Remove old instances that are not found in new instances
	for key := range oldInstances.Iter() {
		curMat.RemoveInstance(key)
	}

	_ = kp.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = kp.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = kp.Metadata.LazySetValueUint64("metrics", "data", metricCount)
	_ = kp.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = kp.Metadata.LazySetValueUint64("bytesRx", "data", kp.Client.Metadata.BytesRx)
	_ = kp.Metadata.LazySetValueUint64("numCalls", "data", kp.Client.Metadata.NumCalls)
	_ = kp.Metadata.LazySetValueUint64("numPartials", "data", numPartials)

	kp.AddCollectCount(metricCount)

	return kp.cookCounters(curMat, prevMat)
}

func (kp *KeyPerf) processPerfRecords(perfRecords []gjson.Result, curMat *matrix.Matrix, oldInstances *set.Set, timestamp int64) (uint64, uint64, time.Duration) {
	var (
		count       uint64
		parseD      time.Duration
		numPartials uint64
	)
	startTime := time.Now()

	count, numPartials = kp.HandleResults(curMat, perfRecords, kp.Prop, false, oldInstances, timestamp)

	parseD = time.Since(startTime)
	return count, numPartials, parseD
}

// validateMatrix ensures that the previous matrix (prevMat) contains all the metrics present in the current matrix (curMat).
// This is crucial for performing accurate comparisons and calculations between the two matrices, especially in scenarios where
// the current matrix may have additional metrics that are not present in the previous matrix, such as after an ONTAP upgrade.
//
// The function iterates over all the metrics in curMat and checks if each metric exists in prevMat. If a metric from curMat
// does not exist in prevMat, it is created in prevMat as a new float64 metric. This prevents potential panics or errors
// when attempting to perform calculations with metrics that are missing in prevMat.
func (kp *KeyPerf) validateMatrix(prevMat *matrix.Matrix, curMat *matrix.Matrix) error {
	var err error
	for k := range curMat.GetMetrics() {
		if prevMat.GetMetric(k) == nil {
			_, err = prevMat.NewMetricFloat64(k)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (kp *KeyPerf) cookCounters(curMat *matrix.Matrix, prevMat *matrix.Matrix) (map[string]*matrix.Matrix, error) {
	var (
		err   error
		skips int
	)
	// skip calculating from delta if no data from previous poll
	if kp.perfProp.isCacheEmpty {
		kp.Logger.Debug("skip postprocessing until next poll (previous cache empty)")
		kp.Matrix[kp.Object] = curMat
		kp.perfProp.isCacheEmpty = false
		return nil, nil
	}

	calcStart := time.Now()

	// cache raw data for next poll
	cachedData := curMat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true, PartialInstances: true})

	orderedNonDenominatorMetrics := make([]*matrix.Metric, 0, len(curMat.GetMetrics()))
	orderedNonDenominatorKeys := make([]string, 0, len(orderedNonDenominatorMetrics))

	orderedDenominatorMetrics := make([]*matrix.Metric, 0, len(curMat.GetMetrics()))
	orderedDenominatorKeys := make([]string, 0, len(orderedDenominatorMetrics))

	counterMap := kp.perfProp.counterInfo

	for key, metric := range curMat.GetMetrics() {
		counter := counterMap[key]
		if counter != nil {
			if counter.denominator == "" {
				// does not require base counter
				orderedNonDenominatorMetrics = append(orderedNonDenominatorMetrics, metric)
				orderedNonDenominatorKeys = append(orderedNonDenominatorKeys, key)
			} else {
				// does require base counter
				orderedDenominatorMetrics = append(orderedDenominatorMetrics, metric)
				orderedDenominatorKeys = append(orderedDenominatorKeys, key)
			}
		} else {
			kp.Logger.Error("Counter is missing or unable to parse", slog.String("counter", metric.GetName()))
			metric.SetExportable(false)
		}
	}

	timestamp := curMat.GetMetric(kp.perfProp.timestampMetricName)
	if timestamp != nil {
		timestamp.SetExportable(false)
	} else {
		return nil, errs.New(errs.ErrConfig, "missing timestamp metric")
	}
	err = kp.validateMatrix(prevMat, curMat)
	if err != nil {
		return nil, err
	}

	// order metrics, such that those requiring base counters are processed last
	orderedMetrics := orderedNonDenominatorMetrics
	orderedMetrics = append(orderedMetrics, orderedDenominatorMetrics...)
	orderedKeys := orderedNonDenominatorKeys
	orderedKeys = append(orderedKeys, orderedDenominatorKeys...)

	var base *matrix.Metric
	var totalSkips int

	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := counterMap[key]
		if counter == nil {
			kp.Logger.Error("Missing counter", slogx.Err(err), slog.String("counter", metric.GetName()))
			continue
		}
		property := counter.counterType
		// used in aggregator plugin
		metric.SetProperty(property)
		// used in volume.go plugin
		metric.SetComment(counter.denominator)

		// raw/string - submit without post-processing
		if property == "raw" || property == "string" {
			continue
		}

		// all other properties - first calculate delta
		if skips, err = curMat.Delta(key, prevMat, cachedData, kp.AllowPartialAggregation, kp.Logger); err != nil {
			kp.Logger.Error("Calculate delta", slogx.Err(err), slog.String("key", key))
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
		if base = curMat.GetMetric(counter.denominator); base == nil {
			kp.Logger.Warn(
				"Base counter missing",
				slog.String("key", key),
				slog.String("property", property),
				slog.String("denominator", counter.denominator),
			)
			skips = curMat.Skip(key)
			totalSkips += skips
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
				skips, err = curMat.DivideWithThreshold(key, counter.denominator, kp.perfProp.latencyIoReqd, cachedData, prevMat, kp.perfProp.timestampMetricName, kp.Logger)
			} else {
				skips, err = curMat.Divide(key, counter.denominator)
			}

			if err != nil {
				kp.Logger.Error("Division by base", slogx.Err(err), slog.String("key", key))
				continue
			}
			totalSkips += skips

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if skips, err = curMat.MultiplyByScalar(key, 100); err != nil {
				kp.Logger.Error("Multiply by scalar", slogx.Err(err), slog.String("key", key))
			} else {
				totalSkips += skips
			}
			continue
		}
		// If we reach here, then one of the earlier clauses should have executed `continue` statement
		kp.Logger.Error(
			"Unknown property",
			slog.String("key", key),
			slog.String("property", property),
		)
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := counterMap[key]
		if counter == nil {
			kp.Logger.Warn("Counter is missing or unable to parse", slog.String("counter", metric.GetName()))
			continue
		}
		property := counter.counterType
		if property == "rate" {
			if skips, err = curMat.Divide(orderedKeys[i], kp.perfProp.timestampMetricName); err != nil {
				kp.Logger.Error(
					"Calculate rate",
					slogx.Err(err),
					slog.Int("i", i),
					slog.String("key", key),
				)
				continue
			}
			totalSkips += skips
		}
	}

	calcD := time.Since(calcStart)
	_ = kp.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = kp.Metadata.LazySetValueInt64("calc_time", "data", calcD.Microseconds())
	_ = kp.Metadata.LazySetValueUint64("skips", "data", uint64(totalSkips)) //nolint:gosec

	// store cache for next poll
	kp.Matrix[kp.Object] = cachedData

	newDataMap := make(map[string]*matrix.Matrix)
	newDataMap[kp.Object] = curMat
	return newDataMap, nil
}

// Interface guards
var (
	_ collector.Collector = (*KeyPerf)(nil)
)
