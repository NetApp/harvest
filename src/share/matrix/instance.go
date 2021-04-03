package matrix

import (
	"goharvest2/share/dict"
)

// Instance struct and related methods

type Instance struct {
	index  int
	labels *dict.Dict
    exportable bool
}

func NewInstance(index int) *Instance {
	i := &Instance{index: index}
	i.labels = dict.New()
	i.exportable = true

	return i
}

func (i *Instance) GetLabels() map[string]string {
	return i.labels.Iter()
}

func (i *Instance) GetLabel(key string) string {
	return i.labels.Get(key)
}

func (i *Instance) SetLabel(key, value string) {
	i.labels.Set(key, value)
}

func (i *Instance) IsExportable() bool {
	return i.exportable
}

func (i *Instance) SetExportable(b bool) {
	i.exportable = b
}