package ssdcachecapacity

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func createMockPlugin() *SsdCacheCapacity {
	params := node.NewS("SsdCacheCapacity")
	p := plugin.New("ESeries", nil, params, nil, "eseries_ssd_cache", nil)
	_ = p.InitAbc()
	s := &SsdCacheCapacity{AbstractPlugin: p}
	s.initVolumeMatrix()
	s.initDriveMatrix()
	return s
}

func TestPopulateMappings_CommaJoinedLabels(t *testing.T) {
	s := createMockPlugin()
	s.volumeNames = map[string]string{
		"vol-1": "repos_0000",
		"vol-2": "prod-db",
	}
	s.driveInfo = map[string]driveData{
		"drive-1": {location: "14", rawCapacity: 1600321314816},
		"drive-2": {location: "15", rawCapacity: 1600321314816},
	}

	data := matrix.New("test", "eseries_ssd_cache", "eseries_ssd_cache")
	inst, err := data.NewInstance("cache-1")
	if err != nil {
		t.Fatalf("failed to create ssd_cache instance: %v", err)
	}
	inst.SetLabel("ssd_cache", "SSD_Cache")
	inst.SetLabel("cached_volume_ids", "vol-1,vol-2")
	inst.SetLabel("drive_ids", "drive-1,drive-2")

	s.populateMappings(data)

	assert.Equal(t, len(s.volumeMat.GetInstances()), 2)
	assert.Equal(t, len(s.driveMat.GetInstances()), 2)

	vol1 := s.volumeMat.GetInstance("cache-1_vol-1")
	if vol1 == nil {
		t.Fatalf("expected volume instance cache-1_vol-1 not found")
	}
	assert.Equal(t, vol1.GetLabel("volume"), "repos_0000")
	assert.Equal(t, vol1.GetLabel("ssd_cache"), "SSD_Cache")

	drv1 := s.driveMat.GetInstance("cache-1_drive-1")
	if drv1 == nil {
		t.Fatalf("expected drive instance cache-1_drive-1 not found")
	}
	assert.Equal(t, drv1.GetLabel("drive"), "14")

	rawCapMetric := s.driveMat.GetMetric("raw_capacity")
	if rawCapMetric == nil {
		t.Fatalf("raw_capacity metric not found")
	}
	val, ok := rawCapMetric.GetValueFloat64(drv1)
	if !ok || val != 1600321314816 {
		t.Fatalf("expected raw_capacity 1600321314816, got %v (ok=%v)", val, ok)
	}
}

func TestPopulateMappings_SingleID(t *testing.T) {
	s := createMockPlugin()

	data := matrix.New("test", "eseries_ssd_cache", "eseries_ssd_cache")
	inst, err := data.NewInstance("cache-1")
	if err != nil {
		t.Fatalf("failed to create ssd_cache instance: %v", err)
	}
	inst.SetLabel("ssd_cache", "SSD_Cache")
	inst.SetLabel("cached_volume_ids", "vol-1")
	inst.SetLabel("drive_ids", "drive-1")

	s.populateMappings(data)

	assert.Equal(t, len(s.volumeMat.GetInstances()), 1)
	assert.Equal(t, len(s.driveMat.GetInstances()), 1)
}

func TestPopulateMappings_EmptyLabels(t *testing.T) {
	s := createMockPlugin()

	data := matrix.New("test", "eseries_ssd_cache", "eseries_ssd_cache")
	inst, err := data.NewInstance("cache-1")
	if err != nil {
		t.Fatalf("failed to create ssd_cache instance: %v", err)
	}
	inst.SetLabel("ssd_cache", "SSD_Cache")

	s.populateMappings(data)

	assert.Equal(t, len(s.volumeMat.GetInstances()), 0)
	assert.Equal(t, len(s.driveMat.GetInstances()), 0)
}
