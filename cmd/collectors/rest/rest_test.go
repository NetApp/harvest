package rest

import (
	"io/ioutil"
	"testing"
)

func Test_getFieldName(t *testing.T) {

	type test struct {
		name   string
		source string
		parent string
		want   int
	}

	var tests = []test{
		{
			name:   "Test1",
			source: readFile("testdata/cluster.json"),
			parent: "",
			want:   52,
		},
		{
			name:   "Test2",
			source: readFile("testdata/sample.json"),
			parent: "",
			want:   3,
		},
		{
			name:   "Test3",
			source: readFile("testdata/test.json"),
			parent: "",
			want:   9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFieldName(tt.source, tt.parent); len(got) != tt.want {
				t.Errorf("length of output slice = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func readFile(path string) string {
	b, _ := ioutil.ReadFile(path)
	return string(b)
}
