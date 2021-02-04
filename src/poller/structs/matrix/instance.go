package matrix

// Instance struct and related methods

type Instance struct {
	Name string
	Index int
	Display string
	Labels map[string]string
}

func NewInstance (index int) *Instance {
    var I Instance
    I = Instance{Index: index}
    I.Labels = map[string]string{}
    return &I
}

func (m *Matrix) AddInstance(key string) (*Instance, error) {
	if _, exists := m.Instances[key]; exists {
		err = errors.New(fmt.Sprintf("Instance [%s] already in cache", key))
		return nil, err
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
