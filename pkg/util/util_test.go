package util

import (
	"testing"
)

func TestIntersection(t *testing.T) {

	type test struct {
		name        string
		a           []string
		b           []string
		matchLength int
		missLength  int
	}

	tests := []test{
		{
			name:        "test1",
			a:           []string{"a", "b", "c"},
			b:           []string{"a", "c"},
			matchLength: 2,
			missLength:  0,
		},
		{
			name:        "test2",
			a:           []string{"a", "b", "c"},
			b:           []string{"a", "e", "d"},
			matchLength: 1,
			missLength:  2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, miss := Intersection(tt.a, tt.b)
			if len(match) != tt.matchLength {
				t.Errorf("Intersection() match length = %v, want %v", len(match), tt.matchLength)
			}
			if len(miss) != tt.missLength {
				t.Errorf("Intersection() miss length = %v, want %v", len(miss), tt.missLength)
			}
		})
	}
}
