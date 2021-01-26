package zapi

import (
	"strings"
	"poller/structs/matrix"
	"poller/yaml"
	"poller/share"
)


func ParseKeyPrefix(keys [][]string) []string {
    var prefix []string
    var i, n int
    n = share.MinLen(keys)-1
    for i=0; i<n; i+=1 {
        if share.AllSame(keys, i) {
            prefix = append(prefix, keys[0][i])
        } else {
            break
        }
    }
    return prefix
}

func ParseCounters(data *matrix.Matrix, elem *yaml.Node, path []string) {
    new_path := append(path, elem.Name)
    Log.Debug("%v Parsing [%s] [%s] with %d values and %d children", new_path, elem.Name, elem.Value, len(elem.Values), len(elem.Children))

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
        data.AddLabel(flat_path, display)
            Log.Trace("Added as Label [%s] [%s]", display, flat_path)
        if value[1] == '^' {
            data.AddInstanceKey(full_path)
            Log.Trace("Added as Key [%s] [%s]", display, flat_path)
        }
    } else {
        data.AddMetric(flat_path, display, true)
            Log.Trace("Added as Metric [%s] [%s]", display, flat_path)
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
