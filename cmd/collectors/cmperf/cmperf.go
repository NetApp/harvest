package cmperf

import (
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/plugins/disk"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/plugins/fabricpool"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/plugins/fcp"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/plugins/fcvi"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/plugins/flexcache"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/plugins/headroom"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/plugins/nic"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/plugins/vscan"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

const (
	latencyIoReqd       = 0
	arrayKeyToken       = "#"
	timestampMetricName = "timestamp"
)

var (
	qosQuery              = "api/cluster/counter/tables/qos"
	qosVolumeQuery        = "api/cluster/counter/tables/qos_volume"
	qosDetailQuery        = "api/cluster/counter/tables/qos_detail"
	qosDetailVolumeQuery  = "api/cluster/counter/tables/qos_detail_volume"
	workloadDetailMetrics = []string{"resource_latency"}
)

var qosQueries = map[string]string{
	qosQuery:       qosQuery,
	qosVolumeQuery: qosVolumeQuery,
}
var qosDetailQueries = map[string]string{
	qosDetailQuery:       qosDetailQuery,
	qosDetailVolumeQuery: qosDetailVolumeQuery,
}

type CmPerf struct {
	*rest2.Rest         // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	perfProp            *perfProp
	archivedMetrics     map[string]*rest2.Metric // Keeps metric definitions that are not found in the counter schema. These metrics may be available in future ONTAP versions.
	hasInstanceSchedule bool
	recordsToSave       int // Number of records to save when using the recorder
}

type counter struct {
	counterType string
	denominator string
}

type perfProp struct {
	isCacheEmpty        bool
	counterInfo         map[string]*counter
	latencyIoReqd       int
	qosLabels           map[string]string
	disableConstituents bool
}

func init() {
	plugin.RegisterModule(&CmPerf{})
}

func (r *CmPerf) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.cmperf",
		New: func() plugin.Module { return new(CmPerf) },
	}
}

func (r *CmPerf) Init(a *collector.AbstractCollector) error {

	var err error

	r.Rest = &rest2.Rest{AbstractCollector: a}

	r.perfProp = &perfProp{}

	r.InitProp()

	r.perfProp.counterInfo = make(map[string]*counter)
	r.archivedMetrics = make(map[string]*rest2.Metric)

	if err := r.InitClient(); err != nil {
		return err
	}

	if r.Prop.TemplatePath, err = r.LoadTemplate(); err != nil {
		return err
	}

	r.InitVars(a.Params)

	if err := collector.Init(r); err != nil {
		return err
	}

	if err := r.InitCache(); err != nil {
		return err
	}

	if err := r.InitMatrix(); err != nil {
		return err
	}

	if err := r.InitQOS(); err != nil {
		return err
	}

	r.InitSchedule()

	r.recordsToSave = collector.RecordKeepLast(r.Params, r.Logger)

	r.Logger.Debug(
		"initialized cache",
		slog.Int("numMetrics", len(r.Prop.Metrics)),
		slog.String("timeout", r.Client.GetTimeout().String()),
	)

	return nil
}

func (r *CmPerf) InitQOS() error {
	if isWorkloadObject(r.Prop.Query) || isWorkloadDetailObject(r.Prop.Query) {
		qosLabels := r.Params.GetChildS("qos_labels")
		if qosLabels == nil {
			return errs.New(errs.ErrMissingParam, "qos_labels")
		}
		r.perfProp.qosLabels = make(map[string]string)
		for _, label := range qosLabels.GetAllChildContentS() {

			display := strings.ReplaceAll(label, "-", "_")
			before, after, found := strings.Cut(label, "=>")
			if found {
				label = strings.TrimSpace(before)
				display = strings.TrimSpace(after)
			}
			r.perfProp.qosLabels[label] = display
		}
	}
	if counters := r.Params.GetChildS("counters"); counters != nil {
		refine := counters.GetChildS("refine")
		if refine != nil {
			withConstituents := refine.GetChildContentS("with_constituents")
			if withConstituents == "false" {
				r.perfProp.disableConstituents = true
			}
			withServiceLatency := refine.GetChildContentS("with_service_latency")
			if withServiceLatency != "false" {
				workloadDetailMetrics = append(workloadDetailMetrics, "service_time_latency")
			}
		}
	}
	return nil
}

