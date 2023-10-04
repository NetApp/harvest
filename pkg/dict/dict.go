/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package dict

import "strings"

func String(m map[string]string) string {
	b := strings.Builder{}
	for k, v := range m {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
		b.WriteString(", ")
	}

	s := b.String()
	if len(s) > 0 {
		return s[:len(s)-2]
	}
	return s
}
