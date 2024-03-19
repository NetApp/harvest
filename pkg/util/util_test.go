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

func TestHasDuplicates(t *testing.T) {
	type test struct {
		testCase      string
		childNames    []string
		expectedArray bool
	}

	tests := []test{
		{
			testCase:      "testcasenochild",
			childNames:    []string{},
			expectedArray: false,
		},
		{
			testCase:      "testcaseonechild",
			childNames:    []string{"type"},
			expectedArray: false,
		},
		{
			testCase:      "testcasemultipleChild",
			childNames:    []string{"type", "style", "aggr"},
			expectedArray: false,
		},
		{
			testCase:      "testcasearray",
			childNames:    []string{"aggr-name", "aggr-name", "aggr-name", "aggr-name"},
			expectedArray: true,
		},
	}

	for _, testcase := range tests {
		isArray := HasDuplicates(testcase.childNames)
		if isArray != testcase.expectedArray {
			t.Errorf("array hasn't been detected properly for %s, isArray is = %v, isArray should = %v", testcase.testCase, isArray, testcase.expectedArray)
		}
	}
}

func TestVersionAtLeast(t *testing.T) {

	testCases := []struct {
		clusterVersion string
		minVersion     string
		expected       bool
	}{
		{"9.11.1", "9.11.1", true},  // equal versions
		{"9.11.1", "9.10.1", true},  // cluster version higher
		{"9.10.1", "9.11.1", false}, // cluster version lower
		{"9.11.1", "bad", false},    // bad minimum version
		{"bad", "9.11.1", false},    // bad cluster version
	}

	for _, tc := range testCases {
		result, err := VersionAtLeast(tc.clusterVersion, tc.minVersion)
		if err != nil && tc.expected {
			t.Errorf("versionAtLeast(%q, %q) returned error: %v", tc.clusterVersion, tc.minVersion, err)
		}
		if result != tc.expected {
			t.Errorf("versionAtLeast(%q, %q) = %v; want %v", tc.clusterVersion, tc.minVersion, result, tc.expected)
		}
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
			t.Errorf("HandleArrayFormat expected: %s, got :%s", tc.expected, got)
		}
	}
}
