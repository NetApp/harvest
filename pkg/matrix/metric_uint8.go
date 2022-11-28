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

type MetricUint8 struct {
	*AbstractMetric
	values []uint8
}

func (u *MetricUint8) Clone(deep bool) Metric {
	clone := MetricUint8{AbstractMetric: u.AbstractMetric.Clone(deep)}
	if deep && len(u.values) != 0 {
		clone.values = make([]uint8, len(u.values))
		copy(clone.values, u.values)
	}
	return &clone
}

// Storage resizing methods

func (u *MetricUint8) Reset(size int) {
	u.record = make([]bool, size)
	u.values = make([]uint8, size)
}

func (u *MetricUint8) Append() {
	u.record = append(u.record, false)
	u.values = append(u.values, 0)
}

// Remove element at index, shift everything to the left
func (u *MetricUint8) Remove(index int) {
	for i := index; i < len(u.values)-1; i++ {
		u.record[i] = u.record[i+1]
		u.values[i] = u.values[i+1]
	}
	u.record = u.record[:len(u.record)-1]
	u.values = u.values[:len(u.values)-1]
}

// Write methods

func (u *MetricUint8) SetValueInt64(i *Instance, v int64) error {
	if v >= 0 {
		u.record[i.index] = true
		u.values[i.index] = uint8(v)
		return nil
	}
	return errs.New(ErrOverflow, fmt.Sprintf("convert int64 (%d) to uint32", v))
}

func (u *MetricUint8) SetValueBool(i *Instance, v bool) error {
	u.record[i.index] = true
	if v {
		u.values[i.index] = 1
	} else {
		u.values[i.index] = 0
	}
	return nil
}

func (u *MetricUint8) SetValueUint8(i *Instance, v uint8) error {
	u.record[i.index] = true
	u.values[i.index] = v
	return nil
}

func (u *MetricUint8) SetValueUint64(i *Instance, v uint64) error {
	u.record[i.index] = true
	u.values[i.index] = uint8(v)
	return nil
}

func (u *MetricUint8) SetValueFloat64(i *Instance, v float64) error {
	if v >= 0 {
		u.record[i.index] = true
		u.values[i.index] = uint8(v)
		return nil
	}
	return errs.New(ErrOverflow, fmt.Sprintf("convert float64 (%f) to uint8", v))
}

func (u *MetricUint8) SetValueString(i *Instance, v string) error {
	var x uint64
	var err error
	if x, err = strconv.ParseUint(v, 10, 8); err == nil {
		u.record[i.index] = true
		u.values[i.index] = uint8(x)
		return nil
	}
	return err
}

func (u *MetricUint8) SetValueBytes(i *Instance, v []byte) error {
	return u.SetValueString(i, string(v))
}

func (u *MetricUint8) AddValueInt64(i *Instance, n int64) error {
	m, _ := u.GetValueInt64(i)
	return u.SetValueInt64(i, m+n)
}

func (u *MetricUint8) AddValueUint8(i *Instance, n uint8) error {
	m, _ := u.GetValueUint8(i)
	return u.SetValueUint8(i, m+n)
}

func (u *MetricUint8) AddValueUint64(i *Instance, n uint64) error {
	m, _ := u.GetValueUint64(i)
	return u.SetValueUint64(i, m+n)
}

func (u *MetricUint8) AddValueFloat64(i *Instance, n float64) error {
	m, _ := u.GetValueFloat64(i)
	return u.SetValueFloat64(i, m+n)
}

// Read methods

func (u *MetricUint8) GetValueInt(i *Instance) (int, bool) {
	return int(u.values[i.index]), u.record[i.index]
}

func (u *MetricUint8) GetValueInt64(i *Instance) (int64, bool) {
	return int64(u.values[i.index]), u.record[i.index]
}

func (u *MetricUint8) GetValueUint8(i *Instance) (uint8, bool) {
	return u.values[i.index], u.record[i.index]
}

func (u *MetricUint8) GetValueUint64(i *Instance) (uint64, bool) {
	return uint64(u.values[i.index]), u.record[i.index]
}

func (u *MetricUint8) GetValueFloat64(i *Instance) (float64, bool) {
	return float64(u.values[i.index]), u.record[i.index]
}

func (u *MetricUint8) GetValueString(i *Instance) (string, bool) {
	return strconv.FormatUint(uint64(u.values[i.index]), 10), u.record[i.index]
}

func (u *MetricUint8) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := u.GetValueString(i)
	return []byte(s), ok
}

// vector arithmetics

func (u *MetricUint8) GetValuesFloat64() []float64 {
	f := make([]float64, len(u.values))
	for i, v := range u.values {
		f[i] = float64(v)
	}
	return f
}

// debug

func (u *MetricUint8) Print() {
	for i := range u.values {
		if u.record[i] {
			fmt.Printf("%s%v%s ", color.Green, u.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, u.values[i], color.End)
		}
	}
}
