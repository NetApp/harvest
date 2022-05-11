/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package matrix

import (
	"github.com/netapp/harvest/v2/pkg/dict"
)

// Instance struct and related methods

type Instance struct {
	index      int
	labels     *dict.Dict
	exportable bool
}

func NewInstance(index int) *Instance {
	me := &Instance{index: index}
	me.labels = dict.New()
	me.exportable = true

	return me
}

func (me *Instance) GetLabel(key string) string {
	return me.labels.Get(key)
}

func (me *Instance) GetLabels() *dict.Dict {
	return me.labels
}

func (me *Instance) SetLabel(key, value string) {
	me.labels.Set(key, value)
}

func (me *Instance) SetLabels(labels *dict.Dict) {
	me.labels = labels
}

func (me *Instance) IsExportable() bool {
	return me.exportable
}

func (me *Instance) SetExportable(b bool) {
	me.exportable = b
}

func (me *Instance) Clone() *Instance {
	clone := NewInstance(me.index)
	clone.labels = me.labels.Copy()
	clone.exportable = me.exportable
	return clone
}
