package svm

import (
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"testing"
)

func TestMCCSVMs(t *testing.T) {
	data := populatedData()
	opts := options.New()
	s := &SVM{AbstractPlugin: plugin.New("svm", opts, nil, nil, "svm", nil)}

	s.updateSVM(data)

	// svm1-mc should be exportable
	instance := data.GetInstance("svm1-mc")
	if exportable := instance.IsExportable(); !exportable {
		t.Fatalf("%s exported should be true, got %t", "svm1-mc", exportable)
	}

	// svm2-mc should not be exportable
	instance = data.GetInstance("svm2-mc")
	if exportable := instance.IsExportable(); exportable {
		t.Fatalf("%s exported should be false, got %t", "svm2-mc", exportable)
	}

	// svm-test should be exportable
	instance = data.GetInstance("svm-test")
	if exportable := instance.IsExportable(); !exportable {
		t.Fatalf("%s exported should be true, got %t", "svm-test", exportable)
	}
}

func populatedData() *matrix.Matrix {
	// Create test data
	data := matrix.New("svm", "svm", "svm")

	// instance in svm-mc with online state
	instance1, _ := data.NewInstance("svm1-mc")
	instance1.SetLabel("svm", "svm1-mc")
	instance1.SetLabel("state", "online")

	// instance in svm-mc with offline state
	instance2, _ := data.NewInstance("svm2-mc")
	instance2.SetLabel("svm", "svm2-mc")
	instance2.SetLabel("state", "offline")

	// instance in other svm with offline state
	instance3, _ := data.NewInstance("svm-test")
	instance3.SetLabel("svm", "svm-test")
	instance3.SetLabel("state", "offline")

	return data
}
