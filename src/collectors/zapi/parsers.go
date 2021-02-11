package main

import (
    "strings"
    
	"goharvest2/share/logger"

    "goharvest2/poller/struct/matrix"
    "goharvest2/poller/struct/yaml"
    "goharvest2/poller/util"
)


func ParseShortestPath(m *matrix.Matrix) []string {

    prefix := make([]string, 0)
    keys := make([][]string, 0)

    for key, _ := range m.Metrics {
        keys = append(keys, strings.Split(key, "."))
    }
    for key, _ := range m.LabelNames.Iter() {
        keys = append(keys, strings.Split(key, "."))
    }

    max := util.MinLen(keys)
    
    for i:=0; i<max; i+=1 {
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
    n = util.MinLen(keys)-1
    for i=0; i<n; i+=1 {
        if util.AllSame(keys, i) {
            prefix = append(prefix, keys[0][i])
        } else {
            break
        }
    }
    return prefix
}

func ParseCounters(data *matrix.Matrix, elem *yaml.Node, path []string) {
    new_path := append(path, elem.Name)
    logger.Debug("", "%v Parsing [%s] [%s] with %d values and %d children", new_path, elem.Name, elem.Value, len(elem.Values), len(elem.Children))

    if elem.Value != "" {
        HandleCounter(data, new_path, elem.Value)
    }
    for _, value := range elem.Values {
        HandleCounter(data, new_path, value)
    }
    for _, child := range elem.Children {
        ParseCounters(data, child, new_path)
    }
}


func HandleCounter(data *matrix.Matrix, path []string, value string) {
    var name, display, flat_path string
    var split_value, full_path []string

    split_value = strings.Split(value, "=>")
    if len(split_value) == 1 {
        name = value
    } else {
        name = split_value[0]
        display = strings.TrimLeft(split_value[1], " ")
    }

    name = strings.TrimLeft(name, "^")
    name = strings.TrimRight(name, " ")

    full_path = append(path[1:], name)
    flat_path = strings.Join(full_path, ".")

    if display == "" {
        display = ParseDisplay(data.Object, full_path)
    }

    if value[0] == '^' {
        data.AddLabelKeyName(flat_path, display)
            logger.Trace("", "%sAdded as Label [%s] [%s]%s => %v", util.Yellow, display, flat_path, util.End, full_path)
        if value[1] == '^' {
            data.AddInstanceKey(full_path[:])
            logger.Trace("", "%sAdded as Key [%s] [%s]%s => %v", util.Red, display, flat_path, util.End, full_path)
        }
    } else {
        data.AddMetric(flat_path, display, true)
            logger.Trace("", "%sAdded as Metric [%s] [%s]%s => %v", util.Blue, display, flat_path, util.End, full_path)
    }
}


func ParseDisplay(obj string, path []string) string {
    var ignore = map[string]int{"attributes" : 0, "info" : 0, "list" : 0, "details" : 0}
    var added = map[string]int{}
    var words []string

    obj_words := strings.Split(obj, "_")
    for _, w := range obj_words {
        ignore[w] = 0
    }

    for _, attribute := range path {
        split := strings.Split(attribute, "-")
        for _, word := range split {
            if word == obj { continue }
            if _, exists := ignore[word]; exists { continue }
            if _, exists := added[word]; exists { continue }
            words = append(words, word)
            added[word] = 0
        }
    }
    return strings.Join(words, "_")
}
