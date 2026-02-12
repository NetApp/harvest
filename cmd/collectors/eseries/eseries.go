package eseries

import (
	"log/slog"
	"strings"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/plugins/hardware"
	"github.com/netapp/harvest/v2/cmd/collectors/eseries/plugins/host"
	"github.com/netapp/harvest/v2/cmd/collectors/eseries/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/eseries/plugins/volumemapping"
	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

type ESeries struct {
	*collector.AbstractCollector
	Client    *rest.Client
	Prop      *Prop
	arrayID   string
	arrayName string
}

type Prop struct {
	Object            string
	Query             string
	TemplatePath      string
	ResponseArrayPath string   // Path to the array in nested responses (e.g., "volumeStats")
	Filter            []string // Query filters to append to URL (e.g., "type=volume")
	InstanceKeys      []string
	InstanceLabels    map[string]string
	Metrics           map[string]*Metric
	Counters          map[string]string
	CacheConfig       *rest.CacheConfig
}

type Metric struct {
	Label      string
	Name       string
	MetricType string
	Exportable bool
}

func init() {
	plugin.RegisterModule(&ESeries{})
}

func (e *ESeries) HarvestModule() plugin.ModuleInfo {
	return plugin.ModuleInfo{
		ID:  "harvest.collector.eseries",
		New: func() plugin.Module { return new(ESeries) },
	}
}

func (e *ESeries) Init(a *collector.AbstractCollector) error {
	var err error

	e.AbstractCollector = a

	e.InitProp()

	if e.Prop.TemplatePath, err = e.LoadTemplate(); err != nil {
		return err
	}

	if err := e.ParseTemplate(ObjectConfig{}); err != nil {
		return err
	}

	// InitClient after ParseTemplate so we know if caching is enabled
	if err := e.InitClient(); err != nil {
		return err
	}

	// In test mode, skip plugin initialization but still initialize Matrix
	if e.Options.IsTest {
		// Initialize Matrix manually (normally done by collector.Init)
		mx := matrix.New(e.Name, e.Object, e.Object)
		if exportOptions := e.Params.GetChildS("export_options"); exportOptions != nil {
			mx.SetExportOptions(exportOptions)
		} else {
			mx.SetExportOptions(matrix.DefaultExportOptions())
		}
		e.Matrix = make(map[string]*matrix.Matrix)
		e.Matrix[e.Object] = mx
	} else {
		if err := collector.Init(e); err != nil {
			return err
		}
	}

	e.InitMatrix()

	e.Logger.Debug(
		"initialized",
		slog.Int("numMetrics", len(e.Prop.Metrics)),
		slog.String("object", e.Prop.Object),
		slog.String("timeout", e.Client.Timeout.String()),
	)

	return nil
}

func (e *ESeries) InitProp() {
	e.Prop = &Prop{
		InstanceKeys:   make([]string, 0),
		InstanceLabels: make(map[string]string),
		Metrics:        make(map[string]*Metric),
		Counters:       make(map[string]string),
	}
}

func (e *ESeries) InitClient() error {
	var err error

	clientTimeout := e.Params.GetChildContentS("client_timeout")
	if clientTimeout == "" {
		clientTimeout = rest.DefaultTimeout
	}

	duration, err := time.ParseDuration(clientTimeout)
	if err != nil {
		e.Logger.Info("Using default timeout", slog.String("timeout", rest.DefaultTimeout))
		duration, _ = time.ParseDuration(rest.DefaultTimeout)
	}

	poller, err := conf.PollerNamed(e.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, e.Logger)

	// Use pooled client only if caching is enabled (checked after template parsing)
	cacheName := ""
	if e.Prop.CacheConfig != nil {
		cacheName = e.Prop.CacheConfig.Name
	}
	if e.Client, err = rest.New(poller, duration, credentials, cacheName); err != nil {
		return err
	}

	if e.Options.IsTest {
		return nil
	}

	if err := e.Client.Init(1, e.Remote); err != nil {
		return err
	}

	e.Remote = e.Client.Remote()

	return nil
}

func (e *ESeries) InitMatrix() {
	mat := e.Matrix[e.Object]
	mat.Object = e.Prop.Object

	if e := e.Params.GetChildS("export_options"); e != nil {
		mat.SetExportOptions(e)
	}

	if e.Params.HasChildS("labels") {
		for _, l := range e.Params.GetChildS("labels").GetChildren() {
			mat.SetGlobalLabel(l.GetNameS(), l.GetContentS())
		}
	}

	e.Logger.Debug(
		"initialized cache",
		slog.Any("instanceKeys", e.Prop.InstanceKeys),
		slog.Int("numMetrics", len(e.Prop.Metrics)),
		slog.Int("numLabels", len(e.Prop.InstanceLabels)),
	)
}

func (e *ESeries) PollCounter() (map[string]*matrix.Matrix, error) {
	var err error

	systems, err := e.Client.GetStorageSystems()
	if err != nil {
		return nil, err
	}

	if len(systems) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no storage system found")
	}

	if len(systems) > 1 {
		e.Logger.Warn("multiple systems found, using first one", slog.Int("count", len(systems)))
	}

	system := systems[0]
	e.arrayID = system.Get("id").ClonedString()
	e.arrayName = system.Get("name").ClonedString()

	if e.arrayID == "" {
		return nil, errs.New(errs.ErrNoInstance, "system missing id")
	}

	if e.arrayName == "" {
		e.arrayName = e.arrayID
	}

	// Store arrayID in Params for plugin access (not exported as global label)
	if e.Params != nil {
		e.Params.NewChildS("array_id", e.arrayID)
		e.Params.NewChildS("array", e.arrayName)
	}

	mat := e.Matrix[e.Object]
	mat.SetGlobalLabel("array", e.arrayName)

	e.Logger.Debug(
		"discovered array",
		slog.String("id", e.arrayID),
		slog.String("name", e.arrayName),
	)

	return nil, nil
}

