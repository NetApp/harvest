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
	expectedCounter     CounterProperty
	expectedCount       int
	expectedInstance    string
	expectedMetric      string
	expectedMinGroupNum int
}

func runCounterTest(t *testing.T, tc testCase) {
	t.Run(tc.name, func(t *testing.T) {
		content, err := os.ReadFile(tc.fileName)
		if err != nil {
			t.Fatalf("Failed to read from file: %v", err)
		}

		sp := &StatPerf{}
		counters, err := sp.ParseCounters(string(content))
		if err != nil {
			t.Fatalf("Unexpected error during parseCounters: %v", err)
		}
		if len(counters) != tc.expectedCount {
			t.Errorf("Expected %d counter, got %d", tc.expectedCount, len(counters))
		}

		cp, exists := counters[tc.expectedCounter.Counter]
		if !exists {
			t.Fatalf("Expected to find counter '%s'", tc.expectedCounter.Counter)
		}

		if cp.BaseCounter != tc.expectedCounter.BaseCounter {
			t.Errorf("Expected BaseCounter '%s', got '%s'", tc.expectedCounter.BaseCounter, cp.BaseCounter)
		}

		if cp.Properties != tc.expectedCounter.Properties {
			t.Errorf("Expected Properties '%s', got '%s'", tc.expectedCounter.Properties, cp.Properties)
		}

		if cp.Type != tc.expectedCounter.Type {
			t.Errorf("Expected Type '%s', got '%s'", tc.expectedCounter.Type, cp.Type)
		}

		if cp.Deprecated != tc.expectedCounter.Deprecated {
			t.Errorf("Expected Deprecated '%s', got '%s'", tc.expectedCounter.Deprecated, cp.Deprecated)
		}

		if cp.ReplacedBy != tc.expectedCounter.ReplacedBy {
			t.Errorf("Expected ReplacedBy '%s', got '%s'", tc.expectedCounter.ReplacedBy, cp.ReplacedBy)
		}

		if cp.LabelCount != tc.expectedCounter.LabelCount {
			t.Errorf("Expected LabelCount %d, got %d", tc.expectedCounter.LabelCount, cp.LabelCount)
		}
	})
}

func TestParseCounters(t *testing.T) {
	testCases := []testCase{
		{
			name:     "admin_login",
			fileName: "testdata/counters.txt",
			expectedCounter: CounterProperty{
				Counter:     "nix_skipped_reason_offline",
				BaseCounter: "-",
				Properties:  "delta,no-zero-values",
				Type:        "-",
				Deprecated:  "false",
				ReplacedBy:  "-",
				LabelCount:  0,
			},
			expectedCount: 29,
		},
		{
			name:     "user_login",
			fileName: "testdata/counters_1.txt",
			expectedCounter: CounterProperty{
				Counter:     "nix_skipped_reason_offline",
				BaseCounter: "-",
				Properties:  "delta,no-zero-values",
				Type:        "-",
				Deprecated:  "false",
				ReplacedBy:  "-",
				LabelCount:  0,
			},
			expectedCount: 29,
		},
		{
			name:     "user_login_permission_changed",
			fileName: "testdata/counters_2.txt",
			expectedCounter: CounterProperty{
				Counter:     "nix_skipped_reason_offline",
				BaseCounter: "-",
				Properties:  "delta,no-zero-values",
				Type:        "-",
				Deprecated:  "false",
				ReplacedBy:  "-",
				LabelCount:  0,
			},
			expectedCount: 29,
		},
		{
			name:     "array",
			fileName: "testdata/array/lun_counters.txt",
			expectedCounter: CounterProperty{
				Counter:     "read_align_histo",
				BaseCounter: "read_ops_sent",
				Properties:  "percent",
				Type:        "array",
				Deprecated:  "false",
				ReplacedBy:  "-",
				LabelCount:  8,
			},
			expectedCount: 404,
		},
		{
			name:     "array_1",
			fileName: "testdata/array/nic_common_counters.txt",
			expectedCounter: CounterProperty{
				Counter:     "rss_matrix",
				BaseCounter: "-",
				Properties:  "delta",
				Type:        "array",
				Deprecated:  "false",
				ReplacedBy:  "-",
				LabelCount:  320,
			},
			expectedCount: 91,
		},
		{
			name:     "array_2",
			fileName: "testdata/array/nic_common_counters.txt",
			expectedCounter: CounterProperty{
				Counter:     "rss_cg_stat",
				BaseCounter: "-",
				Properties:  "raw,no-zero-values",
				Type:        "array",
				Deprecated:  "false",
				ReplacedBy:  "-",
				LabelCount:  64,
			},
			expectedCount: 91,
		},
		{
			name:     "array_3",
			fileName: "testdata/array/wafl_counters.txt",
			expectedCounter: CounterProperty{
				Counter:     "cp_count",
				BaseCounter: "-",
				Properties:  "delta",
				Type:        "array",
				Deprecated:  "false",
				ReplacedBy:  "-",
				LabelCount:  15,
			},
			expectedCount: 553,
		},
	}

	for _, tc := range testCases {
		runCounterTest(t, tc)
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
		{
			name:                "array",
			fileName:            "testdata/array/lun_data.txt",
			expectedInstance:    "/vol/osc_iscsi_vol01/osc_iscsi_vol01",
			expectedMetric:      "read_align_histo",
			expectedMinGroupNum: 1,
		},
		{
			name:                "array",
			fileName:            "testdata/array/wafl_data.txt",
			expectedInstance:    "umeng-aff300-02:kernel:wafl",
			expectedMetric:      "cp_count",
			expectedMinGroupNum: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content, err := os.ReadFile(tc.fileName)
			if err != nil {
				t.Fatalf("Failed to read from file %s: %v", tc.fileName, err)
			}

			s := &StatPerf{
				Rest: &rest2.Rest{
					AbstractCollector: &collector.AbstractCollector{},
				},
			}
			s.Logger = slog.Default()
			if tc.name == "array" {
				s.perfProp = initializePerfPropForArray()
			} else {
				s.perfProp = &perfProp{}
			}
			resp, err := s.parseData(string(content))
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
				if group.Get("instance_name").String() == tc.expectedInstance || group.Get("instance_uuid").String() == tc.expectedInstance {
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

func initializePerfPropForArray() *perfProp {
	return &perfProp{
		counterInfo: map[string]*counter{
			"read_align_histo": {
				name:        "read_align_histo",
				counterType: "array",
				labelCount:  8,
			},
			"write_align_histo": {
				name:        "write_align_histo",
				counterType: "array",
				labelCount:  8,
			},
			"cp_count": {
				name:        "cp_count",
				counterType: "array",
				labelCount:  15,
			},
			"cp_phase_times": {
				name:        "cp_phase_times",
				counterType: "array",
				labelCount:  49,
			},
			"read_io_type": {
				name:        "read_io_type",
				counterType: "array",
				labelCount:  11,
			},
		},
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
			result := FilterNonEmpty(tc.input)
			if !slices.Equal(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
