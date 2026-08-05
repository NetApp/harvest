package maxplugin

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func newMax(t *testing.T, rules ...string) *Max {
	t.Helper()

	params := node.NewS("Max")
	for _, r := range rules {
		params.NewChildS("", r)
	}

	m := New(plugin.New("Test", nil, params, nil, "disk", nil))
	assert.Nil(t, m.Init(conf.Remote{}))

	return m
}

func newDiskMatrix(t *testing.T, busy map[string]float64) *matrix.Matrix {
	t.Helper()

	data := matrix.New("Test", "disk", "disk")
	metric, err := data.NewMetricFloat64("busy")
	assert.Nil(t, err)

	for name, value := range busy {
		instance, err := data.NewInstance(name)
		assert.Nil(t, err)
		instance.SetLabel("aggr", "aggr1")
		instance.SetLabel("disk", name)
		metric.SetValueFloat64(instance, value)
	}

	return data
}

// winningDisk runs the plugin and returns the disk label of the aggr1 max instance
func winningDisk(t *testing.T, m *Max, data *matrix.Matrix) string {
	t.Helper()

	output, _, err := m.Run(map[string]*matrix.Matrix{"disk": data})
	assert.Nil(t, err)
	assert.Equal(t, len(output), 1)

	instance := output[0].GetInstance("aggr1")
	assert.NotNil(t, instance)

	return instance.GetLabel("disk")
}

// Instances with equal values must resolve to the same winner on every run, otherwise
// the exported labels churn between polls as Go randomizes map iteration order.
func TestMaxTiedValuesPickSameInstance(t *testing.T) {
	busy := map[string]float64{
		"1.1.20": 7,
		"1.1.21": 7,
		"1.1.22": 7,
		"1.1.23": 7,
		"1.1.24": 7,
	}

	m := newMax(t, "aggr<>aggr_disk_max ...")
	want := winningDisk(t, m, newDiskMatrix(t, busy))

	for range 20 {
		assert.Equal(t, winningDisk(t, m, newDiskMatrix(t, busy)), want)
	}
}

func TestMaxPicksHighestValue(t *testing.T) {
	busy := map[string]float64{
		"1.1.20": 7,
		"1.1.21": 7,
		"1.1.22": 42,
		"1.1.23": 7,
		"1.1.24": 7,
	}

	m := newMax(t, "aggr<>aggr_disk_max ...")

	for range 20 {
		assert.Equal(t, winningDisk(t, m, newDiskMatrix(t, busy)), "1.1.22")
	}
}
