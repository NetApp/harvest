package ssdcachestats

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func createMockPlugin() *SsdCacheStats {
	params := node.NewS("SsdCacheStats")
	p := plugin.New("EseriesPerf", nil, params, nil, "eseries_ssd_cache", nil)
	_ = p.InitAbc()
	return &SsdCacheStats{AbstractPlugin: p}
}

func createSsdCacheMatrix(t *testing.T) *matrix.Matrix {
	t.Helper()
	mat := matrix.New("test", "eseries_ssd_cache", "eseries_ssd_cache")
	_, _ = mat.NewMetricFloat64("statistics.reads")
	_, _ = mat.NewMetricFloat64("statistics.writes")
	_, _ = mat.NewMetricFloat64("statistics.readBlocks")
	_, _ = mat.NewMetricFloat64("statistics.writeBlocks")
	_, _ = mat.NewMetricFloat64("statistics.fullCacheHits")
	_, _ = mat.NewMetricFloat64("statistics.fullCacheHitBlocks")
	_, _ = mat.NewMetricFloat64("statistics.partialCacheHits")
	_, _ = mat.NewMetricFloat64("statistics.partialCacheHitBlocks")
	_, _ = mat.NewMetricFloat64("statistics.completeCacheMiss")
	_, _ = mat.NewMetricFloat64("statistics.completeCacheMissBlocks")
	_, _ = mat.NewMetricFloat64("statistics.populateOnReads")
	_, _ = mat.NewMetricFloat64("statistics.populateOnWrites")
	_, _ = mat.NewMetricFloat64("statistics.invalidates")
	_, _ = mat.NewMetricFloat64("statistics.recycles")
	_, _ = mat.NewMetricFloat64("statistics.availableBytes")
	_, _ = mat.NewMetricFloat64("statistics.allocatedBytes")
	_, _ = mat.NewMetricFloat64("statistics.populatedCleanBytes")
	_, _ = mat.NewMetricFloat64("statistics.populatedDirtyBytes")
	return mat
}

func setMetric(t *testing.T, mat *matrix.Matrix, name string, inst *matrix.Instance, val float64) {
	t.Helper()
	m := mat.GetMetric(name)
	if m == nil {
		t.Fatalf("metric %s not found", name)
	}
	m.SetValueFloat64(inst, val)
}

func TestSsdCacheStats_CacheHitPercent(t *testing.T) {
	p := createMockPlugin()

	tests := []struct {
		name          string
		reads         float64
		writes        float64
		fullCacheHits float64
		wantSet       bool
	}{
		{
			name:          "Normal case",
			reads:         53.5,
			writes:        23.847,
			fullCacheHits: 17.028,
			wantSet:       true,
		},
		{
			name:          "High hit rate",
			reads:         100.0,
			writes:        0.0,
			fullCacheHits: 90.0,
			wantSet:       true,
		},
		{
			name:          "Zero reads and writes",
			reads:         0.0,
			writes:        0.0,
			fullCacheHits: 0.0,
			wantSet:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat := createSsdCacheMatrix(t)
			inst, _ := mat.NewInstance("ctrl1")

			setMetric(t, mat, "statistics.reads", inst, tt.reads)
			setMetric(t, mat, "statistics.writes", inst, tt.writes)
			setMetric(t, mat, "statistics.fullCacheHits", inst, tt.fullCacheHits)

			p.calculateCacheHitPercent(mat)

			hitPct := mat.GetMetric("hit_percent")
			if hitPct == nil {
				t.Fatal("hit_percent metric was not created")
			}

			val, ok := hitPct.GetValueFloat64(inst)
			if tt.wantSet {
				if !ok {
					t.Fatal("expected hit_percent value to be set")
				}
				assert.Equal(t, val, (tt.fullCacheHits/(tt.reads+tt.writes))*100.0)
			} else if ok && val != 0 {
				t.Errorf("expected no value or zero for hit_percent, got %f", val)
			}
		})
	}
}

