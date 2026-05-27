package cmperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/cmmetrics"
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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	latencyIoReqd       = 0
	arrayKeyToken       = "#"
	timestampMetricName = "timestamp"
)

type CmPerf struct {
	*rest2.Rest     // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	perfProp        *perfProp
	archivedMetrics map[string]*rest2.Metric // Keeps metric definitions that are not found in the counter schema. These metrics may be available in future ONTAP versions.
	recordsToSave   int                      // Number of records to save when using the recorder
	lastTimestamp   time.Time                // tracks last downloaded file timestamp (aggregated: 1 file per object)
}

type counter struct {
	counterType string
	denominator string
}

type perfProp struct {
	isCacheEmpty        bool
	counterInfo         map[string]*counter
	schemaMap           map[uint32]cmmetrics.CounterSchema
	latencyIoReqd       int
	qosLabels           map[string]string
	disableConstituents bool
}

func init() {
	plugin.RegisterModule(&CmPerf{})
}

func (c *CmPerf) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.cmperf",
		New: func() plugin.Module { return new(CmPerf) },
	}
}

func (c *CmPerf) Init(a *collector.AbstractCollector) error {

	var err error

	c.Rest = &rest2.Rest{AbstractCollector: a}

	c.perfProp = &perfProp{}

	c.InitProp()

	c.perfProp.counterInfo = make(map[string]*counter)
	c.archivedMetrics = make(map[string]*rest2.Metric)

	if err := c.InitClient(); err != nil {
		return err
	}

	if c.Prop.TemplatePath, err = c.LoadTemplate(); err != nil {
		return err
	}

	c.InitVars(a.Params)

	if err := collector.Init(c); err != nil {
		return err
	}

	if err := c.InitCache(); err != nil {
		return err
	}

	if err := c.InitMatrix(); err != nil {
		return err
	}

	if err := c.InitQOS(); err != nil {
		return err
	}

	c.recordsToSave = collector.RecordKeepLast(c.Params, c.Logger)

	c.Logger.Debug(
		"initialized cache",
		slog.Int("numMetrics", len(c.Prop.Metrics)),
		slog.String("timeout", c.Client.GetTimeout().String()),
	)

	return nil
}

func (c *CmPerf) InitQOS() error {
	if isWorkloadObject(c.Prop.Query) {
		qosLabels := c.Params.GetChildS("qos_labels")
		if qosLabels == nil {
			return errs.New(errs.ErrMissingParam, "qos_labels")
		}
		c.perfProp.qosLabels = make(map[string]string)
		for _, label := range qosLabels.GetAllChildContentS() {

			display := strings.ReplaceAll(label, "-", "_")
			before, after, found := strings.Cut(label, "=>")
			if found {
				label = strings.TrimSpace(before)
				display = strings.TrimSpace(after)
			}
			c.perfProp.qosLabels[label] = display
		}
	}
	if counters := c.Params.GetChildS("counters"); counters != nil {
		refine := counters.GetChildS("refine")
		if refine != nil {
			withConstituents := refine.GetChildContentS("with_constituents")
			if withConstituents == "false" {
				c.perfProp.disableConstituents = true
			}
		}
	}
	return nil
}

func (c *CmPerf) InitMatrix() error {
	mat := c.Matrix[c.Object]
	// init perf properties
	c.perfProp.latencyIoReqd = c.loadParamInt("latency_io_reqd", latencyIoReqd)
	c.perfProp.isCacheEmpty = true
	// overwrite from abstract collector
	mat.Object = c.Prop.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", c.Remote.Name)
	if c.Params.HasChildS("labels") {
		for _, l := range c.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	_, _ = c.Metadata.NewMetricUint64("skips")
	_, _ = c.Metadata.NewMetricUint64("numPartials")
	_, _ = c.Metadata.NewMetricUint64("downloadFailures")
	return nil
}

// load an int parameter or use defaultValue
func (c *CmPerf) loadParamInt(name string, defaultValue int) int {

	var (
		x string
		n int
		e error
	)

	if x = c.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			c.Logger.Debug("using",
				slog.String("name", name),
				slog.Int("value", n),
			)
			return n
		}
		c.Logger.Warn("invalid parameter (expected integer)", slog.String("name", name), slog.String("value", x))
	}

	c.Logger.Debug("using", slog.String("name", name), slog.Int("defaultValue", defaultValue))
	return defaultValue
}

