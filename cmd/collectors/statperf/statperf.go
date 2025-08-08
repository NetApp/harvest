package statperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/collectors/statperf/plugins/flexcache"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/cmd/tools/rest/clirequestbuilder"
	collector2 "github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	latencyIoReqd       = 10
	defaultBatchSize    = 500
	arrayKeyToken       = "#"
	subLabelToken       = "."
	timestampMetricName = "timestamp"
	endpoint            = "api/private/cli"
	keyToken            = "?#"
)

type StatPerf struct {
	*rest2.Rest       // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	perfProp          *perfProp
	filter            string
	archivedMetrics   map[string]*rest2.Metric // Keeps metric definitions that are not found in the counter schema. These metrics may be available in future ONTAP versions.
	pollInstanceCalls int
	pollDataCalls     int
	recordsToSave     int      // Number of records to save when using the recorder
	instanceNames     *set.Set // required for polldata
	sortedCounters    []string
	batchSize         int
}

type counter struct {
	name        string
	counterType string
	property    string
	denominator string
	description string
	labelCount  int
}

type perfProp struct {
	isCacheEmpty  bool
	counterInfo   map[string]*counter
	latencyIoReqd int
}

func init() {
	plugin.RegisterModule(&StatPerf{})
}

func (s *StatPerf) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.statperf",
		New: func() plugin.Module { return new(StatPerf) },
	}
}

func (s *StatPerf) Init(a *collector.AbstractCollector) error {

	var err error

	s.Rest = &rest2.Rest{AbstractCollector: a}

	s.perfProp = &perfProp{}

	s.InitProp()

	s.perfProp.counterInfo = make(map[string]*counter)
	s.archivedMetrics = make(map[string]*rest2.Metric)
	s.instanceNames = set.New()

	if err := s.InitClient(); err != nil {
		return err
	}

	if s.Prop.TemplatePath, err = s.LoadTemplate(); err != nil {
		return err
	}

	s.InitVars(a.Params)

	if err := collector.Init(s); err != nil {
		return err
	}

	if err := s.InitCache(); err != nil {
		return err
	}

	s.filter = s.loadFilter()
	s.batchSize = s.loadParamInt("batch_size", defaultBatchSize)

	if err := s.InitMatrix(); err != nil {
		return err
	}

	s.recordsToSave = collector.RecordKeepLast(s.Params, s.Logger)

	s.Logger.Debug(
		"initialized cache",
		slog.Int("numMetrics", len(s.Prop.Metrics)),
		slog.String("timeout", s.Client.Timeout.String()),
	)

	return nil
}

func (s *StatPerf) loadFilter() string {

	counters := s.Params.GetChildS("counters")
	if counters != nil {
		if x := counters.GetChildS("filter"); x != nil {
			filter := strings.Join(x.GetAllChildContentS(), ",")
			return filter
		}
	}
	return ""
}

