package eseriesperf

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries"
	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/collectors/eseriesperf/plugins/cachehitratio"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

type EseriesPerf struct {
	*eseries.ESeries
	perfProp      *perfProp
	pollDataCalls int
	recordsToSave int
}

type counter struct {
	name        string
	counterType string
	denominator string
}

type perfProp struct {
	isCacheEmpty         bool
	counterInfo          map[string]*counter
	timestampMetricName  string
	calculateUtilization bool
}

func init() {
	plugin.RegisterModule(&EseriesPerf{})
}

func (ep *EseriesPerf) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.eseriesperf",
		New: func() plugin.Module { return new(EseriesPerf) },
	}
}

func (ep *EseriesPerf) Init(a *collector.AbstractCollector) error {
	var err error

	ep.ESeries = &eseries.ESeries{AbstractCollector: a}
	ep.perfProp = &perfProp{}
	ep.perfProp.counterInfo = make(map[string]*counter)

	ep.InitProp()

	if ep.Prop.TemplatePath, err = ep.LoadTemplate(); err != nil {
		return err
	}

	if err := ep.ParseTemplate(); err != nil {
		return err
	}

	if err := ep.InitClient(); err != nil {
		return err
	}

	ep.Remote = ep.Client.Remote()

	if err := collector.Init(ep); err != nil {
		return err
	}

	ep.InitMatrix()

	ep.buildCounters()

	ep.recordsToSave = collector.RecordKeepLast(ep.Params, ep.Logger)

	ep.Logger.Debug(
		"initialized",
		slog.Int("numMetrics", len(ep.Prop.Metrics)),
		slog.String("object", ep.Prop.Object),
		slog.String("timeout", ep.Client.Timeout.String()),
	)

	return nil
}

