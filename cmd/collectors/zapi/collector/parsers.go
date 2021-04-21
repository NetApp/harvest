//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package zapi

import (
	util "goharvest2/cmd"
	"goharvest2/pkg/color"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strings"
)

func ParseShortestPath(m *matrix.Matrix, l map[string]string) []string {

	prefix := make([]string, 0)
	keys := make([][]string, 0)

	for key := range m.GetMetrics() {
		keys = append(keys, strings.Split(key, "."))
	}
	for key := range l {
		keys = append(keys, strings.Split(key, "."))
	}

	max := util.MinLen(keys)

	for i := 0; i < max; i += 1 {
		if util.AllSame(keys, i) {
			prefix = append(prefix, keys[0][i])
		} else {
			break
		}
	}
	return prefix
}

func ParseKeyPrefix(keys [][]string) []string {
	var prefix []string
	var i, n int
	n = util.MinLen(keys) - 1
	for i = 0; i < n; i += 1 {
		if util.AllSame(keys, i) {
			prefix = append(prefix, keys[0][i])
		} else {
			break
		}
	}
	return prefix
}

func (me *Zapi) LoadCounters(counters *node.Node) (bool, *node.Node) {
	desired := node.NewXmlS("desired-attributes")

	for _, c := range counters.GetChildren() {
		me.ParseCounters(c, desired, []string{})
	}

	//counters.SetXmlNameS("desired-attributes")
	//counters.SetContentS("")
	return len(me.Matrix.GetMetrics()) > 0, desired
}

func (me *Zapi) ParseCounters(elem, desired *node.Node, path []string) {
	//logger.Debug("", "%v Parsing [%s] [%s] with %d values and %d children", new_path, elem.Name, elem.Value, len(elem.Values), len(elem.Children))

	var d *node.Node

	name := elem.GetNameS()
	new_path := path

	if len(elem.GetNameS()) != 0 {
		new_path = append(new_path, name)
		d = node.NewXmlS(name)
	}

	if len(elem.GetContentS()) != 0 {
		if clean := me.HandleCounter(new_path, elem.GetContentS()); clean != "" {
			d = node.NewXmlS(clean)
		}
	}

	if desired != nil && d != nil {
		desired.AddChild(d)
	}
	for _, child := range elem.GetChildren() {
		me.ParseCounters(child, d, new_path)
	}
}

func (me *Zapi) HandleCounter(path []string, content string) string {
	var name, display, key string
	var split_values, full_path []string

	split_values = strings.Split(content, "=>")
	if len(split_values) == 1 {
		name = content
	} else {
		name = split_values[0]
		display = strings.TrimSpace(split_values[1])
	}

	name = strings.TrimSpace(strings.TrimLeft(name, "^"))

	//full_path = append(path[1:], name)
	full_path = append(path, name)
	key = strings.Join(full_path, ".")

	if display == "" {
		display = ParseDisplay(me.Matrix.Object, full_path)
	}

	if content[0] == '^' {
		me.instance_label_paths[key] = display
		//data.AddLabel(key, display)
		logger.Trace(me.Prefix, "%sadd (%s) as label [%s]%s => %v", color.Yellow, key, display, color.End, full_path)
		if content[1] == '^' {
			//data.AddInstanceKey(full_path[:])
			copied := make([]string, len(full_path))
			copy(copied, full_path)
			me.instance_key_paths = append(me.instance_key_paths, copied)
			logger.Trace(me.Prefix, "%sadd (%s) as instance key [%s]%s => %v", color.Red, key, display, color.End, full_path)
		}
	} else {
		metric, err := me.Matrix.NewMetricUint64(key)
		if err != nil {
			logger.Error(me.Prefix, "add ass metric (%s) [%s]: %v", key, display, err)
		} else {
			metric.SetName(display)
			logger.Trace(me.Prefix, "%sadd as metric (%s) [%s]%s => %v", color.Blue, key, display, color.End, full_path)
		}
	}

	//if display == "" {
	//	return ""
	//}
	return name
}

func ParseDisplay(obj string, path []string) string {
	var ignore = map[string]int{"attributes": 0, "info": 0, "list": 0, "details": 0, "storage": 0}
	var added = map[string]int{}
	var words []string

	obj_words := strings.Split(obj, "_")
	for _, w := range obj_words {
		ignore[w] = 0
	}

	for _, attribute := range path {
		split := strings.Split(attribute, "-")
		for _, word := range split {
			if word == obj {
				continue
			}
			if _, exists := ignore[word]; exists {
				continue
			}
			if _, exists := added[word]; exists {
				continue
			}
			words = append(words, word)
			added[word] = 0
		}
	}
	return strings.Join(words, "_")
}
