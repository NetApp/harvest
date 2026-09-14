package collectors

import (
	"log/slog"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

// newVolumeMatrix returns a volume matrix with export options set, as a real
// collector would have them. ProcessFlexGroupFootPrint pops "instance_labels"
// off the export options, so they must not be nil.
func newVolumeMatrix(t *testing.T) *matrix.Matrix {
	t.Helper()
	m := matrix.New("Volume", "volume", "volume")
	m.SetExportOptions(matrix.DefaultExportOptions())
	return m
}

func newVolumeInstance(t *testing.T, m *matrix.Matrix, key string, labels map[string]string) *matrix.Instance {
	t.Helper()
	i, err := m.NewInstance(key)
	if err != nil {
		t.Fatalf("NewInstance(%s) err: %v", key, err)
	}
	for k, v := range labels {
		i.SetLabel(k, v)
	}
	return i
}

// TestProcessFlexGroupDataDuplicateFlexvolKey covers GitHub issue #4095
// "Failed to create new instance in RestPerf:Volume plugin", which shipped with
// the status/untested label.
//
// Two flexvol rows that collapse to the same svm+volume key used to hit
// NewInstance, which returned ErrDuplicateInstanceKey *and* a nil instance,
// logging an error on every poll. The fix switched to GetOrCreateInstance.
func TestProcessFlexGroupDataDuplicateFlexvolKey(t *testing.T) {
	data := newVolumeMatrix(t)

	// Two distinct matrix instances that resolve to the same svm+volume key.
	// This is the #4095 shape: the aggregate key is derived from labels, not
	// from the instance key, so different rows can collide.
	newVolumeInstance(t, data, "uuid-1", map[string]string{"svm": "svm1", "volume": "vol1", "aggr": "aggr1"})
	newVolumeInstance(t, data, "uuid-2", map[string]string{"svm": "svm1", "volume": "vol1", "aggr": "aggr2"})

	volumesMap := map[string]string{"svm1vol1": "flexvol"}

	got, _, err := ProcessFlexGroupData(slog.Default(), data, "style", false, "temp_", volumesMap, true)

	assert.Nil(t, err)
	assert.Equal(t, len(got), 1)

	volumeAggr := got[0]
	// The collision must collapse into exactly one instance, not error out.
	assert.Equal(t, len(volumeAggr.GetInstances()), 1)
	assert.NotNil(t, volumeAggr.GetInstance("svm1.vol1"))

	// The synthetic "labels" metric is still set on the surviving instance.
	labels := volumeAggr.GetMetric("labels")
	assert.NotNil(t, labels)
	v, ok := labels.GetValueFloat64(volumeAggr.GetInstance("svm1.vol1"))
	assert.True(t, ok)
	assert.Equal(t, v, 1.0)
}

// TestProcessFlexGroupDataFlexvolDistinctKeys is the control for the test above:
// genuinely distinct volumes must still produce one instance each, so the
// duplicate-key fix did not collapse unrelated rows.
func TestProcessFlexGroupDataFlexvolDistinctKeys(t *testing.T) {
	data := newVolumeMatrix(t)
	newVolumeInstance(t, data, "uuid-1", map[string]string{"svm": "svm1", "volume": "vol1"})
	newVolumeInstance(t, data, "uuid-2", map[string]string{"svm": "svm1", "volume": "vol2"})

	volumesMap := map[string]string{"svm1vol1": "flexvol", "svm1vol2": "flexvol"}

	got, _, err := ProcessFlexGroupData(slog.Default(), data, "style", false, "temp_", volumesMap, true)

	assert.Nil(t, err)
	assert.Equal(t, len(got), 1)
	assert.Equal(t, len(got[0].GetInstances()), 2)
	assert.NotNil(t, got[0].GetInstance("svm1.vol1"))
	assert.NotNil(t, got[0].GetInstance("svm1.vol2"))
}

// TestProcessFlexGroupDataNilVolumesMap pins the early return: a poller whose
// volumes config has not been populated yet must be a no-op, not an error.
func TestProcessFlexGroupDataNilVolumesMap(t *testing.T) {
	data := newVolumeMatrix(t)
	got, meta, err := ProcessFlexGroupData(slog.Default(), data, "style", false, "temp_", nil, true)

	assert.Nil(t, err)
	assert.Nil(t, got)
	assert.Nil(t, meta)
}

// TestProcessFlexGroupFootPrintMissingTotalFootprint covers GitHub issue #4045
// "Rest:Volume Collector panicked 25.11.0".
//
// A custom template that omits total_footprint left cache.GetMetric
// ("total_footprint") nil while volume_blocks_footprint_bin1 was present. The
// hot-data block was gated only on the capacity-tier metric, so
// totalFootprintMetric.GetValueFloat64 dereferenced a nil *Metric and panicked.
func TestProcessFlexGroupFootPrintMissingTotalFootprint(t *testing.T) {
	data := newVolumeMatrix(t)

	// Present: the capacity-tier metric. Absent: total_footprint.
	bin1, err := data.NewMetricFloat64("volume_blocks_footprint_bin1")
	if err != nil {
		t.Fatalf("NewMetricFloat64 err: %v", err)
	}

	inst := newVolumeInstance(t, data, "uuid-1", map[string]string{
		"svm": "svm1", "volume": "vol1__0001", "aggr": "aggr1", "style": "flexgroup_constituent",
	})
	bin1.SetValueFloat64(inst, 100)

	// Must not panic.
	cache := ProcessFlexGroupFootPrint(data, slog.Default())

	assert.NotNil(t, cache)
	// The FlexGroup rollup instance is still created and the footprint summed.
	fg := cache.GetInstance("svm1.vol1")
	assert.NotNil(t, fg)
	assert.Equal(t, fg.GetLabel("style"), "flexgroup")

	gotBin1, ok := cache.GetMetric("volume_blocks_footprint_bin1").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, gotBin1, 100.0)

	// hot_data cannot be computed without total_footprint, so it must not exist.
	assert.Nil(t, cache.GetMetric("hot_data"))
}

