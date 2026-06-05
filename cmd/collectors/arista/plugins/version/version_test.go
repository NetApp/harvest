package version

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"os"
	"testing"
)

func TestParseVersion(t *testing.T) {
	data, err := os.ReadFile("testdata/version.json")
	assert.Nil(t, err)

	output := gjson.ParseBytes(data).Get("result")
	v := New(&plugin.AbstractPlugin{SLogger: slog.Default()}).(*Version)

	m := matrix.New("version", "version", "version")
	for _, k := range metrics {
		_, _ = m.NewMetricFloat64(k)
	}
	v.parseVersion(output, m)

	assert.Equal(t, len(m.GetInstances()), 1)

	instance := m.GetInstance("JPE17011101")
	assert.NotNil(t, instance)
	assert.Equal(t, instance.GetLabel("model"), "DCS-7010T-48-R")
	assert.Equal(t, instance.GetLabel("osVersion"), "4.18.5M")
	assert.Equal(t, instance.GetLabel("hostname"), "sa-tme-flexpod-kk19-7010T-1g")

	uptimeVal, ok := m.GetMetric(uptime).GetValueFloat64(instance)
	assert.Equal(t, ok, true)
	assert.Equal(t, uptimeVal > 0, true)
}
