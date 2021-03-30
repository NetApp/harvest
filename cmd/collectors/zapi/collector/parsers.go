//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package zapi_collector

import (
    "goharvest2/pkg/logger"
    "goharvest2/pkg/matrix"
    "goharvest2/pkg/tree/node"
    "goharvest2/pkg/util"
    "strings"
)

func ParseShortestPath(m *matrix.Matrix) []string {

    prefix := make([]string, 0)
    keys := make([][]string, 0)

    for key := range m.GetMetrics() {
        keys = append(keys, strings.Split(key, "."))
    }
    for key := range m.GetLabels() {
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

func LoadCounters(data *matrix.Matrix, counters *node.Node) bool {
    path := make([]string, 0)
    ParseCounters(data, counters, path)
    return len(data.GetMetrics()) > 0
}

func ParseCounters(data *matrix.Matrix, elem *node.Node, path []string) {
    //logger.Debug("", "%v Parsing [%s] [%s] with %d values and %d children", new_path, elem.Name, elem.Value, len(elem.Values), len(elem.Children))

    new_path := path
    if len(elem.GetNameS()) != 0 {
        new_path = append(new_path, elem.GetNameS())
    }
    if len(elem.GetContentS()) != 0 {
        HandleCounter(data, new_path, elem.GetContentS())
    }
    for _, child := range elem.GetChildren() {
        ParseCounters(data, child, new_path)
    }
}

func HandleCounter(data *matrix.Matrix, path []string, content string) {
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

    full_path = append(path[1:], name)
    key = strings.Join(full_path, ".")

    if display == "" {
        display = ParseDisplay(data.Object, full_path)
    }

    if content[0] == '^' {
        data.AddLabel(key, display)
        logger.Trace("", "%sAdded as Label [%s] [%s]%s => %v", util.Yellow, display, key, util.End, full_path)
        if content[1] == '^' {
            data.AddInstanceKey(full_path[:])
            logger.Trace("", "%sAdded as Key [%s] [%s]%s => %v", util.Red, display, key, util.End, full_path)
        }
    } else {
        data.AddMetric(key, display, true)
        logger.Trace("", "%sAdded as Metric [%s] [%s]%s => %v", util.Blue, display, key, util.End, full_path)
    }
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