func (c *CmPerf) PollCounter() (map[string]*matrix.Matrix, error) {

	mat := c.Matrix[c.Object]

	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	if mat.GetMetric(timestampMetricName) == nil {
		m, err := mat.NewMetricFloat64(timestampMetricName)
		if err != nil {
			c.Logger.Error("add timestamp metric", slogx.Err(err))
		} else {
			m.SetProperty("raw")
			m.SetExportable(false)
		}
	}

	c.buildCounters()

	return nil, nil
}

// GetOverride override counter property
func (c *CmPerf) GetOverride(counter string) string {
	if o := c.Params.GetChildS("override"); o != nil {
		return o.GetChildContentS(counter)
	}
	return ""
}

func (c *CmPerf) PollData() (map[string]*matrix.Matrix, error) {
	var (
		apiD, parseD time.Duration
		metricCount  uint64
		numPartials  uint64
		startTime    time.Time
		prevMat      *matrix.Matrix
		curMat       *matrix.Matrix
	)

	timestamp := c.Matrix[c.Object].GetMetric(timestampMetricName)
	if timestamp == nil {
		return nil, errs.New(errs.ErrConfig, "missing timestamp metric")
	}

	startTime = time.Now()
	c.RequestMetadata.Reset()
	prevMat = c.Matrix[c.Object]

	// clone matrix without numeric data
	curMat = prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: false, ExportInstances: false})
	curMat.Reset()

	apiD += time.Since(startTime)

	baseDir := os.TempDir()
	if envDir := os.Getenv("HARVEST_CMPERF_TMPDIR"); envDir != "" {
		baseDir = envDir
	}
	tmpDir := filepath.Clean(filepath.Join(baseDir, c.Options.Poller+"-cmperf", "harvest-cmperf-"+c.Prop.Object))
	if mkErr := os.MkdirAll(tmpDir, 0750); mkErr != nil {
		return nil, fmt.Errorf("create CM2 temp dir %s: %w", tmpDir, mkErr)
	}

	startTime = time.Now()
	filesDownloaded, dlErr := c.downloadCM2Files(tmpDir)
	apiD += time.Since(startTime)
	if dlErr != nil {
		_ = c.Metadata.LazySetValueUint64("downloadFailures", "data", 1)
		return nil, dlErr
	}
	_ = c.Metadata.LazySetValueUint64("downloadFailures", "data", 0)

	// TODO: We keep publishing older data from the Prometheus cache until it expires.What is no cm2 files are available for very long time?
	if !filesDownloaded {
		// No new file this poll — leave r.Matrix[r.Object] (prevMat) intact so
		// the next poll with fresh data can diff correctly against it.
		return nil, nil
	}

	startTime = time.Now()
	var pollErr error
	var pollPartials uint64
	metricCount, pollPartials, pollErr = c.pollCM2Files(tmpDir, curMat)
	numPartials += pollPartials
	if pollErr != nil {
		return nil, pollErr
	}
	if len(curMat.GetInstances()) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+c.Prop.Object+" instances on cluster")
	}
	parseD += time.Since(startTime)

	_ = c.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = c.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = c.Metadata.LazySetValueUint64("metrics", "data", metricCount)
	_ = c.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = c.Metadata.LazySetValueUint64("bytesRx", "data", c.RequestMetadata.BytesRx.Load())
	_ = c.Metadata.LazySetValueUint64("numCalls", "data", c.RequestMetadata.NumCalls.Load())
	_ = c.Metadata.LazySetValueUint64("numPartials", "data", numPartials)
	c.AddCollectCount(metricCount)

	return c.cookCounters(curMat, prevMat)
}

