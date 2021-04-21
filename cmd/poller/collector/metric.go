//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
// Parse raw metric name from collector template
//
// Examples:
// Simple name (e.g. "metric_name"), means both name and display are the same
// Custom name (e.g. "metric_name => custom_name") is parsed as display name.
//
package collector

import (
	"strings"
)

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