func (r *CmPerf) InitMatrix() error {
	mat := r.Matrix[r.Object]
	// init perf properties
	r.perfProp.latencyIoReqd = r.loadParamInt("latency_io_reqd", latencyIoReqd)
	r.perfProp.isCacheEmpty = true
	// overwrite from abstract collector
	mat.Object = r.Prop.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", r.Remote.Name)
	if r.Params.HasChildS("labels") {
		for _, l := range r.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	// Add metadata metric for skips/numPartials
	_, _ = r.Metadata.NewMetricUint64("skips")
	_, _ = r.Metadata.NewMetricUint64("numPartials")
	return nil
}

// load an int parameter or use defaultValue
func (r *CmPerf) loadParamInt(name string, defaultValue int) int {

	var (
		x string
		n int
		e error
	)

	if x = r.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			r.Logger.Debug("using",
				slog.String("name", name),
				slog.Int("value", n),
			)
			return n
		}
		r.Logger.Warn("invalid parameter (expected integer)", slog.String("name", name), slog.String("value", x))
	}

	r.Logger.Debug("using", slog.String("name", name), slog.Int("defaultValue", defaultValue))
	return defaultValue
}

func (r *CmPerf) PollCounter() (map[string]*matrix.Matrix, error) {

	mat := r.Matrix[r.Object]

	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	if mat.GetMetric(timestampMetricName) == nil {
		m, err := mat.NewMetricFloat64(timestampMetricName)
		if err != nil {
			r.Logger.Error("add timestamp metric", slogx.Err(err))
		}
		m.SetProperty("raw")
		m.SetExportable(false)
	}

	return nil, nil
}

// GetOverride override counter property
func (r *CmPerf) GetOverride(counter string) string {
	if o := r.Params.GetChildS("override"); o != nil {
		return o.GetChildContentS(counter)
	}
	return ""
}

func (r *CmPerf) PollData() (map[string]*matrix.Matrix, error) {
	var (
		apiD, parseD time.Duration
		metricCount  uint64
		numPartials  uint64
		startTime    time.Time
		prevMat      *matrix.Matrix
		curMat       *matrix.Matrix
	)

	timestamp := r.Matrix[r.Object].GetMetric(timestampMetricName)
	if timestamp == nil {
		return nil, errs.New(errs.ErrConfig, "missing timestamp metric")
	}

	startTime = time.Now()
	r.Client.Metadata.Reset()
	prevMat = r.Matrix[r.Object]

	// clone matrix without numeric data
	curMat = prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	apiD += time.Since(startTime)

	_ = r.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = r.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = r.Metadata.LazySetValueUint64("metrics", "data", metricCount)
	_ = r.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = r.Metadata.LazySetValueUint64("bytesRx", "data", r.Client.Metadata.BytesRx)
	_ = r.Metadata.LazySetValueUint64("numCalls", "data", r.Client.Metadata.NumCalls)
	_ = r.Metadata.LazySetValueUint64("numPartials", "data", numPartials)
	r.AddCollectCount(metricCount)

	return r.cookCounters(curMat, prevMat)
}

