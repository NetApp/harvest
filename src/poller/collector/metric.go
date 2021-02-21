package collector

import (
	"strings"
)

// Parse raw metric name from collector template
// Simple name (e.g. "metric_name"), means both name and display are the same
// Custom name (e.g. "metric_name => custom_name") is parsed as display name.
func ParseMetricName(raw string) (string, string) {

	var name, display string

	name = strings.ReplaceAll(raw, "^", "")	

	if x := strings.Split(name, "=>"); len(x) == 2 {
		name = strings.TrimSpace(x[0])
		display = strings.TrimSpace(x[1])
	} else {
		display = strings.ReplaceAll(name, "-", "_")
	}

	return name, display
}