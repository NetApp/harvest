package lldp

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"os"
	"testing"
)

func TestParseLLDP(t *testing.T) {
	data, err := os.ReadFile("testdata/lldp.json")
	assert.Nil(t, err)

	output := gjson.ParseBytes(data).Get("result")
	l := New(&plugin.AbstractPlugin{SLogger: slog.Default()}).(*LLDP)

	m := matrix.New("lldp", "lldp", "lldp")
	_, _ = m.NewMetricFloat64(labels)
	l.parseLLDP(output, m)

	assert.Equal(t, len(m.GetInstances()), 2)

	instance := m.GetInstance("8c60.4f32.f7a7-Ethernet49")
	assert.NotNil(t, instance)
	assert.Equal(t, instance.GetLabel("local_port"), "Ethernet49")
	assert.Equal(t, instance.GetLabel("remote_name"), "sa-tme-flexpod-5596-kk20-e04963.sa-tme-flexpod.local")
	assert.Equal(t, instance.GetLabel("capabilities"), "bridge")
}

func TestNewLLDPModelTrimsQuotes(t *testing.T) {
	data, err := os.ReadFile("testdata/lldp.json")
	assert.Nil(t, err)

	neighbors := gjson.ParseBytes(data).Get("result.0.lldpNeighbors")
	mgmt := neighbors.Get("Management1.lldpNeighborInfo.0")
	model := NewLLDPModel("Management1", mgmt)

	assert.Equal(t, model.ChassisID, "8480.2d33.eb8a")
	// interfaceId in the fixture is wrapped in escaped quotes ""Eth103/1/41""
	assert.Equal(t, model.RemotePort, "Eth103/1/41")
	assert.Equal(t, model.RemoteName, "sa-tme-flexpod-5596-kk-27-234168")
}
