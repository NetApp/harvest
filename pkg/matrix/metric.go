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
	"github.com/netapp/harvest/v2/pkg/dict"
	"strconv"
)

type Metric struct {
	name        string
	dataType    string
	property    string
	comment     string
	description string
	unit        string
	array       bool
	histogram   bool
	exportable  bool
	labels      *dict.Dict
	buckets     *[]string
	record      []bool
	values      []float64
}

func (m *Metric) Clone(deep bool) *Metric {
	clone := Metric{
		name:        m.name,
		dataType:    m.dataType,
		property:    m.property,
		comment:     m.comment,
		description: m.description,
		unit:        m.unit,
		exportable:  m.exportable,
		array:       m.array,
		histogram:   m.histogram,
		buckets:     m.buckets,
	}
	if m.labels != nil {
		clone.labels = m.labels.Copy()
	}
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

func (m *Metric) GetDescription() string {
	return m.description
}

func (m *Metric) SetDescription(c string) {
	m.description = c
}

func (m *Metric) GetUnit() string {
	return m.unit
}

func (m *Metric) SetUnit(c string) {
	m.unit = c
}

func (m *Metric) IsArray() bool {
	return m.array
}

func (m *Metric) SetArray(c bool) {
	m.array = c
}

func (m *Metric) SetLabel(key, value string) {
	if m.labels == nil {
		m.labels = dict.New()
	}
	m.labels.Set(key, value)
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

func (m *Metric) SetLabels(labels *dict.Dict) {
	m.labels = labels
}

func (m *Metric) GetLabel(key string) string {
	if m.labels != nil {
		return m.labels.Get(key)
	}
	return ""
}

func (m *Metric) GetLabels() *dict.Dict {
	return m.labels

}
func (m *Metric) HasLabels() bool {
	return m.labels != nil && m.labels.Size() != 0
}

func (m *Metric) GetRecords() []bool {
	return m.record
}

func (m *Metric) SetValueNAN(i *Instance) {
	m.record[i.index] = false
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

// Remove element at index, shift everything to the left
func (m *Metric) Remove(index int) {
	for i := index; i < len(m.values)-1; i++ {
		m.record[i] = m.record[i+1]
		m.values[i] = m.values[i+1]
	}
	m.record = m.record[:len(m.record)-1]
	m.values = m.values[:len(m.values)-1]
}

// Write methods

func (m *Metric) SetValueInt64(i *Instance, v int64) error {
	m.record[i.index] = true
	m.values[i.index] = float64(v)
	return nil
}

func (m *Metric) SetValueUint8(i *Instance, v uint8) error {
	m.record[i.index] = true
	m.values[i.index] = float64(v)
	return nil
}

func (m *Metric) SetValueUint64(i *Instance, v uint64) error {
	m.record[i.index] = true
	m.values[i.index] = float64(v)
	return nil
}

func (m *Metric) SetValueFloat64(i *Instance, v float64) error {
	m.record[i.index] = true
	m.values[i.index] = v
	return nil
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

func (m *Metric) AddValueInt64(i *Instance, n int64) error {
	v, _ := m.GetValueInt64(i)
	return m.SetValueInt64(i, v+n)
}

func (m *Metric) AddValueUint8(i *Instance, n uint8) error {
	v, _ := m.GetValueUint8(i)
	return m.SetValueUint8(i, v+n)
}

func (m *Metric) AddValueUint64(i *Instance, n uint64) error {
	v, _ := m.GetValueUint64(i)
	return m.SetValueUint64(i, v+n)
}

func (m *Metric) AddValueFloat64(i *Instance, n float64) error {
	v, _ := m.GetValueFloat64(i)
	return m.SetValueFloat64(i, v+n)
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
		return m.SetValueFloat64(i, x+n)
	}
	return m.SetValueFloat64(i, x)
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
	return strconv.FormatFloat(v, 'f', -1, 64), m.record[i.index]
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