func (s *StatPerf) InitMatrix() error {
	mat := s.Matrix[s.Object]
	// init perf properties
	s.perfProp.latencyIoReqd = s.loadParamInt("latency_io_reqd", latencyIoReqd)
	s.perfProp.isCacheEmpty = true
	// overwrite from abstract collector
	mat.Object = s.Prop.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", s.Remote.Name)
	if s.Params.HasChildS("labels") {
		for _, l := range s.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	// Add metadata metric for skips/numPartials
	_, _ = s.Metadata.NewMetricUint64("skips")
	_, _ = s.Metadata.NewMetricUint64("numPartials")
	return nil
}

// load an int parameter or use defaultValue
func (s *StatPerf) loadParamInt(name string, defaultValue int) int {

	var (
		x string
		n int
		e error
	)

	if x = s.Params.GetChildContentS(name); x != "" {
		if n, e = strconv.Atoi(x); e == nil {
			s.Logger.Debug("using",
				slog.String("name", name),
				slog.Int("value", n),
			)
			return n
		}
		s.Logger.Warn("invalid parameter (expected integer)", slog.String("name", name), slog.String("value", x))
	}

	s.Logger.Debug("using", slog.String("name", name), slog.Int("defaultValue", defaultValue))
	return defaultValue
}

func GetCounterInstanceBaseSet() string {
	baseSetTemplate := `set -showseparator "%s" -showallfields true -rows 0 diagnostic -confirmations off;statistics settings modify -counter-display all;`
	return fmt.Sprintf(baseSetTemplate, collector2.StatPerfSeparator)
}

func getDataBaseSet() string {
	baseSetTemplate := `set -rows 0 diagnostic -showallfields false -confirmations off -units raw;statistics settings modify -counter-display all;`
	return baseSetTemplate
}

func (s *StatPerf) PollCounter() (map[string]*matrix.Matrix, error) {
	var (
		err     error
		records []gjson.Result
	)

	cliCommand, err := clirequestbuilder.New().
		BaseSet(GetCounterInstanceBaseSet()).
		Query("statistics catalog counter show").
		Object(s.Prop.Query).
		Fields([]string{"counter", "base-counter", "properties", "type", "is-deprecated", "replaced-by", "label", "description"}).
		Build()
	if err != nil {
		return nil, err
	}

	s.Logger.Debug("", slog.String("cliCommand", string(cliCommand)))
	if cliCommand == nil {
		return nil, errs.New(errs.ErrConfig, "empty cliCommand")
	}

	apiT := time.Now()
	s.Client.Metadata.Reset()

	records, err = rest.FetchPost(s.Client, "api/private/cli", cliCommand)
	if err != nil {
		return nil, err
	}

	err = s.pollCounter(records, time.Since(apiT))
	if err != nil {
		return nil, err
	}
	keySet := set.New()
	for key := range s.Prop.Counters {
		keySet.Add(key)
	}

	// Add any missing base counters that are not present in the template
	// but are required for calculating values. These metrics need to be requested.
	for key := range s.Prop.Metrics {
		keySet.Add(key)
	}

	s.sortedCounters = keySet.Values()
	slices.Sort(s.sortedCounters)
	return nil, nil
}

func (s *StatPerf) pollCounter(records []gjson.Result, apiD time.Duration) error {
	var (
		parseT time.Time
	)

	mat := s.Matrix[s.Object]

	parseT = time.Now()
	firstRecord := records[0]
	fr := firstRecord.ClonedString()
	if fr == "" {
		return errs.New(errs.ErrConfig, "no data found")
	}

	counters, err := s.ParseCounters(fr)
	if err != nil {
		return err
	}

	seenMetrics := make(map[string]bool)

	// populate denominator metric to prop metrics
	for _, c := range counters {

		name := c.Name
		dataType := c.Type

		if p := s.GetOverride(name); p != "" {
			dataType = p
		}

		// Check if the metric was previously archived and restore it
		if archivedMetric, found := s.archivedMetrics[name]; found {
			s.Prop.Metrics[name] = archivedMetric
			delete(s.archivedMetrics, name) // Remove from archive after restoring
			s.Logger.Info("Metric found in archive. Restore it", slog.String("key", name))
		}

		if mn, has := s.Prop.Metrics[name]; has {
			// handle deprecated counters
			if c.Deprecated == "true" {
				m := &rest2.Metric{Label: "", Name: c.ReplacedBy, MetricType: mn.MetricType, Exportable: mn.Exportable}
				s.Prop.Metrics[c.ReplacedBy] = m
				s.Prop.Counters[c.ReplacedBy] = c.ReplacedBy
				s.Logger.Info(
					"Added replacement for deprecated counter",
					slog.String("deprecated", name),
					slog.String("replacement", c.ReplacedBy),
				)
			}

			if strings.Contains(dataType, "string") {
				if _, ok := s.Prop.InstanceLabels[name]; !ok {
					s.Prop.InstanceLabels[name] = s.Prop.Counters[name]
				}
				// set exportable as false
				s.Prop.Metrics[name].Exportable = false
				continue
			}
			d := c.BaseCounter
			if d != "" && d != "-" {
				if _, has := s.Prop.Metrics[d]; !has {
					// export false
					m := &rest2.Metric{Label: "", Name: d, MetricType: "", Exportable: false}
					s.Prop.Metrics[d] = m
				}
			}
		}
	}

	for _, c := range counters {
		name := c.Name
		if _, has := s.Prop.Metrics[name]; has {
			seenMetrics[name] = true
			if _, ok := s.perfProp.counterInfo[name]; !ok {
				s.perfProp.counterInfo[name] = &counter{
					name:        name,
					counterType: NormalizeCounterValue(c.Type),
					property:    s.parseCounterProperty(name, c.Properties),
					denominator: NormalizeCounterValue(c.BaseCounter),
					description: c.Description,
					labelCount:  c.LabelCount,
				}
				if p := s.GetOverride(name); p != "" {
					s.perfProp.counterInfo[name].property = p
				}
			}
		}
	}

	for name, metric := range s.Prop.Metrics {
		if !seenMetrics[name] {
			s.archivedMetrics[name] = metric
			// Log the metric that is not present in counterSchema.
			s.Logger.Warn("Metric not found in counterSchema", slog.String("key", name))
			delete(s.Prop.Metrics, name)
		}
	}

	// Create an artificial metric to hold timestamp of each instance data.
	// The reason we don't keep a single timestamp for the whole data
	// is because we might get instances in different batches
	if mat.GetMetric(timestampMetricName) == nil {
		m, err := mat.NewMetricFloat64(timestampMetricName)
		if err != nil {
			s.Logger.Error("add timestamp metric", slogx.Err(err))
		}
		m.SetProperty("raw")
		m.SetExportable(false)
	}

	// update metadata for collector logs
	_ = s.Metadata.LazySetValueInt64("api_time", "counter", apiD.Microseconds())
	_ = s.Metadata.LazySetValueInt64("parse_time", "counter", time.Since(parseT).Microseconds())
	_ = s.Metadata.LazySetValueUint64("metrics", "counter", uint64(len(s.perfProp.counterInfo)))
	_ = s.Metadata.LazySetValueUint64("bytesRx", "counter", s.Client.Metadata.BytesRx)
	_ = s.Metadata.LazySetValueUint64("numCalls", "counter", s.Client.Metadata.NumCalls)

	return nil
}

func NormalizeCounterValue(input string) string {
	if input == "-" {
		return ""
	}
	return input
}

func (s *StatPerf) parseCounterProperty(name string, p string) string {
	var (
		property string
	)
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
	case strings.Contains(p, "string"):
		property = "string"
	default:
		s.Logger.Warn(
			"skip counter with unknown property",
			slog.String("name", name),
			slog.String("property", p),
		)
		return ""
	}
	return property
}

