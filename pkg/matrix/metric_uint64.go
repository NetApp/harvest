/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/errs"
	"strconv"
)

type MetricUint64 struct {
	*AbstractMetric
	values []uint64
}

func (m *MetricUint64) Clone(deep bool) Metric {
	clone := MetricUint64{AbstractMetric: m.AbstractMetric.Clone(deep)}
	if deep && len(m.values) != 0 {
		clone.values = make([]uint64, len(m.values))
		copy(clone.values, m.values)
	}
	return &clone
}

// Storage resizing methods

func (m *MetricUint64) Reset(size int) {
	m.record = make([]bool, size)
	m.pass = make([]bool, size)
	m.values = make([]uint64, size)
}

func (m *MetricUint64) Append() {
	m.record = append(m.record, false)
	m.pass = append(m.pass, false)
	m.values = append(m.values, 0)
}

// Remove element at index, shift everything to the left
func (m *MetricUint64) Remove(index int) {
	for i := index; i < len(m.values)-1; i++ {
		m.record[i] = m.record[i+1]
		m.pass[i] = m.pass[i+1]
		m.values[i] = m.values[i+1]
	}
	m.record = m.record[:len(m.record)-1]
	m.pass = m.pass[:len(m.pass)-1]
	m.values = m.values[:len(m.values)-1]
}

// Write methods

func (m *MetricUint64) SetValueInt64(i *Instance, v int64) error {
	if v >= 0 {
		m.record[i.index] = true
		m.pass[i.index] = true
		m.values[i.index] = uint64(v)
		return nil
	}
	return errs.New(ErrOverflow, fmt.Sprintf("convert int64 (%d) to uint64", v))
}

func (m *MetricUint64) SetValueUint8(i *Instance, v uint8) error {
	m.record[i.index] = true
	m.pass[i.index] = true
	m.values[i.index] = uint64(v)
	return nil
}

func (m *MetricUint64) SetValueUint64(i *Instance, v uint64) error {
	m.record[i.index] = true
	m.pass[i.index] = true
	m.values[i.index] = v
	return nil
}

func (m *MetricUint64) SetValueFloat64(i *Instance, v float64) error {
	if v >= 0 {
		m.record[i.index] = true
		m.pass[i.index] = true
		m.values[i.index] = uint64(v)
		return nil
	}
	return errs.New(ErrOverflow, fmt.Sprintf("convert float64 (%f) to uint64", v))
}

func (m *MetricUint64) SetValueString(i *Instance, v string) error {
	var x uint64
	var err error
	if x, err = strconv.ParseUint(v, 10, 64); err == nil {
		m.record[i.index] = true
		m.pass[i.index] = true
		m.values[i.index] = x
		return nil
	}
	return err
}

func (m *MetricUint64) SetValueBytes(i *Instance, v []byte) error {
	return m.SetValueString(i, string(v))
}

func (m *MetricUint64) AddValueInt64(i *Instance, n int64) error {
	v, _, _ := m.GetValueInt64(i)
	return m.SetValueInt64(i, v+n)
}

func (m *MetricUint64) AddValueUint8(i *Instance, n uint8) error {
	v, _, _ := m.GetValueUint8(i)
	return m.SetValueUint8(i, v+n)
}

func (m *MetricUint64) AddValueUint64(i *Instance, n uint64) error {
	v, _, _ := m.GetValueUint64(i)
	return m.SetValueUint64(i, v+n)
}

func (m *MetricUint64) AddValueFloat64(i *Instance, n float64) error {
	v, _, _ := m.GetValueFloat64(i)
	return m.SetValueFloat64(i, v+n)
}

// Read methods

func (m *MetricUint64) GetValueInt(i *Instance) (int, bool, bool) {
	return int(m.values[i.index]), m.record[i.index], m.pass[i.index]
}

func (m *MetricUint64) GetValueInt64(i *Instance) (int64, bool, bool) {
	return int64(m.values[i.index]), m.record[i.index], m.pass[i.index]
}

func (m *MetricUint64) GetValueUint8(i *Instance) (uint8, bool, bool) {
	return uint8(m.values[i.index]), m.record[i.index], m.pass[i.index]
}

func (m *MetricUint64) GetValueUint64(i *Instance) (uint64, bool, bool) {
	return m.values[i.index], m.record[i.index], m.pass[i.index]
}

func (m *MetricUint64) GetValueFloat64(i *Instance) (float64, bool, bool) {
	return float64(m.values[i.index]), m.record[i.index], m.pass[i.index]
}

func (m *MetricUint64) GetValueString(i *Instance) (string, bool, bool) {
	return strconv.FormatUint(m.values[i.index], 10), m.record[i.index], m.pass[i.index]
}

func (m *MetricUint64) GetValueBytes(i *Instance) ([]byte, bool, bool) {
	s, ok, pass := m.GetValueString(i)
	return []byte(s), ok, pass
}

// vector arithmetics

func (m *MetricUint64) GetValuesFloat64() []float64 {
	f := make([]float64, len(m.values))
	for i, v := range m.values {
		f[i] = float64(v)
	}
	return f
}

// debug

func (m *MetricUint64) Print() {
	for i := range m.values {
		if m.record[i] && m.pass[i] {
			fmt.Printf("%s%v%s ", color.Green, m.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, m.values[i], color.End)
		}
	}
}
