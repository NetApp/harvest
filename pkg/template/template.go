package template

import (
	"regexp"
	"strings"
)

// ParseMetric parses display name and type of field and metric type from the raw name of the metric as defined in (sub)template.
// Users can rename a metric with "=>" (e.g., some_long_metric_name => short).
// Trailing "^" characters are ignored/cleaned as they have special meaning in some collectors.
func ParseMetric(rawName string) (string, string, string, string) {
	var name, display string
	metricType := ""
	// Ex: last_transfer_duration(duration) => last_transfer_duration
	if before, after, renamed := splitRename(rawName); renamed {
		display = after
		name, metricType = ParseMetricType(before)
	} else {
		name = rawName
		display = strings.ReplaceAll(rawName, ".", "_")
		display = strings.ReplaceAll(display, "-", "_")
	}

	if after, ok := strings.CutPrefix(name, "^^"); ok {
		return after, strings.TrimPrefix(display, "^^"), "key", ""
	}

	if after, ok := strings.CutPrefix(name, "^"); ok {
		return after, strings.TrimPrefix(display, "^"), "label", ""
	}

	return name, display, "float", metricType
}

// splitRename splits the "name => display" rename form shared by every counter
// template, returning the name, the display name, and whether a rename was
// present. When rawName carries no "=>" the third result is false, the name is
// rawName unchanged and the display is empty. Both halves are whitespace
// trimmed, so "a=>b" and "a => b" parse identically.
func splitRename(rawName string) (string, string, bool) {
	before, after, found := strings.Cut(rawName, "=>")
	if !found {
		return rawName, "", false
	}
	return strings.TrimSpace(before), strings.TrimSpace(after), true
}

// SplitMetricRename returns a counter's name and display name. Unlike ParseMetric
// it applies no display normalization and no "^"/"^^" sigil or "(type)" handling:
// when rawName carries no rename it is returned as both name and display, so names
// keep any "." and "-" characters. Used by the unix and simple collectors, whose
// metric names legitimately contain those characters. A degenerate rename with an
// empty half ("a =>", "=> b") is treated as no rename at all.
func SplitMetricRename(rawName string) (string, string) {
	name, display, renamed := splitRename(rawName)
	if !renamed || name == "" || display == "" {
		return rawName, rawName
	}
	return name, display
}

func ParseMetricType(metricName string) (string, string) {
	metricTypeRegex := regexp.MustCompile(`(.*)\((.*?)\)`)
	match := metricTypeRegex.FindAllStringSubmatch(metricName, -1)
	if match != nil {
		// For last_transfer_duration(duration), name would have 'last_transfer_duration' and metricType would have 'duration'.
		name := match[0][1]
		metricType := match[0][2]
		return name, metricType
	}
	return metricName, ""
}

func ParseZAPIDisplay(obj string, path []string) string {
	var (
		ignore = map[string]int{"attributes": 0, "info": 0, "list": 0, "details": 0, "storage": 0}
		added  = map[string]int{}
		words  []string
	)

	for w := range strings.SplitSeq(obj, "_") {
		ignore[w] = 0
	}

	for _, attribute := range path {
		split := strings.SplitSeq(attribute, "-")
		for word := range split {
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

var arrayRegex = regexp.MustCompile(`^([a-zA-Z][\w.]*)(\.[0-9#])`)

var metricReplacer = strings.NewReplacer("\n", "", " ", "", "\"", "")

func ArrayMetricToString(value string) string {
	s := metricReplacer.Replace(value)

	openBracket := strings.Index(s, "[")
	closeBracket := strings.Index(s, "]")
	if openBracket > -1 && closeBracket > -1 {
		return s[openBracket+1 : closeBracket]
	}
	return value
}

func HandleArrayFormat(name string) string {
	matches := arrayRegex.FindStringSubmatch(name)
	if len(matches) > 2 {
		return matches[1]
	}
	return name
}
