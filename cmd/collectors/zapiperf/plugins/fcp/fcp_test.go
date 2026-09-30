package fcp

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"testing"
)

func TestRunPercents(t *testing.T) {
	data := matrix.New("fcp", "fcp", "fcp")
	read, err := data.NewMetricFloat64("read_data")
	assert.Nil(t, err)
	write, err := data.NewMetricFloat64("write_data")
	assert.Nil(t, err)

	inst, err := data.NewInstance("node1:0a")
	assert.Nil(t, err)
	inst.SetLabel("port", "port.0a")
	inst.SetLabel("speed", "1000")
	read.SetValueFloat64(inst, 100)
	write.SetValueFloat64(inst, 300)

	o := options.Options{IsTest: true}
	f := &Fcp{AbstractPlugin: plugin.New("ZapiPerf", &o, node.NewS("Fcp"), nil, "fcp", nil)}
	f.SLogger = slog.Default()

	_, _, err = f.Run(map[string]*matrix.Matrix{"fcp": data})
	assert.Nil(t, err)

	got := func(name string) float64 {
		v, ok := data.GetMetric(name).GetValueFloat64(inst)
		assert.True(t, ok)
		return v
	}
	assert.Equal(t, got("read_percent"), 0.1)
	assert.Equal(t, got("write_percent"), 0.3)
	assert.Equal(t, got("util_percent"), 0.3)
	assert.Equal(t, inst.GetLabel("port"), "0a")
}