// GetOverride override counter property
func (s *StatPerf) GetOverride(counter string) string {
	if o := s.Params.GetChildS("override"); o != nil {
		return o.GetChildContentS(counter)
	}
	return ""
}

var retryRe = regexp.MustCompile(`retry request with less than (\d+)`)

func (s *StatPerf) PollData() (map[string]*matrix.Matrix, error) {
	var (
		apiD, parseD time.Duration
		metricCount  uint64
		startTime    time.Time
		numPartials  uint64
		prevMat      *matrix.Matrix
		curMat       *matrix.Matrix
		records      []gjson.Result
	)

	if s.instanceNames.Size() == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+s.Object+" instances on cluster")
	}

	timestamp := s.Matrix[s.Object].GetMetric(timestampMetricName)
	if timestamp == nil {
		return nil, errs.New(errs.ErrConfig, "missing timestamp metric")
	}

	startTime = time.Now()
	s.Client.Metadata.Reset()

	s.pollDataCalls++
	if s.pollDataCalls >= s.recordsToSave {
		s.pollDataCalls = 0
	}

	var headers map[string]string

	poller, err := conf.PollerNamed(s.Options.Poller)
	if err != nil {
		slog.Error("failed to find poller", slogx.Err(err), slog.String("poller", s.Options.Poller))
	}

	if poller.IsRecording() {
		headers = map[string]string{
			"From": strconv.Itoa(s.pollDataCalls),
		}
	}

	prevMat = s.Matrix[s.Object]
	// clone matrix without numeric data
	curMat = prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	processBatch := func(perfRecords []gjson.Result) {
		if len(perfRecords) == 0 {
			return
		}

		// Process the current batch of records
		count, np, batchParseD := s.processPerfRecords(perfRecords, curMat, prevMat)
		numPartials += np
		metricCount += count
		parseD += batchParseD
		apiD -= batchParseD
	}

	allInstances := s.instanceNames.Slice()

	for i := 0; i < len(allInstances); i += s.batchSize {
		end := i + s.batchSize
		if end > len(allInstances) {
			end = len(allInstances)
		}
		batchInstances := allInstances[i:end]
		// Build CLI command with the current batch of instances.
		cliCommand, err := clirequestbuilder.New().
			BaseSet(getDataBaseSet()).
			Query("statistics show -raw").
			Object(s.Prop.Query).
			Counters(s.sortedCounters).
			Instances(batchInstances).
			Filter(s.filter).
			Build()

		if err != nil {
			return nil, err
		}
		s.Logger.Debug("", slog.String("cliCommand (batch)", string(cliCommand)))
		if cliCommand == nil {
			return nil, errs.New(errs.ErrConfig, "empty cliCommand")
		}

		// Fetch results (no pagination assumed) for current batch.
		records, err = rest.FetchPost(s.Client, "api/private/cli", cliCommand, headers)
		if err != nil {
			errMsg := strings.ToLower(err.Error())

			// if ONTAP complains about batch size, use a smaller batch size
			if strings.Contains(errMsg, "retry request with less than") && s.batchSize > 100 {
				matches := retryRe.FindStringSubmatch(errMsg)
				if len(matches) == 2 {
					newBatchSize, err := strconv.Atoi(matches[1])
					if err != nil {
						s.Logger.Error("failed to parse batch size", slogx.Err(err), slog.String("errMsg", errMsg))
					} else {
						s.Logger.Warn(
							"batch size too large, reducing",
							slog.Int("currentBatchSize", s.batchSize),
							slog.Int("newBatchSize", newBatchSize),
						)
						s.batchSize = newBatchSize
					}
				} else {
					s.Logger.Error("failed to parse batch size from error message", slog.String("errMsg", errMsg))
				}
			}
			return nil, err
		}

		processBatch(records)
	}
	apiD += time.Since(startTime)

	_ = s.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = s.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = s.Metadata.LazySetValueUint64("metrics", "data", metricCount)
	_ = s.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = s.Metadata.LazySetValueUint64("bytesRx", "data", s.Client.Metadata.BytesRx)
	_ = s.Metadata.LazySetValueUint64("numCalls", "data", s.Client.Metadata.NumCalls)
	_ = s.Metadata.LazySetValueUint64("numPartials", "data", numPartials)
	s.AddCollectCount(metricCount)

	return s.cookCounters(curMat, prevMat)
}