func (c *CmPerf) cookCounters(curMat *matrix.Matrix, prevMat *matrix.Matrix) (map[string]*matrix.Matrix, error) {
	var (
		err   error
		skips int
	)

	// skip calculating from delta if no data from previous poll
	if c.perfProp.isCacheEmpty {
		c.Logger.Debug("skip postprocessing until next poll (previous cache empty)")
		c.Matrix[c.Object] = curMat
		c.perfProp.isCacheEmpty = false
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
			counter := c.counterLookup(metric, key)
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
				c.Logger.Warn("Counter is missing or unable to parse", slog.String("counter", metric.GetName()))
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
	if _, err = curMat.Delta("timestamp", prevMat, cachedData, c.AllowPartialAggregation, c.Logger); err != nil {
		c.Logger.Error("(timestamp) calculate delta:", slogx.Err(err))
	}

	var base *matrix.Metric
	var totalSkips int

	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := c.counterLookup(metric, key)
		if counter == nil {
			c.Logger.Error(
				"Missing counter:",
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
		if skips, err = curMat.Delta(key, prevMat, cachedData, c.AllowPartialAggregation, c.Logger); err != nil {
			c.Logger.Error("Calculate delta:", slogx.Err(err), slog.String("key", key))
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
			c.Logger.Warn(
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
				skips, err = curMat.DivideWithThreshold(key, counter.denominator, c.perfProp.latencyIoReqd, cachedData, prevMat, timestampMetricName, c.Logger)
			} else {
				skips, err = curMat.Divide(key, counter.denominator)
			}

			if err != nil {
				c.Logger.Error("Division by base", slogx.Err(err), slog.String("key", key))
				continue
			}
			totalSkips += skips

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if skips, err = curMat.MultiplyByScalar(key, 100); err != nil {
				c.Logger.Error("Multiply by scalar", slogx.Err(err), slog.String("key", key))
			} else {
				totalSkips += skips
			}
			continue
		}
		// If we reach here, then one of the earlier clauses should have executed `continue` statement
		c.Logger.Error(
			"Unknown property",
			slog.String("key", key),
			slog.String("property", property),
		)
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := c.counterLookup(metric, key)
		if counter != nil {
			property := counter.counterType
			if property == "rate" {
				if skips, err = curMat.Divide(orderedKeys[i], timestampMetricName); err != nil {
					c.Logger.Error(
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
			c.Logger.Warn("Counter is missing or unable to parse", slog.String("counter", metric.GetName()))
			continue
		}
	}

	calcD := time.Since(calcStart)
	_ = c.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = c.Metadata.LazySetValueInt64("calc_time", "data", calcD.Microseconds())
	_ = c.Metadata.LazySetValueUint64("skips", "data", uint64(totalSkips)) //nolint:gosec

	// store cache for next poll
	c.Matrix[c.Object] = cachedData

	newDataMap := make(map[string]*matrix.Matrix)
	newDataMap[c.Object] = curMat
	return newDataMap, nil
}

func (c *CmPerf) counterLookup(metric *matrix.Metric, metricKey string) *counter {
	var co *counter

	if metric.IsArray() {
		name, _, _ := strings.Cut(metricKey, arrayKeyToken)
		co = c.perfProp.counterInfo[name]
	} else {
		co = c.perfProp.counterInfo[metricKey]
	}
	return co
}

func (c *CmPerf) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
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
		c.Logger.Info("no CmPerf plugin found", slog.String("kind", kind))
	}
	return nil
}

func isWorkloadObject(query string) bool {
	return query == "workload" || query == "workload_volume"
}

// Interface guards
var (
	_ collector.Collector = (*CmPerf)(nil)
)
