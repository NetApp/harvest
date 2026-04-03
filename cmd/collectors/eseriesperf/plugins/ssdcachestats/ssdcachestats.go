package ssdcachestats

import (
	"log/slog"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type SsdCacheStats struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SsdCacheStats{AbstractPlugin: p}
}

func (s *SsdCacheStats) Init(_ conf.Remote) error {
	return s.InitAbc()
}

func (s *SsdCacheStats) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[s.Object]
	if data == nil {
		s.SLogger.Debug("No data matrix available")
		return nil, nil, nil
	}

	s.calculateCacheHitPercent(data)
	s.calculateCacheAllocationPercent(data)
	s.calculateCacheUtilizationPercent(data)

	s.computePercent(data, "statistics.fullCacheHits", "statistics.reads", "full_cache_hit_percent")
	s.computePercent(data, "statistics.partialCacheHits", "statistics.reads", "partial_cache_hit_percent")
	s.computePercent(data, "statistics.completeCacheMiss", "statistics.reads", "complete_cache_miss_percent")
	s.computePercent(data, "statistics.populateOnReads", "statistics.reads", "populate_on_read_percent")

	s.computePercent(data, "statistics.fullCacheHitBlocks", "statistics.readBlocks", "full_cache_hit_block_percent")
	s.computePercent(data, "statistics.partialCacheHitBlocks", "statistics.readBlocks", "partial_cache_hit_block_percent")
	s.computePercent(data, "statistics.completeCacheMissBlocks", "statistics.readBlocks", "complete_cache_miss_block_percent")
	s.computePercent(data, "statistics.populateOnReadBlocks", "statistics.readBlocks", "populate_on_read_block_percent")

	s.computePercent(data, "statistics.populateOnWrites", "statistics.writes", "populate_on_write_percent")
	s.computePercent(data, "statistics.invalidates", "statistics.writes", "invalidate_percent")

	s.computePercent(data, "statistics.populateOnWriteBlocks", "statistics.writeBlocks", "populate_on_write_block_percent")

	return nil, nil, nil
}

func (s *SsdCacheStats) computePercent(data *matrix.Matrix, numeratorKey, denominatorKey, resultName string) {
	numerator := data.GetMetric(numeratorKey)
	denominator := data.GetMetric(denominatorKey)

	if numerator == nil || denominator == nil {
		return
	}

	result, err := data.NewMetricFloat64(resultName)
	if err != nil {
		s.SLogger.Warn("Failed to create metric", slog.String("metric", resultName), slog.String("error", err.Error()))
		return
	}

	for _, instance := range data.GetInstances() {
		numVal, numOk := numerator.GetValueFloat64(instance)
		denVal, denOk := denominator.GetValueFloat64(instance)

		if numOk && denOk && denVal > 0 {
			result.SetValueFloat64(instance, (numVal/denVal)*100.0)
		}
	}
}

// calculateCacheHitPercent computes cache hit percentage:
// fullCacheHits / (reads + writes) * 100
func (s *SsdCacheStats) calculateCacheHitPercent(data *matrix.Matrix) {
	reads := data.GetMetric("statistics.reads")
	writes := data.GetMetric("statistics.writes")
	fullCacheHits := data.GetMetric("statistics.fullCacheHits")

	if reads == nil || writes == nil || fullCacheHits == nil {
		s.SLogger.Debug("Missing metrics for cache hit percent")
		return
	}

	cacheHitPct, err := data.NewMetricFloat64("hit_percent")
	if err != nil {
		s.SLogger.Warn("Failed to create hit_percent metric", slog.String("error", err.Error()))
		return
	}

	for _, instance := range data.GetInstances() {
		readsVal, readsOk := reads.GetValueFloat64(instance)
		writesVal, writesOk := writes.GetValueFloat64(instance)
		fullHitsVal, fullOk := fullCacheHits.GetValueFloat64(instance)

		if readsOk && writesOk && fullOk {
			totalIOs := readsVal + writesVal
			if totalIOs > 0 {
				pct := (fullHitsVal / totalIOs) * 100.0
				cacheHitPct.SetValueFloat64(instance, pct)
			}
		}
	}
}

// calculateCacheAllocationPercent computes cache allocation percentage:
// allocatedBytes / (allocatedBytes + availableBytes) * 100
func (s *SsdCacheStats) calculateCacheAllocationPercent(data *matrix.Matrix) {
	allocated := data.GetMetric("statistics.allocatedBytes")
	available := data.GetMetric("statistics.availableBytes")

	if allocated == nil || available == nil {
		s.SLogger.Debug("Missing metrics for cache allocation percent")
		return
	}

	allocPct, err := data.NewMetricFloat64("allocation_percent")
	if err != nil {
		s.SLogger.Warn("Failed to create allocation_percent metric", slog.String("error", err.Error()))
		return
	}

	for _, instance := range data.GetInstances() {
		allocVal, allocOk := allocated.GetValueFloat64(instance)
		availVal, availOk := available.GetValueFloat64(instance)

		if allocOk && availOk {
			total := allocVal + availVal
			if total > 0 {
				pct := (allocVal / total) * 100.0
				allocPct.SetValueFloat64(instance, pct)
			}
		}
	}
}

// calculateCacheUtilizationPercent computes cache utilization percentage:
// (populatedCleanBytes + populatedDirtyBytes) / allocatedBytes * 100
func (s *SsdCacheStats) calculateCacheUtilizationPercent(data *matrix.Matrix) {
	clean := data.GetMetric("statistics.populatedCleanBytes")
	dirty := data.GetMetric("statistics.populatedDirtyBytes")
	allocated := data.GetMetric("statistics.allocatedBytes")

	if clean == nil || dirty == nil || allocated == nil {
		s.SLogger.Debug("Missing metrics for cache utilization percent")
		return
	}

	utilPct, err := data.NewMetricFloat64("utilization_percent")
	if err != nil {
		s.SLogger.Warn("Failed to create utilization_percent metric", slog.String("error", err.Error()))
		return
	}

	for _, instance := range data.GetInstances() {
		cleanVal, cleanOk := clean.GetValueFloat64(instance)
		dirtyVal, dirtyOk := dirty.GetValueFloat64(instance)
		allocVal, allocOk := allocated.GetValueFloat64(instance)

		if cleanOk && dirtyOk && allocOk && allocVal > 0 {
			pct := ((cleanVal + dirtyVal) / allocVal) * 100.0
			utilPct.SetValueFloat64(instance, pct)
		}
	}
}
