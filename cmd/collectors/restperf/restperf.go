package restperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/disk"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/fabricpool"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/fcp"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/fcvi"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/headroom"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/nic"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/volumetag"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/volumetopmetrics"
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/vscan"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	collector2 "github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"iter"
	"log/slog"
	"maps"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	latencyIoReqd          = 10
	arrayKeyToken          = "#"
	objWorkloadClass       = "user_defined|system_defined"
	objWorkloadVolumeClass = "autovolume"
	timestampMetricName    = "timestamp"
	idBatchSize            = 100
)

var (
	constituentRegex      = regexp.MustCompile(`^(.*)__(\d{4})$`)
	qosQuery              = "api/cluster/counter/tables/qos"
	qosVolumeQuery        = "api/cluster/counter/tables/qos_volume"
	qosDetailQuery        = "api/cluster/counter/tables/qos_detail"
	qosDetailVolumeQuery  = "api/cluster/counter/tables/qos_detail_volume"
	qosWorkloadQuery      = "api/storage/qos/workloads"
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

type RestPerf struct {
	*rest2.Rest         // provides: AbstractCollector, Client, Object, Query, TemplateFn, TemplateType
	perfProp            *perfProp
	archivedMetrics     map[string]*rest2.Metric // Keeps metric definitions that are not found in the counter schema. These metrics may be available in future ONTAP versions.
	hasInstanceSchedule bool
	pollInstanceCalls   int
	pollDataCalls       int
	recordsToSave       int // Number of records to save when using the recorder
}

type counter struct {
	name        string
	description string
	counterType string
	unit        string
	denominator string
}

type perfProp struct {
	isCacheEmpty        bool
	counterInfo         map[string]*counter
	latencyIoReqd       int
	qosLabels           map[string]string
	disableConstituents bool
}

type metricResponse struct {
	label   string
	value   string
	isArray bool
}

func init() {
	plugin.RegisterModule(&RestPerf{})
}

func (r *RestPerf) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.restperf",
		New: func() plugin.Module { return new(RestPerf) },
	}
}

func (r *RestPerf) Init(a *collector.AbstractCollector) error {

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
		slog.String("timeout", r.Client.Timeout.String()),
	)

	return nil
}