func (e *ESeries) PollData() (map[string]*matrix.Matrix, error) {
	var (
		count     uint64
		apiTime   time.Duration
		parseTime time.Duration
	)

	mat := e.Matrix[e.Object]

	query := rest.NewURLBuilder().
		APIPath(e.Prop.Query).
		ArrayID(e.arrayID).
		Build()

	var results []gjson.Result
	var err error

	apiStart := time.Now()
	results, err = e.Client.Fetch(e.Client.APIPath+"/"+query, e.Prop.CacheConfig)
	apiTime = time.Since(apiStart)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		e.Logger.Debug("no instances", slog.String("object", e.Object))
		return nil, errs.New(errs.ErrNoInstance, "no instances found")
	}

	// Extract nested array if ResponseArrayPath is specified (for hardware-inventory responses)
	var dataArray []gjson.Result
	if e.Prop.ResponseArrayPath != "" {
		arrayResult := gjson.ParseBytes([]byte(results[0].Raw)).Get(e.Prop.ResponseArrayPath)
		if !arrayResult.Exists() {
			e.Logger.Warn("Response array path not found", slog.String("path", e.Prop.ResponseArrayPath))
			return nil, errs.New(errs.ErrNoInstance, "no instances found")
		}
		dataArray = arrayResult.Array()
	} else {
		dataArray = results
	}

	parseStart := time.Now()
	count = e.pollData(mat, dataArray)
	parseTime = time.Since(parseStart)

	if len(mat.GetInstances()) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no instances found")
	}

	_ = e.Metadata.LazySetValueInt64("api_time", "data", apiTime.Microseconds())
	_ = e.Metadata.LazySetValueInt64("parse_time", "data", parseTime.Microseconds())
	_ = e.Metadata.LazySetValueUint64("metrics", "data", count)
	_ = e.Metadata.LazySetValueUint64("instances", "data", uint64(len(mat.GetInstances())))
	_ = e.Metadata.LazySetValueUint64("bytesRx", "data", e.Client.Metadata.BytesRx)
	_ = e.Metadata.LazySetValueUint64("numCalls", "data", e.Client.Metadata.NumCalls)
	e.AddCollectCount(count)

	return e.Matrix, nil
}

func (e *ESeries) pollData(mat *matrix.Matrix, results []gjson.Result) uint64 {
	var (
		err   error
		count uint64
	)

	oldInstances := set.New()
	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}

	currentInstances := set.New()

	for _, instanceData := range results {
		if !instanceData.IsObject() {
			e.Logger.Warn("instance data is not object, skipping")
			continue
		}

		var instanceKey strings.Builder
		if len(e.Prop.InstanceKeys) > 0 {
			for _, k := range e.Prop.InstanceKeys {
				value := instanceData.Get(k)
				if value.Exists() {
					instanceKey.WriteString(value.ClonedString())
				}
			}
		}

		if instanceKey.String() == "" {
			e.Logger.Warn("empty instance key, skipping",
				slog.String("object", e.Object),
				slog.Any("instanceKeys", e.Prop.InstanceKeys))
			continue
		}

		instKey := instanceKey.String()
		instance := mat.GetInstance(instKey)
		if instance == nil {
			var err error
			if instance, err = mat.NewInstance(instKey); err != nil {
				e.Logger.Error("failed to create instance", slogx.Err(err), slog.String("key", instKey))
				continue
			}
		}

		instance.SetExportable(true)

		if currentInstances.Has(instKey) {
			e.Logger.Error("This instance is already processed. instKey is not unique", slog.String("instKey", instKey))
		} else {
			currentInstances.Add(instKey)
		}

		oldInstances.Remove(instKey)
		instance.ClearLabels()

		for label, display := range e.Prop.InstanceLabels {
			value := instanceData.Get(label)
			if value.Exists() {
				instance.SetLabel(display, value.ClonedString())
			}
		}

		for _, metric := range e.Prop.Metrics {
			value := instanceData.Get(metric.Name)
			if !value.Exists() {
				continue
			}

			metr, ok := mat.GetMetrics()[metric.Name]
			if !ok {
				if metr, err = mat.NewMetricFloat64(metric.Name, metric.Label); err != nil {
					e.Logger.Error(
						"NewMetricFloat64",
						slogx.Err(err),
						slog.String("name", metric.Name),
					)
				} else {
					metr.SetExportable(metric.Exportable)
				}
			}

			metr.SetValueFloat64(instance, value.Float())
			count++
		}
	}

	// Remove old instances that are not found in new instances
	for key := range oldInstances.Iter() {
		mat.RemoveInstance(key)
	}

	return count
}

func (e *ESeries) LoadPlugin(kind string, abc *plugin.AbstractPlugin) plugin.Plugin {
	switch kind {
	case "Hardware":
		return hardware.New(abc)
	case "Host":
		return host.New(abc)
	case "Volume":
		return volume.New(abc)
	case "VolumeMapping":
		return volumemapping.New(abc)
	default:
		e.Logger.Warn("no ESeries plugin found", slog.String("kind", kind))
	}
	return nil
}

func (e *ESeries) GetArray() string {
	return e.arrayID
}

var (
	_ collector.Collector = (*ESeries)(nil)
)