func (s *StatPerf) processPerfRecords(records []gjson.Result, curMat *matrix.Matrix, prevMat *matrix.Matrix) (uint64, uint64, time.Duration) {
	var (
		count        uint64
		parseD       time.Duration
		instanceKeys []string
		numPartials  uint64
		err          error
	)
	instanceKeys = s.Prop.InstanceKeys
	startTime := time.Now()
	firstRecord := records[0]
	fr := firstRecord.ClonedString()
	if fr == "" {
		return 0, 0, 0
	}

	perfRecords, err := s.parseData(fr)
	if err != nil {
		s.Logger.Error("failed to parse data", slogx.Err(err))
		return 0, 0, 0
	}

	perfRecords.ForEach(func(_, data gjson.Result) bool {
		var instanceKey string
		var instanceKeyValues []string
		var instance *matrix.Instance
		if len(instanceKeys) != 0 {
			// extract instance key(s)
			for _, k := range instanceKeys {
				v := data.Get(k).ClonedString()
				if v != "" {
					instanceKeyValues = append(instanceKeyValues, v)
				} else {
					s.Logger.Warn("missing key", slog.String("key", k))
				}
			}

			instanceKey = strings.Join(instanceKeyValues, keyToken)

			if instanceKey == "" {
				return true
			}
		}

		instance = curMat.GetInstance(instanceKey)
		if instance == nil {
			instance, err = curMat.NewInstance(instanceKey)
			if err != nil {
				s.Logger.Error("add instance", slogx.Err(err), slog.String("instanceKey", instanceKey))
				return true
			}
		}

		if data.Get("_aggregation").ClonedString() == "partial_aggregation" {
			if s.AllowPartialAggregation {
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

		ts := data.Get("timestamp").ClonedString()

		data.ForEach(func(k, v gjson.Result) bool {
			var isHistogram bool
			var histogramMetric *matrix.Metric
			metricName := k.ClonedString()
			metricValue := v.ClonedString()
			if metricName == "_aggregation" || metricName == "timestamp" {
				return true
			}

			if counterInfo, exists := s.perfProp.counterInfo[metricName]; exists {
				if counterInfo.counterType == "array" {
					metric := s.Prop.Metrics[metricName]
					result := gjson.Parse(metricValue)

					var labels []string
					var values []string

					// Iterate over the keys and values
					result.ForEach(func(key, value gjson.Result) bool {
						labels = append(labels, key.ClonedString())
						values = append(values, value.ClonedString())
						return true // keep iterating
					})

					if len(labels) != len(values) {
						// warn & skip
						s.Logger.Warn(
							"labels don't match parsed values",
							slog.Any("labels", labels),
							slog.Any("value", values),
						)
						return true
					}

					// ONTAP does not have a `type` for histogram. Harvest tests the `desc` field to determine
					// if a counter is a histogram
					isHistogram = false
					description := strings.ToLower(s.perfProp.counterInfo[metricName].description)
					if strings.Contains(description, "histogram") {
						key := metricName + ".bucket"
						histogramMetric, err = collectors.GetMetric(curMat, prevMat, key, metric.Label)
						if err != nil {
							s.Logger.Error(
								"unable to create histogram metric",
								slogx.Err(err),
								slog.String("key", key),
							)
							return true
						}
						histogramMetric.SetArray(true)
						histogramMetric.SetExportable(metric.Exportable)
						histogramMetric.SetBuckets(&labels)
						isHistogram = true
					}

					for i, label := range labels {
						k := metricName + arrayKeyToken + label
						metr, ok := curMat.GetMetrics()[k]
						if !ok {
							if metr, err = collectors.GetMetric(curMat, prevMat, k, metric.Label); err != nil {
								s.Logger.Error(
									"NewMetricFloat64",
									slogx.Err(err),
									slog.String("name", k),
								)
								continue
							}
							if x := strings.Split(label, subLabelToken); len(x) == 2 {
								// order is reversed to keep it backward compatible with ZapiPerf and RestPerf
								metr.SetLabel("metric", x[1])
								metr.SetLabel("submetric", x[0])
							} else {
								metr.SetLabel("metric", label)
							}
							// differentiate between array and normal counter
							metr.SetArray(true)
							metr.SetExportable(metric.Exportable)
							if isHistogram {
								// Save the index of this label so the labels can be exported in order
								metr.SetLabel("comment", strconv.Itoa(i))
								// Save the bucket name so the flattened metrics can find their bucket when exported
								metr.SetLabel("bucket", metricName+".bucket")
								metr.SetHistogram(true)
							}
						}
						if err = metr.SetValueString(instance, values[i]); err != nil {
							s.Logger.Error(
								"Set value failed",
								slogx.Err(err),
								slog.String("name", metricName),
								slog.String("label", label),
								slog.String("value", values[i]),
							)
							continue
						}
						count++
					}
					return true
				}
			}

			if display, ok := s.Prop.InstanceLabels[metricName]; ok {
				instance.SetLabel(display, NormalizeCounterValue(metricValue))
				count++
			} else {
				if metric, ok := s.Prop.Metrics[metricName]; ok {
					metr, ok := curMat.GetMetrics()[metricName]
					if !ok {
						if metr, err = collectors.GetMetric(curMat, prevMat, metricName, metric.Label); err != nil {
							s.Logger.Error(
								"NewMetricFloat64",
								slogx.Err(err),
								slog.String("name", metricName),
							)
						}
					}
					metr.SetExportable(metric.Exportable)
					if err = metr.SetValueString(instance, metricValue); err != nil {
						s.Logger.Error(
							"Unable to set float key on metric",
							slogx.Err(err),
							slog.String("key", metric.Name),
							slog.String("metric", metric.Label),
							slog.String("metric", metric.Label),
							slog.String("value", metricValue),
						)
					}
					count++
				} else {
					s.Logger.Warn("Counter is missing or unable to parse", slog.String("counter", metricName))
				}
			}
			if err = curMat.GetMetric(timestampMetricName).SetValueString(instance, ts); err != nil {
				s.Logger.Error("Failed to set timestamp", slogx.Err(err))
			}
			return true
		})
		return true
	})
	parseD = time.Since(startTime)
	return count, numPartials, parseD
}

func (s *StatPerf) counterLookup(metric *matrix.Metric, metricKey string) *counter {
	var c *counter
	if metric.IsArray() {
		name, _, _ := strings.Cut(metricKey, arrayKeyToken)
		c = s.perfProp.counterInfo[name]
	} else {
		c = s.perfProp.counterInfo[metricKey]
	}
	return c
}

func (s *StatPerf) cookCounters(curMat *matrix.Matrix, prevMat *matrix.Matrix) (map[string]*matrix.Matrix, error) {
	var (
		err   error
		skips int
	)

	// skip calculating from delta if no data from previous poll
	if s.perfProp.isCacheEmpty {
		s.Logger.Debug("skip postprocessing until next poll (previous cache empty)")
		s.Matrix[s.Object] = curMat
		s.perfProp.isCacheEmpty = false
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
			counter := s.counterLookup(metric, key)
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
				s.Logger.Warn("Counter is missing or unable to parse", slog.String("counter", metric.GetName()))
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
	if _, err = curMat.Delta("timestamp", prevMat, cachedData, s.AllowPartialAggregation, s.Logger); err != nil {
		s.Logger.Error("(timestamp) calculate delta:", slogx.Err(err))
	}

	var base *matrix.Metric
	var totalSkips int

	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := s.counterLookup(metric, key)
		if counter == nil {
			s.Logger.Error(
				"Missing counter:",
				slogx.Err(err),
				slog.String("counter", metric.GetName()),
			)
			continue
		}
		property := counter.property
		// used in aggregator plugin
		metric.SetProperty(property)
		// used in volume.go plugin
		metric.SetComment(counter.denominator)

		// raw/string - submit without post-processing
		if property == "raw" || property == "string" {
			continue
		}

		// all other properties - first calculate delta
		if skips, err = curMat.Delta(key, prevMat, cachedData, s.AllowPartialAggregation, s.Logger); err != nil {
			s.Logger.Error("Calculate delta:", slogx.Err(err), slog.String("key", key))
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
			s.Logger.Warn(
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
				skips, err = curMat.DivideWithThreshold(key, counter.denominator, s.perfProp.latencyIoReqd, cachedData, prevMat, timestampMetricName, s.Logger)
			} else {
				skips, err = curMat.Divide(key, counter.denominator)
			}

			if err != nil {
				s.Logger.Error("Division by base", slogx.Err(err), slog.String("key", key))
				continue
			}
			totalSkips += skips

			if property == "average" {
				continue
			}
		}

		if property == "percent" {
			if skips, err = curMat.MultiplyByScalar(key, 100); err != nil {
				s.Logger.Error("Multiply by scalar", slogx.Err(err), slog.String("key", key))
			} else {
				totalSkips += skips
			}
			continue
		}
		// If we reach here, then one of the earlier clauses should have executed `continue` statement
		s.Logger.Error(
			"Unknown property",
			slog.String("key", key),
			slog.String("property", property),
		)
	}

	// calculate rates (which we deferred to calculate averages/percents first)
	for i, metric := range orderedMetrics {
		key := orderedKeys[i]
		counter := s.counterLookup(metric, key)
		if counter != nil {
			property := counter.property
			if property == "rate" {
				if skips, err = curMat.Divide(orderedKeys[i], timestampMetricName); err != nil {
					s.Logger.Error(
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
			s.Logger.Warn("Counter is missing or unable to parse", slog.String("counter", metric.GetName()))
			continue
		}
	}

	calcD := time.Since(calcStart)
	_ = s.Metadata.LazySetValueUint64("instances", "data", uint64(len(curMat.GetInstances())))
	_ = s.Metadata.LazySetValueInt64("calc_time", "data", calcD.Microseconds())
	_ = s.Metadata.LazySetValueUint64("skips", "data", uint64(totalSkips)) //nolint:gosec

	// store cache for next poll
	s.Matrix[s.Object] = cachedData

	newDataMap := make(map[string]*matrix.Matrix)
	newDataMap[s.Object] = curMat
	return newDataMap, nil
}

func (s *StatPerf) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "FlexCache":
		return flexcache.New(abc)
	default:
		s.Logger.Info("no StatPerf plugin found for %s", slog.String("kind", kind))
	}
	return nil
}

// PollInstance updates instance cache
func (s *StatPerf) PollInstance() (map[string]*matrix.Matrix, error) {
	var (
		err     error
		records []gjson.Result
	)

	s.pollInstanceCalls++
	if s.pollInstanceCalls > s.recordsToSave/3 {
		s.pollInstanceCalls = 0
	}

	var headers map[string]string

	poller, err := conf.PollerNamed(s.Options.Poller)
	if err != nil {
		slog.Error("failed to find poller", slogx.Err(err), slog.String("poller", s.Options.Poller))
	}

	if poller.IsRecording() {
		headers = map[string]string{
			"From": strconv.Itoa(s.pollInstanceCalls),
		}
	}

	cliCommand, err := clirequestbuilder.New().
		BaseSet(GetCounterInstanceBaseSet()).
		Query("statistics catalog instance show").
		Object(s.Prop.Query).
		Fields([]string{"instance", "instanceUUID"}).
		Filter(s.filter).
		Build()
	if err != nil {
		return nil, err
	}

	s.Logger.Debug("", slog.String("cliCommand", string(cliCommand)))
	if cliCommand == nil {
		return nil, errs.New(errs.ErrConfig, "empty cliCommand")
	}

	apiT := time.Now()
	s.Client.Metadata.Reset()
	records, err = rest.FetchPost(s.Client, endpoint, cliCommand, headers)
	if err != nil {
		if errs.IsRestErr(err, errs.EntryNotExist) {
			return nil, errs.New(errs.ErrNoInstance, "no "+s.Object+" instances on cluster")
		}
		return nil, err
	}

	return s.pollInstance(s.Matrix[s.Object], records, time.Since(apiT))
}

func (s *StatPerf) pollInstance(mat *matrix.Matrix, records []gjson.Result, apiD time.Duration) (map[string]*matrix.Matrix, error) {
	var (
		err                              error
		oldInstances                     *set.Set
		oldSize, newSize, removed, added int
		count                            int
	)

	s.instanceNames = set.New()
	oldInstances = set.New()
	parseT := time.Now()
	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}
	oldSize = oldInstances.Size()

	instanceKeys := s.Prop.InstanceKeys
	firstRecord := records[0]
	fr := firstRecord.ClonedString()
	if fr != "" {
		instances, err := s.parseInstances(firstRecord.ClonedString())
		if err != nil {
			return nil, err
		}
		for _, instance := range instances {
			var (
				instanceKey string
			)

			count++
			instanceKey = s.BuildInstanceKey(instance, instanceKeys)
			if instanceKey == "" {
				s.Logger.Error("empty instance key")
				continue
			}
			if oldInstances.Has(instanceKey) {
				// instance already in cache
				oldInstances.Remove(instanceKey)
				s.instanceNames.Add(instance.Instance)
			} else {
				if _, err := mat.NewInstance(instanceKey); err != nil {
					s.Logger.Error("add instance", slogx.Err(err), slog.String("instanceKey", instanceKey))
				} else {
					s.instanceNames.Add(instance.Instance)
				}
			}

		}
	}

	if count == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+s.Object+" instances on cluster")
	}

	for key := range oldInstances.Iter() {
		mat.RemoveInstance(key)
		s.Logger.Debug("removed instance", slog.String("key", key))
	}

	removed = oldInstances.Size()
	newSize = len(mat.GetInstances())
	added = newSize - (oldSize - removed)

	s.Logger.Debug("instances", slog.Int("new", added), slog.Int("removed", removed), slog.Int("total", newSize))

	// update metadata for collector logs
	_ = s.Metadata.LazySetValueInt64("api_time", "instance", apiD.Microseconds())
	_ = s.Metadata.LazySetValueInt64("parse_time", "instance", time.Since(parseT).Microseconds())
	_ = s.Metadata.LazySetValueUint64("instances", "instance", uint64(newSize))
	_ = s.Metadata.LazySetValueUint64("bytesRx", "instance", s.Client.Metadata.BytesRx)
	_ = s.Metadata.LazySetValueUint64("numCalls", "instance", s.Client.Metadata.NumCalls)

	if newSize == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+s.Object+" instances on cluster")
	}

	return nil, err
}

func (s *StatPerf) BuildInstanceKey(inst InstanceInfo, instanceKeys []string) string {
	var keyParts []string

	for _, k := range instanceKeys {
		switch strings.ToLower(k) {
		case "instance_name":
			keyParts = append(keyParts, inst.Instance)
		case "instance_uuid":
			keyParts = append(keyParts, inst.InstanceUUID)
		default:
			s.Logger.Warn("Warning: unknown key", slog.String("key", k))
		}
	}
	return strings.Join(keyParts, keyToken)
}

// Interface guards
var (
	_ collector.Collector = (*StatPerf)(nil)
)
