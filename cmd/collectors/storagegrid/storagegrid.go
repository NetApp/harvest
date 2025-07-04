package storagegrid

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/plugins/bucket"
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/plugins/joinrest"
	srest "github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

type prop struct {
	Object         string
	Query          string
	TemplatePath   string
	InstanceKeys   []string
	InstanceLabels map[string]string
	Metrics        map[string]*Metric
	Counters       map[string]string
	ReturnTimeOut  string
	Fields         []string
	Filter         []string
}

type Metric struct {
	Label      string
	Name       string
	MetricType string
	Exportable bool
}

type StorageGrid struct {
	*collector.AbstractCollector
	client *srest.Client
	Props  *prop
}

func init() {
	plugin.RegisterModule(&StorageGrid{})
}

func (s *StorageGrid) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.storagegrid",
		New: func() plugin.Module { return new(StorageGrid) },
	}
}

func (s *StorageGrid) Init(a *collector.AbstractCollector) error {
	var err error
	s.AbstractCollector = a
	s.InitProp()

	if err := s.initClient(); err != nil {
		return err
	}
	if s.Props.TemplatePath, err = s.LoadTemplate(); err != nil {
		return err
	}
	s.InitAPIPath()
	if err := collector.Init(s); err != nil {
		return err
	}

	if err := s.InitCache(); err != nil {
		return err
	}

	if err := s.InitMatrix(); err != nil {
		return err
	}

	s.Logger.Debug("initialized")
	return nil
}

