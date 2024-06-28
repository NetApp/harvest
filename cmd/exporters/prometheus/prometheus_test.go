/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package prometheus

import (
	"bytes"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"slices"
	"strings"
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

	// output should have exact same lines in the same order
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

func setUpMatrix(object string) *matrix.Matrix {
	m := matrix.New("bike", object, "bike")
	speed, _ := m.NewMetricUint64("max_speed")
	instanceNames := []string{"A", "B"}
	for _, instanceName := range instanceNames {
		instance, _ := m.NewInstance(instanceName)
		_ = speed.SetValueInt64(instance, 3)
	}
	return m
}

func TestRender(t *testing.T) {

	type test struct {
		prefix string
		want   string
	}

	tests := []test{
		{"bike", `bike_max_speed{} 3
bike_max_speed{} 3`},
		{"", `max_speed{} 3
max_speed{} 3`},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			p, err := setupPrometheusExporter()
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			m := setUpMatrix(tt.prefix)

			_, err = p.Export(m)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}

			prom := p.(*Prometheus)
			var lines []string
			for _, metrics := range prom.cache.Get() {
				for _, metric := range metrics {
					lines = append(lines, string(metric))
				}
			}

			slices.Sort(lines)
			if strings.Join(lines, "\n") != tt.want {
				t.Errorf("got = [%s], want = [%s]", strings.Join(lines, "\n"), tt.want)
			}
		})
	}
}

func setupPrometheusExporter() (exporter.Exporter, error) {
	absExp := exporter.New(
		"Prometheus",
		"prom1",
		&options.Options{PromPort: 1},
		conf.Exporter{IsTest: true},
		nil,
	)
	p := New(absExp)
	err := p.Init()
	return p, err
}
