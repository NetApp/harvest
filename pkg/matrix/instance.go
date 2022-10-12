/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

import (
	"github.com/netapp/harvest/v2/pkg/dict"
)

// Instance struct and related methods

type Instance struct {
	key        string
	labels     *dict.Dict
	exportable bool
}

func NewInstance(key string) *Instance {
	me := &Instance{key: key}
	me.labels = dict.New()
	me.exportable = true

	return me
}

func (i *Instance) GetLabel(key string) string {
	return i.labels.Get(key)
}

func (i *Instance) GetLabels() *dict.Dict {
	return i.labels
}

func (i *Instance) SetLabel(key, value string) {
	i.labels.Set(key, value)
}

func (i *Instance) SetLabels(labels *dict.Dict) {
	i.labels = labels
}

func (i *Instance) IsExportable() bool {
	return i.exportable
}

func (i *Instance) SetExportable(b bool) {
	i.exportable = b
}

func (i *Instance) Clone() *Instance {
	clone := NewInstance(i.key)
	clone.labels = i.labels.Copy()
	clone.exportable = i.exportable
	return clone
}
