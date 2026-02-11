package template

import "testing"

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

func TestHandleArrayFormat(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
	}{
		{"ha.partners.0.name", "ha.partners"},                               // with .0
		{"ha.partners.#.name", "ha.partners"},                               // with .#
		{"aggregates.#.name", "aggregates"},                                 // with .#
		{"cloud_storage.stores.#.cloud_store.name", "cloud_storage.stores"}, // with .#
		{"abc.o.1.xyz", "abc.o"},                                            // with .o.1
		{"abc.xyz", "abc.xyz"},                                              // with .xyz
		{"interfaces.#.ip.address", "interfaces"},                           // with .ip.address
	}

	for _, tc := range testCases {
		got := HandleArrayFormat(tc.name)
		if got != tc.expected {
			t.Errorf("HandleArrayFormat got :%s want: %s", got, tc.expected)
		}
	}
}
