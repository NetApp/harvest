/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

import (
	"fmt"
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"strconv"
)

type MetricFloat64 struct {
	*AbstractMetric
	values []float64
}

type VectorSummary struct {
	NegativeCount int
	ZeroCount     int
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
	me.pass = make([]bool, size)
	me.values = make([]float64, size)
}

func (me *MetricFloat64) Append() {
	me.record = append(me.record, false)
	me.pass = append(me.pass, false)
	me.values = append(me.values, 0)
}

// remove element at index, shift everything to left
func (me *MetricFloat64) Remove(index int) {
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

func (me *MetricFloat64) SetValueInt(i *Instance, v int) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueInt32(i *Instance, v int32) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueInt64(i *Instance, v int64) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueUint8(i *Instance, v uint8) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueUint32(i *Instance, v uint32) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueUint64(i *Instance, v uint64) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueFloat32(i *Instance, v float32) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = float64(v)
	return nil
}

func (me *MetricFloat64) SetValueFloat64(i *Instance, v float64) error {
	me.record[i.index] = true
	me.pass[i.index] = true
	me.values[i.index] = v
	return nil
}

func (me *MetricFloat64) SetValueString(i *Instance, v string) error {
	var x float64
	var err error
	if x, err = strconv.ParseFloat(v, 64); err == nil {
		me.record[i.index] = true
		me.pass[i.index] = true
		me.values[i.index] = x
		return nil
	}
	return err
}

func (me *MetricFloat64) SetValueBytes(i *Instance, v []byte) error {
	return me.SetValueString(i, string(v))
}

func (me *MetricFloat64) AddValueInt(i *Instance, n int) error {
	m, _, _ := me.GetValueInt(i)
	return me.SetValueInt(i, m+n)
}

func (me *MetricFloat64) AddValueInt32(i *Instance, n int32) error {
	m, _, _ := me.GetValueInt32(i)
	return me.SetValueInt32(i, m+n)
}

func (me *MetricFloat64) AddValueInt64(i *Instance, n int64) error {
	m, _, _ := me.GetValueInt64(i)
	return me.SetValueInt64(i, m+n)
}

func (me *MetricFloat64) AddValueUint8(i *Instance, n uint8) error {
	m, _, _ := me.GetValueUint8(i)
	return me.SetValueUint8(i, m+n)
}

func (me *MetricFloat64) AddValueUint32(i *Instance, n uint32) error {
	m, _, _ := me.GetValueUint32(i)
	return me.SetValueUint32(i, m+n)
}

func (me *MetricFloat64) AddValueUint64(i *Instance, n uint64) error {
	m, _, _ := me.GetValueUint64(i)
	return me.SetValueUint64(i, m+n)
}

func (me *MetricFloat64) AddValueFloat32(i *Instance, n float32) error {
	m, _, _ := me.GetValueFloat32(i)
	return me.SetValueFloat32(i, m+n)
}

func (me *MetricFloat64) AddValueFloat64(i *Instance, n float64) error {
	m, _, _ := me.GetValueFloat64(i)
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
	if n, has, _ = me.GetValueFloat64(i); has {
		return me.SetValueFloat64(i, x+n)
	}
	return me.SetValueFloat64(i, x)
}

// Read methods

