package dict

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]string
		wantCommas int
	}{
		{name: "empty", args: make(map[string]string), wantCommas: 0},
		{name: "none", args: map[string]string{"a": "a"}, wantCommas: 0},
		{name: "one", args: map[string]string{"a": "a", "b": "b"}, wantCommas: 1},
		{name: "two", args: map[string]string{"a": "a", "b": "b", "c": "c"}, wantCommas: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := String(tt.args)
			gotCommas := strings.Count(s, ",")
			if gotCommas != tt.wantCommas {
				t.Errorf("String() commas got=%d, want=%d", gotCommas, tt.wantCommas)
			}
		})
	}
}