func (ep *EseriesPerf) InitMatrix() {
	mat := ep.Matrix[ep.Object]
	ep.perfProp.isCacheEmpty = true

	mat.Object = ep.Prop.Object

	if e := ep.Params.GetChildS("export_options"); e != nil {
		mat.SetExportOptions(e)
	}

	if ep.Params.HasChildS("labels") {
		for _, l := range ep.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	_, _ = ep.Metadata.NewMetricUint64("skips")
	_, _ = ep.Metadata.NewMetricUint64("numPartials")
}

func (ep *EseriesPerf) LoadTemplate() (string, error) {
	templateName := ep.Params.GetChildS("objects").GetChildContentS(ep.Object)
	if templateName == "" {
		return "", errs.New(errs.ErrMissingParam, "template for object "+ep.Object)
	}

	jitter := ep.Params.GetChildContentS("jitter")
	subTemplate, path, err := ep.ImportSubTemplate([]string{""}, templateName, jitter, ep.Remote.Version)
	if err != nil {
		return "", err
	}

	ep.Params.Union(subTemplate)

	ep.Logger.Debug("loaded template",
		slog.String("object", ep.Object),
		slog.String("template", templateName),
		slog.String("path", path),
	)

	return path, nil
}

// ParseTemplate parses template using performance collector configuration
func (ep *EseriesPerf) ParseTemplate() error {
	objType := ep.Params.GetChildContentS("type")
	if objType == "" {
		return errs.New(errs.ErrMissingParam, "type")
	}

	// Get perf-specific configuration
	config := eseries.GetESeriesPerfObjectConfig(objType)

	if err := ep.ESeries.ParseTemplate(config); err != nil {
		return err
	}

	ep.perfProp.calculateUtilization = config.CalculateUtilization

	return nil
}

func findStaticCounterDefPath() string {
	prodPath := "conf/eseriesperf/static_counter_definitions.yaml"
	testPath := "../../../conf/eseriesperf/static_counter_definitions.yaml"

	if _, err := os.Stat(prodPath); err == nil {
		return prodPath
	}
	return testPath
}

func (ep *EseriesPerf) buildCounters() {
	staticCounterDef, err := LoadStaticCounterDefinitions(ep.Prop.Object, findStaticCounterDefPath(), ep.Logger)
	if err != nil {
		ep.Logger.Error("Failed to load static counter definitions", slogx.Err(err))
	}

	var timestampMetric string
	for metricName := range ep.Prop.Metrics {
		if strings.Contains(metricName, "observedTimeInMS") {
			timestampMetric = metricName
			break
		}
	}

	if timestampMetric != "" {
		if _, exists := ep.Prop.Metrics[timestampMetric]; !exists {
			ep.Prop.Metrics[timestampMetric] = &eseries.Metric{
				Label:      "observedTimeInMS",
				Name:       timestampMetric,
				Exportable: false,
			}
		}
	}

	for k, v := range ep.Prop.Metrics {
		if _, exists := ep.perfProp.counterInfo[k]; exists {
			continue
		}

		counterDef, exists := staticCounterDef.CounterDefinitions[v.Name]
		if !exists {
			ep.Logger.Debug("Skipping metric - not found in static counter definitions", slog.String("name", k))
			continue
		}

		ctr := &counter{
			name:        k,
			counterType: counterDef.Type,
			denominator: counterDef.BaseCounter,
		}

		// Handle timestamp metric
		if strings.Contains(k, "observedTimeInMS") {
			ep.perfProp.timestampMetricName = k
		}

		// Ensure denominator exists in counterInfo if specified
		if counterDef.BaseCounter != "" {
			if _, denomExists := ep.perfProp.counterInfo[counterDef.BaseCounter]; !denomExists {
				var baseCounterType string
				if baseCounterDef, baseCounterExists := staticCounterDef.CounterDefinitions[counterDef.BaseCounter]; baseCounterExists {
					baseCounterType = baseCounterDef.Type
				}
				if baseCounterType != "" {
					ep.perfProp.counterInfo[counterDef.BaseCounter] = &counter{
						name:        counterDef.BaseCounter,
						counterType: staticCounterDef.CounterDefinitions[counterDef.BaseCounter].Type,
					}
					if _, dExists := ep.Prop.Metrics[counterDef.BaseCounter]; !dExists {
						m := &eseries.Metric{Label: "", Name: counterDef.BaseCounter, MetricType: "", Exportable: false}
						ep.Prop.Metrics[counterDef.BaseCounter] = m
					}
				}
			}
		}

		ep.perfProp.counterInfo[k] = ctr
	}
}

func (ep *EseriesPerf) LoadPlugin(kind string, p *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "CacheHitRatio":
		return cachehitratio.New(p)
	default:
		ep.Logger.Info("No eseries plugin found", slog.String("kind", kind))
	}
	return nil
}

func (ep *EseriesPerf) PollCounter() (map[string]*matrix.Matrix, error) {
	return ep.ESeries.PollCounter()
}

func (ep *EseriesPerf) PollData() (map[string]*matrix.Matrix, error) {
	var (
		count       uint64
		numPartials uint64
		apiTime     time.Duration
		parseTime   time.Duration
	)

	ep.Client.Metadata.Reset()

	ep.pollDataCalls++
	if ep.pollDataCalls >= ep.recordsToSave {
		ep.pollDataCalls = 0
	}

	var headers map[string]string

	poller, err := conf.PollerNamed(ep.Options.Poller)
	if err != nil {
		ep.Logger.Error("failed to find poller", slogx.Err(err), slog.String("poller", ep.Options.Poller))
	}

	if poller.IsRecording() {
		headers = map[string]string{
			"From": strconv.Itoa(ep.pollDataCalls),
		}
	}

	oldInstances := set.New()
	prevMat := ep.Matrix[ep.Object]
	for key := range prevMat.GetInstances() {
		oldInstances.Add(key)
	}

	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	systemID := ep.GetArray()

	// Build query - filters are intentionally disabled when using shared cache
	// to prevent cache poisoning (subset of filtered data being cached for all consumers)
	// See applyFilter() in template.go for the filter-disabling logic
	filters := ep.Prop.Filter

	query := rest.NewURLBuilder().
		APIPath(ep.Prop.Query).
		ArrayID(systemID).
		Filter(filters).
		Build()

	var results []gjson.Result
	apiStart := time.Now()

	results, err = ep.Client.Fetch(ep.Client.APIPath+"/"+query, ep.Prop.CacheConfig, headers)
	apiTime = time.Since(apiStart)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		ep.Logger.Debug("no performance instances")
		return nil, errs.New(errs.ErrNoInstance, "no instances found")
	}

	parseStart := time.Now()
	count, numPartials = ep.pollData(curMat, results, oldInstances)
	parseTime = time.Since(parseStart)

	for key := range oldInstances.Iter() {
		curMat.RemoveInstance(key)
	}

	_ = ep.Metadata.LazySetValueInt64("api_time", "data", apiTime.Microseconds())
	_ = ep.Metadata.LazySetValueInt64("parse_time", "data", parseTime.Microseconds())
	_ = ep.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = ep.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = ep.Metadata.LazySetValueUint64("numPartials", "data", numPartials)
	_ = ep.Metadata.LazySetValueUint64("bytesRx", "data", ep.Client.Metadata.BytesRx)
	_ = ep.Metadata.LazySetValueUint64("numCalls", "data", ep.Client.Metadata.NumCalls)

	ep.AddCollectCount(count)

	return ep.cookCounters(curMat, prevMat)
}

