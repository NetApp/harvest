/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package dict

import (
	"github.com/netapp/harvest/v2/pkg/util"
	"reflect"
	"strings"
)

type Dict struct {
	dict map[string]string
}

func New() *Dict {
	d := Dict{}
	d.dict = make(map[string]string)
	return &d
}

func (d *Dict) Copy(labels ...string) *Dict {
	c := &Dict{}
	if len(labels) == 0 {
		c.dict = make(map[string]string, len(d.dict))
		for k, v := range d.dict {
			c.dict[k] = v
		}
	} else {
		c.dict = make(map[string]string, len(labels))
		for _, k := range labels {
			if v, ok := d.dict[k]; ok {
				c.dict[k] = v
			}
		}
	}
	return c
}

func (d *Dict) Set(key, val string) {
	d.dict[key] = val
}

// SetAll sets all global labels that do not already exist
func (d *Dict) SetAll(allKeyVals *Dict) {
	if allKeyVals != nil {
		for key, val := range allKeyVals.dict {
			if _, has := d.dict[key]; !has {
				d.dict[key] = val
			}
		}
	}
}

func (d *Dict) Delete(key string) {
	delete(d.dict, key)
}

func (d *Dict) Get(key string) string {
	if value, has := d.dict[key]; has {
		return value
	}
	return ""
}

func (d *Dict) Pop(key string) string {
	if value, has := d.GetHas(key); has {
		d.Delete(key)
		return value
	}
	return ""
}

func (d *Dict) GetHas(key string) (string, bool) {
	value, has := d.dict[key]
	return value, has
}

func (d *Dict) Has(key string) bool {
	_, has := d.dict[key]
	return has
}

func (d *Dict) Iter() map[string]string {
	return d.dict
}

func (d *Dict) Map() map[string]string {
	return d.dict
}

func (d *Dict) Keys() []string {
	keys := make([]string, 0, len(d.dict))
	for k := range d.dict {
		keys = append(keys, k)
	}
	return keys
}

func (d *Dict) String() string {
	s := make([]string, 0, len(d.dict))
	for k, v := range d.dict {
		s = append(s, k+"="+v)
	}
	return strings.Join(s, ", ")
}

func (d *Dict) Values() []string {
	values := make([]string, 0, len(d.dict))
	for _, v := range d.dict {
		values = append(values, v)
	}
	return values
}

func (d *Dict) IsEmpty() bool {
	return len(d.dict) == 0
}

func (d *Dict) Size() int {
	return len(d.dict)
}

// CompareLabels The function compares the labels in the current Dict with the previous Dict.
// returns current and previous value of a label if values are different. label should exist in []labels
func (d *Dict) CompareLabels(prev *Dict, labels []string) (*Dict, *Dict) {
	cur := New()
	old := New()
	for key, val1 := range d.dict {
		val2, ok := prev.dict[key]
		if util.Contains(labels, key) && (!ok || !reflect.DeepEqual(val1, val2)) {
			cur.dict[key] = val1
			old.dict[key] = val2
		}
	}
	return cur, old
}
