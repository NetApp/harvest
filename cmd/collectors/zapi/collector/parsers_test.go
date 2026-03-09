package zapi

import (
	"testing"

	"github.com/netapp/harvest/v2/pkg/matrix"
)

func TestParseShortestPath(t *testing.T) {
	tests := []struct {
		name     string
		metrics  []string
		labels   map[string]string
		expected []string
	}{
		{
			name:     "single flat key",
			labels:   map[string]string{"cluster-uuid": "uuid"},
			expected: []string{"cluster-uuid"},
		},
		{
			name:     "single nested key",
			labels:   map[string]string{"cluster-identity-info.cluster-uuid": "uuid"},
			expected: []string{"cluster-identity-info"},
		},
		{
			name: "two sibling nested keys",
			labels: map[string]string{
				"cluster-identity-info.cluster-uuid": "uuid",
				"cluster-identity-info.cluster-name": "name",
			},
			expected: []string{"cluster-identity-info"},
		},
		{
			name: "two flat keys with different names",
			labels: map[string]string{
				"cluster-uuid": "uuid",
				"cluster-name": "name",
			},
			expected: nil,
		},
		{
			name: "depth 3 two sibling keys",
			labels: map[string]string{
				"grandparent.parent.leaf-a": "a",
				"grandparent.parent.leaf-b": "b",
			},
			expected: []string{"grandparent", "parent"},
		},
		{
			name:     "depth 3 single key",
			labels:   map[string]string{"grandparent.parent.leaf": "val"},
			expected: []string{"grandparent", "parent"},
		},
		{
			name:     "mixed depth keys with common container",
			metrics:  []string{"container.metric-a"},
			labels:   map[string]string{"container.label-b": "b"},
			expected: []string{"container"},
		},
		{
			name:     "mixed depth keys with no common container",
			metrics:  []string{"info-a.metric"},
			labels:   map[string]string{"info-b.label": "val"},
			expected: nil,
		},
		{
			name: "keys at different depths with common prefix",
			labels: map[string]string{
				"container.sub.deep-leaf": "a",
				"container.shallow-leaf":  "b",
			},
			expected: []string{"container"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := matrix.New("test", "test", "test")
			for _, key := range tt.metrics {
				if _, err := m.NewMetricUint64(key); err != nil {
					t.Fatalf("failed to create metric %q: %v", key, err)
				}
			}
			got := ParseShortestPath(m, tt.labels)
			if !slicesEqual(got, tt.expected) {
				t.Errorf("got = %v, want %v", got, tt.expected)
			}
		})
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
