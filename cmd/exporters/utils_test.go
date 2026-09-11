package exporters

import (
	"log/slog"
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
