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

func (i *Instance) GetIndex() int {
	return i.index
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

func (i *Instance) Clone(isExportable bool, labels ...string) *Instance {
	clone := NewInstance(i.index)
	clone.labels = i.Copy(labels...)
	clone.exportable = isExportable
	return clone
}

func (i *Instance) Copy(labels ...string) map[string]string {
	if len(labels) == 0 {
		return maps.Clone(i.labels)
	}
	m := make(map[string]string, len(labels))
	for _, k := range labels {
		m[k] = i.labels[k]
	}
	return m
}

// CompareDiffs iterates through each key in compareKeys, checking if the receiver and prev have the same value for that key.
// When the values are different, return a new Map with the current and previous value
func (i *Instance) CompareDiffs(prev *Instance, compareKeys []string) (map[string]string, map[string]string) {
	cur := make(map[string]string)
	old := make(map[string]string)

	for _, compareKey := range compareKeys {
		val1, ok1 := i.GetLabels()[compareKey]
		if !ok1 {
			continue
		}
		val2, ok2 := prev.GetLabels()[compareKey]
		if !ok2 || val1 != val2 {
			cur[compareKey] = val1
			old[compareKey] = val2
		}
	}
	return cur, old
}