func (ep *EseriesPerf) pollData(curMat *matrix.Matrix, results []gjson.Result, oldInstances *set.Set) (uint64, uint64) {
	var (
		err         error
		count       uint64
		numPartials uint64
	)

	var dataArray []gjson.Result
	if ep.Prop.ResponseArrayPath != "" {
		if len(results) == 0 {
			ep.Logger.Warn("No results to parse for response array path")
			return 0, 0
		}
		arrayResult := gjson.ParseBytes([]byte(results[0].Raw)).Get(ep.Prop.ResponseArrayPath)
		if !arrayResult.Exists() {
			ep.Logger.Warn("Response array path not found", slog.String("path", ep.Prop.ResponseArrayPath))
			return 0, 0
		}
		dataArray = arrayResult.Array()
	} else {
		dataArray = results
	}

	prevMat := ep.Matrix[ep.Object]

	for _, instanceData := range dataArray {
		if !instanceData.IsObject() {
			ep.Logger.Warn("instance data is not object, skipping")
			continue
		}

		var instanceKey strings.Builder
		if len(ep.Prop.InstanceKeys) > 0 {
			for _, k := range ep.Prop.InstanceKeys {
				value := instanceData.Get(k)
				if value.Exists() {
					instanceKey.WriteString(value.ClonedString())
				}
			}
		}

		if instanceKey.String() == "" {
			ep.Logger.Warn("empty instance key, skipping",
				slog.String("object", ep.Object),
				slog.Any("instanceKeys", ep.Prop.InstanceKeys))
			continue
		}

		instKey := instanceKey.String()

		instance := curMat.GetInstance(instKey)
		if instance == nil {
			var err error
			if instance, err = curMat.NewInstance(instKey); err != nil {
				ep.Logger.Error("failed to create instance", slogx.Err(err), slog.String("key", instKey))
				continue
			}
		}

		instance.SetExportable(true)
		oldInstances.Remove(instKey)

		// Check for tracking field changes to detect partial data
		// Bad data points are ones where lastResetTimeInMS, sourceController, or memberIdsHash
		// has changed between current and previous poll
		if prevMat != nil && !ep.perfProp.isCacheEmpty {
			prevInstance := prevMat.GetInstance(instKey)
			if prevInstance != nil {
				isPartial := false

				// Check if previous observedTimeInMS < current lastResetTimeInMS
				// This indicates a counter reset occurred
				// Note: Previous values are stored in seconds (converted), so we need to convert current values too
				if lastResetTimeMetric := prevMat.GetMetric("lastResetTimeInMS"); lastResetTimeMetric != nil {
					if observedTimeMetric := prevMat.GetMetric("observedTimeInMS"); observedTimeMetric != nil {
						prevObservedTime, prevObsOk := observedTimeMetric.GetValueFloat64(prevInstance)
						prevLastResetTime, prevResetOk := lastResetTimeMetric.GetValueFloat64(prevInstance)

						// Get current lastResetTime from instanceData and convert to seconds
						lastResetCur := instanceData.Get("lastResetTimeInMS")
						curLastResetTime := lastResetCur.Float() / 1000.0
						if lastResetCur.Exists() && prevObsOk {
							if prevObservedTime > 0 && prevObservedTime < curLastResetTime {
								isPartial = true
								ep.Logger.Debug("Partial detected: counter reset detected",
									slog.String("instance", instKey),
									slog.Float64("prevObservedTime", prevObservedTime),
									slog.Float64("curLastResetTime", curLastResetTime))
							}
						}

						// Check if lastResetTimeInMS value changed between polls
						if lastResetCur.Exists() && prevResetOk {
							if prevLastResetTime > 0 && prevLastResetTime != curLastResetTime {
								isPartial = true
								ep.Logger.Debug("Partial detected: lastResetTimeInMS changed",
									slog.String("instance", instKey),
									slog.Float64("prev", prevLastResetTime),
									slog.Float64("cur", curLastResetTime))
							}
						}
					}
				}

				// Check sourceController
				sourceControllerCur := instanceData.Get("sourceController")
				if sourceControllerCur.Exists() {
					sourceControllerPrev := prevInstance.GetLabel("sourceController")
					if sourceControllerPrev != "" && sourceControllerCur.ClonedString() != sourceControllerPrev {
						isPartial = true
						ep.Logger.Debug("Partial detected: sourceController changed",
							slog.String("instance", instKey),
							slog.String("prev", sourceControllerPrev),
							slog.String("cur", sourceControllerCur.ClonedString()))
					}
				}

				// Check memberIdsHash
				memberIDsHashCur := instanceData.Get("memberIdsHash")
				if memberIDsHashCur.Exists() {
					memberIDsHashPrev := prevInstance.GetLabel("memberIdsHash")
					if memberIDsHashPrev != "" && memberIDsHashCur.ClonedString() != memberIDsHashPrev {
						isPartial = true
						ep.Logger.Debug("Partial detected: memberIdsHash changed",
							slog.String("instance", instKey),
							slog.String("prev", memberIDsHashPrev),
							slog.String("cur", memberIDsHashCur.ClonedString()))
					}
				}

				if isPartial {
					instance.SetPartial(true)
					instance.SetExportable(false)
					numPartials++
				}
			}
		}

		for label, display := range ep.Prop.InstanceLabels {
			value := instanceData.Get(label)
			if value.Exists() {
				instance.SetLabel(display, value.ClonedString())
			}
		}

		for _, metric := range ep.Prop.Metrics {
			value := instanceData.Get(metric.Name)
			if !value.Exists() {
				continue
			}

			metr, ok := curMat.GetMetrics()[metric.Name]
			if !ok {
				if metr, err = curMat.NewMetricFloat64(metric.Name, metric.Label); err != nil {
					ep.Logger.Error(
						"NewMetricFloat64",
						slogx.Err(err),
						slog.String("name", metric.Name),
					)
				} else {
					metr.SetExportable(metric.Exportable)
				}
			}

			// Convert milliseconds to seconds for timestamp metrics
			floatValue := value.Float()
			if strings.Contains(metric.Name, "observedTimeInMS") || strings.Contains(metric.Name, "lastResetTimeInMS") {
				floatValue /= 1000.0
			}
			metr.SetValueFloat64(instance, floatValue)
			count++
		}
	}

	return count, numPartials
}

