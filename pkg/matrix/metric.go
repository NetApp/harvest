/*
 * Copyright NetApp Inc, 2021 All rights reserved

Package Description:
   Parse raw metric name from collector template

Examples:
   Simple name (e.g. "metric_name"), means both name and display are the same
   Custom name (e.g. "metric_name => custom_name") is parsed as display name.
*/

package matrix

import (
	"fmt"
	"maps"
	"strconv"
)

type Metric struct {
	name       string
	dataType   string
	property   string
	comment    string
	array      bool
	histogram  bool
	exportable bool
	labels     map[string]string
	buckets    *[]string
	record     []bool
	values     []float64
}

func (m *Metric) Clone(deep bool) *Metric {
	clone := Metric{
		name:       m.name,
		dataType:   m.dataType,
		property:   m.property,
		comment:    m.comment,
		exportable: m.exportable,
		array:      m.array,
		histogram:  m.histogram,
		buckets:    m.buckets,
	}
	clone.labels = maps.Clone(m.labels)
	if deep {
		if len(m.record) != 0 {
			clone.record = make([]bool, len(m.record))
			copy(clone.record, m.record)
		}
		if len(m.values) != 0 {
			clone.values = make([]float64, len(m.values))
			copy(clone.values, m.values)
		}
	}
	return &clone
}

func (m *Metric) GetName() string {
	return m.name
}

func (m *Metric) IsExportable() bool {
	return m.exportable
}

func (m *Metric) SetExportable(b bool) {
	m.exportable = b
}

func (m *Metric) GetType() string {
	return m.dataType
}

func (m *Metric) GetProperty() string {
	return m.property
}

func (m *Metric) SetProperty(p string) {
	m.property = p
}

func (m *Metric) GetComment() string {
	return m.comment
}

func (m *Metric) SetComment(c string) {
	m.comment = c
}

func (m *Metric) IsArray() bool {
	return m.array
}

func (m *Metric) SetArray(c bool) {
	m.array = c
}

func (m *Metric) SetLabel(key, value string) {
	if m.labels == nil {
		m.labels = make(map[string]string)
	}
	m.labels[key] = value
}

func (m *Metric) SetHistogram(b bool) {
	m.histogram = b
}

func (m *Metric) IsHistogram() bool {
	return m.histogram
}

func (m *Metric) Buckets() *[]string {
	return m.buckets
}

func (m *Metric) SetBuckets(buckets *[]string) {
	m.buckets = buckets
}

func (m *Metric) SetLabels(labels map[string]string) {
	m.labels = labels
}

func (m *Metric) GetLabel(key string) string {
	if m.labels != nil {
		return m.labels[key]
	}
	return ""
}

func (m *Metric) GetLabels() map[string]string {
	return m.labels

}
func (m *Metric) HasLabels() bool {
	return len(m.labels) > 0
}

func (m *Metric) GetRecords() []bool {
	return m.record
}

func (m *Metric) SetValueNAN(i *Instance) {
	m.record[i.index] = false
}

func (m *Metric) GetValues() []float64 {
	return m.values
}

// Storage resizing methods

func (m *Metric) Reset(size int) {
	m.record = make([]bool, size)
	m.values = make([]float64, size)
}

func (m *Metric) Append() {
	m.record = append(m.record, false)
	m.values = append(m.values, 0)
}

// Remove an element at index, shift everything to the left
func (m *Metric) Remove(index int) {
	for i := index; i < len(m.values)-1; i++ {
		m.record[i] = m.record[i+1]
		m.values[i] = m.values[i+1]
	}
	m.record = m.record[:len(m.record)-1]
	m.values = m.values[:len(m.values)-1]
}

// Write methods

func (m *Metric) SetValueInt64(i *Instance, v int64) {
	m.record[i.index] = true
	m.values[i.index] = float64(v)
}

func (m *Metric) SetValueUint8(i *Instance, v uint8) {
	m.record[i.index] = true
	m.values[i.index] = float64(v)
}

func (m *Metric) SetValueUint64(i *Instance, v uint64) {
	m.record[i.index] = true
	m.values[i.index] = float64(v)
}

func (m *Metric) SetValueFloat64(i *Instance, v float64) {
	m.record[i.index] = true
	m.values[i.index] = v
}

func (m *Metric) SetValueString(i *Instance, v string) error {
	var x float64
	var err error
	if x, err = strconv.ParseFloat(v, 64); err == nil {
		m.record[i.index] = true
		m.values[i.index] = x
		return nil
	}
	return err
}

func (m *Metric) SetValueBytes(i *Instance, v []byte) error {
	return m.SetValueString(i, string(v))
}

func (m *Metric) AddValueInt64(i *Instance, n int64) {
	v, _ := m.GetValueInt64(i)
	m.SetValueInt64(i, v+n)
}

func (m *Metric) AddValueUint8(i *Instance, n uint8) {
	v, _ := m.GetValueUint8(i)
	m.SetValueUint8(i, v+n)
}

func (m *Metric) AddValueUint64(i *Instance, n uint64) {
	v, _ := m.GetValueUint64(i)
	m.SetValueUint64(i, v+n)
}

func (m *Metric) AddValueFloat64(i *Instance, n float64) {
	v, _ := m.GetValueFloat64(i)
	m.SetValueFloat64(i, v+n)
}

func (m *Metric) AddValueString(i *Instance, v string) error {
	var (
		x, n float64
		err  error
		has  bool
	)
	if x, err = strconv.ParseFloat(v, 64); err != nil {
		return err
	}
	if n, has = m.GetValueFloat64(i); has {
		m.SetValueFloat64(i, x+n)
		return nil
	}
	m.SetValueFloat64(i, x)

	return nil
}

// Read methods

func (m *Metric) GetValueInt(i *Instance) (int, bool) {
	v := m.values[i.index]
	val := int(v)
	return val, m.record[i.index]
}

func (m *Metric) GetValueInt64(i *Instance) (int64, bool) {
	v := m.values[i.index]
	val := int64(v)
	return val, m.record[i.index]
}

func (m *Metric) GetValueUint8(i *Instance) (uint8, bool) {
	v := m.values[i.index]
	return uint8(v), m.record[i.index]
}

func (m *Metric) GetValueUint64(i *Instance) (uint64, bool) {
	v := m.values[i.index]
	val := uint64(v)
	return val, m.record[i.index]
}

func (m *Metric) GetValueFloat64(i *Instance) (float64, bool) {
	v := m.values[i.index]
	return v, m.record[i.index]
}

func (m *Metric) GetValueString(i *Instance) (string, bool) {
	v := m.values[i.index]
	isValid := m.record[i.index]
	switch v {
	case 0:
		return "0", isValid
	case 1:
		return "1", isValid
	}
	return strconv.FormatFloat(v, 'f', -1, 64), isValid
}

func (m *Metric) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := m.GetValueString(i)
	return []byte(s), ok
}

func (m *Metric) Print() {
	for i := range m.values {
		if m.record[i] {
			fmt.Printf("%s%v ", " ", m.values[i])
		} else {
			fmt.Printf("%s%v ", "!", m.values[i])
		}
	}
}
