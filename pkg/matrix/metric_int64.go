/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"strconv"
)

type MetricInt64 struct {
	*AbstractMetric
	values []int64
}

func (m *MetricInt64) Clone(deep bool) Metric {
	clone := MetricInt64{AbstractMetric: m.AbstractMetric.Clone(deep)}
	if deep && len(m.values) != 0 {
		clone.values = make([]int64, len(m.values))
		copy(clone.values, m.values)
	}
	return &clone
}

// Storage resizing methods

func (m *MetricInt64) Reset(size int) {
	m.record = make([]bool, size)
	m.values = make([]int64, size)
}

func (m *MetricInt64) Append() {
	m.record = append(m.record, false)
	m.values = append(m.values, 0)
}

// Remove element at index, shift everything to the left
func (m *MetricInt64) Remove(index int) {
	for i := index; i < len(m.values)-1; i++ {
		m.record[i] = m.record[i+1]
		m.values[i] = m.values[i+1]
	}
	m.record = m.record[:len(m.record)-1]
	m.values = m.values[:len(m.values)-1]
}

// Write methods

func (m *MetricInt64) SetValueInt64(i *Instance, v int64) error {
	m.record[i.index] = true
	m.values[i.index] = v
	return nil
}

func (m *MetricInt64) SetValueUint8(i *Instance, v uint8) error {
	m.record[i.index] = true
	m.values[i.index] = int64(v)
	return nil
}

func (m *MetricInt64) SetValueUint64(i *Instance, v uint64) error {
	m.record[i.index] = true
	m.values[i.index] = int64(v)
	return nil
}

func (m *MetricInt64) SetValueFloat64(i *Instance, v float64) error {
	m.record[i.index] = true
	m.values[i.index] = int64(v)
	return nil
}

func (m *MetricInt64) SetValueString(i *Instance, v string) error {
	var x int64
	var err error
	if x, err = strconv.ParseInt(v, 10, 64); err == nil {
		m.record[i.index] = true
		m.values[i.index] = x
		return nil
	}
	return err
}

func (m *MetricInt64) SetValueBytes(i *Instance, v []byte) error {
	return m.SetValueString(i, string(v))
}

func (m *MetricInt64) AddValueInt64(i *Instance, n int64) error {
	v, _ := m.GetValueInt64(i)
	return m.SetValueInt64(i, v+n)
}

func (m *MetricInt64) AddValueUint8(i *Instance, n uint8) error {
	v, _ := m.GetValueUint8(i)
	return m.SetValueUint8(i, v+n)
}

func (m *MetricInt64) AddValueUint64(i *Instance, n uint64) error {
	v, _ := m.GetValueUint64(i)
	return m.SetValueUint64(i, v+n)
}

func (m *MetricInt64) AddValueFloat64(i *Instance, n float64) error {
	v, _ := m.GetValueFloat64(i)
	return m.SetValueFloat64(i, v+n)
}

// Read methods

func (m *MetricInt64) GetValueInt(i *Instance) (int, bool) {
	return int(m.values[i.index]), m.record[i.index]
}

func (m *MetricInt64) GetValueInt64(i *Instance) (int64, bool) {
	return m.values[i.index], m.record[i.index]
}

func (m *MetricInt64) GetValueUint8(i *Instance) (uint8, bool) {
	return uint8(m.values[i.index]), m.record[i.index]
}

func (m *MetricInt64) GetValueUint64(i *Instance) (uint64, bool) {
	return uint64(m.values[i.index]), m.record[i.index]
}

func (m *MetricInt64) GetValueFloat64(i *Instance) (float64, bool) {
	return float64(m.values[i.index]), m.record[i.index]
}

func (m *MetricInt64) GetValueString(i *Instance) (string, bool) {
	return strconv.FormatInt(m.values[i.index], 10), m.record[i.index]
}

func (m *MetricInt64) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := m.GetValueString(i)
	return []byte(s), ok
}

// vector arithmetics

func (m *MetricInt64) GetValuesFloat64() []float64 {
	f := make([]float64, len(m.values))
	for i, v := range m.values {
		f[i] = float64(v)
	}
	return f
}

// debug

func (m *MetricInt64) Print() {
	for i := range m.values {
		if m.record[i] {
			fmt.Printf("%s%v%s ", color.Green, m.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, m.values[i], color.End)
		}
	}
}
