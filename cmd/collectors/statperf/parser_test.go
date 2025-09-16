package statperf

import (
	"github.com/netapp/harvest/v2/assert"
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
		assert.Nil(t, err)

		sp := &StatPerf{}
		counters, err := sp.ParseCounters(string(content))
		assert.Nil(t, err)
		assert.Equal(t, len(counters), tc.expectedCount)

		cp, exists := counters[tc.expectedCounter.Counter]
		assert.True(t, exists)

		assert.Equal(t, cp.BaseCounter, tc.expectedCounter.BaseCounter)
		assert.Equal(t, cp.Properties, tc.expectedCounter.Properties)
		assert.Equal(t, cp.Type, tc.expectedCounter.Type)
		assert.Equal(t, cp.Deprecated, tc.expectedCounter.Deprecated)
		assert.Equal(t, cp.ReplacedBy, tc.expectedCounter.ReplacedBy)
		assert.Equal(t, cp.LabelCount, tc.expectedCounter.LabelCount)
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
		{
			name:     "space",
			fileName: "testdata/space/counters.txt",
			expectedCounter: CounterProperty{
				Counter:     "write_throughput",
				BaseCounter: "-",
				Properties:  "rate",
				Type:        "-",
				Deprecated:  "false",
				ReplacedBy:  "-",
				LabelCount:  0,
			},
			expectedCount: 74,
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
			assert.Nil(t, err)
			sp := &StatPerf{}
			instances, _ := sp.parseInstances(string(content))

			assert.Equal(t, len(instances), 6)
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
			assert.Nil(t, err)

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
			assert.Nil(t, err)

			assert.True(t, resp.IsArray())
			assert.True(t, len(resp.Array()) >= tc.expectedMinGroupNum)

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

			assert.True(t, foundInstance)
			assert.True(t, foundMetric)
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
			assert.True(t, slices.Equal(result, tc.expected))
		})
	}
}
