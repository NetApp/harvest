/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package zapi

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"strings"
)

func ParseShortestPath(m *matrix.Matrix, l map[string]string) []string {

	var prefix []string
	keys := make([][]string, 0, len(m.GetMetrics())+len(l))

	for key := range m.GetMetrics() {
		keys = append(keys, strings.Split(key, "."))
	}
	for key := range l {
		keys = append(keys, strings.Split(key, "."))
	}

	minLen := node.MinLen(keys)

	for i := range minLen {
		if node.AllSame(keys, i) {
			prefix = append(prefix, keys[0][i])
		} else {
			break
		}
	}
	return prefix
}

func (z *Zapi) LoadCounters(counters *node.Node) (bool, *node.Node) {
	desired := node.NewXMLS("desired-attributes")

	for _, c := range counters.GetChildren() {
		z.ParseCounters(c, desired, []string{})
	}

	return len(z.Matrix[z.Object].GetMetrics()) > 0, desired
}

func (z *Zapi) ParseCounters(elem, desired *node.Node, path []string) {

	var d *node.Node

	name := elem.GetNameS()
	newPath := path

	if elem.GetNameS() != "" {
		newPath = append(newPath, name)
		d = node.NewXMLS(name)
	}

	if elem.GetContentS() != "" {
		if clean := z.HandleCounter(newPath, elem.GetContentS()); clean != "" {
			d = node.NewXMLS(clean)
		}
	}

	if desired != nil && d != nil {
		desired.AddChild(d)
	}
	for _, child := range elem.GetChildren() {
		z.ParseCounters(child, d, newPath)
	}
}

func (z *Zapi) HandleCounter(path []string, content string) string {
	var (
		name, display, key    string
		splitValues, fullPath []string
		err                   error
	)

	mat := z.Matrix[z.Object]
	splitValues = strings.Split(content, "=>")
	if len(splitValues) == 1 {
		name = content
	} else {
		name = splitValues[0]
		display = strings.TrimSpace(splitValues[1])
	}

	name = strings.TrimSpace(strings.TrimLeft(name, "^"))

	fullPath = path
	fullPath = append(fullPath, name)
	key = strings.Join(fullPath, ".")

	if display == "" {
		display = template.ParseZAPIDisplay(mat.Object, fullPath)
	}

	if content[0] == '^' {
		z.instanceLabelPaths[key] = display
		if content[1] == '^' {
			copied := make([]string, len(fullPath))
			copy(copied, fullPath)
			z.instanceKeyPaths = append(z.instanceKeyPaths, copied)
		}
	} else {
		// use user-defined metric type
		if t := z.Params.GetChildContentS("metric_type"); t != "" {
			_, err = mat.NewMetricType(key, t, display)
			// use uint64 as default, since nearly all ZAPI counters are unsigned
		} else {
			_, err = mat.NewMetricUint64(key, display)
		}
		if err != nil {
			z.Logger.Error(
				"Failed to add metric",
				slogx.Err(err),
				slog.String("key", key),
				slog.String("display", display),
			)
		}
	}

	return name
}