func (s *StorageGrid) InitMatrix() error {
	mat := s.Matrix[s.Object]
	// overwrite from abstract collector
	mat.Object = s.Props.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", s.client.Remote.Name)

	if s.Params.HasChildS("labels") {
		for _, l := range s.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	return nil
}

func (s *StorageGrid) InitCache() error {
	var (
		counters *node.Node
	)

	// Unlike other collectors, the storagegrid collector allows a blank object property
	// so that metrics can be exported without a prefix
	if x := s.Params.GetChildContentS("object"); x != "" {
		s.Props.Object = x
	}

	if e := s.Params.GetChildS("export_options"); e != nil {
		s.Matrix[s.Object].SetExportOptions(e)
	}

	if s.Props.Query = s.Params.GetChildContentS("query"); s.Props.Query == "" {
		return errs.New(errs.ErrMissingParam, "query")
	}

	if counters = s.Params.GetChildS("counters"); counters == nil {
		return errs.New(errs.ErrMissingParam, "counters")
	}
	s.ParseCounters(counters, s.Props)

	s.Logger.Debug(
		"Initialized metric cache",
		slog.Any("extracted Instance Keys", s.Props.InstanceKeys),
		slog.Int("numMetrics", len(s.Props.Metrics)),
		slog.Int("numLabels", len(s.Props.InstanceLabels)),
	)

	return nil
}

func (s *StorageGrid) PollData() (map[string]*matrix.Matrix, error) {
	s.client.Metadata.Reset()
	if s.Props.Query == "prometheus" {
		return s.pollPrometheusMetrics()
	}
	return s.pollRest()
}

func (s *StorageGrid) pollPrometheusMetrics() (map[string]*matrix.Matrix, error) {
	var (
		count      uint64
		numRecords int
		startTime  time.Time
		apiD       time.Duration
	)

	metrics := make(map[string]*matrix.Matrix)
	s.Matrix[s.Object].Reset()
	startTime = time.Now()

	for _, metric := range s.Props.Metrics {
		mat, err := s.GetMetric(metric.Name, metric.Label, nil)
		if err != nil {
			s.Logger.Error("failed to get metric", slogx.Err(err), slog.String("metric", metric.Name))
			continue
		}
		metrics[metric.Name] = mat
		numInstances := len(mat.GetInstances())
		if numInstances == 0 {
			s.Logger.Warn("no instances on storagegrid", slog.String("metric", metric.Name))
			continue
		}
		count += uint64(numInstances)
		numRecords += numInstances
	}

	apiD = time.Since(startTime)

	_ = s.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = s.Metadata.LazySetValueInt64("parse_time", "data", 0)
	_ = s.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = s.Metadata.LazySetValueInt64("instances", "data", int64(numRecords))
	_ = s.Metadata.LazySetValueUint64("bytesRx", "data", s.client.Metadata.BytesRx)
	_ = s.Metadata.LazySetValueUint64("numCalls", "data", s.client.Metadata.NumCalls)

	s.AddCollectCount(count)

	return metrics, nil
}

func (s *StorageGrid) makePromMetrics(metricName string, result *[]gjson.Result, tenantNamesByID map[string]string) (*matrix.Matrix, error) {
	var (
		metric   *matrix.Metric
		instance *matrix.Instance
		err      error
	)

	mat := s.Matrix[s.Object].Clone(matrix.With{Data: false, Metrics: false, Instances: false, ExportInstances: true})
	mat.SetExportOptions(matrix.DefaultExportOptions())
	mat.Object = s.Props.Object
	mat.UUID += "." + metricName

	r := (*result)[0]
	resultType := r.Get("resultType").ClonedString()
	if resultType != "vector" {
		return nil, fmt.Errorf("unexpected resultType=[%s]", resultType)
	}

	results := r.Get("result").Array()
	if len(results) == 0 {
		return mat, nil
	}

	if metric, err = mat.NewMetricFloat64(metricName); err != nil {
		return nil, fmt.Errorf("failed to create newMetric float64 metric=[%s]", metricName)
	}

	instances := r.Get("result").Array()
	for i, rr := range instances {
		if instance, err = mat.NewInstance(metricName + "-" + strconv.Itoa(i)); err != nil {
			s.Logger.Error(
				"",
				slogx.Err(err),
				slog.String("metricName", metricName),
				slog.Int("i", i),
			)
			continue
		}
		rr.Get("metric").ForEach(func(kk, vv gjson.Result) bool {
			key := kk.ClonedString()
			value := vv.ClonedString()

			if key == "__name__" {
				return true
			}
			if key == "instance" {
				key = "node"
			}
			if tenantNamesByID != nil && key == "tenant_id" {
				tenantName, ok := tenantNamesByID[value]
				if ok {
					instance.SetLabel("tenant", tenantName)
				}
			}
			instance.SetLabel(key, value)
			return true
		})

		// copy Prometheus metric value into new metric
		valueArray := rr.Get("value").Array()
		if len(valueArray) > 0 {
			metric.SetValueFloat64(instance, valueArray[1].Float())
		}
	}
	return mat, nil
}

func (s *StorageGrid) pollRest() (map[string]*matrix.Matrix, error) {
	var (
		count        uint64
		apiD, parseD time.Duration
		startTime    time.Time
		records      []gjson.Result
	)

	s.Matrix[s.Object].Reset()
	startTime = time.Now()

	if err := s.getRest(s.Props.Query, &records); err != nil {
		return nil, err
	}

	apiD = time.Since(startTime)

	if len(records) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+s.Object+" instances on cluster")
	}

	startTime = time.Now()
	count = s.handleResults(records)
	parseD = time.Since(startTime)

	numRecords := len(s.Matrix[s.Object].GetInstances())

	_ = s.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = s.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = s.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = s.Metadata.LazySetValueInt64("instances", "data", int64(numRecords))
	_ = s.Metadata.LazySetValueUint64("bytesRx", "data", s.client.Metadata.BytesRx)
	_ = s.Metadata.LazySetValueUint64("numCalls", "data", s.client.Metadata.NumCalls)

	s.AddCollectCount(count)

	return s.Matrix, nil
}

func (s *StorageGrid) getRest(href string, result *[]gjson.Result) error {
	s.Logger.Debug("", slog.String("href", href))
	if href == "" {
		return errs.New(errs.ErrConfig, "empty url")
	}
	return s.client.Fetch(href, result)
}

