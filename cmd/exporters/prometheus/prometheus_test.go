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

    output := FilterMetaTags(example)

    if len(output) != len(expected) {
        t.Fatalf("filtered data should have %d, but got %d lines", len(expected), len(output))
    }

    // output should have exact same lines in same order
    for i, _ := range output {
        if ! bytes.Equal(output[i], expected[i]) {
            t.Fatalf("line:%d - output = [%s], expected = [%s]", i, string(output[i]), string(expected[i]))
        }
    }

    t.Log("OK - output is exactly what is expected")
}

