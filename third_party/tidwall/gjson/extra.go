package gjson

import "strings"

func (t Result) ClonedString() string {
	return strings.Clone(t.String())
}
