package exporters

import (
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

// newLabeledMatrix returns a matrix with one instance carrying 7 labels,
// gathered via the include_all_labels map-iteration path, so their order is
// randomized per call unless sortLabels is requested.
func newLabeledMatrix(t *testing.T) *matrix.Matrix {
	t.Helper()
	m := matrix.New("sort_labels_test", "sort_labels_test", "sort_labels_test")

	metric, err := m.NewMetricUint64("value")
	assert.Nil(t, err)
	instance, err := m.NewInstance("A")
	assert.Nil(t, err)
	metric.SetValueInt64(instance, 42)

	instance.SetLabel("aaa", "1")
	instance.SetLabel("bbb", "2")
	instance.SetLabel("ccc", "3")
	instance.SetLabel("ddd", "4")
	instance.SetLabel("eee", "5")
	instance.SetLabel("fff", "6")
	instance.SetLabel("ggg", "7")

	return m
}

func renderOnce(t *testing.T, m *matrix.Matrix, sortLabels bool) string {
	t.Helper()
	rendered, _, _ := Render(m, false, sortLabels, "netapp_", slog.Default(), "")

	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, string(line))
	}
	return strings.Join(lines, "\n")
}

// TestRenderSortedIsDeterministic verifies that repeated renders with
// sortLabels enabled produce byte-identical output.
func TestRenderSortedIsDeterministic(t *testing.T) {
	m := newLabeledMatrix(t)

	first := renderOnce(t, m, true)
	assert.True(t, first != "")

	for range 50 {
		got := renderOnce(t, m, true)
		assert.Equal(t, got, first)
	}
}

// TestRenderUnsortedIsNotDeterministic verifies that without sortLabels,
// repeated renders disagree, since label order follows Go's randomized map
// iteration. Go randomizes the starting position and walks from there, so
// the orderings are rotations rather than arbitrary permutations — but even
// a handful of reachable orderings makes 50 identical runs effectively
// impossible.
func TestRenderUnsortedIsNotDeterministic(t *testing.T) {
	m := newLabeledMatrix(t)

	seen := make(map[string]struct{})
	for range 50 {
		seen[renderOnce(t, m, false)] = struct{}{}
	}

	assert.True(t, len(seen) > 1)
}

// newRenderMatrix returns a matrix that exercises every kind of rendered line:
// labels, scalar, labeled array, normalized histogram and raw histogram.
func newRenderMatrix(t *testing.T) *matrix.Matrix {
	t.Helper()
	m := matrix.New("disk", "disk", "disk")
	m.SetGlobalLabel("datacenter", "dc1")
	m.SetExportOptions(matrix.NewExportOptionsWithLabels([]string{"node", "disk"}, []string{"model"}))

	readOps, err := m.NewMetricUint64("read_ops")
	assert.Nil(t, err)
	opLatency, err := m.NewMetricUint64("op_latency.read", "op_latency")
	assert.Nil(t, err)
	opLatency.SetLabel("op", "read")

	addHistogram := func(name string, buckets []string) []*matrix.Metric {
		bucket, err := m.NewMetricUint64(name+".bucket", name)
		assert.Nil(t, err)
		bucket.SetExportable(false)
		bucket.SetBuckets(&buckets)
		metrics := make([]*matrix.Metric, 0, len(buckets))
		for i := range buckets {
			h, err := m.NewMetricUint64(name+"."+strconv.Itoa(i), name)
			assert.Nil(t, err)
			h.SetHistogram(true)
			h.SetLabel("bucket", name+".bucket")
			h.SetLabel("comment", strconv.Itoa(i))
			metrics = append(metrics, h)
		}
		return metrics
	}
	latencyHist := addHistogram("latency_hist", []string{"<1ms", "<2ms", ">2ms"})
	alignHist := addHistogram("align_hist", []string{"0", "1"})

	for i, name := range []string{"A", "B"} {
		instance, err := m.NewInstance(name)
		assert.Nil(t, err)
		instance.SetLabel("node", "node-"+name)
		instance.SetLabel("disk", "disk-"+name)
		instance.SetLabel("model", `X"1`)
		readOps.SetValueInt64(instance, int64(10+i))
		opLatency.SetValueInt64(instance, int64(20+i))
		for j, h := range latencyHist {
			h.SetValueInt64(instance, int64(j+1))
		}
		for j, h := range alignHist {
			h.SetValueInt64(instance, int64(100+j))
		}
	}
	return m
}

