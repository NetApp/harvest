/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package prometheus

import (
	"github.com/google/go-cmp/cmp"
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

	expected := `# HELP some_metric help text
# TYPE some_metric type
some_metric{node="node_1"} 0.0
# HELP some_other_metric help text
# TYPE some_other_metric type
some_other_metric{node="node_2"} 0.0
some_other_metric{node="node_3"} 0.0
`
	p := Prometheus{}
	seen := make(map[string]struct{})
	var w strings.Builder
	_ = p.writeMetrics(&w, example, seen)

	diff := cmp.Diff(w.String(), expected)
	if diff != "" {
		t.Errorf("Mismatch (-got +want):\n%s", diff)
	}
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

func setUpChangeMatrix() *matrix.Matrix {
	m := matrix.New("change", "change", "change")
	// Create a metric with a metric value change
	log, _ := m.NewMetricUint64("log")
	instance, _ := m.NewInstance("A")
	_ = log.SetValueInt64(instance, 3)
	instance.SetLabel("category", "metric")
	instance.SetLabel("cluster", "umeng-aff300-01-02")
	instance.SetLabel("object", "volume")
	instance.SetLabel("op", "metric_change")
	instance.SetLabel("track", "volume_size_total")

	// Create a metric with a label change
	instance2, _ := m.NewInstance("B")
	_ = log.SetValueInt64(instance2, 3)
	instance2.SetLabel("category", "label")
	instance2.SetLabel("cluster", "umeng-aff300-01-02")
	instance2.SetLabel("new_value", "offline")
	instance2.SetLabel("object", "volume")
	instance2.SetLabel("old_value", "online")
	instance2.SetLabel("op", "update")
	instance2.SetLabel("track", "state")

	return m
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
			p, err := setUpPrometheusExporter("")
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

func TestGlobalPrefixWithChangelog(t *testing.T) {

	type test struct {
		prefix string
		want   string
	}

	tests := []test{
		{"prefix", `
netapp_change_log{category="label",cluster="umeng-aff300-01-02",new_value="offline",object="volume",old_value="online",op="update",track="state"} 3
netapp_change_log{category="metric",cluster="umeng-aff300-01-02",object="volume",op="metric_change",track="netapp_volume_size_total"} 3`},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			p, err := setUpPrometheusExporter("netapp")

			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
			m := setUpChangeMatrix()

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
			diff := cmp.Diff(strings.TrimSpace(tt.want), strings.Join(lines, "\n"))
			if diff != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func setUpPrometheusExporter(prefix string) (exporter.Exporter, error) {

	absExp := exporter.New(
		"Prometheus",
		"prom1",
		&options.Options{PromPort: 1},
		conf.Exporter{
			IsTest:     true,
			SortLabels: true,
		},
		nil,
	)

	if prefix != "" {
		absExp.Params.GlobalPrefix = &prefix
	}
	p := New(absExp)
	err := p.Init()
	return p, err
}
