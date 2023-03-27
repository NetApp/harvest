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
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
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

	if err = s.initClient(); err != nil {
		return err
	}
	if s.Props.TemplatePath, err = s.LoadTemplate(); err != nil {
		return err
	}
	s.InitAPIPath()
	if err = collector.Init(s); err != nil {
		return err
	}

	if err = s.InitCache(); err != nil {
		return err
	}

	if err = s.InitMatrix(); err != nil {
		return err
	}

	s.Logger.Info().Msg("initialized")
	return nil
}

func (s *StorageGrid) InitMatrix() error {
	mat := s.Matrix[s.Object]
	// overwrite from abstract collector
	mat.Object = s.Props.Object
	// Add system (cluster) name
	mat.SetGlobalLabel("cluster", s.client.Cluster.Name)

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

	if x := s.Params.GetChildContentS("object"); x != "" {
		s.Props.Object = x
	} else {
		s.Props.Object = strings.ToLower(s.Object)
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

	s.Logger.Debug().
		Strs("extracted Instance Keys", s.Props.InstanceKeys).
		Int("numMetrics", len(s.Props.Metrics)).
		Int("numLabels", len(s.Props.InstanceLabels)).
		Msg("Initialized metric cache")

	return nil
}

func (s *StorageGrid) PollData() (map[string]*matrix.Matrix, error) {
	if s.Props.Query == "prometheus" {
		return s.pollPrometheusMetrics()
	} else {
		return s.pollRest()
	}
}

func (s *StorageGrid) pollPrometheusMetrics() (map[string]*matrix.Matrix, error) {
	var (
		count      uint64
		numRecords int
		startTime  time.Time
		apiD       time.Duration
	)

	metrics := make(map[string]*matrix.Matrix)
	s.Logger.Debug().Msg("starting data poll")
	s.Matrix[s.Object].Reset()
	startTime = time.Now()

	for _, metric := range s.Props.Metrics {
		mat, err := s.GetMetric(metric.Name, metric.Label, nil)
		if err != nil {
			s.Logger.Error().Err(err).Str("metric", metric.Name).Msg("failed to get metric")
			continue
		}
		metrics[metric.Name] = mat
		numInstances := len(mat.GetInstances())
		if numInstances == 0 {
			s.Logger.Warn().Str("metric", metric.Name).Msg("no instances on storagegrid")
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
	s.AddCollectCount(count)

	return metrics, nil
}

func (s *StorageGrid) makePromMetrics(metricName string, result *[]gjson.Result, tenantNamesByID map[string]string) (*matrix.Matrix, error) {
	var (
		metric   *matrix.Metric
		instance *matrix.Instance
		err      error
	)

	mat := s.Matrix[s.Object].Clone(false, false, false)
	mat.SetExportOptions(matrix.DefaultExportOptions())
	mat.Object = s.Props.Object
	mat.UUID += "." + metricName

	r := (*result)[0]
	resultType := r.Get("resultType").String()
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
			s.Logger.Error().Err(err).Str("instanceKey", metricName+"-"+strconv.Itoa(i)).Send()
			continue
		}
		rr.Get("metric").ForEach(func(kk, vv gjson.Result) bool {
			key := kk.String()
			value := vv.String()

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
			err = metric.SetValueFloat64(instance, valueArray[1].Float())
			if err != nil {
				s.Logger.Error().Err(err).Str("metric", metricName).Msg("Unable to set float key on metric")
				continue
			}
		}
	}
	return mat, nil
}

func (s *StorageGrid) pollRest() (map[string]*matrix.Matrix, error) {
	var (
		count        uint64
		apiD, parseD time.Duration
		startTime    time.Time
		err          error
		records      []gjson.Result
	)

	s.Logger.Debug().Msg("starting data poll")
	s.Matrix[s.Object].Reset()
	startTime = time.Now()

	if err = s.getRest(s.Props.Query, &records); err != nil {
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
	s.AddCollectCount(count)

	return s.Matrix, nil
}

func (s *StorageGrid) getRest(href string, result *[]gjson.Result) error {
	s.Logger.Debug().Str("href", href).Send()
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

	for _, instanceData := range result {
		var (
			instanceKey string
			instance    *matrix.Instance
		)

		if !instanceData.IsObject() {
			s.Logger.Warn().Str("type", instanceData.Type.String()).Msg("Instance data is not object, skipping")
			continue
		}

		// extract instance key(s)
		for _, k := range s.Props.InstanceKeys {
			value := instanceData.Get(k)
			if value.Exists() {
				instanceKey += value.String()
			} else {
				s.Logger.Warn().Str("key", k).Msg("skip instance, missing key")
				break
			}
		}

		if instanceKey == "" {
			s.Logger.Trace().Msg("Instance key is empty, skipping")
			continue
		}

		instance = mat.GetInstance(instanceKey)

		if instance == nil {
			if instance, err = mat.NewInstance(instanceKey); err != nil {
				s.Logger.Error().Err(err).Str("instanceKey", instanceKey).Send()
				continue
			}
		}

		for label, display := range s.Props.InstanceLabels {
			value := instanceData.Get(label)
			if value.Exists() {
				if value.IsArray() {
					var labelArray []string
					for _, r := range value.Array() {
						labelString := r.String()
						labelArray = append(labelArray, labelString)
					}
					instance.SetLabel(display, strings.Join(labelArray, ","))
				} else {
					instance.SetLabel(display, value.String())
				}
				count++
			} else {
				s.Logger.Trace().
					Str("instanceKey", instanceKey).
					Str("label", label).
					Msg("Missing label value")
			}
		}

		for _, metric := range s.Props.Metrics {
			metr, ok := mat.GetMetrics()[metric.Name]
			if !ok {
				if metr, err = mat.NewMetricFloat64(metric.Name, metric.Label); err != nil {
					s.Logger.Error().Err(err).
						Str("name", metric.Name).
						Msg("NewMetricFloat64")
				}
			}
			f := instanceData.Get(metric.Name)
			if f.Exists() {
				var floatValue float64
				switch metric.MetricType {
				case "":
					floatValue = f.Float()
				default:
					s.Logger.Warn().
						Str("type", metric.MetricType).
						Str("metric", metric.Name).
						Msg("unknown metric type")
				}

				if err = metr.SetValueFloat64(instance, floatValue); err != nil {
					s.Logger.Error().Err(err).
						Str("key", metric.Name).
						Str("metric", metric.Label).
						Msg("Unable to set float key on metric")
				}
				count++
			}
		}

	}
	return count
}

func (s *StorageGrid) initClient() error {
	var err error

	if s.client, err = srest.NewClient(s.Options.Poller, s.Params.GetChildContentS("client_timeout"), s.Auth); err != nil {
		return err
	}

	if err = s.client.Init(5); err != nil {
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
			name, display, kind, metricType = util.ParseMetric(c)
			s.Logger.Debug().
				Str("kind", kind).
				Str("name", name).
				Str("display", display).
				Msg("Collected")

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
		template     *node.Node
		templatePath string
		err          error
	)

	// import template

	template, templatePath, err = s.ImportSubTemplate(
		"",
		rest.TemplateFn(s.Params, s.Object),
		s.client.Cluster.Version,
	)
	if err != nil {
		return "", err
	}

	s.Params.Union(template)
	return templatePath, nil
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
		s.Logger.Warn().Str("kind", kind).Msg("plugin not found")
	}
	return nil
}

// InitAPIPath reads the REST API version from the template and uses it instead of
// the DefaultAPIVersion
func (s *StorageGrid) InitAPIPath() {
	apiVersion := s.Params.GetChildContentS("api")
	if !strings.HasSuffix(s.client.APIPath, apiVersion) {
		cur := s.client.APIPath
		s.client.APIPath = "/apiVersion/" + apiVersion
		s.Logger.Debug().
			Str("clientAPI", cur).
			Str("templateAPI", apiVersion).
			Msg("Use template apiVersion")
	}
}

func (s *StorageGrid) CollectAutoSupport(p *collector.Payload) {
	var exporterTypes []string
	for _, exporter := range s.Exporters {
		exporterTypes = append(exporterTypes, exporter.GetClass())
	}

	var counters = make([]string, 0)
	for k := range s.Props.Counters {
		counters = append(counters, k)
	}

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
	})

	version := s.client.Cluster.Version
	p.Target.Version = strconv.Itoa(version[0]) + "." + strconv.Itoa(version[1]) + "." + strconv.Itoa(version[2])
	p.Target.Model = "storagegrid"
	p.Target.ClusterUUID = s.client.Cluster.UUID

	if p.Nodes == nil {
		nodeIds, err := s.getNodeUuids()
		if err != nil {
			// log the error, but don't exit method so subsequent info is collected
			s.Logger.Error().Err(err).Msg("Unable to get nodes.")
			nodeIds = make([]collector.ID, 0)
		}
		p.Nodes = &collector.InstanceInfo{
			Ids:   nodeIds,
			Count: int64(len(nodeIds)),
		}
	}

	if s.Object == "Tenant" {
		md := s.GetMetadata()
		info := collector.InstanceInfo{
			Count:      md.LazyValueInt64("instances", "data"),
			DataPoints: md.LazyValueInt64("metrics", "data"),
			PollTime:   md.LazyValueInt64("poll_time", "data"),
			APITime:    md.LazyValueInt64("api_time", "data"),
			ParseTime:  md.LazyValueInt64("parse_time", "data"),
			PluginTime: md.LazyValueInt64("plugin_time", "data"),
		}
		p.Tenants = &info
	}
}

func (s *StorageGrid) getNodeUuids() ([]collector.ID, error) {
	var (
		err    error
		infos  []collector.ID
		health []byte
	)

	health, err = s.client.GetGridRest("grid/node-health")
	if err != nil {
		return nil, err
	}
	data := gjson.GetBytes(health, "data").Array()

	for _, each := range data {
		infos = append(infos, collector.ID{
			SerialNumber: each.Get("id").String(),
			SystemID:     each.Get("siteId").String(),
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
		s.Logger.Debug().Str("metric", metric).Msg("no metrics on cluster")
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
