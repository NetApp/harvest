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
	values map[string]int64
}

func (m *MetricInt64) Clone(deep bool) Metric {
	clone := MetricInt64{AbstractMetric: m.AbstractMetric.Clone(deep)}
	if deep && len(m.values) != 0 {
		clone.values = make(map[string]int64, len(m.values))
		for key, element := range m.values {
			clone.values[key] = element
		}
	} else {
		clone.values = make(map[string]int64)
	}
	return &clone
}

// Storage resizing methods

func (m *MetricInt64) Reset(size int) {
	m.skip = make(map[string]bool)
	m.values = make(map[string]int64, size)
}

// Remove element at key, shift everything to the left
func (m *MetricInt64) Remove(key string) {
	delete(m.skip, key)
	delete(m.values, key)
}

// Write methods

func (m *MetricInt64) SetValueInt64(i *Instance, v int64) error {
	m.values[i.key] = v
	return nil
}

func (m *MetricInt64) SetValueUint8(i *Instance, v uint8) error {
	m.values[i.key] = int64(v)
	return nil
}

func (m *MetricInt64) SetValueUint64(i *Instance, v uint64) error {
	m.values[i.key] = int64(v)
	return nil
}

func (m *MetricInt64) SetValueFloat64(i *Instance, v float64) error {
	m.values[i.key] = int64(v)
	return nil
}

func (m *MetricInt64) SetValueString(i *Instance, v string) error {
	var x int64
	var err error
	if x, err = strconv.ParseInt(v, 10, 64); err == nil {
		m.values[i.key] = x
		return nil
	}
	m.skip[i.key] = true
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
	// check if key exists in value map
	val, has := m.values[i.key]
	return int(val), has && !m.skip[i.key]
}

func (m *MetricInt64) GetValueInt64(i *Instance) (int64, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return val, has && !m.skip[i.key]
}

func (m *MetricInt64) GetValueUint8(i *Instance) (uint8, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return uint8(val), has && !m.skip[i.key]
}

func (m *MetricInt64) GetValueUint64(i *Instance) (uint64, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return uint64(val), has && !m.skip[i.key]
}

func (m *MetricInt64) GetValueFloat64(i *Instance) (float64, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return float64(val), has && !m.skip[i.key]
}

func (m *MetricInt64) GetValueString(i *Instance) (string, bool) {
	// check if key exists in value map
	val, has := m.values[i.key]
	return strconv.FormatInt(val, 10), has && !m.skip[i.key]
}

func (m *MetricInt64) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := m.GetValueString(i)
	return []byte(s), ok
}

// vector arithmetics

func (m *MetricInt64) GetValuesFloat64() map[string]float64 {
	f := make(map[string]float64, len(m.values))
	for i := range m.values {
		f[i] = float64(m.values[i])
	}
	return f
}

// debug

func (m *MetricInt64) Print() {
	for i := range m.values {
		if !m.skip[i] {
			fmt.Printf("%s%v%s ", color.Green, m.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, m.values[i], color.End)
		}
	}
}