func TestSsdCacheStats_AllocationPercent(t *testing.T) {
	p := createMockPlugin()

	tests := []struct {
		name      string
		allocated float64
		available float64
		wantSet   bool
	}{
		{
			name:      "Normal allocation",
			allocated: 13316915200.0,
			available: 1599784091648.0,
			wantSet:   true,
		},
		{
			name:      "Fully allocated",
			allocated: 100.0,
			available: 0.0,
			wantSet:   true,
		},
		{
			name:      "Zero total",
			allocated: 0.0,
			available: 0.0,
			wantSet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat := createSsdCacheMatrix(t)
			inst, _ := mat.NewInstance("ctrl1")

			setMetric(t, mat, "statistics.allocatedBytes", inst, tt.allocated)
			setMetric(t, mat, "statistics.availableBytes", inst, tt.available)

			p.calculateCacheAllocationPercent(mat)

			allocPct := mat.GetMetric("allocation_percent")
			if allocPct == nil {
				t.Fatal("allocation_percent metric was not created")
			}

			val, ok := allocPct.GetValueFloat64(inst)
			if tt.wantSet {
				if !ok {
					t.Fatal("expected allocation_percent value to be set")
				}
				assert.Equal(t, val, (tt.allocated/(tt.allocated+tt.available))*100.0)
			} else if ok && val != 0 {
				t.Errorf("expected no value or zero, got %f", val)
			}
		})
	}
}

func TestSsdCacheStats_UtilizationPercent(t *testing.T) {
	p := createMockPlugin()

	tests := []struct {
		name      string
		clean     float64
		dirty     float64
		allocated float64
		wantSet   bool
	}{
		{
			name:      "Normal utilization",
			clean:     3460448256.0,
			dirty:     0.0,
			allocated: 13316915200.0,
			wantSet:   true,
		},
		{
			name:      "With dirty bytes",
			clean:     3000000000.0,
			dirty:     1000000000.0,
			allocated: 10000000000.0,
			wantSet:   true,
		},
		{
			name:      "Zero allocated",
			clean:     100.0,
			dirty:     50.0,
			allocated: 0.0,
			wantSet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat := createSsdCacheMatrix(t)
			inst, _ := mat.NewInstance("ctrl1")

			setMetric(t, mat, "statistics.populatedCleanBytes", inst, tt.clean)
			setMetric(t, mat, "statistics.populatedDirtyBytes", inst, tt.dirty)
			setMetric(t, mat, "statistics.allocatedBytes", inst, tt.allocated)

			p.calculateCacheUtilizationPercent(mat)

			utilPct := mat.GetMetric("utilization_percent")
			if utilPct == nil {
				t.Fatal("utilization_percent metric was not created")
			}

			val, ok := utilPct.GetValueFloat64(inst)
			if tt.wantSet {
				if !ok {
					t.Fatal("expected utilization_percent value to be set")
				}
				assert.Equal(t, val, ((tt.clean+tt.dirty)/tt.allocated)*100.0)
			} else if ok && val != 0 {
				t.Errorf("expected no value or zero, got %f", val)
			}
		})
	}
}

func TestSsdCacheStats_ComputePercent(t *testing.T) {
	p := createMockPlugin()

	tests := []struct {
		name       string
		numKey     string
		denKey     string
		resultName string
		numVal     float64
		denVal     float64
		wantSet    bool
	}{
		{
			name:       "full_cache_hit_percent",
			numKey:     "statistics.fullCacheHits",
			denKey:     "statistics.reads",
			resultName: "full_cache_hit_percent",
			numVal:     17.028,
			denVal:     53.5,
			wantSet:    true,
		},
		{
			name:       "partial_cache_hit_percent",
			numKey:     "statistics.partialCacheHits",
			denKey:     "statistics.reads",
			resultName: "partial_cache_hit_percent",
			numVal:     36.472,
			denVal:     53.5,
			wantSet:    true,
		},
		{
			name:       "zero denominator",
			numKey:     "statistics.fullCacheHits",
			denKey:     "statistics.reads",
			resultName: "zero_test_percent",
			numVal:     10.0,
			denVal:     0.0,
			wantSet:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat := createSsdCacheMatrix(t)
			inst, _ := mat.NewInstance("ctrl1")

			setMetric(t, mat, tt.numKey, inst, tt.numVal)
			setMetric(t, mat, tt.denKey, inst, tt.denVal)

			p.computePercent(mat, tt.numKey, tt.denKey, tt.resultName)

			result := mat.GetMetric(tt.resultName)
			if result == nil {
				t.Fatalf("metric %s was not created", tt.resultName)
			}

			val, ok := result.GetValueFloat64(inst)
			if tt.wantSet {
				if !ok {
					t.Fatalf("expected %s value to be set", tt.resultName)
				}
				assert.Equal(t, val, (tt.numVal/tt.denVal)*100.0)
			} else if ok && val != 0 {
				t.Errorf("expected no value or zero for %s, got %f", tt.resultName, val)
			}
		})
	}
}