func (r *RestPerf) InitQOS() error {
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

func (r *RestPerf) InitMatrix() error {
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

// load workload_class or use defaultValue
func (r *RestPerf) loadWorkloadClassQuery(defaultValue string) string {

	var x *node.Node

	name := "workload_class"

	if x = r.Params.GetChildS(name); x != nil {
		v := x.GetAllChildContentS()
		if len(v) == 0 {
			r.Logger.Debug(
				"",
				slog.String("name", name),
				slog.String("defaultValue", defaultValue),
			)
			return defaultValue
		}
		slices.Sort(v)
		s := strings.Join(v, "|")
		r.Logger.Debug("", slog.String("name", name), slog.String("value", s))
		return s
	}
	r.Logger.Debug("", slog.String("name", name), slog.String("defaultValue", defaultValue))
	return defaultValue
}

// load an int parameter or use defaultValue
func (r *RestPerf) loadParamInt(name string, defaultValue int) int {

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

func (r *RestPerf) PollCounter() (map[string]*matrix.Matrix, error) {
	var (
		err     error
		records []gjson.Result
	)

	href := rest.NewHrefBuilder().
		APIPath(r.Prop.Query).
		MaxRecords(r.BatchSize).
		ReturnTimeout(r.Prop.ReturnTimeOut).
		Build()
	r.Logger.Debug("", slog.String("href", href))
	if href == "" {
		return nil, errs.New(errs.ErrConfig, "empty url")
	}

	apiT := time.Now()
	r.Client.Metadata.Reset()

	records, err = rest.FetchAll(r.Client, href)
	if err != nil {
		return r.handleError(err, href)
	}

	return r.pollCounter(records, time.Since(apiT))
}

func (r *RestPerf) pollCounter(records []gjson.Result, apiD time.Duration) (map[string]*matrix.Matrix, error) {
	var (
		err           error
		counterSchema gjson.Result
		parseT        time.Time
	)

	mat := r.Matrix[r.Object]
	firstRecord := records[0]

	parseT = time.Now()

	if firstRecord.Exists() {
		counterSchema = firstRecord.Get("counter_schemas")
	} else {
		return nil, errs.New(errs.ErrConfig, "no data found")
	}
	seenMetrics := make(map[string]bool)

	// populate denominator metric to prop metrics
	counterSchema.ForEach(func(_, c gjson.Result) bool {
		if !c.IsObject() {
			r.Logger.Warn("Counter is not object, skipping", slog.String("type", c.Type.String()))
			return true
		}

		name := c.Get("name").ClonedString()
		dataType := c.Get("type").ClonedString()

		if p := r.GetOverride(name); p != "" {
			dataType = p
		}

		// Check if the metric was previously archived and restore it
		if archivedMetric, found := r.archivedMetrics[name]; found {
			r.Prop.Metrics[name] = archivedMetric
			delete(r.archivedMetrics, name) // Remove from archive after restoring
			r.Logger.Info("Metric found in archive. Restore it", slog.String("key", name))
		}

		if _, has := r.Prop.Metrics[name]; has {
			if strings.Contains(dataType, "string") {
				if _, ok := r.Prop.InstanceLabels[name]; !ok {
					r.Prop.InstanceLabels[name] = r.Prop.Counters[name]
				}
				// set exportable as false
				r.Prop.Metrics[name].Exportable = false
				return true
			}
			d := c.Get("denominator.name").ClonedString()
			if d != "" {
				if _, has := r.Prop.Metrics[d]; !has {
					if isWorkloadDetailObject(r.Prop.Query) {
						// It is not needed because 'ops' is used as the denominator in latency calculations.
						if d == "visits" {
							return true
						}
					}
					// export false
					m := &rest2.Metric{Label: "", Name: d, MetricType: "", Exportable: false}
					r.Prop.Metrics[d] = m
				}
			}
		}
		return true
	})

	counterSchema.ForEach(func(_, c gjson.Result) bool {

		if !c.IsObject() {
			r.Logger.Warn("Counter is not object, skipping", slog.String("type", c.Type.String()))
			return true
		}

		name := c.Get("name").ClonedString()
		if _, has := r.Prop.Metrics[name]; has {
			seenMetrics[name] = true
			if _, ok := r.perfProp.counterInfo[name]; !ok {
				r.perfProp.counterInfo[name] = &counter{
					name:        name,
					description: c.Get("description").ClonedString(),
					counterType: c.Get("type").ClonedString(),
					unit:        c.Get("unit").ClonedString(),
					denominator: c.Get("denominator.name").ClonedString(),
				}
				if p := r.GetOverride(name); p != "" {
					r.perfProp.counterInfo[name].counterType = p
				}
			}
		}

		return true
	})

	for name, metric := range r.Prop.Metrics {
		if !seenMetrics[name] {
			r.archivedMetrics[name] = metric
			// Log the metric that is not present in counterSchema.
			r.Logger.Warn("Metric not found in counterSchema", slog.String("key", name))
			delete(r.Prop.Metrics, name)
		}
	}

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

	_, err = r.processWorkLoadCounter()
	if err != nil {
		return nil, err
	}

	// update metadata for collector logs
	_ = r.Metadata.LazySetValueInt64("api_time", "counter", apiD.Microseconds())
	_ = r.Metadata.LazySetValueInt64("parse_time", "counter", time.Since(parseT).Microseconds())
	_ = r.Metadata.LazySetValueUint64("metrics", "counter", uint64(len(r.perfProp.counterInfo)))
	_ = r.Metadata.LazySetValueUint64("bytesRx", "counter", r.Client.Metadata.BytesRx)
	_ = r.Metadata.LazySetValueUint64("numCalls", "counter", r.Client.Metadata.NumCalls)

	return nil, nil
}

func parseProps(instanceData gjson.Result) map[string]gjson.Result {
	var props = map[string]gjson.Result{
		"id": gjson.Get(instanceData.ClonedString(), "id"),
	}

	instanceData.ForEach(func(key, v gjson.Result) bool {
		keyS := key.ClonedString()
		if keyS == "properties" {
			v.ForEach(func(_, each gjson.Result) bool {
				key := each.Get("name").ClonedString()
				value := each.Get("value")
				props[key] = value
				return true
			})
			return false
		}
		return true
	})
	return props
}

func parseProperties(instanceData gjson.Result, property string) gjson.Result {
	var (
		result gjson.Result
	)

	if property == "id" {
		value := gjson.Get(instanceData.ClonedString(), "id")
		return value
	}

	instanceData.ForEach(func(key, v gjson.Result) bool {
		keyS := key.ClonedString()
		if keyS == "properties" {
			v.ForEach(func(_, each gjson.Result) bool {
				if each.Get("name").ClonedString() == property {
					value := each.Get("value")
					result = value
					return false
				}
				return true
			})
			return false
		}
		return true
	})

	return result
}

func parseMetricResponses(instanceData gjson.Result, metric map[string]*rest2.Metric) map[string]*metricResponse {
	var (
		mapMetricResponses = make(map[string]*metricResponse)
		numWant            = len(metric)
		numSeen            = 0
	)
	instanceData.ForEach(func(key, v gjson.Result) bool {
		keyS := key.ClonedString()
		if keyS == "counters" {
			v.ForEach(func(_, each gjson.Result) bool {
				if numSeen == numWant {
					return false
				}
				name := each.Get("name").ClonedString()
				_, ok := metric[name]
				if !ok {
					return true
				}
				value := each.Get("value").ClonedString()
				if value != "" {
					mapMetricResponses[name] = &metricResponse{value: value, label: ""}
					numSeen++
					return true
				}
				values := each.Get("values").ClonedString()
				labels := each.Get("labels").ClonedString()
				if values != "" {
					mapMetricResponses[name] = &metricResponse{
						value:   template.ArrayMetricToString(values),
						label:   template.ArrayMetricToString(labels),
						isArray: true,
					}
					numSeen++
					return true
				}
				subCounters := each.Get("counters")
				if !subCounters.IsArray() {
					return true
				}

				// handle sub metrics
				subLabelsS := labels
				subLabelsS = template.ArrayMetricToString(subLabelsS)
				subLabelSlice := strings.Split(subLabelsS, ",")
				var finalLabels []string
				var finalValues []string
				subCounters.ForEach(func(_, subCounter gjson.Result) bool {
					label := subCounter.Get("label").ClonedString()
					subValues := subCounter.Get("values").ClonedString()
					m := template.ArrayMetricToString(subValues)
					ms := strings.Split(m, ",")
					if len(ms) > len(subLabelSlice) {
						return false
					}
					for i := range ms {
						finalLabels = append(finalLabels, subLabelSlice[i]+arrayKeyToken+label)
					}
					finalValues = append(finalValues, ms...)
					return true
				})
				if len(finalLabels) == len(finalValues) {
					mr := metricResponse{
						value:   strings.Join(finalValues, ","),
						label:   strings.Join(finalLabels, ","),
						isArray: true,
					}
					mapMetricResponses[name] = &mr
				}
				return true
			})
		}
		return true
	})
	return mapMetricResponses
}

func parseMetricResponse(instanceData gjson.Result, metric string) *metricResponse {
	instanceDataS := instanceData.ClonedString()
	t := gjson.Get(instanceDataS, "counters.#.name")
	for _, name := range t.Array() {
		if name.ClonedString() != metric {
			continue
		}
		metricPath := "counters.#(name=" + metric + ")"
		many := gjson.Parse(instanceDataS)
		value := many.Get(metricPath + ".value")
		values := many.Get(metricPath + ".values")
		labels := many.Get(metricPath + ".labels")
		subLabels := many.Get(metricPath + ".counters.#.label")
		subValues := many.Get(metricPath + ".counters.#.values")
		if value.ClonedString() != "" {
			return &metricResponse{value: value.ClonedString(), label: ""}
		}
		if values.ClonedString() != "" {
			return &metricResponse{
				value:   template.ArrayMetricToString(values.ClonedString()),
				label:   template.ArrayMetricToString(labels.ClonedString()),
				isArray: true,
			}
		}

		// check for sub metrics
		if subLabels.ClonedString() != "" {
			var finalLabels []string
			var finalValues []string
			subLabelsS := labels.ClonedString()
			subLabelsS = template.ArrayMetricToString(subLabelsS)
			subLabelSlice := strings.Split(subLabelsS, ",")
			ls := subLabels.Array()
			vs := subValues.Array()
			var vLen int
			for i, v := range vs {
				label := ls[i].ClonedString()
				m := template.ArrayMetricToString(v.ClonedString())
				ms := strings.Split(m, ",")
				for range ms {
					finalLabels = append(finalLabels, label+arrayKeyToken+subLabelSlice[vLen])
					vLen++
				}
				if vLen > len(subLabelSlice) {
					break
				}
				finalValues = append(finalValues, ms...)
			}
			if vLen == len(subLabelSlice) {
				return &metricResponse{value: strings.Join(finalValues, ","), label: strings.Join(finalLabels, ","), isArray: true}
			}
		}
	}
	return &metricResponse{}
}

// GetOverride override counter property
func (r *RestPerf) GetOverride(counter string) string {
	if o := r.Params.GetChildS("override"); o != nil {
		return o.GetChildContentS(counter)
	}
	return ""
}

func (r *RestPerf) processWorkLoadCounter() (map[string]*matrix.Matrix, error) {
	var err error
	mat := r.Matrix[r.Object]
	// for these two objects, we need to create latency/ops counters for each of the workload layers
	// their original counters will be discarded
	if isWorkloadDetailObject(r.Prop.Query) {

		for name, metric := range r.Prop.Metrics {
			metr, ok := mat.GetMetrics()[name]
			if !ok {
				if metr, err = mat.NewMetricFloat64(name, metric.Label); err != nil {
					r.Logger.Error("NewMetricFloat64", slogx.Err(err), slog.String("name", name))
				}
			}
			metr.SetExportable(metric.Exportable)
		}

		var service, wait, ops *matrix.Metric

		if service = mat.GetMetric("service_time"); service == nil {
			r.Logger.Error("metric [service_time] required to calculate workload missing")
		}

		if wait = mat.GetMetric("wait_time"); wait == nil {
			r.Logger.Error("metric [wait-time] required to calculate workload missing")
		}

		if service == nil || wait == nil {
			return nil, errs.New(errs.ErrMissingParam, "workload metrics")
		}

		if ops = mat.GetMetric("ops"); ops == nil {
			if _, err = mat.NewMetricFloat64("ops"); err != nil {
				return nil, err
			}
			r.perfProp.counterInfo["ops"] = &counter{
				name:        "ops",
				description: "",
				counterType: "rate",
				unit:        "per_sec",
				denominator: "",
			}
		}

		service.SetExportable(false)
		wait.SetExportable(false)

		resourceMap := r.Params.GetChildS("resource_map")
		if resourceMap == nil {
			return nil, errs.New(errs.ErrMissingParam, "resource_map")
		}
		for _, x := range resourceMap.GetChildren() {
			for _, wm := range workloadDetailMetrics {
				name := x.GetNameS() + wm
				resource := x.GetContentS()

				if m := mat.GetMetric(name); m != nil {
					continue
				}
				m, err := mat.NewMetricFloat64(name, wm)
				if err != nil {
					return nil, err
				}
				r.perfProp.counterInfo[name] = &counter{
					name:        wm,
					description: "",
					counterType: r.perfProp.counterInfo[service.GetName()].counterType,
					unit:        r.perfProp.counterInfo[service.GetName()].unit,
					denominator: "ops",
				}
				m.SetLabel("resource", resource)
			}
		}
	}
	return r.Matrix, nil
}

// batchIDs splits the id values into smaller batches
func batchIDs(ids []string) [][]string {
	var batches [][]string
	for i := 0; i < len(ids); i += idBatchSize {
		end := i + idBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		batches = append(batches, ids[i:end])
	}
	return batches
}

func (r *RestPerf) PollData() (map[string]*matrix.Matrix, error) {
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

	dataQuery := path.Join(r.Prop.Query, "rows")

	var filter []string
	// Sort metrics so that the href is deterministic
	metrics := slices.Sorted(maps.Keys(r.Prop.Metrics))

	filter = append(filter, "counters.name="+strings.Join(metrics, "|"))

	if isWorkloadDetailObject(r.Prop.Query) {
		resourceMap := r.Params.GetChildS("resource_map")
		if resourceMap == nil {
			return nil, errs.New(errs.ErrMissingParam, "resource_map")
		}
		workloadDetailFilter := make([]string, 0, len(resourceMap.GetChildren()))
		for _, x := range resourceMap.GetChildren() {
			workloadDetailFilter = append(workloadDetailFilter, x.GetNameS())
		}
		if len(workloadDetailFilter) > 0 {
			filter = append(filter,
				"query="+strings.Join(workloadDetailFilter, "|"),
				"query_fields=properties.value")
		}
	}

	// filter is applied to api/storage/qos/workloads for workload objects in pollInstance method
	if !isWorkloadObject(r.Prop.Query) && !isWorkloadDetailObject(r.Prop.Query) {
		filter = append(filter, r.Prop.Filter...)
	}

	// No batching needed, use a single batch with all ids
	idBatches := [][]string{nil}
	if isWorkloadObject(r.Prop.Query) || isWorkloadDetailObject(r.Prop.Query) {
		// if filtering is enabled then collected uuid in pollInstance need to be passed as filter as id
		if r.Prop.Filter != nil {
			if isWorkloadObject(r.Prop.Query) {
				// id is of form workloadName:uuid
				// Test-wid13148:8333dc06-8b7a-11ed-86dd-00a098d390f2
				var ids []string
				for uuid, instance := range r.Matrix[r.Object].GetInstances() {
					id := instance.GetLabel("workload") + ":" + uuid
					ids = append(ids, id)
				}
				idBatches = batchIDs(ids)
			} else if isWorkloadDetailObject(r.Prop.Query) {
				// id is of form node:workloadName:subsystemname
				// umeng-aff300-02:Test-wid8678.CPU_dblade
				var ids []string
				for _, instance := range r.Matrix[r.Object].GetInstances() {
					id := "*" + instance.GetLabel("workload") + "*"
					ids = append(ids, id)
				}
				idBatches = batchIDs(ids)
			}
		}
	}

	for _, idBatch := range idBatches {
		if idBatch != nil {
			idFilter := "id=" + strings.Join(idBatch, "|")
			filter = append(filter, idFilter)
		}
		href := rest.NewHrefBuilder().
			APIPath(dataQuery).
			Fields([]string{"*"}).
			Filter(filter).
			MaxRecords(r.BatchSize).
			ReturnTimeout(r.Prop.ReturnTimeOut).
			Build()

		r.Logger.Debug("", slog.String("href", href))
		if href == "" {
			return nil, errs.New(errs.ErrConfig, "empty url")
		}

		r.pollDataCalls++
		if r.pollDataCalls >= r.recordsToSave {
			r.pollDataCalls = 0
		}

		var headers map[string]string

		poller, err := conf.PollerNamed(r.Options.Poller)
		if err != nil {
			slog.Error("failed to find poller", slogx.Err(err), slog.String("poller", r.Options.Poller))
		}

		if poller.IsRecording() {
			headers = map[string]string{
				"From": strconv.Itoa(r.pollDataCalls),
			}
		}

		// Track old instances before processing batches
		oldInstances := set.New()
		// The creation and deletion of objects with an instance schedule are managed through pollInstance.
		if !r.hasInstanceSchedule {
			for key := range r.Matrix[r.Object].GetInstances() {
				oldInstances.Add(key)
			}
		}

		prevMat = r.Matrix[r.Object]
		// clone matrix without numeric data
		curMat = prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
		curMat.Reset()

		processBatch := func(perfRecords []rest.PerfRecord) error {
			if len(perfRecords) == 0 {
				return nil
			}

			// Process the current batch of records
			count, np, batchParseD := r.processPerfRecords(perfRecords, curMat, prevMat, oldInstances)
			numPartials += np
			metricCount += count
			parseD += batchParseD
			return nil
		}

		err = rest.FetchRestPerfDataStream(r.Client, href, processBatch, headers)
		apiD += time.Since(startTime)

		if err != nil {
			return nil, err
		}

		if !r.hasInstanceSchedule {
			// Remove old instances that are not found in new instances
			for key := range oldInstances.Iter() {
				curMat.RemoveInstance(key)
			}
		}
	}

	if isWorkloadDetailObject(r.Prop.Query) {
		if err := r.getParentOpsCounters(curMat); err != nil {
			// no point to continue as we can't calculate the other counters
			return nil, err
		}
	}

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

func (r *RestPerf) processPerfRecords(perfRecords []rest.PerfRecord, curMat *matrix.Matrix, prevMat *matrix.Matrix, oldInstances *set.Set) (uint64, uint64, time.Duration) {
	var (
		count        uint64
		parseD       time.Duration
		instanceKeys []string
		numPartials  uint64
		ts           float64
		err          error
	)
	instanceKeys = r.Prop.InstanceKeys
	startTime := time.Now()

	// init current time
	ts = float64(startTime.UnixNano()) / collector2.BILLION
	for _, perfRecord := range perfRecords {
		pr := perfRecord.Records
		t := perfRecord.Timestamp

		if t != 0 {
			ts = float64(t) / collector2.BILLION
		} else {
			r.Logger.Warn("Missing timestamp in response")
		}

		pr.ForEach(func(_, instanceData gjson.Result) bool {
			var (
				instanceKey     string
				instance        *matrix.Instance
				isHistogram     bool
				histogramMetric *matrix.Metric
			)

			if !instanceData.IsObject() {
				r.Logger.Warn("Instance data is not object, skipping", slog.String("type", instanceData.Type.String()))
				return true
			}

			props := parseProps(instanceData)

			if len(instanceKeys) != 0 {
				// extract instance key(s)
				for _, k := range instanceKeys {
					value, ok := props[k]
					if ok {
						instanceKey += value.ClonedString()
					} else {
						r.Logger.Warn("missing key", slog.String("key", k))
					}
				}

				if instanceKey == "" {
					return true
				}
			}

			var layer = "" // latency layer (resource) for workloads

			// special case for these two objects
			// we need to process each latency layer for each instance/counter
			if isWorkloadDetailObject(r.Prop.Query) {

				// example instanceKey : umeng-aff300-02:test-wid12022.CPU_dblade
				i := strings.Index(instanceKey, ":")
				instanceKey = instanceKey[i+1:]
				before, after, found := strings.Cut(instanceKey, ".")
				if found {
					instanceKey = before
					layer = after
				} else {
					r.Logger.Warn("instanceKey has unexpected format", slog.String("instanceKey", instanceKey))
					return true
				}

				for _, wm := range workloadDetailMetrics {
					mLayer := layer + wm
					if l := curMat.GetMetric(mLayer); l == nil {
						return true
					}
				}
			}

			if r.Params.GetChildContentS("only_cluster_instance") != "true" {
				if instanceKey == "" {
					return true
				}
			}

			instance = curMat.GetInstance(instanceKey)
			if instance == nil {
				if isWorkloadObject(r.Prop.Query) || isWorkloadDetailObject(r.Prop.Query) {
					return true
				}
				instance, err = curMat.NewInstance(instanceKey)
				if err != nil {
					r.Logger.Error("add instance", slogx.Err(err), slog.String("instanceKey", instanceKey))
					return true
				}
			}

			oldInstances.Remove(instanceKey)

			// check for partial aggregation
			if instanceData.Get("aggregation.complete").ClonedString() == "false" {
				if r.AllowPartialAggregation {
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

			for label, display := range r.Prop.InstanceLabels {
				value, ok := props[label]
				if ok {
					if value.IsArray() {
						var labelArray []string
						value.ForEach(func(_, r gjson.Result) bool {
							labelString := r.ClonedString()
							labelArray = append(labelArray, labelString)
							return true
						})
						instance.SetLabel(display, strings.Join(labelArray, ","))
					} else {
						instance.SetLabel(display, value.ClonedString())
					}
					count++
				} else {
					// check for label value in metric
					f := parseMetricResponse(instanceData, label)
					if f.value != "" {
						instance.SetLabel(display, f.value)
						count++
					} else {
						// ignore physical_disk_id logging as in some of 9.12 versions, this field may be absent
						if r.Prop.Query == "api/cluster/counter/tables/disk:constituent" && label == "physical_disk_id" {
							r.Logger.Debug("Missing label value", slog.String("instanceKey", instanceKey), slog.String("label", label))
						} else {
							r.Logger.Error("Missing label value", slog.String("instanceKey", instanceKey), slog.String("label", label))
						}
					}
				}
			}

			metricResponses := parseMetricResponses(instanceData, r.Prop.Metrics)

			for name, metric := range r.Prop.Metrics {
				f, ok := metricResponses[name]
				if ok {
					// special case for workload_detail
					if isWorkloadDetailObject(r.Prop.Query) {
						for _, wm := range workloadDetailMetrics {
							wMetric := curMat.GetMetric(layer + wm)
							switch {
							case wm == "resource_latency" && (name == "wait_time" || name == "service_time"):
								if err := wMetric.AddValueString(instance, f.value); err != nil {
									r.Logger.Error(
										"Add resource_latency failed",
										slogx.Err(err),
										slog.String("name", name),
										slog.String("value", f.value),
									)
								} else {
									count++
								}
								continue
							case wm == "service_time_latency" && name == "service_time":
								if err = wMetric.SetValueString(instance, f.value); err != nil {
									r.Logger.Error(
										"Add service_time_latency failed",
										slogx.Err(err),
										slog.String("name", name),
										slog.String("value", f.value),
									)
								} else {
									count++
								}
							case wm == "wait_time_latency" && name == "wait_time":
								if err = wMetric.SetValueString(instance, f.value); err != nil {
									r.Logger.Error(
										"Add wait_time_latency failed",
										slogx.Err(err),
										slog.String("name", name),
										slog.String("value", f.value),
									)
								} else {
									count++
								}
							}
						}
						continue
					}
					if f.isArray {
						labels := strings.Split(f.label, ",")
						values := strings.Split(f.value, ",")

						if len(labels) != len(values) {
							// warn & skip
							r.Logger.Warn(
								"labels don't match parsed values",
								slog.String("labels", f.label),
								slog.String("value", f.value),
							)
							continue
						}

						// ONTAP does not have a `type` for histogram. Harvest tests the `desc` field to determine
						// if a counter is a histogram
						isHistogram = false
						description := strings.ToLower(r.perfProp.counterInfo[name].description)
						if len(labels) > 0 && strings.Contains(description, "histogram") {
							key := name + ".bucket"
							histogramMetric, err = collectors.GetMetric(curMat, prevMat, key, metric.Label)
							if err != nil {
								r.Logger.Error(
									"unable to create histogram metric",
									slogx.Err(err),
									slog.String("key", key),
								)
								continue
							}
							histogramMetric.SetArray(true)
							histogramMetric.SetExportable(metric.Exportable)
							histogramMetric.SetBuckets(&labels)
							isHistogram = true
						}

						for i, label := range labels {
							k := name + arrayKeyToken + label
							metr, ok := curMat.GetMetrics()[k]
							if !ok {
								if metr, err = collectors.GetMetric(curMat, prevMat, k, metric.Label); err != nil {
									r.Logger.Error(
										"NewMetricFloat64",
										slogx.Err(err),
										slog.String("name", k),
									)
									continue
								}
								if x := strings.Split(label, arrayKeyToken); len(x) == 2 {
									metr.SetLabel("metric", x[0])
									metr.SetLabel("submetric", x[1])
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
									metr.SetLabel("bucket", name+".bucket")
									metr.SetHistogram(true)
								}
							}
							if err = metr.SetValueString(instance, values[i]); err != nil {
								r.Logger.Error(
									"Set value failed",
									slogx.Err(err),
									slog.String("name", name),
									slog.String("label", label),
									slog.String("value", values[i]),
								)
								continue
							}
							count++
						}
					} else {
						metr, ok := curMat.GetMetrics()[name]
						if !ok {
							if metr, err = collectors.GetMetric(curMat, prevMat, name, metric.Label); err != nil {
								r.Logger.Error(
									"NewMetricFloat64",
									slogx.Err(err),
									slog.String("name", name),
								)
							}
						}
						metr.SetExportable(metric.Exportable)
						if c, err := strconv.ParseFloat(f.value, 64); err == nil {
							metr.SetValueFloat64(instance, c)
						} else {
							r.Logger.Error(
								"Unable to parse float value",
								slogx.Err(err),
								slog.String("key", metric.Name),
								slog.String("metric", metric.Label),
							)
						}
						count++
					}
				} else {
					r.Logger.Warn("Counter is missing or unable to parse", slog.String("counter", name))
				}
			}
			curMat.GetMetric(timestampMetricName).SetValueFloat64(instance, ts)

			return true
		})
	}
	parseD = time.Since(startTime)
	return count, numPartials, parseD
}

func (r *RestPerf) cookCounters(curMat *matrix.Matrix, prevMat *matrix.Matrix) (map[string]*matrix.Matrix, error) {
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

func perfToJSON(records []rest.PerfRecord) iter.Seq[gjson.Result] {
	return func(yield func(gjson.Result) bool) {
		for _, record := range records {
			if record.Records.IsArray() {
				record.Records.ForEach(func(_, r gjson.Result) bool {
					return yield(r)
				})
			}
		}
	}
}

// Poll counter "ops" of the related/parent object, required for objects
// workload_detail and workload_detail_volume. This counter is already
// collected by the other collectors, so this poll is redundant
// (until we implement some sort of inter-collector communication).
func (r *RestPerf) getParentOpsCounters(data *matrix.Matrix) error {

	var (
		ops       *matrix.Metric
		object    string
		dataQuery string
		err       error
		records   []gjson.Result
	)

	if r.Prop.Query == qosDetailQuery {
		dataQuery = path.Join(qosQuery, "rows")
		object = "qos"
	} else {
		dataQuery = path.Join(qosVolumeQuery, "rows")
		object = "qos_volume"
	}

	if ops = data.GetMetric("ops"); ops == nil {
		r.Logger.Error("ops counter not found in cache")
		return errs.New(errs.ErrMissingParam, "counter ops")
	}

	var filter []string
	filter = append(filter, "counters.name=ops")
	href := rest.NewHrefBuilder().
		APIPath(dataQuery).
		Fields([]string{"*"}).
		Filter(filter).
		ReturnTimeout(r.Prop.ReturnTimeOut).
		Build()

	r.Logger.Debug("", slog.String("href", href))
	if href == "" {
		return errs.New(errs.ErrConfig, "empty url")
	}

	records, err = rest.FetchAll(r.Client, href)
	if err != nil {
		r.Logger.Error("Failed to fetch data", slogx.Err(err), slog.String("href", href))
		return err
	}

	if len(records) == 0 {
		return errs.New(errs.ErrNoInstance, "no "+object+" instances on cluster")
	}

	for _, instanceData := range records {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			r.Logger.Warn("Instance data is not object, skipping", slog.String("type", instanceData.Type.String()))
			continue
		}

		value := parseProperties(instanceData, "name")
		if value.Exists() {
			instanceKey += value.ClonedString()
		} else {
			r.Logger.Warn("skip instance, missing key", slog.String("key", "name"))
			continue
		}
		instance = data.GetInstance(instanceKey)
		if instance == nil {
			continue
		}

		counterName := "ops"
		f := parseMetricResponse(instanceData, counterName)
		if f.value != "" {
			if err = ops.SetValueString(instance, f.value); err != nil {
				r.Logger.Error(
					"set metric",
					slogx.Err(err),
					slog.String("metric", counterName),
					slog.String("value", value.ClonedString()),
				)
			}
		}
	}

	return nil
}

func (r *RestPerf) counterLookup(metric *matrix.Metric, metricKey string) *counter {
	var c *counter

	if metric.IsArray() {
		name, _, _ := strings.Cut(metricKey, arrayKeyToken)
		c = r.perfProp.counterInfo[name]
	} else {
		c = r.perfProp.counterInfo[metricKey]
	}
	return c
}

func (r *RestPerf) LoadPlugin(kind string, p *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Nic":
		return nic.New(p)
	case "Fcp":
		return fcp.New(p)
	case "Headroom":
		return headroom.New(p)
	case "Volume":
		return volume.New(p)
	case "VolumeTag":
		return volumetag.New(p)
	case "VolumeTopClients":
		return volumetopmetrics.New(p)
	case "Disk":
		return disk.New(p)
	case "Vscan":
		return vscan.New(p)
	case "FabricPool":
		return fabricpool.New(p)
	case "FCVI":
		return fcvi.New(p)
	default:
		r.Logger.Info("no Restperf plugin found", slog.String("kind", kind))
	}
	return nil
}

// PollInstance updates instance cache
func (r *RestPerf) PollInstance() (map[string]*matrix.Matrix, error) {
	var (
		err     error
		records []gjson.Result
	)

	// The PollInstance method is only needed for `workload` and `workload_detail` objects.
	if !isWorkloadObject(r.Prop.Query) && !isWorkloadDetailObject(r.Prop.Query) {
		return nil, nil
	}

	dataQuery := path.Join(r.Prop.Query, "rows")
	fields := "properties"
	var filter []string

	if isWorkloadObject(r.Prop.Query) || isWorkloadDetailObject(r.Prop.Query) {
		fields = "*"
		dataQuery = qosWorkloadQuery
		if r.Prop.Query == qosVolumeQuery || r.Prop.Query == qosDetailVolumeQuery {
			filter = append(filter, "workload_class="+r.loadWorkloadClassQuery(objWorkloadVolumeClass))
		} else {
			filter = append(filter, "workload_class="+r.loadWorkloadClassQuery(objWorkloadClass))
		}
		filter = append(filter, r.Prop.Filter...)
	}

	r.pollInstanceCalls++
	if r.pollInstanceCalls > r.recordsToSave/3 {
		r.pollInstanceCalls = 0
	}

	var headers map[string]string

	poller, err := conf.PollerNamed(r.Options.Poller)
	if err != nil {
		slog.Error("failed to find poller", slogx.Err(err), slog.String("poller", r.Options.Poller))
	}

	if poller.IsRecording() {
		headers = map[string]string{
			"From": strconv.Itoa(r.pollInstanceCalls),
		}
	}

	href := rest.NewHrefBuilder().
		APIPath(dataQuery).
		Fields([]string{fields}).
		Filter(filter).
		MaxRecords(r.BatchSize).
		ReturnTimeout(r.Prop.ReturnTimeOut).
		Build()

	r.Logger.Debug("", slog.String("href", href))
	if href == "" {
		return nil, errs.New(errs.ErrConfig, "empty url")
	}

	apiT := time.Now()
	r.Client.Metadata.Reset()
	records, err = rest.FetchAll(r.Client, href, headers)
	if err != nil {
		return r.handleError(err, href)
	}

	return r.pollInstance(r.Matrix[r.Object], slices.Values(records), time.Since(apiT))
}

func (r *RestPerf) pollInstance(mat *matrix.Matrix, records iter.Seq[gjson.Result], apiD time.Duration) (map[string]*matrix.Matrix, error) {
	var (
		err                              error
		oldInstances                     *set.Set
		oldSize, newSize, removed, added int
		count                            int
	)

	oldInstances = set.New()
	parseT := time.Now()
	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}
	oldSize = oldInstances.Size()

	instanceKeys := r.Prop.InstanceKeys
	if isWorkloadObject(r.Prop.Query) {
		instanceKeys = []string{"uuid"}
	}
	if isWorkloadDetailObject(r.Prop.Query) {
		instanceKeys = []string{"name"}
	}

	for instanceData := range records {
		var (
			instanceKey string
		)

		count++

		if !instanceData.IsObject() {
			r.Logger.Warn("Instance data is not object, skipping", slog.String("type", instanceData.Type.String()))
			continue
		}

		if isWorkloadObject(r.Prop.Query) || isWorkloadDetailObject(r.Prop.Query) {
			// The API endpoint api/storage/qos/workloads lacks an is_constituent filter, unlike qos-workload-get-iter. As a result, we must perform client-side filtering.
			// Although the api/private/cli/qos/workload endpoint includes this filter, it doesn't provide an option to fetch all records, both constituent and flexgroup types.
			if r.perfProp.disableConstituents {
				if constituentRegex.MatchString(instanceData.Get("volume").ClonedString()) {
					// skip constituent
					continue
				}
			}
		}

		// extract instance key(s)
		for _, k := range instanceKeys {
			var value gjson.Result
			if isWorkloadObject(r.Prop.Query) || isWorkloadDetailObject(r.Prop.Query) {
				value = instanceData.Get(k)
			} else {
				value = parseProperties(instanceData, k)
			}
			if value.Exists() {
				instanceKey += value.ClonedString()
			} else {
				r.Logger.Warn("skip instance, missing key", slog.String("key", k))
				break
			}
		}

		if oldInstances.Has(instanceKey) {
			// instance already in cache
			oldInstances.Remove(instanceKey)
			instance := mat.GetInstance(instanceKey)
			r.updateQosLabels(instanceData, instance)
		} else if instance, err := mat.NewInstance(instanceKey); err != nil {
			r.Logger.Error("add instance", slogx.Err(err), slog.String("instanceKey", instanceKey))
		} else {
			r.updateQosLabels(instanceData, instance)
		}
	}

	if count == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+r.Object+" instances on cluster")
	}

	for key := range oldInstances.Iter() {
		mat.RemoveInstance(key)
		r.Logger.Debug("removed instance", slog.String("key", key))
	}

	removed = oldInstances.Size()
	newSize = len(mat.GetInstances())
	added = newSize - (oldSize - removed)

	r.Logger.Debug("instances", slog.Int("new", added), slog.Int("removed", removed), slog.Int("total", newSize))

	// update metadata for collector logs
	_ = r.Metadata.LazySetValueInt64("api_time", "instance", apiD.Microseconds())
	_ = r.Metadata.LazySetValueInt64("parse_time", "instance", time.Since(parseT).Microseconds())
	_ = r.Metadata.LazySetValueUint64("instances", "instance", uint64(newSize))
	_ = r.Metadata.LazySetValueUint64("bytesRx", "instance", r.Client.Metadata.BytesRx)
	_ = r.Metadata.LazySetValueUint64("numCalls", "instance", r.Client.Metadata.NumCalls)

	if newSize == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+r.Object+" instances on cluster")
	}

	return nil, err
}

func (r *RestPerf) updateQosLabels(qos gjson.Result, instance *matrix.Instance) {
	if isWorkloadObject(r.Prop.Query) || isWorkloadDetailObject(r.Prop.Query) {
		for label, display := range r.perfProp.qosLabels {
			// lun,file,qtree may not always exist for workload
			if value := qos.Get(label); value.Exists() {
				instance.SetLabel(display, value.ClonedString())
			}
		}
	}
}

func (r *RestPerf) handleError(err error, href string) (map[string]*matrix.Matrix, error) {
	if errs.IsRestErr(err, errs.TableNotFound) || errs.IsRestErr(err, errs.APINotFound) {
		// the table or API does not exist. return ErrAPIRequestRejected so the task goes to stand-by
		return nil, fmt.Errorf(
			"polling href=[%s] err: %w",
			collectors.TruncateURL(href),
			errs.New(errs.ErrAPIRequestRejected, err.Error()),
		)
	}
	return nil, fmt.Errorf("failed to fetch data. href=[%s] err: %w", collectors.TruncateURL(href), err)
}

func (r *RestPerf) InitSchedule() {
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
	_ collector.Collector = (*RestPerf)(nil)
)
