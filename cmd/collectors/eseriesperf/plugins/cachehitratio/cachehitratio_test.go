package cachehitratio

import (
	"testing"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func createMockAbstractPlugin() *plugin.AbstractPlugin {
	params := node.NewS("CacheHitRatio")
	p := plugin.New("EseriesPerf", nil, params, nil, "volume_stats", nil)
	_ = p.InitAbc()
	return p
}

func createVolumeMatrix(t *testing.T) *matrix.Matrix {
	t.Helper()
	mat := matrix.New("test", "volume_stats", "volume_stats")
	_, _ = mat.NewMetricFloat64("readOps")
	_, _ = mat.NewMetricFloat64("writeOps")
	_, _ = mat.NewMetricFloat64("readHitOps")
	_, _ = mat.NewMetricFloat64("writeHitOps")
	return mat
}

func createControllerMatrix(t *testing.T) *matrix.Matrix {
	t.Helper()
	mat := matrix.New("test", "controller_stats", "controller_stats")
	_, _ = mat.NewMetricFloat64("readOps")
	_, _ = mat.NewMetricFloat64("writeOps")
	_, _ = mat.NewMetricFloat64("cacheHitsIopsTotal")
	return mat
}

func TestGetCacheHitRatio(t *testing.T) {
	p := &CacheHitRatio{AbstractPlugin: createMockAbstractPlugin()}

	tests := []struct {
		name     string
		hitOps   float64
		totalOps float64
		expected float64
	}{
		{"Normal case", 50.0, 100.0, 50.0},
		{"Perfect hit ratio", 100.0, 100.0, 100.0},
		{"No hits", 0.0, 100.0, 0.0},
		{"No operations", 0.0, 0.0, 0.0},
		{"Over 100% - should cap", 110.0, 100.0, 100.0},
		{"Negative hit ops", -10.0, 100.0, -1.0},
		{"Negative total ops", 50.0, -100.0, -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := p.getCacheHitRatio(tt.hitOps, tt.totalOps); result != tt.expected {
				t.Errorf("got %f; want %f", result, tt.expected)
			}
		})
	}
}

func TestDetectObjectType(t *testing.T) {
	p := &CacheHitRatio{AbstractPlugin: createMockAbstractPlugin()}

	tests := []struct {
		name      string
		setupFunc func() *matrix.Matrix
		wantType  string
	}{
		{
			name: "Volume object",
			setupFunc: func() *matrix.Matrix {
				mat := matrix.New("test", "test", "test")
				_, _ = mat.NewMetricFloat64("readHitOps")
				_, _ = mat.NewMetricFloat64("writeHitOps")
				return mat
			},
			wantType: "volume",
		},
		{
			name: "Controller object",
			setupFunc: func() *matrix.Matrix {
				mat := matrix.New("test", "test", "test")
				_, _ = mat.NewMetricFloat64("cacheHitsIopsTotal")
				return mat
			},
			wantType: "controller",
		},
		{
			name: "Unknown object",
			setupFunc: func() *matrix.Matrix {
				return matrix.New("test", "test", "test")
			},
			wantType: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mat := tt.setupFunc()
			if got := p.detectObjectType(mat); got != tt.wantType {
				t.Errorf("got '%s'; want '%s'", got, tt.wantType)
			}
		})
	}
}

func TestCalculateVolumeRatios(t *testing.T) {
	p := &CacheHitRatio{AbstractPlugin: createMockAbstractPlugin()}
	mat := createVolumeMatrix(t)

	// Create test instance with 80% read hit, 60% write hit
	instance, _ := mat.NewInstance("vol1")
	mat.GetMetric("readOps").SetValueFloat64(instance, 100.0)
	mat.GetMetric("writeOps").SetValueFloat64(instance, 100.0)
	mat.GetMetric("readHitOps").SetValueFloat64(instance, 80.0)
	mat.GetMetric("writeHitOps").SetValueFloat64(instance, 60.0)

	p.calculateVolumeRatios(mat)

	// Helper to verify metric existence and value
	checkMetric := func(name string, expected float64) {
		t.Helper()
		metric := mat.GetMetric(name)
		if metric == nil {
			t.Errorf("%s metric was not created", name)
			return
		}
		if val, ok := metric.GetValueFloat64(instance); !ok {
			t.Errorf("Failed to get %s value", name)
		} else if val != expected {
			t.Errorf("%s = %f; want %f", name, val, expected)
		}
	}

	checkMetric("readCacheHitRatio", 80.0)
	checkMetric("writeCacheHitRatio", 60.0)
	checkMetric("totalCacheHitRatio", 70.0) // (80 + 60) / 200 * 100
}

func TestCalculateAggregatedRatios(t *testing.T) {
	p := &CacheHitRatio{AbstractPlugin: createMockAbstractPlugin()}
	mat := createControllerMatrix(t)

	// Create test instance with 75% total hit ratio
	instance, _ := mat.NewInstance("controller1")
	mat.GetMetric("readOps").SetValueFloat64(instance, 100.0)
	mat.GetMetric("writeOps").SetValueFloat64(instance, 100.0)
	mat.GetMetric("cacheHitsIopsTotal").SetValueFloat64(instance, 150.0)

	p.calculateAggregatedRatios(mat)

	// Verify: 150 / 200 * 100 = 75%
	metric := mat.GetMetric("totalCacheHitRatio")
	if metric == nil {
		t.Fatal("totalCacheHitRatio metric was not created")
	}
	if val, ok := metric.GetValueFloat64(instance); !ok {
		t.Error("Failed to get totalCacheHitRatio value")
	} else if val != 75.0 {
		t.Errorf("totalCacheHitRatio = %f; want 75.0", val)
	}
}

func TestRun_VolumeObject(t *testing.T) {
	p := &CacheHitRatio{AbstractPlugin: createMockAbstractPlugin()}
	p.Object = "volume_stats"

	mat := createVolumeMatrix(t)
	dataMap := map[string]*matrix.Matrix{"volume_stats": mat}

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

	// Verify all cache hit ratio metrics were created
	for _, name := range []string{"readCacheHitRatio", "writeCacheHitRatio", "totalCacheHitRatio"} {
		if mat.GetMetric(name) == nil {
			t.Errorf("%s metric was not created by Run()", name)
		}
	}
}
