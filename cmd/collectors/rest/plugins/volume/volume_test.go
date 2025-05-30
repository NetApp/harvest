package volume

import (
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
	instance1.SetLabel("state", "online")
	volumesMap["vol_test1"+"svm-mc"] = volumeInfo{}

	// instance in svm-mc with offline state
	instance2, _ := data.NewInstance("vol_test2")
	instance2.SetLabel("volume", "vol_test2")
	instance2.SetLabel("svm", "svm-mc")
	instance2.SetLabel("state", "offline")
	volumesMap["vol_test2"+"svm-mc"] = volumeInfo{}

	// instance in other svm with offline state
	instance3, _ := data.NewInstance("vol_test3")
	instance3.SetLabel("volume", "vol_test3")
	instance3.SetLabel("svm", "svm-test")
	instance3.SetLabel("state", "offline")
	volumesMap["vol_test3"+"svm-test"] = volumeInfo{}

	v.updateVolumeLabels(data, volumesMap)

	// vol_test1 should be exportable
	instance := data.GetInstance("vol_test1")
	if exportable := instance.IsExportable(); !exportable {
		t.Fatalf("%s exported should be true, got %t", "vol_test1", exportable)
	}

	// vol_test2 should not be exportable
	instance = data.GetInstance("vol_test2")
	if exportable := instance.IsExportable(); exportable {
		t.Fatalf("%s exported should be false, got %t", "vol_test2", exportable)
	}

	// vol_test3 should be exportable
	instance = data.GetInstance("vol_test3")
	if exportable := instance.IsExportable(); !exportable {
		t.Fatalf("%s exported should be true, got %t", "vol_test3", exportable)
	}

}
