/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package matrix

import (
	"maps"
)

// Instance struct and related methods

type Instance struct {
	index      int
	labels     map[string]string
	exportable bool
}

func NewInstance(index int) *Instance {
	me := &Instance{index: index}
	me.labels = make(map[string]string)
	me.exportable = true
	return me
}

func (i *Instance) GetLabel(key string) string {
	return i.labels[key]
}

func (i *Instance) GetLabels() map[string]string {
	return i.labels
}

func (i *Instance) ClearLabels() {
	clear(i.labels)
}

func (i *Instance) SetLabel(key, value string) {
	i.labels[key] = value
}

func (i *Instance) SetLabels(labels map[string]string) {
	i.labels = labels
}

func (i *Instance) IsExportable() bool {
	return i.exportable
}

func (i *Instance) SetExportable(b bool) {
	i.exportable = b
}

func (i *Instance) Clone(isExportable bool) *Instance {
	clone := NewInstance(i.index)
	clone.labels = maps.Clone(i.labels)
	clone.exportable = isExportable
	return clone
}
