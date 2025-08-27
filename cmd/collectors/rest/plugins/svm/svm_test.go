package svm

import (
	"github.com/netapp/harvest/v2/assert"
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
	assert.True(t, instance.IsExportable())

	// svm2-mc should not be exportable
	instance = data.GetInstance("svm2-mc")
	assert.False(t, instance.IsExportable())

	// svm-test should be exportable
	instance = data.GetInstance("svm-test")
	assert.True(t, instance.IsExportable())
}

func populatedData() *matrix.Matrix {
	// Create test data
	data := matrix.New("svm", "svm", "svm")

	instance1, _ := data.NewInstance("svm1-mc")
	instance1.SetLabel("svm", "svm1-mc")
	instance1.SetLabel("state", "running")

	instance2, _ := data.NewInstance("svm2-mc")
	instance2.SetLabel("svm", "svm2-mc")
	instance2.SetLabel("state", "stopped")

	instance3, _ := data.NewInstance("svm-test")
	instance3.SetLabel("svm", "svm-test")
	instance3.SetLabel("state", "stopped")

	return data
}
