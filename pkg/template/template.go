package template

import (
	"regexp"
	"strings"
)

// ParseMetric parses display name and type of field and metric type from the raw name of the metric as defined in (sub)template.
// Users can rename a metric with "=>" (e.g., some_long_metric_name => short).
// Trailing "^" characters are ignored/cleaned as they have special meaning in some collectors.
func ParseMetric(rawName string) (string, string, string, string) {
	var (
		name, display string
		values        []string
	)
	metricType := ""
	// Ex: last_transfer_duration(duration) => last_transfer_duration
	if values = strings.SplitN(rawName, "=>", 2); len(values) == 2 {
		name = strings.TrimSpace(values[0])
		display = strings.TrimSpace(values[1])
		name, metricType = ParseMetricType(name)
	} else {
		name = rawName
		display = strings.ReplaceAll(rawName, ".", "_")
		display = strings.ReplaceAll(display, "-", "_")
	}

	if strings.HasPrefix(name, "^^") {
		return strings.TrimPrefix(name, "^^"), strings.TrimPrefix(display, "^^"), "key", ""
	}

	if strings.HasPrefix(name, "^") {
		return strings.TrimPrefix(name, "^"), strings.TrimPrefix(display, "^"), "label", ""
	}

	return name, display, "float", metricType
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

	for _, w := range strings.Split(obj, "_") {
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
