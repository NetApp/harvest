package volume

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"testing"
)

func TestMCCVolumes(t *testing.T) {
	volumesMap := make(map[string]volumeInfo)
	opts := options.New()
	v := &Volume{AbstractPlugin: plugin.New("volume", opts, nil, nil, "volume", nil)}
	v.volTagMap = make(map[string]volumeTag)

	// Create test data
	data := matrix.New("volume", "volume", "volume")
	// instance in svm-mc with online state
	instance1, _ := data.NewInstance("vol_test1")
	instance1.SetLabel("volume", "vol_test1")
	instance1.SetLabel("svm", "svm-mc")
	volumesMap["vol_test1"+"svm-mc"] = volumeInfo{}

	// instance in svm-mc with online state
	instance2, _ := data.NewInstance("vol_test2")
	instance2.SetLabel("volume", "vol_test2")
	instance2.SetLabel("svm", "svm-mc")
	instance2.SetLabel("state", "online")
	volumesMap["vol_test2"+"svm-mc"] = volumeInfo{}

	// instance in other svm with offline state
	instance3, _ := data.NewInstance("vol_test3")
	instance3.SetLabel("volume", "vol_test3")
	instance3.SetLabel("svm", "svm-test")
	instance3.SetLabel("state", "offline")
	volumesMap["vol_test3"+"svm-test"] = volumeInfo{}

	v.updateVolumeLabels(data, volumesMap)

	// vol_test1 should not be exported
	instance := data.GetInstance("vol_test1")
	assert.False(t, instance.IsExportable())

	// vol_test2 should be exported
	instance = data.GetInstance("vol_test2")
	assert.True(t, instance.IsExportable())

	// vol_test3 should be exported
	instance = data.GetInstance("vol_test3")
	assert.True(t, instance.IsExportable())
}
