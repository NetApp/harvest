package networkinterface

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"os"
	"testing"
)

func newTestMatrix() *matrix.Matrix {
	m := matrix.New("interface", "interface", "interface")
	for _, k := range metrics {
		_, _ = m.NewMetricFloat64(k)
	}
	return m
}

func TestParseInterface(t *testing.T) {
	data, err := os.ReadFile("testdata/interfaces.json")
	assert.Nil(t, err)

	output := gjson.ParseBytes(data).Get("result")
	i := New(&plugin.AbstractPlugin{SLogger: slog.Default()}).(*Interface)

	m := newTestMatrix()
	i.parseInterface(output, m)

	// Only ethernet interfaces are collected
	assert.Equal(t, len(m.GetInstances()), 53)

	instance := m.GetInstance("Ethernet8_28:99:3a:29:5d:fd")
	assert.NotNil(t, instance)
	assert.Equal(t, instance.GetLabel("interface"), "Ethernet8")

	rxBytes, ok := m.GetMetric(receiveBytes).GetValueFloat64(instance)
	assert.Equal(t, ok, true)
	assert.Equal(t, rxBytes, float64(602))

	txDiscards, ok := m.GetMetric(transmitDrops).GetValueFloat64(instance)
	assert.Equal(t, ok, true)
	assert.Equal(t, txDiscards, float64(2617825))

	// Ethernet8 is notconnect (admin enabled, line down) => error_status 1, up 0, admin_up 1
	upVal, _ := m.GetMetric(up).GetValueFloat64(instance)
	assert.Equal(t, upVal, float64(0))
	adminVal, _ := m.GetMetric(adminUp).GetValueFloat64(instance)
	assert.Equal(t, adminVal, float64(1))
	errVal, _ := m.GetMetric(errorStatus).GetValueFloat64(instance)
	assert.Equal(t, errVal, float64(1))

	// Descriptions returned by the device may include literal surrounding quotes.
	// They must be stripped so the exported label is valid OpenMetrics.
	quoted := m.GetInstance("Ethernet22_28:99:3a:29:5e:0b")
	assert.NotNil(t, quoted)
	assert.Equal(t, quoted.GetLabel("description"), "c220-m5- kk20 ru 12")
}
