package environment

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
	m := matrix.New("environment", "environment", "environment")
	for _, k := range metrics {
		_, _ = m.NewMetricFloat64(k)
	}
	return m
}

func TestParseEnvironment(t *testing.T) {
	data, err := os.ReadFile("testdata/environment.json")
	assert.Nil(t, err)

	result := gjson.ParseBytes(data).Get("result")
	e := New(&plugin.AbstractPlugin{SLogger: slog.Default()}).(*Environment)

	m := newTestMatrix()
	e.parseTemperature(result.Get("0"), m)
	e.parsePower(result.Get("1"), m)
	e.parseCooling(result.Get("2"), m)

	// 8 temp sensors + 2 power supplies + 2 fans + 1 ambient = 13
	assert.Equal(t, len(m.GetInstances()), 13)

	sensor := m.GetInstance("temp_TempSensor1")
	assert.NotNil(t, sensor)
	tempVal, ok := m.GetMetric(sensorTemp).GetValueFloat64(sensor)
	assert.Equal(t, ok, true)
	assert.Equal(t, tempVal > 0, true)

	// Power supply 2 is in powerLoss state => power_up 0
	ps2 := m.GetInstance("power_2")
	assert.NotNil(t, ps2)
	upVal, _ := m.GetMetric(powerUp).GetValueFloat64(ps2)
	assert.Equal(t, upVal, float64(0))

	ps1 := m.GetInstance("power_1")
	assert.NotNil(t, ps1)
	up1, _ := m.GetMetric(powerUp).GetValueFloat64(ps1)
	assert.Equal(t, up1, float64(1))

	// Ambient temperature instance
	ambient := m.GetInstance("cooling_ambient")
	assert.NotNil(t, ambient)
	ambVal, ok := m.GetMetric(ambientTemp).GetValueFloat64(ambient)
	assert.Equal(t, ok, true)
	assert.Equal(t, ambVal > 0, true)

	// Fan speed
	fan := m.GetInstance("fan_1/1")
	assert.NotNil(t, fan)
	fanUpVal, _ := m.GetMetric(fanUp).GetValueFloat64(fan)
	assert.Equal(t, fanUpVal, float64(1))
}
