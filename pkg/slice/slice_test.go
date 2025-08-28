package slice

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
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
			assert.Equal(t, len(match), tt.matchLength)
			assert.Equal(t, len(miss), tt.missLength)
		})
	}
}

func TestHasDuplicates(t *testing.T) {
	type test struct {
		testCase      string
		childNames    []string
		expectedArray bool
	}

	tests := []test{
		{
			testCase:      "testcasenochild",
			childNames:    []string{},
			expectedArray: false,
		},
		{
			testCase:      "testcaseonechild",
			childNames:    []string{"type"},
			expectedArray: false,
		},
		{
			testCase:      "testcasemultipleChild",
			childNames:    []string{"type", "style", "aggr"},
			expectedArray: false,
		},
		{
			testCase:      "testcasearray",
			childNames:    []string{"aggr-name", "aggr-name", "aggr-name", "aggr-name"},
			expectedArray: true,
		},
	}

	for _, testcase := range tests {
		isArray := HasDuplicates(testcase.childNames)
		assert.Equal(t, isArray, testcase.expectedArray)
	}
}
