
package matrix

import "goharvest2/share/errors"

func (m *Matrix) LazySetValueInt(mkey, ikey string, v int) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueInt(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (m *Matrix) LazySetValueInt32(mkey, ikey string, v int32) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueInt32(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (m *Matrix) LazySetValueInt64(mkey, ikey string, v int64) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueInt64(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (m *Matrix) LazySetValueUint32(mkey, ikey string, v uint32) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueUint32(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (m *Matrix) LazySetValueUint64(mkey, ikey string, v uint64) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueUint64(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (m *Matrix) LazySetValueFloat32(mkey, ikey string, v float32) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueFloat32(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}

func (m *Matrix) LazySetValueFloat64(mkey, ikey string, v float64) error {
	if instance := m.GetInstance(ikey); instance != nil {
		if metric := m.GetMetric(mkey); metric != nil {
			return metric.SetValueFloat64(instance, v)
		}
		return errors.New(INVALID_METRIC_KEY, mkey)
	}
	return errors.New(INVALID_INSTANCE_KEY, ikey)
}