const wantRenderGolden = `# HELP netapp_disk_align_hist Metric for disk
# HELP netapp_disk_labels Pseudo-metric for disk labels
# HELP netapp_disk_latency_hist Metric for disk
# HELP netapp_disk_op_latency Metric for disk
# HELP netapp_disk_read_ops Metric for disk
# TYPE netapp_disk_align_hist histogram
# TYPE netapp_disk_labels gauge
# TYPE netapp_disk_latency_hist histogram
# TYPE netapp_disk_op_latency gauge
# TYPE netapp_disk_read_ops gauge
netapp_disk_align_hist{datacenter="dc1",disk="disk-A",node="node-A",metric="0"} 100 1700000000000
netapp_disk_align_hist{datacenter="dc1",disk="disk-A",node="node-A",metric="1"} 101 1700000000000
netapp_disk_align_hist{datacenter="dc1",disk="disk-B",node="node-B",metric="0"} 100 1700000000000
netapp_disk_align_hist{datacenter="dc1",disk="disk-B",node="node-B",metric="1"} 101 1700000000000
netapp_disk_labels{datacenter="dc1",disk="disk-A",model="X\\\"1",node="node-A"} 1.0 1700000000000
netapp_disk_labels{datacenter="dc1",disk="disk-B",model="X\\\"1",node="node-B"} 1.0 1700000000000
netapp_disk_latency_hist_bucket{datacenter="dc1",disk="disk-A",node="node-A",le="+Inf"} 6 1700000000000
netapp_disk_latency_hist_bucket{datacenter="dc1",disk="disk-A",node="node-A",le="1000"} 1 1700000000000
netapp_disk_latency_hist_bucket{datacenter="dc1",disk="disk-A",node="node-A",le="2000"} 3 1700000000000
netapp_disk_latency_hist_bucket{datacenter="dc1",disk="disk-B",node="node-B",le="+Inf"} 6 1700000000000
netapp_disk_latency_hist_bucket{datacenter="dc1",disk="disk-B",node="node-B",le="1000"} 1 1700000000000
netapp_disk_latency_hist_bucket{datacenter="dc1",disk="disk-B",node="node-B",le="2000"} 3 1700000000000
netapp_disk_latency_hist_count{datacenter="dc1",disk="disk-A",node="node-A"} 6
netapp_disk_latency_hist_count{datacenter="dc1",disk="disk-B",node="node-B"} 6
netapp_disk_latency_hist_sum{datacenter="dc1",disk="disk-A",node="node-A"} 5000
netapp_disk_latency_hist_sum{datacenter="dc1",disk="disk-B",node="node-B"} 5000
netapp_disk_op_latency{datacenter="dc1",disk="disk-A",node="node-A",op="read"} 20 1700000000000
netapp_disk_op_latency{datacenter="dc1",disk="disk-B",node="node-B",op="read"} 21 1700000000000
netapp_disk_read_ops{datacenter="dc1",disk="disk-A",node="node-A"} 10 1700000000000
netapp_disk_read_ops{datacenter="dc1",disk="disk-B",node="node-B"} 11 1700000000000`

type recordingSink struct {
	grow  int
	lines []string
}

func (r *recordingSink) Grow(n int)        { r.grow += n }
func (r *recordingSink) Write(line []byte) { r.lines = append(r.lines, string(line)) }

// TestRenderGolden pins the rendered lines, including HELP and TYPE.
func TestRenderGolden(t *testing.T) {
	rendered, stats, metricNames := Render(newRenderMatrix(t), true, true, "netapp_", slog.Default(), "1700000000000")

	lines := make([]string, 0, len(rendered))
	var renderedBytes uint64
	for _, line := range rendered {
		lines = append(lines, string(line))
		renderedBytes += uint64(len(line)) + 1
	}

	// HELP and TYPE come first, and once, for every metric name.
	typeLine := make(map[string]int)
	for i, line := range lines {
		name, ok := strings.CutPrefix(line, "# TYPE ")
		if !ok {
			continue
		}
		name = strings.Fields(name)[0]
		assert.True(t, strings.HasPrefix(lines[i-1], "# HELP "+name+" "))
		_, seen := typeLine[name]
		assert.False(t, seen)
		typeLine[name] = i
	}
	for i, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		name := line[:strings.IndexAny(line, "{")]
		name = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, "_bucket"), "_count"), "_sum")
		typeAt, ok := typeLine[name]
		assert.True(t, ok)
		assert.True(t, typeAt < i)
	}

	sorted := slices.Clone(lines)
	slices.Sort(sorted)
	assert.Equal(t, strings.Join(sorted, "\n"), wantRenderGolden)
	assert.Equal(t, stats.InstancesExported, uint64(2))
	assert.Equal(t, stats.MetricsExported, uint64(len(lines)))
	assert.Equal(t, stats.RenderedBytes, renderedBytes)
	assert.Equal(t, metricNames.Size(), 4)
}

func TestRenderToMatchesRender(t *testing.T) {
	m := newRenderMatrix(t)
	for _, sortLabels := range []bool{true, false} {
		rendered, wantStats, wantNames := Render(m, true, sortLabels, "netapp_", slog.Default(), "")

		var sink recordingSink
		stats, names := RenderTo(m, true, sortLabels, "netapp_", slog.Default(), "", &sink)

		want := make([]string, 0, len(rendered))
		for _, line := range rendered {
			want = append(want, string(line))
		}
		got := slices.Clone(sink.lines)
		slices.Sort(want)
		slices.Sort(got)
		assert.Equal(t, strings.Join(got, "\n"), strings.Join(want, "\n"))
		assert.Equal(t, stats, wantStats)
		assert.Equal(t, names.Size(), wantNames.Size())
		assert.True(t, sink.grow > 0)
	}
}
