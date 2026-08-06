package cmperf

import (
	"log/slog"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/cmmetrics"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func TestIsCompleteCollection(t *testing.T) {
	tests := []struct {
		name     string
		statuses []cmmetrics.StatusCode
		want     bool
	}{
		{
			name:     "complete only",
			statuses: []cmmetrics.StatusCode{{Code: cmmetrics.CompleteCollection}},
			want:     true,
		},
		{
			name: "complete then no-additional",
			statuses: []cmmetrics.StatusCode{
				{Code: cmmetrics.CompleteCollection},
				{Code: cmmetrics.NoAdditionalStatus},
			},
			want: true,
		},
		{
			name: "no-additional then complete",
			statuses: []cmmetrics.StatusCode{
				{Code: cmmetrics.NoAdditionalStatus},
				{Code: cmmetrics.CompleteCollection},
			},
			want: true,
		},
		{
			name: "complete plus secondary metrics file is not complete",
			statuses: []cmmetrics.StatusCode{
				{Code: cmmetrics.CompleteCollection},
				{Code: cmmetrics.SecondaryMetricsFile},
			},
			want: false,
		},
		{
			name:     "secondary metrics file alone",
			statuses: []cmmetrics.StatusCode{{Code: cmmetrics.SecondaryMetricsFile}},
			want:     false,
		},
		{
			name:     "empty statuses",
			statuses: nil,
			want:     false,
		},
		{
			name:     "no-additional alone, no complete",
			statuses: []cmmetrics.StatusCode{{Code: cmmetrics.NoAdditionalStatus}},
			want:     false,
		},
		{
			name:     "partial collection",
			statuses: []cmmetrics.StatusCode{{Code: cmmetrics.PartialCollection}},
			want:     false,
		},
		{
			name: "complete plus partial plus no-additional",
			statuses: []cmmetrics.StatusCode{
				{Code: cmmetrics.CompleteCollection},
				{Code: cmmetrics.PartialCollection},
				{Code: cmmetrics.NoAdditionalStatus},
			},
			want: false,
		},
		{
			name: "multiple complete entries with different nodes",
			statuses: []cmmetrics.StatusCode{
				{Code: cmmetrics.CompleteCollection, Nodes: []string{"node1"}},
				{Code: cmmetrics.CompleteCollection, Nodes: []string{"node2"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCompleteCollection(tt.statuses)
			assert.Equal(t, got, tt.want)
		})
	}
}

func newTestCmPerf(t *testing.T) *CmPerf {
	t.Helper()
	c := &CmPerf{Rest: &rest2.Rest{AbstractCollector: &collector.AbstractCollector{
		Object: "test",
		Logger: slog.Default(),
	}}}
	c.InitProp()
	c.Prop.Object = "test"
	c.Params = node.NewS("root")
	c.Matrix = map[string]*matrix.Matrix{c.Object: matrix.New("test", "test", "test")}
	c.perfProp = &perfProp{
		counterInfo:       make(map[string]*counter),
		histogramCounters: make(map[string]bool),
	}
	return c
}

func TestHistogramCounters(t *testing.T) {
	yml := `
histograms:
  - smb2_latency_by_size
`
	root, err := tree.LoadYaml([]byte(yml))
	assert.Nil(t, err)

	tests := []struct {
		name          string
		labels        []string
		values        []uint64
		wantHistogram bool
	}{
		{name: "read_latency_hist", labels: []string{"0-1ms", "1-2ms", "2-5ms"}, values: []uint64{10, 20, 30}, wantHistogram: true}, // name heuristic
		{name: "smb2_latency_by_size", labels: []string{"0-1ms", "1-2ms"}, values: []uint64{5, 6}, wantHistogram: true},             // template override only
		{name: "read_ops_by_size", labels: []string{"small", "large"}, values: []uint64{1, 2}, wantHistogram: false},                // plain array, no marker
		{name: "read_ops", wantHistogram: false}, // scalar, never a histogram
		{name: "state_reference_history", labels: []string{"a", "b"}, values: []uint64{1, 2}, wantHistogram: false}, // "_hist" substring, not a real histogram suffix
	}

	c := newTestCmPerf(t)
	c.Params = root
	for _, name := range root.GetChildS("histograms").GetAllChildContentS() {
		c.perfProp.histogramCounters[name] = true
	}

	schema := cmmetrics.ObjectSchema{}
	for i, tt := range tests {
		schema.CounterSchema = append(schema.CounterSchema, cmmetrics.CounterSchema{Index: uint32(i + 1), Name: tt.name, LabelsX: tt.labels}) //nolint:gosec
		c.Prop.Metrics[tt.name] = &rest2.Metric{Label: tt.name, Exportable: true}
	}
	c.buildCountersFromSchema(schema, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			co := c.perfProp.counterInfo[tt.name]
			if co == nil {
				t.Fatalf("counter %s missing from counterInfo", tt.name)
			}
			assert.Equal(t, co.isHistogram, tt.wantHistogram)

			if len(tt.labels) == 0 {
				return
			}

			curMat := matrix.New("test", "test", "test")
			prevMat := matrix.New("test", "test", "test")
			inst, err := curMat.NewInstance("inst1")
			assert.Nil(t, err)

			cs := cmmetrics.CounterSchema{Name: tt.name, LabelsX: tt.labels}
			count := c.populateArrayCounter(curMat, prevMat, inst, cs, tt.values)
			assert.Equal(t, count, uint64(len(tt.values)))

			bucket := curMat.GetMetric(tt.name + ".bucket")
			if tt.wantHistogram {
				if bucket == nil {
					t.Fatal("expected bucket metric to be created")
				}
				assert.True(t, bucket.IsHistogram())
			} else if bucket != nil {
				t.Fatal("did not expect a bucket metric for a plain array counter")
			}

			for i, label := range tt.labels {
				m := curMat.GetMetric(tt.name + arrayKeyToken + label)
				if m == nil {
					t.Fatalf("expected flattened metric for label %q", label)
				}
				assert.Equal(t, m.IsHistogram(), tt.wantHistogram)
				val, ok := m.GetValueFloat64(inst)
				assert.True(t, ok)
				assert.Equal(t, val, float64(tt.values[i]))
			}
		})
	}
}

func TestPopulateArrayCounterCreatesInPrevMat(t *testing.T) {
	c := newTestCmPerf(t)
	name := "read_latency_hist"
	labels := []string{"0-1ms", "1-2ms"}
	values := []uint64{10, 20}

	c.perfProp.counterInfo[name] = &counter{isHistogram: true}
	c.Prop.Metrics[name] = &rest2.Metric{Label: name, Exportable: true}

	curMat := matrix.New("test", "test", "test")
	prevMat := matrix.New("test", "test", "test")
	inst, err := curMat.NewInstance("inst1")
	assert.Nil(t, err)

	cs := cmmetrics.CounterSchema{Name: name, LabelsX: labels}
	count := c.populateArrayCounter(curMat, prevMat, inst, cs, values)
	assert.Equal(t, count, uint64(len(values)))

	bucketKey := name + ".bucket"
	if curMat.GetMetric(bucketKey) == nil {
		t.Fatal("expected bucket metric in curMat")
	}
	if prevMat.GetMetric(bucketKey) == nil {
		t.Fatal("expected bucket metric to also exist in prevMat")
	}

	for _, label := range labels {
		key := name + arrayKeyToken + label
		if curMat.GetMetric(key) == nil {
			t.Fatalf("expected %s in curMat", key)
		}
		if prevMat.GetMetric(key) == nil {
			t.Fatalf("expected %s to also exist in prevMat", key)
		}
	}
}
