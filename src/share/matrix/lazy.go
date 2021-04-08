
package matrix

import "goharvest2/share/errors"

func (me *Matrix) LazySetValueInt(mkey, ikey string, v int) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueInt(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (me *Matrix) LazySetValueInt32(mkey, ikey string, v int32) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueInt32(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (me *Matrix) LazySetValueInt64(mkey, ikey string, v int64) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueInt64(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (me *Matrix) LazySetValueUint8(mkey, ikey string, v uint8) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueUint8(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (me *Matrix) LazySetValueUint32(mkey, ikey string, v uint32) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueUint32(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (me *Matrix) LazySetValueUint64(mkey, ikey string, v uint64) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueUint64(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (me *Matrix) LazySetValueFloat32(mkey, ikey string, v float32) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueFloat32(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (me *Matrix) LazySetValueFloat64(mkey, ikey string, v float64) error {
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			return metric.SetValueFloat64(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (me *Matrix) LazyGetValueInt64(mkey, ikey string) int64 {
	var v int64
	if instance := me.GetInstance(ikey); instance != nil {
		if metric := me.GetMetric(mkey); metric != nil {
			v, _ = metric.GetValueInt64(instance)
		}
	}
	return v
}