func (s *StorageGrid) handleResults(result []gjson.Result) uint64 {
	var (
		err   error
		count uint64
	)

	mat := s.Matrix[s.Object]

	// Keep track of old instances
	oldInstances := make(map[string]bool)
	for key := range mat.GetInstances() {
		oldInstances[key] = true
	}

	for _, instanceData := range result {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			s.Logger.Warn("Instance data is not object, skipping", slog.String("type", instanceData.Type.String()))
			continue
		}

		// extract instance key(s)
		for _, k := range s.Props.InstanceKeys {
			value := instanceData.Get(k)
			if value.Exists() {
				instanceKey += value.ClonedString()
			} else {
				s.Logger.Warn("skip instance, missing key", slog.String("key", k))
				break
			}
		}

		if instanceKey == "" {
			if s.Params.GetChildContentS("only_cluster_instance") == "true" {
				instanceKey = "cluster"
			} else {
				continue
			}
		}

		instance = mat.GetInstance(instanceKey)

		if instance == nil {
			if instance, err = mat.NewInstance(instanceKey); err != nil {
				s.Logger.Error("", slogx.Err(err), slog.String("instanceKey", instanceKey))
				continue
			}
		}

		delete(oldInstances, instanceKey)

		for label, display := range s.Props.InstanceLabels {
			value := instanceData.Get(label)
			if value.Exists() {
				if value.IsArray() {
					var labelArray []string
					for _, r := range value.Array() {
						labelString := r.ClonedString()
						labelArray = append(labelArray, labelString)
					}
					instance.SetLabel(display, strings.Join(labelArray, ","))
				} else {
					instance.SetLabel(display, value.ClonedString())
				}
				count++
			}
		}

		for _, metric := range s.Props.Metrics {
			metr, ok := mat.GetMetrics()[metric.Name]
			if !ok {
				if metr, err = mat.NewMetricFloat64(metric.Name, metric.Label); err != nil {
					s.Logger.Error(
						"NewMetricFloat64",
						slogx.Err(err),
						slog.String("name", metric.Name),
					)
				}
			}
			f := instanceData.Get(metric.Name)
			if f.Exists() {
				var floatValue float64
				switch metric.MetricType {
				case "":
					floatValue = f.Float()
				default:
					s.Logger.Warn(
						"unknown metric type",
						slog.String("type", metric.MetricType),
						slog.String("metric", metric.Name),
					)
				}

				metr.SetValueFloat64(instance, floatValue)
				count++
			}
		}
	}
	// Remove instances not present in the new set
	for key := range oldInstances {
		mat.RemoveInstance(key)
		s.Logger.Debug("removed instance", slog.String("key", key))
	}
	return count
}

func (s *StorageGrid) initClient() error {
	var err error

	if s.client, err = srest.NewClientFunc(s.Options.Poller, s.Params.GetChildContentS("client_timeout"), s.Auth); err != nil {
		return err
	}

	if s.Options.IsTest {
		return nil
	}

	if err := s.client.Init(5, s.Remote); err != nil {
		return err
	}
	s.client.TraceLogSet(s.Name, s.Params)

	return nil
}

func (s *StorageGrid) ParseCounters(counter *node.Node, prop *prop) {
	var (
		display, name, kind, metricType string
	)

	for _, c := range counter.GetAllChildContentS() {
		if c != "" {
			name, display, kind, metricType = template.ParseMetric(c)
			s.Logger.Debug(
				"Collected",
				slog.String("kind", kind),
				slog.String("name", name),
				slog.String("display", display),
			)

			prop.Counters[name] = display
			switch kind {
			case "key":
				prop.InstanceLabels[name] = display
				prop.InstanceKeys = append(prop.InstanceKeys, name)
			case "label":
				prop.InstanceLabels[name] = display
			case "float":
				m := &Metric{Label: display, Name: name, MetricType: metricType, Exportable: true}
				prop.Metrics[name] = m
			}
		}
	}
}

func (s *StorageGrid) InitProp() {
	s.Props = &prop{
		InstanceKeys:   make([]string, 0),
		InstanceLabels: make(map[string]string),
		Counters:       make(map[string]string),
		Metrics:        make(map[string]*Metric),
	}
}

func (s *StorageGrid) LoadTemplate() (string, error) {
	var (
		subTemplate *node.Node
		path        string
		err         error
	)

	jitter := s.Params.GetChildContentS("jitter")

	subTemplate, path, err = s.ImportSubTemplate("", rest.TemplateFn(s.Params, s.Object), jitter, s.Remote.Version)
	if err != nil {
		return "", err
	}

	s.Params.Union(subTemplate)
	return path, nil
}

func (s *StorageGrid) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Bucket":
		return bucket.New(abc)
	case "Tenant":
		return NewTenant(abc, s)
	case "JoinRest":
		return joinrest.New(abc)
	default:
		s.Logger.Warn("plugin not found", slog.String("kind", kind))
	}
	return nil
}