func TestSsdCacheStats_ZeroDenominator(t *testing.T) {
	p := createMockPlugin()
	mat := createSsdCacheMatrix(t)

	inst, _ := mat.NewInstance("ctrl1")
	for _, name := range []string{
		"statistics.reads", "statistics.writes",
		"statistics.readBlocks", "statistics.writeBlocks",
		"statistics.fullCacheHits", "statistics.fullCacheHitBlocks",
		"statistics.partialCacheHits", "statistics.partialCacheHitBlocks",
		"statistics.completeCacheMiss", "statistics.completeCacheMissBlocks",
		"statistics.populateOnReads",
		"statistics.populateOnWrites",
		"statistics.invalidates", "statistics.recycles",
		"statistics.availableBytes", "statistics.allocatedBytes",
		"statistics.populatedCleanBytes", "statistics.populatedDirtyBytes",
	} {
		setMetric(t, mat, name, inst, 0.0)
	}

	dataMap := map[string]*matrix.Matrix{"eseries_ssd_cache": mat}
	result, metadata, err := p.Run(dataMap)

	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if result != nil {
		t.Error("Run() should return nil matrices")
	}
	if metadata != nil {
		t.Error("Run() should return nil metadata")
	}

	hitPct := mat.GetMetric("hit_percent")
	if hitPct != nil {
		val, ok := hitPct.GetValueFloat64(inst)
		if ok && val != 0 {
			t.Errorf("hit_percent should be unset for zero denominators, got %f", val)
		}
	}
}

func TestSsdCacheStats_MissingMetrics(t *testing.T) {
	p := createMockPlugin()

	mat := matrix.New("test", "eseries_ssd_cache", "eseries_ssd_cache")
	_, _ = mat.NewInstance("ctrl1")

	dataMap := map[string]*matrix.Matrix{"eseries_ssd_cache": mat}

	result, metadata, err := p.Run(dataMap)

	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if result != nil {
		t.Error("Run() should return nil matrices")
	}
	if metadata != nil {
		t.Error("Run() should return nil metadata")
	}
}

