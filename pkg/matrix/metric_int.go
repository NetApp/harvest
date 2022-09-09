/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"strconv"
)

type MetricInt struct {
	*AbstractMetric
	values []int
}

func (me *MetricInt) Clone(deep bool) Metric {
	clone := MetricInt{AbstractMetric: me.AbstractMetric.Clone(deep)}
	if deep && len(me.values) != 0 {
		clone.values = make([]int, len(me.values))
		copy(clone.values, me.values)
	}
	return &clone
}

// Storage resizing methods

func (me *MetricInt) Reset(size int) {
	me.record = make([]bool, size)
	me.pass = make([]bool, size)
	me.values = make([]int, size)
}

func (me *MetricInt) Append() {
	me.record = append(me.record, false)
	me.pass = append(me.pass, false)
	me.values = append(me.values, 0)
}

// remove element at index, shift everything to left
func (me *MetricInt) Remove(index int) {
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

func (me *MetricInt) SetValueInt(i *Instance, v int) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = v
	return nil
}

func (me *MetricInt) SetValueInt32(i *Instance, v int32) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueInt64(i *Instance, v int64) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueUint8(i *Instance, v uint8) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueUint32(i *Instance, v uint32) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueUint64(i *Instance, v uint64) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueFloat32(i *Instance, v float32) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueFloat64(i *Instance, v float64) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueString(i *Instance, v string) error {
	var x int
	var err error
	if x, err = strconv.Atoi(v); err == nil {
		me.record[i.index] = true
		me.pass[i.index] = true
		me.values[i.index] = x
		return nil
	}
	return err
}

func (me *MetricInt) SetValueBytes(i *Instance, v []byte) error {
	return me.SetValueString(i, string(v))
}

func (me *MetricInt) AddValueInt(i *Instance, n int) error {
	m, _, _ := me.GetValueInt(i)
	return me.SetValueInt(i, m+n)
}

func (me *MetricInt) AddValueInt32(i *Instance, n int32) error {
	m, _, _ := me.GetValueInt32(i)
	return me.SetValueInt32(i, m+n)
}

func (me *MetricInt) AddValueInt64(i *Instance, n int64) error {
	m, _, _ := me.GetValueInt64(i)
	return me.SetValueInt64(i, m+n)
}

func (me *MetricInt) AddValueUint8(i *Instance, n uint8) error {
	m, _, _ := me.GetValueUint8(i)
	return me.SetValueUint8(i, m+n)
}

func (me *MetricInt) AddValueUint32(i *Instance, n uint32) error {
	m, _, _ := me.GetValueUint32(i)
	return me.SetValueUint32(i, m+n)
}

func (me *MetricInt) AddValueUint64(i *Instance, n uint64) error {
	m, _, _ := me.GetValueUint64(i)
	return me.SetValueUint64(i, m+n)
}

func (me *MetricInt) AddValueFloat32(i *Instance, n float32) error {
	m, _, _ := me.GetValueFloat32(i)
	return me.SetValueFloat32(i, m+n)
}

func (me *MetricInt) AddValueFloat64(i *Instance, n float64) error {
	m, _, _ := me.GetValueFloat64(i)
	return me.SetValueFloat64(i, m+n)
}

// Read methods

func (me *MetricInt) GetValueInt(i *Instance) (int, bool, bool) {
	return me.values[i.index], me.record[i.index], me.pass[i.index]
}

func (me *MetricInt) GetValueInt32(i *Instance) (int32, bool, bool) {
	return int32(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricInt) GetValueInt64(i *Instance) (int64, bool, bool) {
	return int64(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricInt) GetValueUint8(i *Instance) (uint8, bool, bool) {
	return uint8(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricInt) GetValueUint32(i *Instance) (uint32, bool, bool) {
	return uint32(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricInt) GetValueUint64(i *Instance) (uint64, bool, bool) {
	return uint64(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricInt) GetValueFloat32(i *Instance) (float32, bool, bool) {
	return float32(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricInt) GetValueFloat64(i *Instance) (float64, bool, bool) {
	return float64(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricInt) GetValueString(i *Instance) (string, bool, bool) {
	return strconv.FormatInt(int64(me.values[i.index]), 10), me.record[i.index], me.pass[i.index]
}

func (me *MetricInt) GetValueBytes(i *Instance) ([]byte, bool, bool) {
	s, ok, pass := me.GetValueString(i)
	return []byte(s), ok, pass
}

// vector arithmetics

func (me *MetricInt) GetValuesFloat64() []float64 {
	f := make([]float64, len(me.values))
	for i, v := range me.values {
		f[i] = float64(v)
	}
	return f
}

// debug
func (me *MetricInt) Print() {
	for i := range me.values {
		if me.record[i] && me.pass[i] {
			fmt.Printf("%s%v%s ", color.Green, me.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, me.values[i], color.End)
		}
	}
}
