package networkinterface

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"os"
	"testing"
)

func TestInterface_parseInterface(t *testing.T) {
	// Read the file from the testdata directory
	filename := "testdata/N9K-C9336C-FX2_10.2.4-show_interface.json"
	data, err := os.ReadFile(filename)
	assert.Nil(t, err)

	result := gjson.ParseBytes(data)

	i := &Interface{AbstractPlugin: plugin.New("cisco_interface", options.New(), nil, nil, "cisco_interface", nil)}
	i.SLogger = slog.Default()
	m, err := i.initMatrix("cisco_interface")
	assert.Nil(t, err)
	i.matrix = m

	i.parseInterface(result, m)

	assert.Equal(t, len(m.GetInstances()), 3)

	type metricValue struct {
		metric string
		value  string
	}

	type wants struct {
		instance string
		metrics  []metricValue
	}

	// Define the expected instances and their metrics
	tests := []wants{
		{instance: "mgmt0_dead.beef.c1d0", metrics: []metricValue{
			{metric: adminUp, value: "1"},
			{metric: up, value: "0"},
		}},
		{instance: "Ethernet1/1_dead.beef.c1d8", metrics: []metricValue{
			{metric: adminUp, value: "1"},
			{metric: crcErrors, value: "0"},
			{metric: errorStatus, value: "0"},
			{metric: ethOutDiscard, value: "8"},
			{metric: receiveBroadcast, value: "705871"},
			{metric: receiveBytes, value: "65962757720684"},
			{metric: receiveErrors, value: "0"},
			{metric: receiveMulticast, value: "1027477"},
			{metric: receiveDrops, value: "0"},
			{metric: transmitBytes, value: "68349245749232"},
			{metric: transmitErrors, value: "0"},
			{metric: transmitDrops, value: "0"},
		}},
		{instance: "Ethernet1/2_dead.beef.c1dc", metrics: []metricValue{
			{metric: adminUp, value: "1"},
			{metric: receiveBytes, value: "68347824469373"},
			{metric: errorStatus, value: "0"},
		}},
	}

	for _, tt := range tests {
		assert.NotNil(t, m.GetInstance(tt.instance))

		for _, metric := range tt.metrics {
			name := tt.instance + "/" + metric.metric
			t.Run(name, func(t *testing.T) {
				valueString, b := m.GetMetric(metric.metric).GetValueString(m.GetInstance(tt.instance))

				assert.True(t, b)
				assert.Equal(t, valueString, metric.value)
			})
		}
	}
}
