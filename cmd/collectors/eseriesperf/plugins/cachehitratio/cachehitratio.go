package cachehitratio

import (
	"log/slog"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type CacheHitRatio struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &CacheHitRatio{AbstractPlugin: p}
}

func (c *CacheHitRatio) Init(_ conf.Remote) error {
	return c.InitAbc()
}

func (c *CacheHitRatio) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[c.Object]
	if data == nil {
		c.SLogger.Debug("No data matrix available")
		return nil, nil, nil
	}

	// Determine object type by checking available metrics
	objectType := c.detectObjectType(data)

	switch objectType {
	case "volume":
		c.calculateVolumeRatios(data)
	case "controller", "system":
		c.calculateAggregatedRatios(data)
	default:
		c.SLogger.Warn("Unknown object type, skipping cache hit ratio calculation", slog.String("object", c.Object))
	}

	return nil, nil, nil
}

func (c *CacheHitRatio) detectObjectType(data *matrix.Matrix) string {
	// Volume objects have separate read and write hit ops
	if data.GetMetric("readHitOps") != nil && data.GetMetric("writeHitOps") != nil {
		return "volume"
	}

	if data.GetMetric("cacheHitsIopsTotal") != nil {
		return "controller"
	}

	return "unknown"
}

// calculateVolumeRatios calculates cache hit ratios for volume objects
func (c *CacheHitRatio) calculateVolumeRatios(data *matrix.Matrix) {
	readHitOps := data.GetMetric("readHitOps")
	writeHitOps := data.GetMetric("writeHitOps")
	readOps := data.GetMetric("readOps")
	writeOps := data.GetMetric("writeOps")

	if readHitOps == nil || writeHitOps == nil || readOps == nil || writeOps == nil {
		c.SLogger.Debug("Missing required metrics for volume cache hit ratio")
		return
	}

	readCacheHitRatio, err := data.NewMetricFloat64("read_cache_hit_ratio")
	if err != nil {
		c.SLogger.Warn("Failed to create read_cache_hit_ratio metric", slog.String("error", err.Error()))
		return
	}

	writeCacheHitRatio, err := data.NewMetricFloat64("write_cache_hit_ratio")
	if err != nil {
		c.SLogger.Warn("Failed to create write_cache_hit_ratio metric", slog.String("error", err.Error()))
		return
	}

	totalCacheHitRatio, err := data.NewMetricFloat64("total_cache_hit_ratio")
	if err != nil {
		c.SLogger.Warn("Failed to create total_cache_hit_ratio metric", slog.String("error", err.Error()))
		return
	}

	for _, instance := range data.GetInstances() {
		// Calculate read cache hit ratio
		readOpsVal, readOpsOk := readOps.GetValueFloat64(instance)
		readHitVal, readHitOk := readHitOps.GetValueFloat64(instance)

		if readOpsOk && readHitOk {
			ratio := c.getCacheHitRatio(readHitVal, readOpsVal)
			if ratio != -1.0 {
				readCacheHitRatio.SetValueFloat64(instance, ratio)
			}
		}

		// Calculate write cache hit ratio
		writeOpsVal, writeOpsOk := writeOps.GetValueFloat64(instance)
		writeHitVal, writeHitOk := writeHitOps.GetValueFloat64(instance)

		if writeOpsOk && writeHitOk {
			ratio := c.getCacheHitRatio(writeHitVal, writeOpsVal)
			if ratio != -1.0 {
				writeCacheHitRatio.SetValueFloat64(instance, ratio)
			}
		}

		// Calculate total cache hit ratio
		if readOpsOk && writeOpsOk && readHitOk && writeHitOk {
			totalOps := readOpsVal + writeOpsVal
			totalHits := readHitVal + writeHitVal
			ratio := c.getCacheHitRatio(totalHits, totalOps)
			if ratio != -1.0 {
				totalCacheHitRatio.SetValueFloat64(instance, ratio)
			}
		}
	}

	c.SLogger.Debug("Calculated volume cache hit ratios", slog.Int("instances", len(data.GetInstances())))
}

// calculateAggregatedRatios calculates cache hit ratio for controller/system objects
func (c *CacheHitRatio) calculateAggregatedRatios(data *matrix.Matrix) {
	cacheHitsIopsTotal := data.GetMetric("cacheHitsIopsTotal")
	readOps := data.GetMetric("readOps")
	writeOps := data.GetMetric("writeOps")

	if cacheHitsIopsTotal == nil || readOps == nil || writeOps == nil {
		c.SLogger.Debug("Missing required metrics for controller/system cache hit ratio")
		return
	}

	totalCacheHitRatio, err := data.NewMetricFloat64("total_cache_hit_ratio")
	if err != nil {
		c.SLogger.Warn("Failed to create total_cache_hit_ratio metric", slog.String("error", err.Error()))
		return
	}

	for _, instance := range data.GetInstances() {
		readOpsVal, readOpsOk := readOps.GetValueFloat64(instance)
		writeOpsVal, writeOpsOk := writeOps.GetValueFloat64(instance)
		cacheHits, cacheHitsOk := cacheHitsIopsTotal.GetValueFloat64(instance)

		if readOpsOk && writeOpsOk && cacheHitsOk {
			totalOps := readOpsVal + writeOpsVal
			ratio := c.getCacheHitRatio(cacheHits, totalOps)
			if ratio != -1.0 {
				totalCacheHitRatio.SetValueFloat64(instance, ratio)
			}
		}
	}

	c.SLogger.Debug("Calculated controller/system cache hit ratios", slog.Int("instances", len(data.GetInstances())))
}

// getCacheHitRatio calculates cache hit ratio as a percentage
// Formula: (hitOps / totalOps) * 100
// Returns:
//   - -1.0 if either parameter is negative (invalid input)
//   - 0.0 if totalOps is zero (no operations)
//   - Value between 0-100, capped at 100%
func (c *CacheHitRatio) getCacheHitRatio(hitOps, totalOps float64) float64 {
	if hitOps < 0 || totalOps < 0 {
		return -1.0
	}

	if totalOps == 0 {
		return 0.0
	}

	ratio := (hitOps / totalOps) * 100.0

	if ratio > 100.0 {
		return 100.0
	}

	return ratio
}
