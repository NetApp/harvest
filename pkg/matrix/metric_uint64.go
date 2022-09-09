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

func (me *MetricUint64) Clone(deep bool) Metric {
	clone := MetricUint64{AbstractMetric: me.AbstractMetric.Clone(deep)}
	if deep && len(me.values) != 0 {
		clone.values = make([]uint64, len(me.values))
		copy(clone.values, me.values)
	}
	return &clone
}

// Storage resizing methods

func (me *MetricUint64) Reset(size int) {
	me.record = make([]bool, size)
	me.pass = make([]bool, size)
	me.values = make([]uint64, size)
}

func (me *MetricUint64) Append() {
	me.record = append(me.record, false)
	me.pass = append(me.pass, false)
	me.values = append(me.values, 0)
}

// remove element at index, shift everything to left
func (me *MetricUint64) Remove(index int) {
	for i := index; i < len(me.values)-1; i++ {
		me.record[i] = me.record[i+1]
		me.pass[i] = me.pass[i+1]
		me.values[i] = me.values[i+1]
	}
	me.record = me.record[:len(me.record)-1]
	me.pass = me.pass[:len(me.pass)-1]
	me.values = me.values[:len(me.values)-1]
}

// Write methods

func (me *MetricUint64) SetValueInt(i *Instance, v int) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = uint64(v)
	return nil
}

func (me *MetricUint64) SetValueInt32(i *Instance, v int32) error {
	if v >= 0 {
		me.record[i.index] = true
		me.pass[i.index] = true
		me.values[i.index] = uint64(v)
		return nil
	}
	return errs.New(ErrOverflow, fmt.Sprintf("convert int32 (%d) to uint64", v))
}

func (me *MetricUint64) SetValueInt64(i *Instance, v int64) error {
	if v >= 0 {
		me.record[i.index] = true
		me.pass[i.index] = true
		me.values[i.index] = uint64(v)
		return nil
	}
	return errs.New(ErrOverflow, fmt.Sprintf("convert int64 (%d) to uint64", v))
}

func (me *MetricUint64) SetValueUint8(i *Instance, v uint8) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = uint64(v)
	return nil
}

func (me *MetricUint64) SetValueUint32(i *Instance, v uint32) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = uint64(v)
	return nil
}

func (me *MetricUint64) SetValueUint64(i *Instance, v uint64) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = v
	return nil
}

func (me *MetricUint64) SetValueFloat32(i *Instance, v float32) error {
	if v >= 0 {
		me.record[i.index] = true
		me.pass[i.index] = true
		me.values[i.index] = uint64(v)
		return nil
	}
	return errs.New(ErrOverflow, fmt.Sprintf("convert float32 (%f) to uint64", v))
}

func (me *MetricUint64) SetValueFloat64(i *Instance, v float64) error {
	if v >= 0 {
		me.record[i.index] = true
		me.pass[i.index] = true
		me.values[i.index] = uint64(v)
		return nil
	}
	return errs.New(ErrOverflow, fmt.Sprintf("convert float64 (%f) to uint64", v))
}

func (me *MetricUint64) SetValueString(i *Instance, v string) error {
	var x uint64
	var err error
	if x, err = strconv.ParseUint(v, 10, 64); err == nil {
		me.record[i.index] = true
		me.pass[i.index] = true
		me.values[i.index] = x
		return nil
	}
	return err
}

func (me *MetricUint64) SetValueBytes(i *Instance, v []byte) error {
	return me.SetValueString(i, string(v))
}

func (me *MetricUint64) AddValueInt(i *Instance, n int) error {
	m, _, _ := me.GetValueInt(i)
	return me.SetValueInt(i, m+n)
}

func (me *MetricUint64) AddValueInt32(i *Instance, n int32) error {
	m, _, _ := me.GetValueInt32(i)
	return me.SetValueInt32(i, m+n)
}

func (me *MetricUint64) AddValueInt64(i *Instance, n int64) error {
	m, _, _ := me.GetValueInt64(i)
	return me.SetValueInt64(i, m+n)
}

func (me *MetricUint64) AddValueUint8(i *Instance, n uint8) error {
	m, _, _ := me.GetValueUint8(i)
	return me.SetValueUint8(i, m+n)
}

func (me *MetricUint64) AddValueUint32(i *Instance, n uint32) error {
	m, _, _ := me.GetValueUint32(i)
	return me.SetValueUint32(i, m+n)
}

func (me *MetricUint64) AddValueUint64(i *Instance, n uint64) error {
	m, _, _ := me.GetValueUint64(i)
	return me.SetValueUint64(i, m+n)
}

func (me *MetricUint64) AddValueFloat32(i *Instance, n float32) error {
	m, _, _ := me.GetValueFloat32(i)
	return me.SetValueFloat32(i, m+n)
}

func (me *MetricUint64) AddValueFloat64(i *Instance, n float64) error {
	m, _, _ := me.GetValueFloat64(i)
	return me.SetValueFloat64(i, m+n)
}

// Read methods

func (me *MetricUint64) GetValueInt(i *Instance) (int, bool, bool) {
	return int(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricUint64) GetValueInt32(i *Instance) (int32, bool, bool) {
	return int32(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricUint64) GetValueInt64(i *Instance) (int64, bool, bool) {
	return int64(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricUint64) GetValueUint8(i *Instance) (uint8, bool, bool) {
	return uint8(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricUint64) GetValueUint32(i *Instance) (uint32, bool, bool) {
	return uint32(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricUint64) GetValueUint64(i *Instance) (uint64, bool, bool) {
	return me.values[i.index], me.record[i.index], me.pass[i.index]
}

func (me *MetricUint64) GetValueFloat32(i *Instance) (float32, bool, bool) {
	return float32(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricUint64) GetValueFloat64(i *Instance) (float64, bool, bool) {
	return float64(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricUint64) GetValueString(i *Instance) (string, bool, bool) {
	return strconv.FormatUint(me.values[i.index], 10), me.record[i.index], me.pass[i.index]
}

func (me *MetricUint64) GetValueBytes(i *Instance) ([]byte, bool, bool) {
	s, ok, pass := me.GetValueString(i)
	return []byte(s), ok, pass
}

// vector arithmetics

func (me *MetricUint64) GetValuesFloat64() []float64 {
	f := make([]float64, len(me.values))
	for i, v := range me.values {
		f[i] = float64(v)
	}
	return f
}

// debug

func (me *MetricUint64) Print() {
	for i := range me.values {
		if me.record[i] && me.pass[i] {
			fmt.Printf("%s%v%s ", color.Green, me.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, me.values[i], color.End)
		}
	}
}
