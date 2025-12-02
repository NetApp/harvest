package prometheus

import (
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/assert"
	"strings"
	"testing"
)

func Test_memcache_streamMetrics(t *testing.T) {
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
	m := memCache{}
	seen := make(map[string]struct{})
	var w strings.Builder
	_ = m.writeMetrics(&w, example, seen)

	diff := cmp.Diff(w.String(), expected)
	assert.Equal(t, diff, "")
}
