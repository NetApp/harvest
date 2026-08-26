package cmperf

import (
	"log/slog"
	"os"
	"path/filepath"
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
		counterInfo:          make(map[string]*counter),
		histogramCounters:    make(map[string]bool),
		arrayShapeMismatches: make(map[string]int),
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
	c.buildCountersFromSchema(schema, matrix.New("test", "test", "test"), matrix.New("test", "test", "test"))

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

func TestBuildCountersFromSchema_CreatesOnCurMat(t *testing.T) {
	c := newTestCmPerf(t)
	c.Prop.Metrics["rx_bytes"] = &rest2.Metric{Label: "rx_bytes", Exportable: true}
	c.Prop.Metrics["read_io_type"] = &rest2.Metric{Label: "read_io_type", Exportable: true}

	schema := cmmetrics.ObjectSchema{CounterSchema: []cmmetrics.CounterSchema{
		{Index: 1, Name: "rx_bytes"}, // plain scalar, no denominator
		{Index: 2, Name: "read_io_type", BaseIndex: 3, HasBaseIndex: true, LabelsX: []string{"cache", "pmem"}}, // array numerator with denominator
		{Index: 3, Name: "read_io_type_base"}, // scalar denominator
	}}

	curMat := matrix.New("test", "test", "test")
	prevMat := matrix.New("test", "test", "test")
	c.buildCountersFromSchema(schema, curMat, prevMat)

	if curMat.GetMetric("rx_bytes") == nil {
		t.Fatal("expected scalar counter rx_bytes to be created on curMat")
	}
	if curMat.GetMetric("read_io_type") != nil {
		t.Fatal("did not expect a bare placeholder for array-shaped read_io_type")
	}
	if curMat.GetMetric("read_io_type_base") == nil {
		t.Fatal("expected denominator read_io_type_base to be created on curMat even though its numerator is array-shaped")
	}
	if co := c.perfProp.counterInfo["read_io_type"]; co == nil || co.denominator != "read_io_type_base" {
		t.Fatalf("expected read_io_type.denominator == read_io_type_base, got %+v", co)
	}
	if m := curMat.GetMetric("read_io_type_base"); m.IsExportable() {
		t.Fatal("expected synthesized denominator read_io_type_base to be non-exportable")
	}
}

func TestCookCounters_ArrayShapedBaseShipsRaw(t *testing.T) {
	c := newTestCmPerf(t)
	c.Metadata = matrix.New("test", "metadata", "metadata")
	if _, err := c.Metadata.NewInstance("data"); err != nil {
		t.Fatal(err)
	}
	_, _ = c.Metadata.NewMetricUint64("skips")
	_, _ = c.Metadata.NewMetricUint64("numPartials")
	_, _ = c.Metadata.NewMetricUint64("instances")
	_, _ = c.Metadata.NewMetricInt64("calc_time")

	c.Prop.Metrics["service_time"] = &rest2.Metric{Label: "service_time", Exportable: true}

	schema := cmmetrics.ObjectSchema{CounterSchema: []cmmetrics.CounterSchema{
		// scalar numerator, average type, base is array-shaped -> must be forced to raw
		{Index: 1, Name: "service_time", Type: cmmetrics.CookAverage, BaseIndex: 2, HasBaseIndex: true},
		{Index: 2, Name: "visits", LabelsX: []string{"cpu0", "cpu1"}},
	}}

	curMat := matrix.New("test", "test", "test")
	prevMat := matrix.New("test", "test", "test")
	c.buildCountersFromSchema(schema, curMat, prevMat)

	if co := c.perfProp.counterInfo["service_time"]; co == nil || co.counterType != "raw" || co.denominator != "" {
		t.Fatalf("expected service_time forced to raw with no denominator, got %+v", co)
	}

	inst, err := curMat.NewInstance("inst1")
	assert.Nil(t, err)
	curMat.MustGetMetric("service_time").SetValueFloat64(inst, 100)
	tsMetric, err := curMat.NewMetricFloat64(timestampMetricName)
	assert.Nil(t, err)
	tsMetric.SetProperty("raw")
	tsMetric.SetValueFloat64(inst, 2)

	prevInst, err := prevMat.NewInstance("inst1")
	assert.Nil(t, err)
	prevMat.MustGetMetric("service_time").SetValueFloat64(prevInst, 40)
	prevTs, err := prevMat.NewMetricFloat64(timestampMetricName)
	assert.Nil(t, err)
	prevTs.SetValueFloat64(prevInst, 1)

	newData, err := c.cookCounters(curMat, prevMat)
	assert.Nil(t, err)

	got, ok := newData[c.Object].GetMetric("service_time").GetValueFloat64(inst)
	assert.True(t, ok)
	// raw ships the value untouched (no delta, no divide) - NOT (100-40)=60 and NOT divided.
	assert.Equal(t, got, float64(100))
}

func TestCookCounters_ArrayNumeratorWithScalarBaseDivides(t *testing.T) {
	c := newTestCmPerf(t)
	c.Metadata = matrix.New("test", "metadata", "metadata")
	if _, err := c.Metadata.NewInstance("data"); err != nil {
		t.Fatal(err)
	}
	_, _ = c.Metadata.NewMetricUint64("skips")
	_, _ = c.Metadata.NewMetricUint64("numPartials")
	_, _ = c.Metadata.NewMetricUint64("instances")
	_, _ = c.Metadata.NewMetricInt64("calc_time")

	c.Prop.Metrics["read_io_type"] = &rest2.Metric{Label: "read_io_type", Exportable: true}

	schema := cmmetrics.ObjectSchema{CounterSchema: []cmmetrics.CounterSchema{
		{Index: 1, Name: "read_io_type", Type: cmmetrics.CookPercent, BaseIndex: 2, HasBaseIndex: true, LabelsX: []string{"cache", "pmem"}},
		{Index: 2, Name: "read_io_type_base", Type: cmmetrics.CookDelta}, // base's own delta must run before the numerator's divide
	}}

	curMat := matrix.New("test", "test", "test")
	prevMat := matrix.New("test", "test", "test")
	c.buildCountersFromSchema(schema, curMat, prevMat)

	if co := c.perfProp.counterInfo["read_io_type"]; co == nil || co.counterType != "percent" || co.denominator != "read_io_type_base" {
		t.Fatalf("expected read_io_type to be percent with denominator read_io_type_base, got %+v", co)
	}

	inst, err := curMat.NewInstance("inst1")
	assert.Nil(t, err)
	prevInst, err := prevMat.NewInstance("inst1")
	assert.Nil(t, err)

	arrCs := cmmetrics.CounterSchema{Name: "read_io_type", LabelsX: []string{"cache", "pmem"}}
	count := c.populateArrayCounter(curMat, prevMat, inst, arrCs, []uint64{30, 5})
	assert.Equal(t, count, uint64(2))
	prevMat.MustGetMetric("read_io_type#cache").SetValueFloat64(prevInst, 10)
	prevMat.MustGetMetric("read_io_type#pmem").SetValueFloat64(prevInst, 2)

	curMat.MustGetMetric("read_io_type_base").SetValueFloat64(inst, 100)
	prevMat.MustGetMetric("read_io_type_base").SetValueFloat64(prevInst, 50)

	tsMetric, err := curMat.NewMetricFloat64(timestampMetricName)
	assert.Nil(t, err)
	tsMetric.SetProperty("raw")
	tsMetric.SetValueFloat64(inst, 2)
	prevTs, err := prevMat.NewMetricFloat64(timestampMetricName)
	assert.Nil(t, err)
	prevTs.SetValueFloat64(prevInst, 1)

	newData, err := c.cookCounters(curMat, prevMat)
	assert.Nil(t, err)

	// (30-10)/(100-50)*100 = 40, (5-2)/(100-50)*100 = 6
	cache, ok := newData[c.Object].GetMetric("read_io_type#cache").GetValueFloat64(inst)
	assert.True(t, ok)
	assert.Equal(t, cache, float64(40))
	pmem, ok := newData[c.Object].GetMetric("read_io_type#pmem").GetValueFloat64(inst)
	assert.True(t, ok)
	assert.Equal(t, pmem, float64(6))
}

func TestBuildCountersFromSchema_NoBaseIndexNotConfusedWithZero(t *testing.T) {
	c := newTestCmPerf(t)
	c.Prop.Metrics["tx_total_errors"] = &rest2.Metric{Label: "tx_total_errors", Exportable: true}

	schema := cmmetrics.ObjectSchema{CounterSchema: []cmmetrics.CounterSchema{
		{Index: 0, Name: "instance_name", Type: cmmetrics.CookString}, // occupies index 0
		{Index: 43, Name: "tx_total_errors"},                          // no base counter at all (HasBaseIndex left false)
	}}

	c.buildCountersFromSchema(schema, matrix.New("test", "test", "test"), matrix.New("test", "test", "test"))

	co := c.perfProp.counterInfo["tx_total_errors"]
	if co == nil {
		t.Fatal("expected counterInfo for tx_total_errors")
	}
	if co.denominator != "" {
		t.Fatalf("expected no denominator, got %q", co.denominator)
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

func TestPopulateArrayCounter2D(t *testing.T) {
	c := newTestCmPerf(t)
	name := "rss_matrix"
	labelsX := []string{"queue_0", "queue_1"}
	labelsY := []string{"rx_bytes", "tx_bytes"}
	values := []uint64{100, 200, 300, 400} // row-major: q0/rx, q0/tx, q1/rx, q1/tx

	c.Prop.Metrics[name] = &rest2.Metric{Label: name, Exportable: true}

	curMat := matrix.New("test", "test", "test")
	prevMat := matrix.New("test", "test", "test")
	inst, err := curMat.NewInstance("inst1")
	assert.Nil(t, err)

	cs := cmmetrics.CounterSchema{Name: name, LabelsX: labelsX, LabelsY: labelsY}
	count := c.populateArrayCounter(curMat, prevMat, inst, cs, values)
	assert.Equal(t, count, uint64(len(values)))

	want := map[string]float64{
		"queue_0#rx_bytes": 100,
		"queue_0#tx_bytes": 200,
		"queue_1#rx_bytes": 300,
		"queue_1#tx_bytes": 400,
	}
	for suffix, wantVal := range want {
		key := name + arrayKeyToken + suffix
		m := curMat.GetMetric(key)
		if m == nil {
			t.Fatalf("expected %s in curMat", key)
		}
		val, ok := m.GetValueFloat64(inst)
		assert.True(t, ok)
		assert.Equal(t, val, wantVal)

		if prevMat.GetMetric(key) == nil {
			t.Fatalf("expected %s to also exist in prevMat", key)
		}
	}
}

func TestPopulateArrayCounter_ShapeMismatchesAreSkippedAndCounted(t *testing.T) {
	curMat := matrix.New("test", "test", "test")
	prevMat := matrix.New("test", "test", "test")

	tests := []struct {
		name   string
		cs     cmmetrics.CounterSchema
		values []uint64
	}{
		{
			name:   "2D empty LabelsX",
			cs:     cmmetrics.CounterSchema{Name: "a", LabelsX: nil, LabelsY: []string{"rx_bytes", "tx_bytes"}},
			values: []uint64{1, 2},
		},
		{
			name:   "2D value count mismatch",
			cs:     cmmetrics.CounterSchema{Name: "b", LabelsX: []string{"queue_0", "queue_1"}, LabelsY: []string{"rx_bytes", "tx_bytes"}},
			values: []uint64{1, 2, 3}, // expected 4
		},
		{
			name:   "1D value count mismatch",
			cs:     cmmetrics.CounterSchema{Name: "c", LabelsX: []string{"cache", "pmem"}},
			values: []uint64{1},
		},
	}

	c := newTestCmPerf(t)
	inst, err := curMat.NewInstance("inst1")
	assert.Nil(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Prop.Metrics[tt.cs.Name] = &rest2.Metric{Label: tt.cs.Name, Exportable: true}
			count := c.populateArrayCounter(curMat, prevMat, inst, tt.cs, tt.values)
			assert.Equal(t, count, uint64(0))
			if c.perfProp.arrayShapeMismatches[tt.cs.Name] != 1 {
				t.Fatalf("expected arrayShapeMismatches[%s] == 1, got %d", tt.cs.Name, c.perfProp.arrayShapeMismatches[tt.cs.Name])
			}
		})
	}
}

func TestRetainCmperfFiles(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want int
	}{
		{name: "empty", env: "", want: 0},
		{name: "zero", env: "0", want: 0},
		{name: "positive", env: "3", want: 3},
		{name: "invalid", env: "abc", want: 0},
		{name: "negative", env: "-1", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(cmperfRetainFilesEnv, tt.env)
			assert.Equal(t, retainCmperfFiles(), tt.want)
		})
	}
}