func (ep *EseriesPerf) cookCounters(curMat *matrix.Matrix, prevMat *matrix.Matrix) (map[string]*matrix.Matrix, error) {
	var (
		err   error
		skips int
	)

	if ep.perfProp.isCacheEmpty {
		ep.Logger.Debug("skip postprocessing until next poll (previous cache empty)")
		ep.Matrix[ep.Object] = curMat
		ep.perfProp.isCacheEmpty = false
		return nil, nil
	}

	calcStart := time.Now()

	cachedData := curMat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true, PartialInstances: true})

	orderedNonDenominatorMetrics := make([]*matrix.Metric, 0)
	orderedNonDenominatorKeys := make([]string, 0)
	orderedDenominatorMetrics := make([]*matrix.Metric, 0)
	orderedDenominatorKeys := make([]string, 0)

	counterMap := ep.perfProp.counterInfo

	for key, metric := range curMat.GetMetrics() {
		counter := counterMap[key]
		if counter != nil {
			if counter.denominator == "" {
				orderedNonDenominatorMetrics = append(orderedNonDenominatorMetrics, metric)
				orderedNonDenominatorKeys = append(orderedNonDenominatorKeys, key)
			} else {
				orderedDenominatorMetrics = append(orderedDenominatorMetrics, metric)
				orderedDenominatorKeys = append(orderedDenominatorKeys, key)
			}
		} else {
			ep.Logger.Debug("Counter metadata not found", slog.String("counter", metric.GetName()))
			metric.SetExportable(false)
		}
	}

	timestamp := curMat.GetMetric(ep.perfProp.timestampMetricName)
	if timestamp != nil {
		timestamp.SetExportable(false)
	} else {
		return nil, errs.New(errs.ErrConfig, "missing timestamp metric")
	}

	// Don't export lastResetTimeInMS it's only used for partial detection
	lastResetTime := curMat.GetMetric("lastResetTimeInMS")
	if lastResetTime != nil {
		lastResetTime.SetExportable(false)
	}

	// Don't export idleTime it's only used for utilization
	idleTime := curMat.GetMetric("idleTime")
	if idleTime != nil {
		idleTime.SetExportable(false)
	}

	err = ep.validateMatrix(prevMat, curMat)
	if err != nil {
		return nil, err
	}

	orderedMetrics := orderedNonDenominatorMetrics
	orderedMetrics = append(orderedMetrics, orderedDenominatorMetrics...)
	orderedKeys := orderedNonDenominatorKeys
	orderedKeys = append(orderedKeys, orderedDenominatorKeys...)

	var totalSkips int

	// First pass: Calculate deltas only
	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := counterMap[key]
		if counter == nil {
			continue
		}

		property := counter.counterType
		metric.SetProperty(property)
		metric.SetComment(counter.denominator)

		if property == "raw" {
			continue
		}

		if skips, err = curMat.Delta(key, prevMat, cachedData, false, ep.Logger); err != nil {
			ep.Logger.Error("Calculate delta", slogx.Err(err), slog.String("key", key))
			continue
		}
		totalSkips += skips
	}

	// Calculate utilization if template flag is set (after all deltas are done)
	if ep.perfProp.calculateUtilization {
		if skips, err := ep.calculateUtilization(curMat); err != nil {
			ep.Logger.Error("Calculate utilization", slogx.Err(err))
		} else {
			totalSkips += skips
		}
	}

	// Second pass: Apply transformations (average, percent)
	for i := range orderedMetrics {
		key := orderedKeys[i]
		counter := counterMap[key]
		if counter == nil {
			continue
		}

		property := counter.counterType

		if property == "raw" || property == "delta" || property == "rate" {
			continue
		}

		if property == "average" || property == "percent" {
			base := curMat.GetMetric(counter.denominator)
			if base == nil {
				ep.Logger.Warn("Base counter missing", slog.String("key", key), slog.String("denominator", counter.denominator))
				skips = curMat.Skip(key)
				totalSkips += skips
				continue
			}

			if strings.Contains(key, "Latency") {
				skips, err = curMat.DivideWithThreshold(key, counter.denominator, 0, cachedData, prevMat, ep.perfProp.timestampMetricName, ep.Logger)
			} else {
				skips, err = curMat.Divide(key, counter.denominator)
			}

			if err != nil {
				ep.Logger.Error("Division by base", slogx.Err(err), slog.String("key", key))
				continue
			}
			totalSkips += skips

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if skips, err = curMat.MultiplyByScalar(key, 100); err != nil {
				ep.Logger.Error("Multiply by scalar", slogx.Err(err), slog.String("key", key))
			} else {
				totalSkips += skips
			}
			continue
		}
	}

	// Third pass: Calculate rates (divide by timestamp delta)
	for i := range orderedMetrics {
		key := orderedKeys[i]
		counter := counterMap[key]
		if counter == nil {
			continue
		}

		property := counter.counterType
		if property == "rate" {
			if skips, err = curMat.Divide(orderedKeys[i], ep.perfProp.timestampMetricName); err != nil {
				ep.Logger.Error("Calculate rate", slogx.Err(err), slog.String("key", key))
				continue
			}
			totalSkips += skips
		}
	}

	calcD := time.Since(calcStart)
	_ = ep.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = ep.Metadata.LazySetValueInt64("calc_time", "data", calcD.Microseconds())
	if totalSkips >= 0 {
		_ = ep.Metadata.LazySetValueUint64("skips", "data", uint64(totalSkips))
	}

	ep.Matrix[ep.Object] = cachedData

	newDataMap := make(map[string]*matrix.Matrix)
	newDataMap[ep.Object] = curMat
	return newDataMap, nil
}

