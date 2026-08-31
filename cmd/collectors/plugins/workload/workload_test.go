package workload

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"log/slog"
	"testing"
)

func TestRunSetsThroughputMetrics(t *testing.T) {
	data := matrix.New("workload", "workload", "workload")
	instance, err := data.NewInstance("workload-1")
	assert.Nil(t, err)
	instance.SetLabel("max_xput", "100IOPS,1GB/s")
	instance.SetLabel("min_xput", "50IOPS,500MB/s")

	w := &Workload{AbstractPlugin: &plugin.AbstractPlugin{
		Object:  "workload",
		SLogger: slog.Default(),
	}}
	_, _, err = w.Run(map[string]*matrix.Matrix{"workload": data})
	assert.Nil(t, err)

	assert.Equal(t, instance.GetLabel("max_throughput_iops"), "100")
	assert.Equal(t, instance.GetLabel("max_throughput_mbps"), "1000")
	assert.Equal(t, instance.GetLabel("min_throughput_iops"), "50")
	assert.Equal(t, instance.GetLabel("min_throughput_mbps"), "500")
}
