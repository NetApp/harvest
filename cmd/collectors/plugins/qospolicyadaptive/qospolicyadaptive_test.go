package qospolicyadaptive

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"log/slog"
	"testing"
)

func TestRunConvertsAdaptiveIOPS(t *testing.T) {
	data := matrix.New("qos", "qos_policy_adaptive", "qos")
	exportable, err := data.NewInstance("exportable")
	assert.Nil(t, err)
	exportable.SetLabel("absolute_min_iops", "6144IOPS/TB")
	exportable.SetLabel("expected_iops", "6144IOPS/GB")
	exportable.SetLabel("peak_iops", "100IOPS")

	skipped, err := data.NewInstance("skipped")
	assert.Nil(t, err)
	skipped.SetExportable(false)
	skipped.SetLabel("peak_iops", "200IOPS")

	p := &QosPolicyAdaptive{AbstractPlugin: &plugin.AbstractPlugin{
		Object:  "qos_policy_adaptive",
		SLogger: slog.Default(),
	}}
	_, _, err = p.Run(map[string]*matrix.Matrix{"qos_policy_adaptive": data})
	assert.Nil(t, err)

	assert.Equal(t, exportable.GetLabel("absolute_min_iops"), "6144")
	assert.Equal(t, exportable.GetLabel("expected_iops"), "6144000")
	assert.Equal(t, exportable.GetLabel("peak_iops"), "100")
	assert.Equal(t, skipped.GetLabel("peak_iops"), "200IOPS")
}
