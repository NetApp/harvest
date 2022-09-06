/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

// this file provides methods got quick read/writes to the matrix
// except for using to update metadata, it's unsafe to use these methods,
// and they may be deprecated in the future

package matrix

import "github.com/netapp/harvest/v2/pkg/errs"

func (me *Matrix) LazySetValueInt(mkey, ikey string, v int) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueInt(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (me *Matrix) LazySetValueInt32(mkey, ikey string, v int32) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueInt32(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (me *Matrix) LazySetValueInt64(mkey, ikey string, v int64) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueInt64(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (me *Matrix) LazyGetValueInt64(m, i string) (int64, bool, bool) {
	if metric := me.GetMetric(m); metric != nil {
		if instance := me.GetInstance(i); instance != nil {
			return metric.GetValueInt64(instance)
		}
	}
	return 0, false, false
}

func (me *Matrix) LazyValueInt64(m, i string) int64 {
	valueInt64, _, _ := me.LazyGetValueInt64(m, i)
	return valueInt64
}

func (me *Matrix) LazyAddValueInt64(m, i string, v int64) error {
	if metric := me.GetMetric(m); metric != nil {
		if instance := me.GetInstance(i); instance != nil {
			return metric.AddValueInt64(instance, v)
		}
		return errs.New(ErrInvalidInstanceKey, i)
	}
	return errs.New(ErrInvalidMetricKey, m)
}

func (me *Matrix) LazySetValueUint8(mkey, ikey string, v uint8) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueUint8(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (me *Matrix) LazySetValueUint32(mkey, ikey string, v uint32) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueUint32(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (me *Matrix) LazySetValueUint64(mkey, ikey string, v uint64) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueUint64(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (me *Matrix) LazySetValueFloat32(mkey, ikey string, v float32) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueFloat32(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (me *Matrix) LazySetValueFloat64(mkey, ikey string, v float64) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueFloat64(instance, v)
		}
		return errs.New(ErrInvalidMetricKey, mkey)
	}
	return errs.New(ErrInvalidInstanceKey, ikey)
}

func (me *Matrix) LazyGetValueFloat64(m, i string) (float64, bool, bool) {
	if metric := me.GetMetric(m); metric != nil {
		if instance := me.GetInstance(i); instance != nil {
			return metric.GetValueFloat64(instance)
		}
	}
	return 0.0, false, false
}

func (me *Matrix) LazyValueFloat64(m, i string) float64 {
	valueFloat64, _, _ := me.LazyGetValueFloat64(m, i)
	return valueFloat64
}
