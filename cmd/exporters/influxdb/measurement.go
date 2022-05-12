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
	tagSet      []string
	fieldSet    []string
	timestamp   string
}

func NewMeasurement(name string, size int) *Measurement {
	m := Measurement{measurement: name}
	m.tagSet = make([]string, size)
	m.fieldSet = make([]string, 0)
	return &m
}

func (m *Measurement) String() string {
	format := "\nmeasurement=%s \ntag_set=%v \nfield_set=%v \ntimestamp=%s"
	return fmt.Sprintf(format, m.measurement, m.tagSet, m.fieldSet, m.timestamp)
}

func (m *Measurement) AddTag(key, value string) {
	m.tagSet = append(m.tagSet, escape(key)+"="+escape(value))
}

func (m *Measurement) AddField(key, value string) {
	m.fieldSet = append(m.fieldSet, escape(key)+"="+value)
}

func (m *Measurement) AddFieldString(key, value string) {
	value = strings.ReplaceAll(value, "\\", `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	m.AddField(key, `"`+value+`"`)
}

func (m *Measurement) SetTimestamp(t string) {
	m.timestamp = t
}

func (m *Measurement) Render() (string, error) {
	var sep1, sep2 string

	if len(m.tagSet) == 0 {
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
		strings.Join(m.tagSet, ","),
		" ",
		strings.Join(m.fieldSet, ","),
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
