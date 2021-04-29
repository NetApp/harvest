/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
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
