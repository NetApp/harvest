package matrix

import (
	"goharvest2/share/errors"
	"goharvest2/poller/struct/dict"

)

// Instance struct and related methods

type Instance struct {
	Index int
	Labels *dict.Dict
}
          
func (m *Matrix) AddInstance(key string) (*Instance, error) {
	if _, exists := m.Instances[key]; exists {
		return nil, errors.New(errors.MATRIX_HASH, "instance [" + key + "] already in cache")
	}
	i := &Instance{Index : len(m.Instances)}
	i.Labels = dict.New()
	m.Instances[key] = i

	if !m.IsEmpty() {
		for i:=0; i<m.MetricsIndex; i+=1 {
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
