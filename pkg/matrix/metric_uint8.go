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
	values map[string]uint8
}

func (u *MetricUint8) Clone(deep bool) Metric {
	clone := MetricUint8{AbstractMetric: u.AbstractMetric.Clone(deep)}
	if deep && len(u.values) != 0 {
		if len(u.values) != 0 {
			clone.values = make(map[string]uint8, len(u.values))
			for key, element := range u.values {
				clone.values[key] = element
			}
		}
	} else {
		clone.values = make(map[string]uint8)
	}
	return &clone
}

// Storage resizing methods

func (u *MetricUint8) Reset(size int) {
	u.skip = make(map[string]bool)
	u.values = make(map[string]uint8, size)
}

// Remove element at key, shift everything to the left
func (u *MetricUint8) Remove(key string) {
	delete(u.skip, key)
	delete(u.values, key)
}

// Write methods

func (u *MetricUint8) SetValueInt64(i *Instance, v int64) error {
	if v >= 0 {
		u.values[i.key] = uint8(v)
		return nil
	}
	u.skip[i.key] = true
	return errs.New(ErrOverflow, fmt.Sprintf("convert int64 (%d) to uint32", v))
}

func (u *MetricUint8) SetValueBool(i *Instance, v bool) error {
	if v {
		u.values[i.key] = 1
	} else {
		u.values[i.key] = 0
	}
	return nil
}

func (u *MetricUint8) SetValueUint8(i *Instance, v uint8) error {
	u.values[i.key] = v
	return nil
}

func (u *MetricUint8) SetValueUint64(i *Instance, v uint64) error {
	u.values[i.key] = uint8(v)
	return nil
}

func (u *MetricUint8) SetValueFloat64(i *Instance, v float64) error {
	if v >= 0 {
		u.values[i.key] = uint8(v)
		return nil
	}
	u.skip[i.key] = true
	return errs.New(ErrOverflow, fmt.Sprintf("convert float64 (%f) to uint8", v))
}

func (u *MetricUint8) SetValueString(i *Instance, v string) error {
	var x uint64
	var err error
	if x, err = strconv.ParseUint(v, 10, 8); err == nil {
		u.values[i.key] = uint8(x)
		return nil
	}
	u.skip[i.key] = true
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
	// check if key exists in value map
	val, has := u.values[i.key]
	return int(val), has && !u.skip[i.key]
}

func (u *MetricUint8) GetValueInt64(i *Instance) (int64, bool) {
	// check if key exists in value map
	val, has := u.values[i.key]
	return int64(val), has && !u.skip[i.key]
}

func (u *MetricUint8) GetValueUint8(i *Instance) (uint8, bool) {
	// check if key exists in value map
	val, has := u.values[i.key]
	return val, has && !u.skip[i.key]
}

func (u *MetricUint8) GetValueUint64(i *Instance) (uint64, bool) {
	// check if key exists in value map
	val, has := u.values[i.key]
	return uint64(val), has && !u.skip[i.key]
}

func (u *MetricUint8) GetValueFloat64(i *Instance) (float64, bool) {
	// check if key exists in value map
	val, has := u.values[i.key]
	return float64(val), has && !u.skip[i.key]
}

func (u *MetricUint8) GetValueString(i *Instance) (string, bool) {
	// check if key exists in value map
	val, has := u.values[i.key]
	return strconv.FormatUint(uint64(val), 10), has && !u.skip[i.key]
}

func (u *MetricUint8) GetValueBytes(i *Instance) ([]byte, bool) {
	s, ok := u.GetValueString(i)
	return []byte(s), ok
}

// vector arithmetics
func (u *MetricUint8) GetValuesFloat64() map[string]float64 {
	f := make(map[string]float64, len(u.values))
	for i := range u.values {
		f[i] = float64(u.values[i])
	}
	return f
}

// debug

func (u *MetricUint8) Print() {
	for i := range u.values {
		if !u.skip[i] {
			fmt.Printf("%s%v%s ", color.Green, u.values[i], color.End)
		} else {
			fmt.Printf("%s%v%s ", color.Red, u.values[i], color.End)
		}
	}
}
