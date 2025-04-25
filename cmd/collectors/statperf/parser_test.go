package statperf

import (
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"log/slog"
	"os"
	"slices"
	"testing"
)

type testCase struct {
	name                string
	fileName            string
	expectedInstance    string
	expectedMetric      string
	expectedMinGroupNum int
}

func TestParseCounters(t *testing.T) {

	testCases := []testCase{
		{
			name:     "admin_login",
			fileName: "testdata/counters.txt",
		}, {
			name:     "user_login",
			fileName: "testdata/counters_1.txt",
		},
		{
			name:     "user_login_permission_changed",
			fileName: "testdata/counters_2.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(tc.fileName)
			if err != nil {
				t.Fatalf("Failed to read from file: %v", err)
			}

			sp := &StatPerf{}
			counters, err := sp.parseCounters(string(content))
			if err != nil {
				t.Fatalf("Unexpected error during parseCounters: %v", err)
			}
			if len(counters) != 32 {
				t.Errorf("Expected 32 counter, got %d", len(counters))
			}

			cp, exists := counters["nix_skipped_reason_offline"]
			if !exists {
				t.Fatalf("Expected to find counter 'nix_skipped_reason_offline'")
			}

			if cp.BaseCounter != "-" {
				t.Errorf("Expected BaseCounter '-', got '%s'", cp.BaseCounter)
			}

			if cp.Properties != "delta,no-zero-values" {
				t.Errorf("Expected Properties 'delta,no-zero-values', got '%s'", cp.Properties)
			}

			if cp.Type != "-" {
				t.Errorf("Expected Type '-', got '%s'", cp.Type)
			}
			if cp.Deprecated != "false" {
				t.Errorf("Expected Deprecated 'false', got '%s'", cp.Deprecated)
			}
			if cp.ReplacedBy != "-" {
				t.Errorf("Expected ReplacedBy '-', got '%s'", cp.ReplacedBy)
			}
		})
	}
}

func TestParseInstances(t *testing.T) {

	testCases := []testCase{
		{
			name:     "admin_login",
			fileName: "testdata/instances.txt",
		}, {
			name:     "user_login",
			fileName: "testdata/instances_1.txt",
		},
		{
			name:     "user_login_permission_changed",
			fileName: "testdata/instances_2.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(tc.fileName)
			if err != nil {
				t.Fatalf("Failed to read from file: %v", err)
			}
			sp := &StatPerf{}
			instances, _ := sp.parseInstances(string(content))

			if len(instances) != 6 {
				t.Errorf("Expected 6 instances, got %d", len(instances))
			}
		})
	}
}

func TestParseData(t *testing.T) {
	testCases := []testCase{
		{
			name:                "admin_login",
			fileName:            "testdata/data.txt",
			expectedInstance:    "hs_1",
			expectedMetric:      "evict_rw_cache_skipped_reason_disconnected",
			expectedMinGroupNum: 6,
		},
		{
			name:                "user_login",
			fileName:            "testdata/data_1.txt",
			expectedInstance:    "hs_1",
			expectedMetric:      "evict_rw_cache_skipped_reason_disconnected",
			expectedMinGroupNum: 6,
		},
		{
			name:                "user_login_permission_changed",
			fileName:            "testdata/data_2.txt",
			expectedInstance:    "hs_1",
			expectedMetric:      "evict_rw_cache_skipped_reason_disconnected",
			expectedMinGroupNum: 6,
		},
		{
			name:                "multiline",
			fileName:            "testdata/multiline_data.txt",
			expectedInstance:    "trident_pvc_12d750da_fd62_4f3a_a711_70688b3844cb",
			expectedMetric:      "wvblk_saved_fsinfo_private_inos_reserve",
			expectedMinGroupNum: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(tc.fileName)
			if err != nil {
				t.Fatalf("Failed to read from file %s: %v", tc.fileName, err)
			}

			sp := &StatPerf{
				Rest: &rest2.Rest{
					AbstractCollector: &collector.AbstractCollector{},
				},
			}
			sp.Logger = slog.Default()

			resp, err := sp.parseData(string(content))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !resp.IsArray() {
				t.Fatalf("Expected JSON result to be an array, got: %v", resp.Type)
			}

			groupsCount := len(resp.Array())
			if groupsCount < tc.expectedMinGroupNum {
				t.Errorf("Expected at least %d groups, got %d", tc.expectedMinGroupNum, groupsCount)
			}

			foundInstance := false
			foundMetric := false
			for _, group := range resp.Array() {
				if group.Get("instance_name").String() == tc.expectedInstance {
					foundInstance = true
				}
				if group.Get(tc.expectedMetric).String() != "" {
					foundMetric = true
				}
			}

			if !foundInstance {
				t.Errorf("Expected to find a group with instance_name '%s'", tc.expectedInstance)
			}
			if !foundMetric {
				t.Errorf("Expected to find a group with metric '%s'", tc.expectedMetric)
			}
		})
	}
}

func TestFilterNonEmpty(t *testing.T) {
	// Define test cases.
	tests := []struct {
		name     string
		input    string
		expected []string
		hasError bool
	}{
		{
			name: "Simple input with empty lines",
			input: `line one

line two
    `,
			expected: []string{"line one", "line two"},
			hasError: false,
		},
		{
			name: "Input with all empty lines",
			input: `    


			`,
			expected: []string{},
			hasError: false,
		},
		{
			name: "Input with no newline at the end",
			input: `line one
line two`,
			expected: []string{"line one", "line two"},
			hasError: false,
		},
		{
			name: "Normal input",
			input: `object·instance·counter·value·
Object·Instance·Counter·Text Value·
flexcache_per_volume·Test·blocks_requested_from_client·637069129383·`,
			expected: []string{
				"object·instance·counter·value·",
				"Object·Instance·Counter·Text Value·",
				"flexcache_per_volume·Test·blocks_requested_from_client·637069129383·",
			},
			hasError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := filterNonEmpty(tc.input)
			if !slices.Equal(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
