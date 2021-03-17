package matrix

import (
	"goharvest2/share/dict"
	"goharvest2/share/errors"
)

// Instance struct and related methods

type Instance struct {
	Index  int
	Labels *dict.Dict
}

func (m *Matrix) AddInstance(key string) (*Instance, error) {
	if _, exists := m.Instances[key]; exists {
		return nil, errors.New(errors.MATRIX_HASH, "instance ["+key+"] already in cache")
	}
	i := &Instance{Index: len(m.Instances)}
	i.Labels = dict.New()
	m.Instances[key] = i

	if !m.IsEmpty() {
		for i := 0; i < m.SizeMetrics(); i += 1 {
			m.Data[i] = append(m.Data[i], NAN)
		}
	}

	return i, nil
}

func (m *Matrix) GetInstance(key string) *Instance {
	if instance, found := m.Instances[key]; found {
		return instance
	}
	return nil
}
