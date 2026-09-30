package fabricpool

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"testing"
)

// TestRunWeightedLatency checks that the flexgroup latency is weighted by the ops counter.
// RestPerf keys the ops counter as cloud_bin_op (only its display is renamed to cloud_bin_operation),
// while CmPerf keys it as cloud_bin_operation, so the plugin has to look up a different key in each.
func TestRunWeightedLatency(t *testing.T) {
	tests := []struct {
		name   string
		parent string
		opsKey string
	}{
		{name: "RestPerf", parent: "RestPerf", opsKey: "cloud_bin_op#GET"},
		{name: "CmPerf", parent: "CmPerf", opsKey: "cloud_bin_operation#GET"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const latencyKey = "cloud_bin_op_latency_average#GET"

			data := matrix.New("fabricpool", "fabricpool", "fabricpool")
			latency, err := data.NewMetricFloat64(latencyKey)
			assert.Nil(t, err)
			ops, err := data.NewMetricFloat64(tt.opsKey)
			assert.Nil(t, err)

			// two constituents of the flexgroup "fg"
			constituents := []struct {
				name    string
				latency float64
				ops     float64
			}{
				{name: "aggr1_fg__0001_bin_0", latency: 100, ops: 10},
				{name: "aggr1_fg__0002_bin_1", latency: 200, ops: 30},
			}
			for _, c := range constituents {
				i, err := data.NewInstance(c.name)
				assert.Nil(t, err)
				i.SetLabel("fabricpool", c.name)
				i.SetLabel("comp_aggr_name", "aggr1")
				i.SetLabel("svm", "svm1")
				i.SetLabel("cloud_target", "target1")
				latency.SetValueFloat64(i, c.latency)
				ops.SetValueFloat64(i, c.ops)
			}

			o := options.Options{IsTest: true}
			f := &FabricPool{AbstractPlugin: plugin.New(tt.parent, &o, node.NewS("FabricPool"), nil, "fabricpool", nil)}
			f.SLogger = slog.Default()

			output, _, err := f.Run(map[string]*matrix.Matrix{"fabricpool": data})
			assert.Nil(t, err)
			assert.Equal(t, len(output), 1)

			fg := output[0].GetInstance("svm1.fg.target1")
			assert.NotNil(t, fg)

			// The weighted average of the two constituents is 175, the ops total is 40
			got, ok := output[0].GetMetric(latencyKey).GetValueFloat64(fg)
			assert.True(t, ok)
			assert.Equal(t, got, 175.0)

			gotOps, ok := output[0].GetMetric(tt.opsKey).GetValueFloat64(fg)
			assert.True(t, ok)
			assert.Equal(t, gotOps, 40.0)
		})
	}
}
