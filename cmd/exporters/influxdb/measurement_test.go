/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package influxdb

import (
	"github.com/netapp/harvest/v2/assert"
	"testing"
)

/* All examples from :
https://docs.influxdata.com/influxdb/v1.8/write_protocols/line_protocol_tutorial/
*/

func TestMeasurementA(t *testing.T) {

	expecting := `weather,location=us-midwest temperature=82 1465839830100400200`

	m := NewMeasurement("weather", 0)
	m.AddTag("location", "us-midwest")
	m.AddField("temperature", "82")
	m.SetTimestamp("1465839830100400200")

	out, err := m.Render()
	assert.Nil(t, err)
	assert.Equal(t, out, expecting)
}

func TestMeasurementB(t *testing.T) {

	expecting := `weather temperature=82 1465839830100400200`

	m := NewMeasurement("weather", 0)
	m.AddField("temperature", "82")
	m.SetTimestamp("1465839830100400200")

	out, err := m.Render()
	assert.Nil(t, err)
	assert.Equal(t, out, expecting)
}

func TestMeasurementC(t *testing.T) {

	expecting := `weather,location=us-midwest temperature=82`

	m := NewMeasurement("weather", 0)
	m.AddTag("location", "us-midwest")
	m.AddField("temperature", "82")

	out, err := m.Render()
	assert.Nil(t, err)
	assert.Equal(t, out, expecting)
}

func TestMeasurementD(t *testing.T) {

	expecting := `weather,location=us-midwest temperature=82,humidity=71 1465839830100400200`

	m := NewMeasurement("weather", 0)
	m.AddTag("location", "us-midwest")
	m.AddField("temperature", "82")
	m.AddField("humidity", "71")
	m.SetTimestamp("1465839830100400200")

	out, err := m.Render()
	assert.Nil(t, err)
	assert.Equal(t, out, expecting)
}

func TestMeasurementE(t *testing.T) {

	expecting := `weather,location=us\,midwest temperature="too warm" 1465839830100400200`

	m := NewMeasurement("weather", 0)
	m.AddTag("location", "us,midwest")
	m.AddFieldString("temperature", "too warm")
	m.SetTimestamp("1465839830100400200")

	out, err := m.Render()
	assert.Nil(t, err)
	assert.Equal(t, out, expecting)
}
