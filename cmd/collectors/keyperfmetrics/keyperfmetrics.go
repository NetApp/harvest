package keyperfmetrics

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

const (
	latencyIoReqd       = 10
	timestampMetricName = "statistics.timestamp"
)

type KeyPerfMetrics struct {
	*rest.Rest // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	perfProp   *perfProp
}

type counter struct {
	name        string
	counterType string
	unit        string
	denominator string
}

type perfProp struct {
	isCacheEmpty  bool
	counterInfo   map[string]*counter
	latencyIoReqd int
}

func init() {
	plugin.RegisterModule(&KeyPerfMetrics{})
}

func (kp *KeyPerfMetrics) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.keyperfmetrics",
		New: func() plugin.Module { return new(KeyPerfMetrics) },
	}
}

func (kp *KeyPerfMetrics) Init(a *collector.AbstractCollector) error {

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

	kp.Logger.Debug().
		Int("numMetrics", len(kp.Prop.Metrics)).
		Str("timeout", kp.Client.Timeout.String()).
		Msg("initialized cache")
	return nil
}

func (kp *KeyPerfMetrics) InitMatrix() error {
	mat := kp.Matrix[kp.Object]
	// init perf properties
	kp.perfProp.latencyIoReqd = kp.loadParamInt("latency_io_reqd", latencyIoReqd)
	kp.perfProp.isCacheEmpty = true
	// overwrite from abstract collector
	mat.Object = kp.Prop.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", kp.Client.Cluster().Name)
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
func (kp *KeyPerfMetrics) loadParamInt(name string, defaultValue int) int {

	var (
		x string
		n int
		e error
	)

	if x = kp.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			kp.Logger.Debug().Msgf("using %s = [%d]", name, n)
			return n
		}
		kp.Logger.Warn().Msgf("invalid parameter %s = [%s] (expected integer)", name, x)
	}

	kp.Logger.Debug().Str("name", name).Str("defaultValue", strconv.Itoa(defaultValue)).Msg("using values")
	return defaultValue
}

func (kp *KeyPerfMetrics) buildCounters() {
	for k := range kp.Prop.Metrics {
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
			case strings.Contains(k, timestampMetricName):
				ctr = &counter{
					name:        k,
					counterType: "delta",
					unit:        "sec",
				}
			}

			if ctr != nil {
				kp.perfProp.counterInfo[k] = ctr
			}
		}
	}
}

func (kp *KeyPerfMetrics) PollData() (map[string]*matrix.Matrix, error) {
	var (
		err         error
		perfRecords []gjson.Result
		startTime   time.Time
	)
	startTime = time.Now()
	kp.Client.Metadata.Reset()

	href := kp.Prop.Href
	kp.Logger.Debug().Str("href", href).Send()
	if href == "" {
		return nil, errs.New(errs.ErrConfig, "empty url")
	}

	perfRecords, err = kp.GetRestData(href)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch href=%s %w", href, err)
	}

	return kp.pollData(startTime, perfRecords, func(e *rest.EndPoint) ([]gjson.Result, time.Duration, error) {
		return kp.ProcessEndPoint(e)
	})
}

