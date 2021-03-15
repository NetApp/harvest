package main

import (
	"strings"
)

type Measurement struct {
	measurement string
	tag_set     []string
	field_set   []string
	timestamp   string
}

func NewMeasurement(name string) *Measurement {
	m := Measurement{measurement: name}
	m.tag_set = make([]string, 0)
	m.field_set = make([]string, 0)
	return &m
}

func (m *Measurement) AddTag(key, value string) {
	m.tag_set = append(m.tag_set, escape(key)+"="+escape(value))
}

func (m *Measurement) AddField(key, value string) {
	m.field_set = append(m.field_set, escape(key)+"="+value)
}

func (m *Measurement) AddFieldString(key, value string) {
	value = strings.ReplaceAll(value, `"`, `\"`)
	m.AddField(key, `"`+value+`"`)
}

func (m *Measurement) SetTimestamp(t string) {
	m.timestamp = t
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
		strings.Join(m.field_set, ","),
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
