/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package influxdb

import (
	"fmt"
	"strings"
)

type Measurement struct {
	measurement string
	tag_set     []string
	field_set   map[string]string
	timestamp   string
}

func NewMeasurement(name string, size int) *Measurement {
	m := Measurement{measurement: name}
	m.tag_set = make([]string, size)
	m.field_set = make(map[string]string)
	return &m
}

func (m *Measurement) String() string {
	format := "\nmeasurement=%s \ntag_set=%v \nfield_set=%v \ntimestamp=%s"
	return fmt.Sprintf(format, m.measurement, m.tag_set, m.field_set, m.timestamp)
}

func (m *Measurement) AddTag(key, value string) {
	m.tag_set = append(m.tag_set, escape(key)+"="+escape(value))
}

// Returns "true" if field key already exists
func (m *Measurement) HasFieldKey(key string) bool {
	_, has := m.field_set[key]
	return has
}

func (m *Measurement) AddField(key, value string) {
	m.field_set[key] = value
}

func (m *Measurement) AddFieldString(key, value string) {
	value = strings.ReplaceAll(value, "\\", `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	m.AddField(key, `"`+value+`"`)
}

func (m *Measurement) SetTimestamp(t string) {
	m.timestamp = t
}

func (m *Measurement) renderFields() []string {
	r := make([]string, 0)
	for k, v := range m.field_set {
		r = append(r, escape(k)+"="+v)
	}
	return r
}

func (m *Measurement) Render() (string, error) {
	var sep1, sep2 string

	if len(m.tag_set) == 0 {
		sep1 = ""
	} else {
		sep1 = ","
	}

	if len(m.timestamp) == 0 {
		sep2 = ""
	} else {
		sep2 = " "
	}

	return joinAll(
		m.measurement,
		sep1,
		strings.Join(m.tag_set, ","),
		" ",
		strings.Join(m.renderFields(), ","),
		sep2,
		m.timestamp,
	), nil
}

func joinAll(slices ...string) string {
	return strings.Join(slices, "")
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "\\", `\\`)
	s = strings.ReplaceAll(s, " ", `\ `)
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, "=", `\=`)
	return strings.ReplaceAll(s, "\n", `\n`)
}