// InitAPIPath reads the REST API version from the template and uses it instead of
// the DefaultAPIVersion
func (s *StorageGrid) InitAPIPath() {
	apiVersion := s.Params.GetChildContentS("api")
	if !strings.HasSuffix(s.client.APIPath, apiVersion) {
		cur := s.client.APIPath
		s.client.APIPath = "/api/" + apiVersion
		s.Logger.Debug(
			"Use template apiVersion",
			slog.String("clientAPI", cur),
			slog.String("templateAPI", apiVersion),
		)
	}
}

func (s *StorageGrid) CollectAutoSupport(p *collector.Payload) {
	exporterTypes := make([]string, 0, len(s.Exporters))
	for _, exporter := range s.Exporters {
		exporterTypes = append(exporterTypes, exporter.GetClass())
	}

	var counters = make([]string, 0, len(s.Props.Counters))
	for k := range s.Props.Counters {
		counters = append(counters, k)
	}
	slices.Sort(counters)

	var schedules = make([]collector.Schedule, 0)
	tasks := s.Params.GetChildS("schedule")
	if tasks != nil && len(tasks.GetChildren()) > 0 {
		for _, task := range tasks.GetChildren() {
			schedules = append(schedules, collector.Schedule{
				Name:     task.GetNameS(),
				Schedule: task.GetContentS(),
			})
		}
	}

	// Add collector information
	md := s.GetMetadata()
	info := collector.InstanceInfo{
		Count:      md.LazyValueInt64("instances", "data"),
		DataPoints: md.LazyValueInt64("metrics", "data"),
		PollTime:   md.LazyValueInt64("poll_time", "data"),
		APITime:    md.LazyValueInt64("api_time", "data"),
		ParseTime:  md.LazyValueInt64("parse_time", "data"),
		PluginTime: md.LazyValueInt64("plugin_time", "data"),
	}

	p.AddCollectorAsup(collector.AsupCollector{
		Name:      s.Name,
		Query:     s.Props.Query,
		Exporters: exporterTypes,
		Counters: collector.Counters{
			Count: len(counters),
			List:  counters,
		},
		Schedules:     schedules,
		ClientTimeout: s.client.Timeout.String(),
		InstanceInfo:  &info,
	})

	version := s.Remote.Version
	p.Target.Version = version
	p.Target.Model = "storagegrid"
	p.Target.ClusterUUID = s.Remote.UUID

	if p.Nodes == nil {
		nodeIDs, err := s.getNodeUuids()
		if err != nil {
			// log the error, but don't exit method so later info is collected
			s.Logger.Error("Unable to get nodes", slogx.Err(err))
			nodeIDs = make([]collector.ID, 0)
		}
		p.Nodes = &collector.InstanceInfo{
			Ids:   nodeIDs,
			Count: int64(len(nodeIDs)),
		}
	}

	if s.Object == "Tenant" {
		p.Tenants = &info
	}
}

func (s *StorageGrid) getNodeUuids() ([]collector.ID, error) {
	var (
		err    error
		health []byte
	)

	health, err = s.client.GetGridRest("grid/node-health")
	if err != nil {
		return nil, err
	}
	data := gjson.GetBytes(health, "data").Array()

	infos := make([]collector.ID, 0, len(data))
	for _, each := range data {
		infos = append(infos, collector.ID{
			SerialNumber: each.Get("id").ClonedString(),
			SystemID:     each.Get("siteId").ClonedString(),
		})
	}

	// Sort to make diffing easier
	sort.SliceStable(infos, func(i, j int) bool {
		return infos[i].SerialNumber < infos[j].SerialNumber
	})
	return infos, nil
}

func (s *StorageGrid) GetMetric(metric string, display string, tenantNamesByID map[string]string) (*matrix.Matrix, error) {
	var records []gjson.Result
	err := s.client.GetMetricQuery(metric, &records)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric=[%s] error: %w", metric, err)
	}
	if len(records) == 0 {
		s.Logger.Debug("no metrics on cluster", slog.String("metric", metric))
		return nil, nil
	}
	nameOfMetric := metric
	if display != "" {
		nameOfMetric = display
	}
	return s.makePromMetrics(nameOfMetric, &records, tenantNamesByID)
}

// Interface guards
var (
	_ collector.Collector = (*StorageGrid)(nil)
)
