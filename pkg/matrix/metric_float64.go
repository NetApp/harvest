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

type MetricFloat64 struct {
	*AbstractMetric
	values []float64
}

func (me *MetricFloat64) Clone(deep bool) Metric {
	clone := MetricFloat64{AbstractMetric: me.AbstractMetric.Clone(deep)}
	if deep && len(me.values) != 0 {
		clone.values = make([]float64, len(me.values))
		copy(clone.values, me.values)
	}
	return &clone
}

// Storage resizing methods

func (me *MetricFloat64) Reset(size int) {
	me.record = make([]bool, size)
	me.values = make([]float64, size)
}

func (me *MetricFloat64) Append() {
	me.record = append(me.record, false)
	me.values = append(me.values, 0)
}

// remove element at index, shift everything to left
func (me *MetricFloat64) Remove(index int) {
	for i := index; i < len(me.values)-1; i++ {
		me.record[i] = me.record[i+1]
		me.values[i] = me.values[i+1]
	}
	me.record = me.record[:len(me.record)-1]
	me.values = me.values[:len(me.values)-1]
}

// Write methods

func (me *MetricFloat64) SetValueInt(i *Instance, v int) error {
	me.record[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueInt32(i *Instance, v int32) error {
	me.record[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueInt64(i *Instance, v int64) error {
	me.record[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueUint8(i *Instance, v uint8) error {
	me.record[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueUint32(i *Instance, v uint32) error {
	me.record[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueUint64(i *Instance, v uint64) error {
	me.record[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueFloat32(i *Instance, v float32) error {
	me.record[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueFloat64(i *Instance, v float64) error {
	me.record[i.index] = true
	me.values[i.index] = v
	return nil
}

func (me *MetricFloat64) SetValueString(i *Instance, v string) error {
	var x float64
	var err error
	if x, err = strconv.ParseFloat(v, 64); err == nil {
		me.record[i.index] = true
		me.values[i.index] = x
		return nil
	}
	return err
}

func (me *MetricFloat64) SetValueBytes(i *Instance, v []byte) error {
	return me.SetValueString(i, string(v))
}

func (me *MetricFloat64) AddValueInt(i *Instance, n int) error {
	m, _ := me.GetValueInt(i)
	return me.SetValueInt(i, m+n)
}

func (me *MetricFloat64) AddValueInt32(i *Instance, n int32) error {
	m, _ := me.GetValueInt32(i)
	return me.SetValueInt32(i, m+n)
}

func (me *MetricFloat64) AddValueInt64(i *Instance, n int64) error {
	m, _ := me.GetValueInt64(i)
	return me.SetValueInt64(i, m+n)
}

func (me *MetricFloat64) AddValueUint8(i *Instance, n uint8) error {
	m, _ := me.GetValueUint8(i)
	return me.SetValueUint8(i, m+n)
}

func (me *MetricFloat64) AddValueUint32(i *Instance, n uint32) error {
	m, _ := me.GetValueUint32(i)
	return me.SetValueUint32(i, m+n)
}

func (me *MetricFloat64) AddValueUint64(i *Instance, n uint64) error {
	m, _ := me.GetValueUint64(i)
	return me.SetValueUint64(i, m+n)
}

func (me *MetricFloat64) AddValueFloat32(i *Instance, n float32) error {
	m, _ := me.GetValueFloat32(i)
	return me.SetValueFloat32(i, m+n)
}

func (me *MetricFloat64) AddValueFloat64(i *Instance, n float64) error {
	m, _ := me.GetValueFloat64(i)
	return me.SetValueFloat64(i, m+n)
}

func (me *MetricFloat64) AddValueString(i *Instance, v string) error {
	var (
		x, n float64
		err  error
		has  bool
	)
	if x, err = strconv.ParseFloat(v, 64); err != nil {
		return err
	}
	if n, has = me.GetValueFloat64(i); has {
		return me.SetValueFloat64(i, x+n)
	}
	return me.SetValueFloat64(i, x)
}

// Read methods

func (me *MetricFloat64) GetValueInt(i *Instance) (int, bool) {
	return int(me.values[i.index]), me.record[i.index]
}

func (me *MetricFloat64) GetValueInt32(i *Instance) (int32, bool) {
	return int32(me.values[i.index]), me.record[i.index]
}

func (me *MetricFloat64) GetValueInt64(i *Instance) (int64, bool) {
	return int64(me.values[i.index]), me.record[i.index]
}

func (me *MetricFloat64) GetValueUint8(i *Instance) (uint8, bool) {
	return uint8(me.values[i.index]), me.record[i.index]
}

func (me *MetricFloat64) GetValueUint32(i *Instance) (uint32, bool) {
	return uint32(me.values[i.index]), me.record[i.index]
}

func (me *MetricFloat64) GetValueUint64(i *Instance) (uint64, bool) {
	return uint64(me.values[i.index]), me.record[i.index]
}

func (me *MetricFloat64) GetValueFloat32(i *Instance) (float32, bool) {
	return float32(me.values[i.index]), me.record[i.index]
}

func (me *MetricFloat64) GetValueFloat64(i *Instance) (float64, bool) {
	return me.values[i.index], me.record[i.index]
}

func (me *MetricFloat64) GetValueString(i *Instance) (string, bool) {
	return strconv.FormatFloat(me.values[i.index], 'f', -1, 64), me.record[i.index]
}

func (me *MetricFloat64) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := me.GetValueString(i)
	return []byte(s), ok
}

// vector arithmetics

func (me *MetricFloat64) GetValuesFloat64() []float64 {
	return me.values
}

func (me *MetricFloat64) Delta(s Metric) error {
	sValues := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	if len(me.values) != len(sValues) {
		return errs.New(ErrUnequalVectors, fmt.Sprintf("minuend=%d, subtrahend=%d", len(me.values), len(sValues)))
	}
	for i := range me.values {
		if me.record[i] && sRecord[i] {
			me.values[i] -= sValues[i]
		}
	}
	return nil
}

func (me *MetricFloat64) Divide(s Metric) error {
	sValues := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	if len(me.values) != len(sValues) {
		return errs.New(ErrUnequalVectors, fmt.Sprintf("minuend=%d, subtrahend=%d", len(me.values), len(sValues)))
	}
	for i := 0; i < len(me.values); i++ {
		if me.record[i] && sRecord[i] && sValues[i] != 0 {
			me.values[i] /= sValues[i]
		}
	}
	return nil
}

func (me *MetricFloat64) DivideWithThreshold(s Metric, t int) error {
	x := float64(t)
	sValues := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	if len(me.values) != len(sValues) {
		return errs.New(ErrUnequalVectors, fmt.Sprintf("minuend=%d, subtrahend=%d", len(me.values), len(sValues)))
	}
	for i := 0; i < len(me.values); i++ {
		if me.record[i] && sRecord[i] && sValues[i] >= x {
			me.values[i] /= sValues[i]
		} else {
			me.values[i] = 0
		}
	}
	return nil
}

func (me *MetricFloat64) MultiplyByScalar(s int) error {
	x := float64(s)
	for i := 0; i < len(me.values); i++ {
		if me.record[i] {
			me.values[i] *= x
		}
	}
	return nil
}

func (me *MetricFloat64) Print() {
	for i := range me.values {
		if me.record[i] {
			fmt.Printf("%s%v%s ", color.Green, me.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, me.values[i], color.End)
		}
	}
}