func (me *MetricFloat64) GetValueInt(i *Instance) (int, bool, bool) {
	return int(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricFloat64) GetValueInt32(i *Instance) (int32, bool, bool) {
	return int32(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricFloat64) GetValueInt64(i *Instance) (int64, bool, bool) {
	return int64(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricFloat64) GetValueUint8(i *Instance) (uint8, bool, bool) {
	return uint8(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricFloat64) GetValueUint32(i *Instance) (uint32, bool, bool) {
	return uint32(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricFloat64) GetValueUint64(i *Instance) (uint64, bool, bool) {
	return uint64(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricFloat64) GetValueFloat32(i *Instance) (float32, bool, bool) {
	return float32(me.values[i.index]), me.record[i.index], me.pass[i.index]
}

func (me *MetricFloat64) GetValueFloat64(i *Instance) (float64, bool, bool) {
	return me.values[i.index], me.record[i.index], me.pass[i.index]
}

func (me *MetricFloat64) GetValueString(i *Instance) (string, bool, bool) {
	return strconv.FormatFloat(me.values[i.index], 'f', -1, 64), me.record[i.index], me.pass[i.index]
}

func (me *MetricFloat64) GetValueBytes(i *Instance) ([]byte, bool, bool) {
	s, ok, pass := me.GetValueString(i)
	return []byte(s), ok, pass
}

// vector arithmetics

func (me *MetricFloat64) GetValuesFloat64() []float64 {
	return me.values
}

func (me *MetricFloat64) Delta(s Metric, logger *logging.Logger) (VectorSummary, error) {
	var vs VectorSummary
	prevRaw := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	pass := me.GetPass()
	if len(me.values) != len(prevRaw) || len(pass) != len(prevRaw) {
		return vs, errs.New(ErrUnequalVectors, fmt.Sprintf("minuend=%d, subtrahend=%d", len(me.values), len(prevRaw)))
	}
	for i := range me.values {
		if me.record[i] && sRecord[i] {
			v := me.values[i]
			// reset pass
			pass[i] = true
			//if current and previous raw are <= 0
			if me.values[i] <= 0 || prevRaw[i] <= 0 {
				pass[i] = false
				if me.values[i] < 0 {
					logger.Trace().
						Str("metric", me.GetName()).
						Float64("current", me.values[i]).
						Float64("previous", prevRaw[i]).
						Msg("Negative raw values detected")
				}
			}
			me.values[i] -= prevRaw[i]
			//if cooked value is <= 0 then pass delta
			if me.values[i] <= 0 {
				pass[i] = false
				if me.values[i] < 0 {
					vs.NegativeCount += 1
					logger.Trace().
						Str("metric", me.GetName()).
						Float64("current", v).
						Float64("previous", prevRaw[i]).
						Msg("Negative cooked value detected")
				} else if me.values[i] == 0 {
					vs.ZeroCount += 1
				}
			}
		}
	}
	return vs, nil
}

func (me *MetricFloat64) Divide(s Metric, logger *logging.Logger) (VectorSummary, error) {
	var vs VectorSummary
	sValues := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	pass := me.GetPass()
	if len(me.values) != len(sValues) || len(pass) != len(sValues) {
		return vs, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d", len(me.values), len(sValues)))
	}
	for i := 0; i < len(me.values); i++ {
		if me.record[i] && sRecord[i] && sValues[i] != 0 {
			v := me.values[i]
			// reset pass
			pass[i] = true
			// if numerator/denominator raw is 0 or negative
			if me.values[i] <= 0 || sValues[i] <= 0 {
				pass[i] = false
				if me.values[i] < 0 || sValues[i] < 0 {
					logger.Trace().
						Str("metric", me.GetName()).
						Float64("numerator", me.values[i]).
						Float64("denominator", sValues[i]).
						Msg("Negative raw values detected")
				}
			}
			me.values[i] /= sValues[i]
			// if cooked value is 0 or negative then pass delta
			if me.values[i] <= 0 {
				pass[i] = false
				if me.values[i] < 0 {
					logger.Trace().
						Str("metric", me.GetName()).
						Float64("numerator", v).
						Float64("denominator", sValues[i]).
						Msg("Negative cooked value detected")
					vs.NegativeCount += 1
				} else if me.values[i] == 0 {
					vs.ZeroCount += 1
				}
			}
		}
	}
	return vs, nil
}

func (me *MetricFloat64) DivideWithThreshold(s Metric, t int, logger *logging.Logger) (VectorSummary, error) {
	var vs VectorSummary
	x := float64(t)
	sValues := s.GetValuesFloat64()
	sRecord := s.GetRecords()
	pass := me.GetPass()
	if len(me.values) != len(sValues) || len(pass) != len(sValues) {
		return vs, errs.New(ErrUnequalVectors, fmt.Sprintf("numerator=%d, denominator=%d", len(me.values), len(sValues)))
	}
	for i := 0; i < len(me.values); i++ {
		v := me.values[i]
		// reset pass
		pass[i] = true
		// if numerator/denominator raw is 0 or negative
		if me.values[i] <= 0 || sValues[i] <= 0 {
			pass[i] = false
			if me.values[i] < 0 || sValues[i] < 0 {
				logger.Trace().
					Str("metric", me.GetName()).
					Float64("numerator", v).
					Float64("denominator", sValues[i]).
					Msg("Negative raw values detected")
			}
		}
		if me.record[i] && sRecord[i] && sValues[i] >= x {
			me.values[i] /= sValues[i]
		}
		// if cooked value is 0 or negative then pass delta
		if me.values[i] <= 0 {
			pass[i] = false
			if me.values[i] < 0 {
				logger.Trace().
					Str("metric", me.GetName()).
					Float64("numerator", v).
					Float64("denominator", sValues[i]).
					Msg("Negative cooked value detected")
				vs.NegativeCount += 1
			} else if me.values[i] == 0 {
				vs.ZeroCount += 1
			}
		}
	}
	return vs, nil
}

func (me *MetricFloat64) MultiplyByScalar(s int, logger *logging.Logger) (VectorSummary, error) {
	var vs VectorSummary
	x := float64(s)
	pass := me.GetPass()
	for i := 0; i < len(me.values); i++ {
		// reset pass
		pass[i] = true
		if me.record[i] {
			me.values[i] *= x
		}
		// if cooked value is 0 or negative then pass delta
		if me.values[i] <= 0 {
			pass[i] = false
			if me.values[i] < 0 {
				logger.Trace().
					Str("metric", me.GetName()).
					Float64("current", me.values[i]).
					Msg("Negative cooked value detected")
				vs.NegativeCount += 1
			} else if me.values[i] == 0 {
				vs.ZeroCount += 1
			}
		}
	}
	return vs, nil
}

func (me *MetricFloat64) Print() {
	for i := range me.values {
		if me.record[i] && me.pass[i] {
			fmt.Printf("%s%v%s ", color.Green, me.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, me.values[i], color.End)
		}
	}
}
