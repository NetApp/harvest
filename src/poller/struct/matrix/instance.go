package matrix

import (
	"goharvest2/poller/errors"
	"goharvest2/poller/struct/dict"

)

// Instance struct and related methods

type Instance struct {
	Name string
	Index int
	Display string
	Labels *dict.Dict
}

func NewInstance (index int) *Instance {
    var I Instance
    I = Instance{Index: index}
    I.Labels = dict.New()
    return &I
}

func (m *Matrix) AddInstance(key string) (*Instance, error) {
	if _, exists := m.Instances[key]; exists {
		return nil, errors.New(errors.MATRIX_HASH, "instance [" + key + "] already in cache")
	}

	i := NewInstance(len(m.Instances))
	m.Instances[key] = i

	return i, nil
}

func (m *Matrix) GetInstance(key string) *Instance {
	if instance, found := m.Instances[key]; found {
		return instance
	}
	return nil
}