func (r *CmPerf) cookCounters(curMat *matrix.Matrix, prevMat *matrix.Matrix) (map[string]*matrix.Matrix, error) {
	var (
		err   error
		skips int
	)

	// skip calculating from delta if no data from previous poll
	if r.perfProp.isCacheEmpty {
		r.Logger.Debug("skip postprocessing until next poll (previous cache empty)")
		r.Matrix[r.Object] = curMat
		r.perfProp.isCacheEmpty = false
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
		if metric.GetName() != timestampMetricName && metric.Buckets() == nil {
			counter := r.counterLookup(metric, key)
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
				r.Logger.Warn("Counter is missing or unable to parse", slog.String("counter", metric.GetName()))
			}
		}
	}

	// order metrics, such that those requiring base counters are processed last
	orderedMetrics := orderedNonDenominatorMetrics
	orderedMetrics = append(orderedMetrics, orderedDenominatorMetrics...)
	orderedKeys := orderedNonDenominatorKeys
	orderedKeys = append(orderedKeys, orderedDenominatorKeys...)

	// Calculate timestamp delta first since many counters require it for postprocessing.
	// Timestamp has "raw" property, so it isn't post-processed automatically
	if _, err = curMat.Delta("timestamp", prevMat, cachedData, r.AllowPartialAggregation, r.Logger); err != nil {
		r.Logger.Error("(timestamp) calculate delta:", slogx.Err(err))
	}

	var base *matrix.Metric
	var totalSkips int

	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := r.counterLookup(metric, key)
		if counter == nil {
			r.Logger.Error(
				"Missing counter:",
				slogx.Err(err),
				slog.String("counter", metric.GetName()),
			)
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
		if skips, err = curMat.Delta(key, prevMat, cachedData, r.AllowPartialAggregation, r.Logger); err != nil {
			r.Logger.Error("Calculate delta:", slogx.Err(err), slog.String("key", key))
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
			if isWorkloadDetailObject(r.Prop.Query) {
				// The workload detail generates metrics at the resource level. The 'service_time' and 'wait_time' metrics are used as raw values for these resource-level metrics. Their denominator, 'visits', is not collected; therefore, a check is added here to prevent warnings.
				// There is no need to cook these metrics further.
				if key == "service_time" || key == "wait_time" {
					continue
				}
			}
			r.Logger.Warn(
				"Base counter missing",
				slog.String("key", key),
				slog.String("property", property),
				slog.String("denominator", counter.denominator),
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
				skips, err = curMat.DivideWithThreshold(key, counter.denominator, r.perfProp.latencyIoReqd, cachedData, prevMat, timestampMetricName, r.Logger)
			} else {
				skips, err = curMat.Divide(key, counter.denominator)
			}

			if err != nil {
				r.Logger.Error("Division by base", slogx.Err(err), slog.String("key", key))
				continue
			}
			totalSkips += skips

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if skips, err = curMat.MultiplyByScalar(key, 100); err != nil {
				r.Logger.Error("Multiply by scalar", slogx.Err(err), slog.String("key", key))
			} else {
				totalSkips += skips
			}
			continue
		}
		// If we reach here, then one of the earlier clauses should have executed `continue` statement
		r.Logger.Error(
			"Unknown property",
			slog.String("key", key),
			slog.String("property", property),
		)
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := r.counterLookup(metric, key)
		if counter != nil {
			property := counter.counterType
			if property == "rate" {
				if skips, err = curMat.Divide(orderedKeys[i], timestampMetricName); err != nil {
					r.Logger.Error(
						"Calculate rate",
						slogx.Err(err),
						slog.Int("i", i),
						slog.String("metric", metric.GetName()),
						slog.String("key", key),
					)
					continue
				}
				totalSkips += skips
			}
		} else {
			r.Logger.Warn("Counter is missing or unable to parse", slog.String("counter", metric.GetName()))
			continue
		}
	}

	calcD := time.Since(calcStart)
	_ = r.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = r.Metadata.LazySetValueInt64("calc_time", "data", calcD.Microseconds())
	_ = r.Metadata.LazySetValueUint64("skips", "data", uint64(totalSkips)) //nolint:gosec

	// store cache for next poll
	r.Matrix[r.Object] = cachedData

	newDataMap := make(map[string]*matrix.Matrix)
	newDataMap[r.Object] = curMat
	return newDataMap, nil
}

func (r *CmPerf) counterLookup(metric *matrix.Metric, metricKey string) *counter {
	var c *counter

	if metric.IsArray() {
		name, _, _ := strings.Cut(metricKey, arrayKeyToken)
		c = r.perfProp.counterInfo[name]
	} else {
		c = r.perfProp.counterInfo[metricKey]
	}
	return c
}

func (r *CmPerf) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Vscan":
		return vscan.New(abc)
	case "FlexCache":
		return flexcache.New(abc)
	case "Disk":
		return disk.New(abc)
	case "Nic":
		return nic.New(abc)
	case "Headroom":
		return headroom.New(abc)
	case "Fcp":
		return fcp.New(abc)
	case "FCVI":
		return fcvi.New(abc)
	case "FabricPool":
		return fabricpool.New(abc)
	case "Volume":
		return volume.New(abc)

	default:
		r.Logger.Info("no CmPerf plugin found", slog.String("kind", kind))
	}
	return nil
}

func (r *CmPerf) InitSchedule() {
	if r.Schedule == nil {
		return
	}
	tasks := r.Schedule.GetTasks()
	for _, task := range tasks {
		if task.Name == "instance" {
			r.hasInstanceSchedule = true
			return
		}
	}
}

func isWorkloadObject(query string) bool {
	_, ok := qosQueries[query]
	return ok
}

func isWorkloadDetailObject(query string) bool {
	_, ok := qosDetailQueries[query]
	return ok
}

// Interface guards
var (
	_ collector.Collector = (*CmPerf)(nil)
)
