/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/errors"
	"strconv"
)

type MetricUint8 struct {
	*AbstractMetric
	values []uint8
}

func (me *MetricUint8) Clone(deep bool) Metric {
	clone := MetricUint8{AbstractMetric: me.AbstractMetric.Clone(deep)}
	if deep && len(me.values) != 0 {
		clone.values = make([]uint8, len(me.values))
		for i, v := range me.values {
			clone.values[i] = v
		}
	}
	return &clone
}

// Storage resizing methods

func (me *MetricUint8) Reset(size int) {
	me.record = make([]bool, size)
	me.values = make([]uint8, size)
}

func (me *MetricUint8) Append() {
	me.record = append(me.record, false)
	me.values = append(me.values, 0)
}

// remove element at index, shift everything to left
func (me *MetricUint8) Remove(index int) {
	for i := index; i < len(me.values)-1; i++ {
		me.record[i] = me.record[i+1]
		me.values[i] = me.values[i+1]
	}
	me.record = me.record[:len(me.record)-1]
	me.values = me.values[:len(me.values)-1]
}

// Write methods

func (me *MetricUint8) SetValueInt(i *Instance, v int) error {
	me.record[i.index] = true
	me.values[i.index] = uint8(v)
	return nil
}

func (me *MetricUint8) SetValueInt32(i *Instance, v int32) error {
	if v >= 0 {
		me.record[i.index] = true
		me.values[i.index] = uint8(v)
		return nil
	}
	return errors.New(OVERFLOW_ERROR, fmt.Sprintf("convert int32 (%d) to uint32", v))
}

func (me *MetricUint8) SetValueInt64(i *Instance, v int64) error {
	if v >= 0 {
		me.record[i.index] = true
		me.values[i.index] = uint8(v)
		return nil
	}
	return errors.New(OVERFLOW_ERROR, fmt.Sprintf("convert int64 (%d) to uint32", v))
}

func (me *MetricUint8) SetValueBool(i *Instance, v bool) error {
	me.record[i.index] = true
	if v {
		me.values[i.index] = 1
	} else {
		me.values[i.index] = 0
	}
	return nil
}

func (me *MetricUint8) SetValueUint8(i *Instance, v uint8) error {
	me.record[i.index] = true
	me.values[i.index] = v
	return nil
}

func (me *MetricUint8) SetValueUint32(i *Instance, v uint32) error {
	me.record[i.index] = true
	me.values[i.index] = uint8(v)
	return nil
}

func (me *MetricUint8) SetValueUint64(i *Instance, v uint64) error {
	me.record[i.index] = true
	me.values[i.index] = uint8(v)
	return nil
}

func (me *MetricUint8) SetValueFloat32(i *Instance, v float32) error {
	if v >= 0 {
		me.record[i.index] = true
		me.values[i.index] = uint8(v)
		return nil
	}
	return errors.New(OVERFLOW_ERROR, fmt.Sprintf("convert float32 (%f) to uint8", v))
}

func (me *MetricUint8) SetValueFloat64(i *Instance, v float64) error {
	if v >= 0 {
		me.record[i.index] = true
		me.values[i.index] = uint8(v)
		return nil
	}
	return errors.New(OVERFLOW_ERROR, fmt.Sprintf("convert float64 (%f) to uint8", v))
}

func (me *MetricUint8) SetValueString(i *Instance, v string) error {
	var x uint64
	var err error
	if x, err = strconv.ParseUint(v, 10, 8); err == nil {
		me.record[i.index] = true
		me.values[i.index] = uint8(x)
		return nil
	}
	return err
}

func (me *MetricUint8) SetValueBytes(i *Instance, v []byte) error {
	return me.SetValueString(i, string(v))
}

func (me *MetricUint8) AddValueInt(i *Instance, n int) error {
	m, _ := me.GetValueInt(i)
	return me.SetValueInt(i, m+n)
}

func (me *MetricUint8) AddValueInt32(i *Instance, n int32) error {
	m, _ := me.GetValueInt32(i)
	return me.SetValueInt32(i, m+n)
}

func (me *MetricUint8) AddValueInt64(i *Instance, n int64) error {
	m, _ := me.GetValueInt64(i)
	return me.SetValueInt64(i, m+n)
}

func (me *MetricUint8) AddValueUint8(i *Instance, n uint8) error {
	m, _ := me.GetValueUint8(i)
	return me.SetValueUint8(i, m+n)
}

func (me *MetricUint8) AddValueUint32(i *Instance, n uint32) error {
	m, _ := me.GetValueUint32(i)
	return me.SetValueUint32(i, m+n)
}

func (me *MetricUint8) AddValueUint64(i *Instance, n uint64) error {
	m, _ := me.GetValueUint64(i)
	return me.SetValueUint64(i, m+n)
}

func (me *MetricUint8) AddValueFloat32(i *Instance, n float32) error {
	m, _ := me.GetValueFloat32(i)
	return me.SetValueFloat32(i, m+n)
}

func (me *MetricUint8) AddValueFloat64(i *Instance, n float64) error {
	m, _ := me.GetValueFloat64(i)
	return me.SetValueFloat64(i, m+n)
}

// Read methods

func (me *MetricUint8) GetValueInt(i *Instance) (int, bool) {
	return int(me.values[i.index]), me.record[i.index]
}

func (me *MetricUint8) GetValueInt32(i *Instance) (int32, bool) {
	return int32(me.values[i.index]), me.record[i.index]
}

func (me *MetricUint8) GetValueInt64(i *Instance) (int64, bool) {
	return int64(me.values[i.index]), me.record[i.index]
}

func (me *MetricUint8) GetValueUint8(i *Instance) (uint8, bool) {
	return me.values[i.index], me.record[i.index]
}

func (me *MetricUint8) GetValueUint32(i *Instance) (uint32, bool) {
	return uint32(me.values[i.index]), me.record[i.index]
}

func (me *MetricUint8) GetValueUint64(i *Instance) (uint64, bool) {
	return uint64(me.values[i.index]), me.record[i.index]
}

func (me *MetricUint8) GetValueFloat32(i *Instance) (float32, bool) {
	return float32(me.values[i.index]), me.record[i.index]
}

func (me *MetricUint8) GetValueFloat64(i *Instance) (float64, bool) {
	return float64(me.values[i.index]), me.record[i.index]
}

func (me *MetricUint8) GetValueString(i *Instance) (string, bool) {
	return strconv.FormatUint(uint64(me.values[i.index]), 10), me.record[i.index]
}

func (me *MetricUint8) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := me.GetValueString(i)
	return []byte(s), ok
}

// vector arithmetics

func (me *MetricUint8) GetValuesFloat64() []float64 {
	f := make([]float64, len(me.values))
	for i, v := range me.values {
		f[i] = float64(v)
	}
	return f
}

// debug

func (me *MetricUint8) Print() {
	for i := range me.values {
		if me.record[i] {
			fmt.Printf("%s%v%s ", color.Green, me.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, me.values[i], color.End)
		}
	}
}
