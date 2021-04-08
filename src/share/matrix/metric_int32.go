package matrix

import (
	"strconv"
	"fmt"
	"goharvest2/share/util"
)

type MetricInt32 struct {
	*AbstractMetric
	values []int32
}

func (me *MetricInt32) Clone(deep bool) Metric {
	clone := MetricInt32{AbstractMetric: me.AbstractMetric.Clone(deep)}
	if deep && len(me.values) != 0 {
		clone.values = make([]int32, len(me.values))
		for i,v := range me.values {
			clone.values[i] = v
		}
	}
	return &clone
}

// Storage resizing methods

func (me *MetricInt32) Reset(size int) {
	me.record = make([]bool, size)
	me.values = make([]int32, size)
}

func (me *MetricInt32) Append() {
	me.record = append(me.record, false)
	me.values = append(me.values, 0)
}

// remove element at index, shift everything to left
func (me *MetricInt32) Remove(index int) {
	for i := index; i < len(me.values)-1; i++ {
		me.record[i] = me.record[i+1]
		me.values[i] = me.values[i+1]
	}
	me.record = me.record[:len(me.record)-1]
	me.values = me.values[:len(me.values)-1]
}

// Write methods 

func (me *MetricInt32) SetValueInt(i *Instance, v int) error {
	me.record[i.index] = true
	me.values[i.index] = int32(v)
	return nil
}

func (me *MetricInt32) SetValueInt32(i *Instance, v int32) error {
	me.record[i.index] = true
	me.values[i.index] = v
	return nil
}

func (me *MetricInt32) SetValueInt64(i *Instance, v int64) error {
	me.record[i.index] = true
	me.values[i.index] = int32(v)
	return nil
}

func (me *MetricInt32) SetValueUint8(i *Instance, v uint8) error {
	me.record[i.index] = true
	me.values[i.index] = int32(v)
	return nil
}

func (me *MetricInt32) SetValueUint32(i *Instance, v uint32) error {
	me.record[i.index] = true
	me.values[i.index] = int32(v)
	return nil
}

func (me *MetricInt32) SetValueUint64(i *Instance, v uint64) error {
	me.record[i.index] = true
	me.values[i.index] = int32(v)
	return nil
}

func (me *MetricInt32) SetValueFloat32(i *Instance, v float32) error{
	me.record[i.index] = true
	me.values[i.index] = int32(v)
	return nil
}

func (me *MetricInt32) SetValueFloat64(i *Instance, v float64) error {
	me.record[i.index] = true
	me.values[i.index] = int32(v)
	return nil
}

func (me *MetricInt32) SetValueString(i *Instance, v string) error {
	var x int64
	var err error
	if x, err = strconv.ParseInt(v, 10, 32); err == nil {
		me.record[i.index] = true
		me.values[i.index] = int32(x)
		return nil
	}
	return err
}

func (me *MetricInt32) SetValueBytes(i *Instance, v []byte) error {
	return me.SetValueString(i, string(v))
}

// Read methods

func (me *MetricInt32) GetValueInt32(i *Instance) (int32, bool) {
	return me.values[i.index], me.record[i.index]
}

func (me *MetricInt32) GetValueInt64(i *Instance) (int64, bool) {
	return int64(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt32) GetValueUint8(i *Instance) (uint8, bool) {
	return uint8(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt32) GetValueUint32(i *Instance) (uint32, bool) {
	return uint32(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt32) GetValueUint64(i *Instance) (uint64, bool) {
	return uint64(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt32) GetValueFloat32(i *Instance) (float32, bool) {
	return float32(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt32) GetValueFloat64(i *Instance) (float64, bool) {
	return float64(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt32) GetValueString(i *Instance) (string, bool) {
	return strconv.FormatInt(int64(me.values[i.index]), 10), me.record[i.index]
}

func (me *MetricInt32) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := me.GetValueString(i)
	return []byte(s), ok
}

// vector arithmetics

func (me *MetricInt32) GetValuesFloat64() []float64 {
	f := make([]float64, len(me.values))
	for i, v := range me.values {
		f[i] = float64(v)
	}
	return f
}

// debug

func (me *MetricInt32) Print() {
	for i := range me.values {
		if me.record[i] {
			fmt.Printf("%s%v%s ", util.Green, me.values[i], util.End)
		} else {
			fmt.Printf("%s%v%s ", util.Red, me.values[i], util.End)
		}
	}
}