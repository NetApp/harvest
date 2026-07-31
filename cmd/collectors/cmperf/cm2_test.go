package cmperf

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors/cmperf/cmmetrics"
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