// TestProcessFlexGroupFootPrintHotData is the control: when both metrics are
// present, hot_data is still derived as total - capacity-tier. This proves the
// #4045 guard did not disable the feature it protects.
func TestProcessFlexGroupFootPrintHotData(t *testing.T) {
	data := newVolumeMatrix(t)

	bin1, err := data.NewMetricFloat64("volume_blocks_footprint_bin1")
	if err != nil {
		t.Fatalf("NewMetricFloat64 err: %v", err)
	}
	total, err := data.NewMetricFloat64("total_footprint")
	if err != nil {
		t.Fatalf("NewMetricFloat64 err: %v", err)
	}

	inst := newVolumeInstance(t, data, "uuid-1", map[string]string{
		"svm": "svm1", "volume": "vol1__0001", "aggr": "aggr1", "style": "flexgroup_constituent",
	})
	bin1.SetValueFloat64(inst, 30)
	total.SetValueFloat64(inst, 100)

	cache := ProcessFlexGroupFootPrint(data, slog.Default())

	fg := cache.GetInstance("svm1.vol1")
	assert.NotNil(t, fg)

	hotData := cache.GetMetric("hot_data")
	assert.NotNil(t, hotData)
	got, ok := hotData.GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, got, 70.0)
}

// TestProcessFlexGroupFootPrintConstituentsAreSummed pins the rollup across
// multiple constituents of one FlexGroup, including the sorted aggr label that
// downstream dashboards join on.
func TestProcessFlexGroupFootPrintConstituentsAreSummed(t *testing.T) {
	data := newVolumeMatrix(t)

	bin1, err := data.NewMetricFloat64("volume_blocks_footprint_bin1")
	if err != nil {
		t.Fatalf("NewMetricFloat64 err: %v", err)
	}
	total, err := data.NewMetricFloat64("total_footprint")
	if err != nil {
		t.Fatalf("NewMetricFloat64 err: %v", err)
	}

	// Two constituents, deliberately given aggrs out of sorted order.
	c1 := newVolumeInstance(t, data, "uuid-1", map[string]string{
		"svm": "svm1", "volume": "vol1__0001", "aggr": "aggrB", "style": "flexgroup_constituent",
	})
	c2 := newVolumeInstance(t, data, "uuid-2", map[string]string{
		"svm": "svm1", "volume": "vol1__0002", "aggr": "aggrA", "style": "flexgroup_constituent",
	})
	// A flexvol that must be ignored entirely.
	newVolumeInstance(t, data, "uuid-3", map[string]string{
		"svm": "svm1", "volume": "plain", "aggr": "aggrC", "style": "flexvol",
	})

	bin1.SetValueFloat64(c1, 10)
	bin1.SetValueFloat64(c2, 20)
	total.SetValueFloat64(c1, 100)
	total.SetValueFloat64(c2, 200)

	cache := ProcessFlexGroupFootPrint(data, slog.Default())

	assert.Equal(t, len(cache.GetInstances()), 1)
	fg := cache.GetInstance("svm1.vol1")
	assert.NotNil(t, fg)
	assert.Equal(t, fg.GetLabel("aggr"), "aggrA,aggrB")

	gotTotal, ok := cache.GetMetric("total_footprint").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, gotTotal, 300.0)

	// Summed total footprint minus summed capacity-tier footprint.
	gotHot, ok := cache.GetMetric("hot_data").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, gotHot, 270.0)
}
