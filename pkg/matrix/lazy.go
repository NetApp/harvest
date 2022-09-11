/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

// this file provides methods got quick read/writes to the matrix
// except for using to update metadata, it's unsafe to use these methods,
// and they may be deprecated in the future

package matrix

import "github.com/netapp/harvest/v2/pkg/errs"

func (m *Matrix) LazySetValueInt64(mkey, ikey string, v int64) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueInt64(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (m *Matrix) LazyGetValueInt64(key, i string) (int64, bool, bool) {
	if metric := m.GetMetric(key); metric != nil {
		if instance := m.GetInstance(i); instance != nil {
			return metric.GetValueInt64(instance)
		}
	}
	return 0, false, false
}

func (m *Matrix) LazyValueInt64(key, i string) int64 {
	valueInt64, _, _ := m.LazyGetValueInt64(key, i)
	return valueInt64
}

func (m *Matrix) LazyAddValueInt64(key, i string, v int64) error {
	if metric := m.GetMetric(key); metric != nil {
		if instance := m.GetInstance(i); instance != nil {
			return metric.AddValueInt64(instance, v)
		}
		return errs.New(ErrInvalidInstanceKey, i)
	}
	return errs.New(ErrInvalidMetricKey, key)
}

func (m *Matrix) LazySetValueUint8(mkey, ikey string, v uint8) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueUint8(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (m *Matrix) LazySetValueUint64(mkey, ikey string, v uint64) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueUint64(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (m *Matrix) LazySetValueFloat64(mkey, ikey string, v float64) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueFloat64(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (m *Matrix) LazyGetValueFloat64(key, i string) (float64, bool, bool) {
	if metric := m.GetMetric(key); metric != nil {
		if instance := m.GetInstance(i); instance != nil {
			return metric.GetValueFloat64(instance)
		}
	}
	return 0.0, false, false
}

func (m *Matrix) LazyValueFloat64(key, i string) float64 {
	valueFloat64, _, _ := m.LazyGetValueFloat64(key, i)
	return valueFloat64
}
