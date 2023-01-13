/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package zapi

import (
	"github.com/netapp/harvest/v2/pkg/color"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
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

	for i := 0; i < max; i++ {
		if util.AllSame(keys, i) {
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

	//counters.SetXMLNameS("desired-attributes")
	//counters.SetContentS("")
	return len(z.Matrix[z.Object].GetMetrics()) > 0, desired
}

func (z *Zapi) ParseCounters(elem, desired *node.Node, path []string) {
	//logger.Debug("", "%v Parsing [%s] [%s] with %d values and %d children", new_path, elem.Name, elem.Value, len(elem.Values), len(elem.Children))

	var d *node.Node

	name := elem.GetNameS()
	newPath := path

	if len(elem.GetNameS()) != 0 {
		newPath = append(newPath, name)
		d = node.NewXMLS(name)
	}

	if len(elem.GetContentS()) != 0 {
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

	fullPath = append(path, name)
	key = strings.Join(fullPath, ".")

	if display == "" {
		display = util.ParseZAPIDisplay(mat.Object, fullPath)
	}

	if content[0] == '^' {
		z.instanceLabelPaths[key] = display
		//data.AddLabel(key, display)
		z.Logger.Trace().Msgf("%sadd (%s) as label [%s]%s => %v", color.Yellow, key, display, color.End, fullPath)
		if content[1] == '^' {
			copied := make([]string, len(fullPath))
			copy(copied, fullPath)
			z.instanceKeyPaths = append(z.instanceKeyPaths, copied)
			z.Logger.Trace().Msgf("%sadd (%s) as instance key [%s]%s => %v", color.Red, key, display, color.End, fullPath)
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
			z.Logger.Error().Err(err).Str("key", key).Str("display", display).Msg("Failed to add metric")
		} else {
			z.Logger.Trace().Str("key", key).Str("display", display).Strs("fullPath", fullPath).
				Msg("Add metric")
		}
	}

	return name
}