func TestPruneCmperfTempDir(t *testing.T) {
	names := []string{
		"1000_volume.pb",
		"2000_volume.pb",
		"3000_volume.pb",
		"4000_volume.pb",
	}

	t.Run("retain 0 deletes all", func(t *testing.T) {
		sub := t.TempDir()
		for _, name := range names {
			if err := os.WriteFile(filepath.Join(sub, name), []byte("x"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		assert.Nil(t, pruneCmperfTempDir(sub, 0, nil))
		entries, err := os.ReadDir(sub)
		assert.Nil(t, err)
		assert.Equal(t, len(entries), 0)
	})

	t.Run("retain 2 keeps newest", func(t *testing.T) {
		sub := t.TempDir()
		for _, name := range names {
			if err := os.WriteFile(filepath.Join(sub, name), []byte("x"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		assert.Nil(t, pruneCmperfTempDir(sub, 2, nil))
		entries, err := os.ReadDir(sub)
		assert.Nil(t, err)
		assert.Equal(t, len(entries), 2)
		kept := map[string]bool{}
		for _, e := range entries {
			kept[e.Name()] = true
		}
		assert.True(t, kept["4000_volume.pb"])
		assert.True(t, kept["3000_volume.pb"])
		assert.False(t, kept["2000_volume.pb"])
		assert.False(t, kept["1000_volume.pb"])
	})
}