// calculateUtilization calculates utilization percentages based on time delta metrics:
// readUtilization = (deltaReadTime / totalTime) * 100
// writeUtilization = (deltaWriteTime / totalTime) * 100
// totalUtilization = ((deltaReadTime + deltaWriteTime) / totalTime) * 100
// where totalTime = deltaReadTime + deltaWriteTime + deltaIdleTime
// This calculation must be done in the collector, not the plugin,
// because idleTime cannot be computed without the missing denominator configuration.
// This function creates three new metrics in the matrix.
func (ep *EseriesPerf) calculateUtilization(curMat *matrix.Matrix) (int, error) {
	readTimeMetric := curMat.GetMetric("readTimeTotal")
	writeTimeMetric := curMat.GetMetric("writeTimeTotal")
	idleTimeMetric := curMat.GetMetric("idleTime")

	if readTimeMetric == nil || writeTimeMetric == nil || idleTimeMetric == nil {
		return 0, errs.New(errs.ErrMissingParam, "missing time metrics (read_time_total, write_time_total, idle_time) for utilization")
	}

	// Create utilization metrics
	readUtilMetric, err := curMat.NewMetricFloat64("read_utilization")
	if err != nil {
		return 0, err
	}
	readUtilMetric.SetProperty("percent")

	writeUtilMetric, err := curMat.NewMetricFloat64("write_utilization")
	if err != nil {
		return 0, err
	}
	writeUtilMetric.SetProperty("percent")

	totalUtilMetric, err := curMat.NewMetricFloat64("total_utilization")
	if err != nil {
		return 0, err
	}
	totalUtilMetric.SetProperty("percent")

	skips := 0
	for _, instance := range curMat.GetInstances() {
		readTime, readOk := readTimeMetric.GetValueFloat64(instance)
		writeTime, writeOk := writeTimeMetric.GetValueFloat64(instance)
		idleTime, idleOk := idleTimeMetric.GetValueFloat64(instance)

		if !readOk || !writeOk || !idleOk {
			continue
		}

		totalTime := readTime + writeTime + idleTime
		if totalTime > 0 {
			readUtilization := (readTime / totalTime) * 100.0
			writeUtilization := (writeTime / totalTime) * 100.0
			totalUtilization := ((readTime + writeTime) / totalTime) * 100.0

			readUtilMetric.SetValueFloat64(instance, readUtilization)
			writeUtilMetric.SetValueFloat64(instance, writeUtilization)
			totalUtilMetric.SetValueFloat64(instance, totalUtilization)
		} else {
			skips++
		}
	}

	return skips, nil
}

// validateMatrix ensures that the previous matrix (prevMat) contains all the metrics present in the current matrix (curMat).
// This is crucial for performing accurate comparisons and calculations between the two matrices, especially in scenarios where
// the current matrix may have additional metrics that are not present in the previous matrix, such as after an ONTAP upgrade.
//
// The function iterates over all the metrics in curMat and checks if each metric exists in prevMat. If a metric from curMat
// does not exist in prevMat, it is created in prevMat as a new float64 metric. This prevents potential panics or errors
// when attempting to perform calculations with metrics that are missing in prevMat.
func (ep *EseriesPerf) validateMatrix(prevMat *matrix.Matrix, curMat *matrix.Matrix) error {
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
