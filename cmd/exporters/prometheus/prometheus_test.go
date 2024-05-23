/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package prometheus

import (
	"bytes"
	"testing"
)

func TestFilterMetaTags(t *testing.T) {

	example := [][]byte{
		[]byte(`# HELP some_metric help text`),
		[]byte(`# TYPE some_metric type`),
		[]byte(`some_metric{node="node_1"} 0.0`),
		[]byte(`# HELP some_other_metric help text`),
		[]byte(`# TYPE some_other_metric type`),
		[]byte(`some_other_metric{node="node_2"} 0.0`),
		[]byte(`# HELP some_other_metric DUPLICATE help text`),
		[]byte(`# TYPE some_other_metric type`),
		[]byte(`some_other_metric{node="node_3"} 0.0`),
	}

	expected := [][]byte{
		[]byte(`# HELP some_metric help text`),
		[]byte(`# TYPE some_metric type`),
		[]byte(`some_metric{node="node_1"} 0.0`),
		[]byte(`# HELP some_other_metric help text`),
		[]byte(`# TYPE some_other_metric type`),
		[]byte(`some_other_metric{node="node_2"} 0.0`),
		[]byte(`some_other_metric{node="node_3"} 0.0`),
	}

	output := filterMetaTags(example)

	if len(output) != len(expected) {
		t.Fatalf("filtered data should have %d, but got %d lines", len(expected), len(output))
	}

	// output should have exact same lines in same order
	for i := range output {
		if !bytes.Equal(output[i], expected[i]) {
			t.Fatalf("line:%d - output = [%s], expected = [%s]", i, string(output[i]), string(expected[i]))
		}
	}

	t.Log("OK - output is exactly what is expected")
}

func TestEscape(t *testing.T) {
	replacer := newReplacer()

	type test struct {
		key   string
		value string
		want  string
	}

	tests := []test{
		{key: `abc`, value: `abc`, want: `abc="abc"`},
		{key: `abc`, value: `a"b"c`, want: `abc="a\\\"b\\\"c"`},
		{key: `abc`, value: `a\c`, want: `abc="a\\\\c"`},
		{key: `abc`, value: `a\nc`, want: `abc="a\\\\nc"`},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := escape(replacer, tc.key, tc.value)
			if got != tc.want {
				t.Errorf("escape failed got=[%s] want=[%s] for key=[%s] value=[%s]", got, tc.want, tc.key, tc.value)
			}
		})
	}
}

func BenchmarkEscape(b *testing.B) {
	replacer := newReplacer()
	for range b.N {
		escape(replacer, "abc", `a\c"foo"\ndef`)
	}
}