func TestSsdCacheStats_Run(t *testing.T) {
	p := createMockPlugin()
	mat := createSsdCacheMatrix(t)

	ctrl1, _ := mat.NewInstance("ctrl1")
	ctrl2, _ := mat.NewInstance("ctrl2")

	setMetric(t, mat, "statistics.reads", ctrl1, 53.5)
	setMetric(t, mat, "statistics.writes", ctrl1, 23.847)
	setMetric(t, mat, "statistics.readBlocks", ctrl1, 41343.36)
	setMetric(t, mat, "statistics.writeBlocks", ctrl1, 36389.22)
	setMetric(t, mat, "statistics.fullCacheHits", ctrl1, 17.028)
	setMetric(t, mat, "statistics.fullCacheHitBlocks", ctrl1, 1958.19)
	setMetric(t, mat, "statistics.partialCacheHits", ctrl1, 36.472)
	setMetric(t, mat, "statistics.partialCacheHitBlocks", ctrl1, 39382.72)
	setMetric(t, mat, "statistics.completeCacheMiss", ctrl1, 0.0)
	setMetric(t, mat, "statistics.completeCacheMissBlocks", ctrl1, 0.0)
	setMetric(t, mat, "statistics.populateOnReads", ctrl1, 9.097)
	setMetric(t, mat, "statistics.populateOnWrites", ctrl1, 1.556)
	setMetric(t, mat, "statistics.invalidates", ctrl1, 23.847)
	setMetric(t, mat, "statistics.recycles", ctrl1, 0.0)
	setMetric(t, mat, "statistics.availableBytes", ctrl1, 1599784091648.0)
	setMetric(t, mat, "statistics.allocatedBytes", ctrl1, 13316915200.0)
	setMetric(t, mat, "statistics.populatedCleanBytes", ctrl1, 3460448256.0)
	setMetric(t, mat, "statistics.populatedDirtyBytes", ctrl1, 0.0)

	setMetric(t, mat, "statistics.reads", ctrl2, 55.845)
	setMetric(t, mat, "statistics.writes", ctrl2, 25.592)
	setMetric(t, mat, "statistics.readBlocks", ctrl2, 47025.38)
	setMetric(t, mat, "statistics.writeBlocks", ctrl2, 39547.97)
	setMetric(t, mat, "statistics.fullCacheHits", ctrl2, 10.901)
	setMetric(t, mat, "statistics.fullCacheHitBlocks", ctrl2, 953.07)
	setMetric(t, mat, "statistics.partialCacheHits", ctrl2, 44.930)
	setMetric(t, mat, "statistics.partialCacheHitBlocks", ctrl2, 46072.42)
	setMetric(t, mat, "statistics.completeCacheMiss", ctrl2, 0.0)
	setMetric(t, mat, "statistics.completeCacheMissBlocks", ctrl2, 0.0)
	setMetric(t, mat, "statistics.populateOnReads", ctrl2, 7.521)
	setMetric(t, mat, "statistics.populateOnWrites", ctrl2, 0.592)
	setMetric(t, mat, "statistics.invalidates", ctrl2, 25.592)
	setMetric(t, mat, "statistics.recycles", ctrl2, 0.0)
	setMetric(t, mat, "statistics.availableBytes", ctrl2, 1599784091648.0)
	setMetric(t, mat, "statistics.allocatedBytes", ctrl2, 13421772800.0)
	setMetric(t, mat, "statistics.populatedCleanBytes", ctrl2, 1750466560.0)
	setMetric(t, mat, "statistics.populatedDirtyBytes", ctrl2, 0.0)

	dataMap := map[string]*matrix.Matrix{"eseries_ssd_cache": mat}
	result, metadata, err := p.Run(dataMap)

	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if result != nil {
		t.Error("Run() should return nil matrices")
	}
	if metadata != nil {
		t.Error("Run() should return nil metadata")
	}

	expectedMetrics := []string{
		"hit_percent",
		"allocation_percent",
		"utilization_percent",
		"full_cache_hit_percent",
		"partial_cache_hit_percent",
		"complete_cache_miss_percent",
	}

	for _, name := range expectedMetrics {
		if mat.GetMetric(name) == nil {
			t.Errorf("metric %s was not created by Run()", name)
		}
	}

	checkMetricValue := func(name string, inst *matrix.Instance, expected float64) {
		t.Helper()
		m := mat.GetMetric(name)
		if m == nil {
			t.Errorf("metric %s not found", name)
			return
		}
		val, ok := m.GetValueFloat64(inst)
		if !ok {
			t.Errorf("no value for %s", name)
			return
		}
		assert.Equal(t, val, expected)
	}

	reads1, writes1, hits1 := 53.5, 23.847, 17.028
	alloc1, avail1, clean1 := 13316915200.0, 1599784091648.0, 3460448256.0
	reads2, writes2, hits2 := 55.845, 25.592, 10.901
	checkMetricValue("hit_percent", ctrl1, (hits1/(reads1+writes1))*100.0)
	checkMetricValue("allocation_percent", ctrl1, (alloc1/(alloc1+avail1))*100.0)
	checkMetricValue("utilization_percent", ctrl1, (clean1/alloc1)*100.0)
	checkMetricValue("full_cache_hit_percent", ctrl1, (hits1/reads1)*100.0)
	checkMetricValue("hit_percent", ctrl2, (hits2/(reads2+writes2))*100.0)
}

func TestSsdCacheStats_NilDataMap(t *testing.T) {
	p := createMockPlugin()

	dataMap := map[string]*matrix.Matrix{"wrong_object": matrix.New("x", "x", "x")}
	result, metadata, err := p.Run(dataMap)

	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if result != nil {
		t.Error("Run() should return nil for missing object")
	}
	if metadata != nil {
		t.Error("Run() should return nil metadata")
	}
}
