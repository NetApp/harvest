package eseries

import (
	"log/slog"
	"strings"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/template"
)

func (e *ESeries) LoadTemplate() (string, error) {
	templateName := e.Params.GetChildS("objects").GetChildContentS(e.Object)
	if templateName == "" {
		return "", errs.New(errs.ErrMissingParam, "template for object "+e.Object)
	}

	jitter := e.Params.GetChildContentS("jitter")
	subTemplate, path, err := e.ImportSubTemplate([]string{""}, templateName, jitter, e.Remote.Version)
	if err != nil {
		return "", err
	}

	// Merge template into params
	e.Params.Union(subTemplate)

	e.Logger.Debug("loaded template",
		slog.String("object", e.Object),
		slog.String("template", templateName),
		slog.String("path", path),
	)

	return path, nil
}

// ObjectConfig holds configuration for different object types
type ObjectConfig struct {
	ArrayPath                  string
	Filter                     string
	CalculateUtilization       bool
	CalculateQueueDepthAverage bool
	UsesSharedCache            bool
}

func GetESeriesPerfObjectConfig(objType string) ObjectConfig {
	configs := map[string]ObjectConfig{
		"controller":  {ArrayPath: "controllerStats", Filter: "type=controller", UsesSharedCache: true},
		"pool":        {ArrayPath: "poolStats", Filter: "type=storagePool", UsesSharedCache: true},
		"volume":      {ArrayPath: "volumeStats", Filter: "type=volume", UsesSharedCache: true, CalculateQueueDepthAverage: true},
		"drive":       {ArrayPath: "diskStats", Filter: "type=drive", UsesSharedCache: true, CalculateUtilization: true},
		"interface":   {ArrayPath: "interfaceStats", Filter: "type=ioInterface", UsesSharedCache: true},
		"application": {ArrayPath: "applicationStats", Filter: "type=application", UsesSharedCache: true},
		"workload":    {ArrayPath: "workloadStats", Filter: "type=workload", UsesSharedCache: true},
		"array":       {ArrayPath: "systemStats", Filter: "type=storageSystem", UsesSharedCache: true},
	}
	if config, ok := configs[objType]; ok {
		return config
	}
	return ObjectConfig{}
}

// ExtractCacheNameFromQuery Example: "storage-systems/{array_id}/live-statistics" -> "live-statistics"
func ExtractCacheNameFromQuery(query string) string {
	parts := strings.Split(query, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func (e *ESeries) getCacheTTL() time.Duration {
	schedule := e.Params.GetChildS("schedule")
	if schedule == nil {
		return 1 * time.Minute
	}

	dataInterval := schedule.GetChildS("data")
	if dataInterval == nil {
		return 1 * time.Minute
	}

	if ttl, err := time.ParseDuration(dataInterval.GetContentS()); err == nil {
		// Expire cache 2 seconds before next poll to ensure fresh fetch every poll cycle
		// This makes sure that cache is expired before every new poll.
		bufferTime := 2 * time.Second
		if ttl > bufferTime {
			return ttl - bufferTime
		}
		// For very short intervals, use 90% of the interval
		return time.Duration(float64(ttl) * 0.9)
	}
	return 1 * time.Minute
}

// setupSharedCache enables shared cache if configured
func (e *ESeries) setupSharedCache(config ObjectConfig) {
	if !config.UsesSharedCache || e.Prop.CacheConfig != nil {
		return
	}

	cacheName := ExtractCacheNameFromQuery(e.Prop.Query)
	if cacheName == "" {
		return
	}

	e.Prop.CacheConfig = &rest.CacheConfig{
		Name: cacheName,
		TTL:  e.getCacheTTL(),
	}

	e.Logger.Debug("shared cache auto-enabled",
		slog.String("name", cacheName),
		slog.String("ttl", e.Prop.CacheConfig.TTL.String()))
}

func (e *ESeries) applyFilter(config ObjectConfig) {
	if e.Prop.CacheConfig == nil && config.Filter != "" {
		e.Prop.Filter = append(e.Prop.Filter, config.Filter)
	} else if e.Prop.CacheConfig != nil && config.Filter != "" {
		e.Logger.Debug("filter disabled due to shared cache",
			slog.String("skipped_filter", config.Filter),
			slog.String("cache", e.Prop.CacheConfig.Name))
	}
}

func (e *ESeries) ParseTemplate(config ObjectConfig) error {
	e.Prop.Object = e.Params.GetChildContentS("object")
	e.Prop.Query = e.Params.GetChildContentS("query")

	// Validate required fields
	if e.Prop.Object == "" {
		return errs.New(errs.ErrMissingParam, "object")
	}
	if e.Prop.Query == "" {
		return errs.New(errs.ErrMissingParam, "query")
	}

	if config != (ObjectConfig{}) {
		e.Prop.ResponseArrayPath = config.ArrayPath

		// Setup shared cache and filter
		e.setupSharedCache(config)
		e.applyFilter(config)
	}

	return e.parseCounters()
}

// parseCounters parses the counters section from template
func (e *ESeries) parseCounters() error {
	counters := e.Params.GetChildS("counters")
	if counters == nil {
		return nil
	}

	for _, counter := range counters.GetChildren() {
		content := counter.GetContentS()
		if content == "" {
			continue
		}

		name, display, kind, metricType := template.ParseMetric(content)

		switch kind {
		case "key":
			e.Prop.InstanceKeys = append(e.Prop.InstanceKeys, name)
			e.Prop.InstanceLabels[name] = display
		case "label":
			e.Prop.InstanceLabels[name] = display
		default:
			metric := &Metric{
				Label:      display,
				Name:       name,
				MetricType: metricType,
				Exportable: true,
			}
			e.Prop.Metrics[name] = metric
			e.Prop.Counters[name] = display
		}
	}

	// Only parse filter from counters if type field wasn't used AND cache is disabled
	if e.Params.GetChildContentS("type") == "" {
		if e.Prop.CacheConfig == nil {
			if x := counters.GetChildS("filter"); x != nil {
				e.Prop.Filter = append(e.Prop.Filter, x.GetAllChildContentS()...)
			}
		}
	}

	return nil
}