// validateMatrix ensures that the previous matrix (prevMat) contains all the metrics present in the current matrix (curMat).
// This is crucial for performing accurate comparisons and calculations between the two matrices, especially in scenarios where
// the current matrix may have additional metrics that are not present in the previous matrix, such as after an ONTAP upgrade.
//
// The function iterates over all the metrics in curMat and checks if each metric exists in prevMat. If a metric from curMat
// does not exist in prevMat, it is created in prevMat as a new float64 metric. This prevents potential panics or errors
// when attempting to perform calculations with metrics that are missing in prevMat.
func (kp *KeyPerfMetrics) validateMatrix(prevMat *matrix.Matrix, curMat *matrix.Matrix) error {
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

func (kp *KeyPerfMetrics) pollData(
	startTime time.Time,
	perfRecords []gjson.Result,
	endpointFunc func(e *rest.EndPoint) ([]gjson.Result, time.Duration, error),
) (map[string]*matrix.Matrix, error) {
	var (
		count        uint64
		apiD, parseD time.Duration
		err          error
		skips        int
		numPartials  uint64
		instIndex    int
		prevMat      *matrix.Matrix
		curMat       *matrix.Matrix
	)

	prevMat = kp.Matrix[kp.Object]

	// clone matrix without numeric data
	curMat = prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	apiD = time.Since(startTime)

	startTime = time.Now()

	if len(perfRecords) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+kp.Object+" instances on cluster")
	}
	count, numPartials = kp.HandleResults(curMat, perfRecords, kp.Prop, false)

	// process endpoints
	eCount, endpointAPID := kp.ProcessEndPoints(curMat, endpointFunc)
	count += eCount

	parseD = time.Since(startTime)
	_ = kp.Metadata.LazySetValueInt64("api_time", "data", (apiD + endpointAPID).Microseconds())
	_ = kp.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = kp.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = kp.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = kp.Metadata.LazySetValueUint64("bytesRx", "data", kp.Client.Metadata.BytesRx)
	_ = kp.Metadata.LazySetValueUint64("numCalls", "data", kp.Client.Metadata.NumCalls)
	_ = kp.Metadata.LazySetValueUint64("numPartials", "data", numPartials)

	kp.AddCollectCount(count)

	// skip calculating from delta if no data from previous poll
	if kp.perfProp.isCacheEmpty {
		kp.Logger.Debug().Msg("skip postprocessing until next poll (previous cache empty)")
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

	for key, metric := range curMat.GetMetrics() {
		counter := kp.counterLookup(key)
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
			kp.Logger.Warn().Str("counter", metric.GetName()).Msg("Counter is missing or unable to parse")
		}
	}

	timestamp := curMat.GetMetric(timestampMetricName)
	if timestamp != nil {
		timestamp.SetExportable(false)
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
		counter := kp.counterLookup(key)
		if counter == nil {
			kp.Logger.Error().Err(err).Str("counter", metric.GetName()).Msg("Missing counter:")
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
		if skips, err = curMat.Delta(key, prevMat, kp.Logger); err != nil {
			kp.Logger.Error().Err(err).Str("key", key).Msg("Calculate delta")
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
			kp.Logger.Warn().
				Str("key", key).
				Str("property", property).
				Str("denominator", counter.denominator).
				Int("instIndex", instIndex).
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
				skips, err = curMat.DivideWithThreshold(key, counter.denominator, kp.perfProp.latencyIoReqd, cachedData, prevMat, timestampMetricName, kp.Logger)
			} else {
				skips, err = curMat.Divide(key, counter.denominator)
			}

			if err != nil {
				kp.Logger.Error().Err(err).Str("key", key).Msg("Division by base")
				continue
			}
			totalSkips += skips

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if skips, err = curMat.MultiplyByScalar(key, 100); err != nil {
				kp.Logger.Error().Err(err).Str("key", key).Msg("Multiply by scalar")
			} else {
				totalSkips += skips
			}
			continue
		}
		// If we reach here then one of the earlier clauses should have executed `continue` statement
		kp.Logger.Error().Err(err).
			Str("key", key).
			Str("property", property).
			Int("instIndex", instIndex).
			Msg("Unknown property")
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := kp.counterLookup(key)
		if counter != nil {
			property := counter.counterType
			if property == "rate" {
				if skips, err = curMat.Divide(orderedKeys[i], timestampMetricName); err != nil {
					kp.Logger.Error().Err(err).
						Int("i", i).
						Str("metric", metric.GetName()).
						Str("key", orderedKeys[i]).
						Int("instIndex", instIndex).
						Msg("Calculate rate")
					continue
				}
				totalSkips += skips
			}
		} else {
			kp.Logger.Warn().Str("counter", metric.GetName()).Msg("Counter is missing or unable to parse ")
			continue
		}
	}

	calcD := time.Since(calcStart)
	_ = kp.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = kp.Metadata.LazySetValueInt64("calc_time", "data", calcD.Microseconds())
	_ = kp.Metadata.LazySetValueUint64("skips", "data", uint64(totalSkips))

	// store cache for next poll
	kp.Matrix[kp.Object] = cachedData

	newDataMap := make(map[string]*matrix.Matrix)
	newDataMap[kp.Object] = curMat
	return newDataMap, nil
}

func (kp *KeyPerfMetrics) counterLookup(metricKey string) *counter {
	c := kp.perfProp.counterInfo[metricKey]
	return c
}

// Interface guards
var (
	_ collector.Collector = (*KeyPerfMetrics)(nil)
)
