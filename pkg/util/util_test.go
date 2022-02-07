package util

import (
	"testing"
)

func TestIntersection(t *testing.T) {

	type test struct {
		name        string
		a           []string
		b           []string
		matchLength int
		missLength  int
	}

	tests := []test{
		{
			name:        "test1",
			a:           []string{"a", "b", "c"},
			b:           []string{"a", "c"},
			matchLength: 2,
			missLength:  0,
		},
		{
			name:        "test2",
			a:           []string{"a", "b", "c"},
			b:           []string{"a", "e", "d"},
			matchLength: 1,
			missLength:  2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, miss := Intersection(tt.a, tt.b)
			if len(match) != tt.matchLength {
				t.Errorf("Intersection() match length = %v, want %v", len(match), tt.matchLength)
			}
			if len(miss) != tt.missLength {
				t.Errorf("Intersection() miss length = %v, want %v", len(miss), tt.missLength)
			}
		})
	}
}

func TestParseMetricType(t *testing.T) {

	type test struct {
		metricName         string
		expectedName       string
		expectedMetricType string
	}

	tests := []test{
		{
			metricName:         "last_transfer_duration(duration)",
			expectedName:       "last_transfer_duration",
			expectedMetricType: "duration",
		},
		{
			metricName:         "newest_snapshot_timestamp(timestamp)",
			expectedName:       "newest_snapshot_timestamp",
			expectedMetricType: "timestamp",
		},
		{
			metricName:         "resync_successful_count",
			expectedName:       "resync_successful_count",
			expectedMetricType: "",
		},
		{
			metricName:         "total_transfer_bytes()",
			expectedName:       "total_transfer_bytes",
			expectedMetricType: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.metricName, func(t *testing.T) {
			name, metricType := ParseMetricType(tt.metricName)
			if name != tt.expectedName {
				t.Errorf("metricname not matching, actual name = %v, want %v", name, tt.expectedName)
			}
			if metricType != tt.expectedMetricType {
				t.Errorf("metrictype not matching, actual metric type = %v, want %v", metricType, tt.expectedMetricType)
			}
		})
	}
}
