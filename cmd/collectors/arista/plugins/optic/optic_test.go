package optic

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
	m := matrix.New("optic", "optic", "optic")
	for _, k := range metrics {
		_, _ = m.NewMetricFloat64(k)
	}
	return m
}

func TestParseOpticPopulated(t *testing.T) {
	data, err := os.ReadFile("testdata/transceiver.json")
	assert.Nil(t, err)

	output := gjson.ParseBytes(data).Get("result")
	o := New(&plugin.AbstractPlugin{SLogger: slog.Default()}).(*Optic)

	m := newTestMatrix()
	o.parseOptic(output, m)

	// Only interfaces with DOM data are collected (Ethernet49, Ethernet50)
	assert.Equal(t, len(m.GetInstances()), 2)

	instance := m.GetInstance("Ethernet49")
	assert.NotNil(t, instance)

	rxVal, _ := m.GetMetric(rx).GetValueFloat64(instance)
	assert.Equal(t, rxVal, -3.14)
	txVal, _ := m.GetMetric(tx).GetValueFloat64(instance)
	assert.Equal(t, txVal, -2.5)
	tempVal, _ := m.GetMetric(temperature).GetValueFloat64(instance)
	assert.Equal(t, tempVal, 35.2)
}

func TestParseOpticEmpty(t *testing.T) {
	data, err := os.ReadFile("testdata/transceiver_empty.json")
	assert.Nil(t, err)

	output := gjson.ParseBytes(data).Get("result")
	o := New(&plugin.AbstractPlugin{SLogger: slog.Default()}).(*Optic)

	m := newTestMatrix()
	o.parseOptic(output, m)

	// No DOM data on copper/empty ports
	assert.Equal(t, len(m.GetInstances()), 0)
}
