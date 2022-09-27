package storagegrid

import (
	"github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/plugins/bucket"
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/plugins/tenant"
	srest "github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
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
	if err := collector.Init(s); err != nil {
		return err
	}

	if err := s.InitCache(); err != nil {
		return err
	}

	if err := s.InitMatrix(); err != nil {
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

	// create metric cache
	if counters = s.Params.GetChildS("counters"); counters == nil {
		return errs.New(errs.ErrMissingParam, "counters")
	}

	s.ParseCounters(counters, s.Props)
	_, _ = s.Metadata.NewMetricUint64("datapoint_count")

	s.Logger.Debug().
		Strs("extracted Instance Keys", s.Props.InstanceKeys).
		Int("numMetrics", len(s.Props.Metrics)).
		Int("numLabels", len(s.Props.InstanceLabels)).
		Msg("Initialized metric cache")

	return nil
}

func (s *StorageGrid) PollData() (map[string]*matrix.Matrix, error) {
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

	s.Logger.Info().
		Int("instances", numRecords).
		Uint64("metrics", count).
		Str("apiD", apiD.Round(time.Millisecond).String()).
		Str("parseD", parseD.Round(time.Millisecond).String()).
		Msg("Collected")

	_ = s.Metadata.LazySetValueInt64("count", "data", int64(numRecords))
	_ = s.Metadata.LazySetValueInt64("api_time", "data", apiD.Microseconds())
	_ = s.Metadata.LazySetValueInt64("parse_time", "data", parseD.Microseconds())
	_ = s.Metadata.LazySetValueUint64("datapoint_count", "data", count)
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
				if metr, err = mat.NewMetricFloat64(metric.Name); err != nil {
					s.Logger.Error().Err(err).
						Str("name", metric.Name).
						Msg("NewMetricFloat64")
				}
			}
			f := instanceData.Get(metric.Name)
			if f.Exists() {
				metr.SetName(metric.Label)

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
	a := s.AbstractCollector

	if s.client, err = srest.NewClient(s.Options.Poller, a.Params.GetChildContentS("client_timeout")); err != nil {
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
		return tenant.New(abc)
	default:
		s.Logger.Warn().Str("kind", kind).Msg("plugin not found")
	}
	return nil
}

// Interface guards
var (
	_ collector.Collector = (*StorageGrid)(nil)
)
