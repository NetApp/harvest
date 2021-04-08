package matrix

import (
	"fmt"
	"strconv"
	"goharvest2/share/util"
)

type MetricInt struct {
	*AbstractMetric
	values []int
}

func (me *MetricInt) Clone(deep bool) Metric {
	clone := MetricInt{AbstractMetric: me.AbstractMetric.Clone(deep)}
	if deep && len(me.values) != 0 {
		clone.values = make([]int, len(me.values))
		for i,v := range me.values {
			clone.values[i] = v
		}
	}
	return &clone
}

// Storage resizing methods

func (me *MetricInt) Reset(size int) {
	me.record = make([]bool, size)
	me.values = make([]int, size)
}

func (me *MetricInt) Append() {
	me.record = append(me.record, false)
	me.values = append(me.values, 0)
}

// remove element at index, shift everything to left
func (me *MetricInt) Remove(index int) {
	for i := index; i < len(me.values)-1; i++ {
		me.record[i] = me.record[i+1]
		me.values[i] = me.values[i+1]
	}
	me.record = me.record[:len(me.record)-1]
	me.values = me.values[:len(me.values)-1]
}

// Write methods 

func (me *MetricInt) SetValueInt(i *Instance, v int) error {
	me.record[i.index] = true
	me.values[i.index] = v
	return nil
}

func (me *MetricInt) SetValueInt32(i *Instance, v int32) error {
	me.record[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueInt64(i *Instance, v int64) error {
	me.record[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueUint8(i *Instance, v uint8) error {
	me.record[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueUint32(i *Instance, v uint32) error {
	me.record[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueUint64(i *Instance, v uint64) error {
	me.record[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueFloat32(i *Instance, v float32) error{
	me.record[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueFloat64(i *Instance, v float64) error {
	me.record[i.index] = true
	me.values[i.index] = int(v)
	return nil
}

func (me *MetricInt) SetValueString(i *Instance, v string) error {
	var x int
	var err error
	if x, err = strconv.Atoi(v); err == nil {
		me.record[i.index] = true
		me.values[i.index] = x
		return nil
	}
	return err
}

func (me *MetricInt) SetValueBytes(i *Instance, v []byte) error {
	return me.SetValueString(i, string(v))
}

// Read methods

func (me *MetricInt) GetValueInt32(i *Instance) (int32, bool) {
	return int32(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt) GetValueInt64(i *Instance) (int64, bool) {
	return int64(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt) GetValueUint8(i *Instance) (uint8, bool) {
	return uint8(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt) GetValueUint32(i *Instance) (uint32, bool) {
	return uint32(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt) GetValueUint64(i *Instance) (uint64, bool) {
	return uint64(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt) GetValueFloat32(i *Instance) (float32, bool) {
	return float32(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt) GetValueFloat64(i *Instance) (float64, bool) {
	return float64(me.values[i.index]), me.record[i.index]
}

func (me *MetricInt) GetValueString(i *Instance) (string, bool) {
	return strconv.FormatInt(int64(me.values[i.index]), 10), me.record[i.index]
}

func (me *MetricInt) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := me.GetValueString(i)
	return []byte(s), ok
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
		if me.record[i] {
			fmt.Printf("%s%v%s ", util.Green, me.values[i], util.End)
		} else {
			fmt.Printf("%s%v%s ", util.Red, me.values[i], util.End)
		}
	}
}