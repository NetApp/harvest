package statperf

import (
	"github.com/netapp/harvest/v2/assert"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"log/slog"
	"os"
	"slices"
	"strings"
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
				Rest: &rest2.Rest{AbstractCollector: &collector.AbstractCollector{}},
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

func TestParseRows_MultiLineContinuations(t *testing.T) {
	tests := []struct {
		name                  string
		input                 string
		expectedGroups        int
		expectedCounter       string
		expectedCounterValue  string
		expectedInstanceValue string
	}{
		{
			name: "single_token_value_continuation",
			input: `Object: volume
Instance: test_volume
Start-time: 11/7/2025 10:00:00
End-time: 11/7/2025 10:00:00
Scope: cluster-01

    Counter                                                     Value
    -------------------------------- --------------------------------
    instance_name                     trident_pvc_12d750da_fd62_4f3a_
                                                    a711_70688b3844cb
    read_data                                                   104KB
3 entries were displayed.`,
			expectedGroups:        1,
			expectedCounter:       "instance_name",
			expectedInstanceValue: "trident_pvc_12d750da_fd62_4f3a_a711_70688b3844cb",
		},
		{
			name: "multi_token_value_continuation",
			input: `Object: nvm_mirror
Instance: NVM Mirror
Start-time: 11/7/2025 12:16:52
End-time: 11/7/2025 12:16:52
Scope: sa-tme-flexpod-a800-rdma-01

    Counter                                                     Value
    -------------------------------- --------------------------------
    instance_name                                          NVM Mirror
    instance_uuid                        sa-tme-flexpod-a800-rdma-01:
                                                    kernel:NVM Mirror
    node_name                             sa-tme-flexpod-a800-rdma-01
4 entries were displayed.`,
			expectedGroups:        1,
			expectedCounter:       "instance_uuid",
			expectedInstanceValue: "sa-tme-flexpod-a800-rdma-01:kernel:NVM Mirror",
		},
		{
			name: "single_token_counter_continuation",
			input: `Object: volume
Instance: test_volume
Start-time: 11/7/2025 10:00:00
End-time: 11/7/2025 10:00:00
Scope: cluster-01

    Counter                                                     Value
    -------------------------------- --------------------------------
    instance_name                                        test_volume
    wvblk_saved_fsinfo_private_inos_                             1638
    reserve
3 entries were displayed.`,
			expectedGroups:       1,
			expectedCounter:      "wvblk_saved_fsinfo_private_inos_reserve",
			expectedCounterValue: "1638",
		},
		{
			name: "multiple_instances_with_continuations",
			input: `Object: nvm_mirror
Instance: NVM Mirror
Start-time: 11/7/2025 12:16:52
End-time: 11/7/2025 12:16:52
Scope: sa-tme-flexpod-a800-rdma-01

    Counter                                                     Value
    -------------------------------- --------------------------------
    instance_name                                          NVM Mirror
    instance_uuid                        sa-tme-flexpod-a800-rdma-01:
                                                    kernel:NVM Mirror
    write_throughput                                   13847639326958

Object: nvm_mirror
Instance: NVM Mirror
Start-time: 11/7/2025 12:16:52
End-time: 11/7/2025 12:16:52
Scope: sa-tme-flexpod-a800-rdma-02

    Counter                                                     Value
    -------------------------------- --------------------------------
    instance_name                                          NVM Mirror
    instance_uuid                        sa-tme-flexpod-a800-rdma-02:
                                                    kernel:NVM Mirror
    write_throughput                                     629627001678
8 entries were displayed.`,
			expectedGroups: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &StatPerf{
				Rest: &rest2.Rest{AbstractCollector: &collector.AbstractCollector{}},
			}
			s.Logger = slog.Default()
			s.perfProp = &perfProp{
				counterInfo: map[string]*counter{
					"instance_name":    {name: "instance_name"},
					"instance_uuid":    {name: "instance_uuid"},
					"node_name":        {name: "node_name"},
					"write_throughput": {name: "write_throughput"},
					"read_data":        {name: "read_data"},
					"wvblk_saved_fsinfo_private_inos_reserve": {name: "wvblk_saved_fsinfo_private_inos_reserve"},
				},
			}

			groups, err := s.parseRows(tc.input)
			assert.Nil(t, err)
			assert.Equal(t, len(groups), tc.expectedGroups)

			if tc.expectedCounter != "" {
				found := false
				for _, group := range groups {
					val, ok := group[tc.expectedCounter]
					if !ok {
						continue
					}
					found = true
					valStr, _ := val.(string)
					if tc.expectedCounterValue != "" {
						assert.Equal(t, valStr, tc.expectedCounterValue)
					}
					if tc.expectedInstanceValue != "" {
						assert.Equal(t, valStr, tc.expectedInstanceValue)
					}
					break
				}
				assert.True(t, found)
			}
		})
	}
}

func TestParseRows_BothCounterAndValueContinuation(t *testing.T) {
	// Test case where a continuation line has content in both columns
	input := `Object: volume
Instance: test_volume
Start-time: 11/7/2025 10:00:00
End-time: 11/7/2025 10:00:00
Scope: cluster-01

    Counter                                                     Value
    -------------------------------- --------------------------------
    very_long_counter_name_that_sp                  some_very_long_v
    ans_multiple_lines                              alue_text_here
    normal_counter                                              12345
3 entries were displayed.`

	s := &StatPerf{
		Rest: &rest2.Rest{AbstractCollector: &collector.AbstractCollector{}},
	}
	s.Logger = slog.Default()
	s.perfProp = &perfProp{
		counterInfo: map[string]*counter{
			"very_long_counter_name_that_spans_multiple_lines": {name: "very_long_counter_name_that_spans_multiple_lines"},
			"normal_counter": {name: "normal_counter"},
		},
	}

	groups, err := s.parseRows(input)
	assert.Nil(t, err)
	assert.Equal(t, len(groups), 1)

	val1, ok1 := groups[0]["very_long_counter_name_that_spans_multiple_lines"]
	assert.True(t, ok1)
	assert.Equal(t, val1, "some_very_long_value_text_here")

	val2, ok2 := groups[0]["normal_counter"]
	assert.True(t, ok2)
	assert.Equal(t, val2, "12345")
}

func TestParseRows_WrappedArrayCounter(t *testing.T) {
	input := `Object: workload_queue_combined
Instance: fg-wid26474
Start-time: 10/21/2026 15:33:33
End-time: 10/21/2026 15:33:33
Scope: cluster_1
Number of Constituents: 2 (complete_aggregation)
    Counter                                                     Value
    -------------------------------- --------------------------------
    cpu_dblade_background_service_                                  -
    time
                              direct                             1397
                             wafl_cp                                0
                   background_delete                                0
                  readahead_x_stripe                                0
    disk_ssd_background_service_time                                -
                              direct                                5
                             wafl_cp                                0
                   background_delete                                0
                  readahead_x_stripe                                0
    iops_total                                              184097303
3 entries were displayed.`

	s := &StatPerf{
		Rest: &rest2.Rest{AbstractCollector: &collector.AbstractCollector{}},
	}
	s.Logger = slog.Default()
	s.perfProp = &perfProp{
		counterInfo: map[string]*counter{
			"cpu_dblade_background_service_time": {name: "cpu_dblade_background_service_time", counterType: "array", labelCount: 4},
			"disk_ssd_background_service_time":   {name: "disk_ssd_background_service_time", counterType: "array", labelCount: 4},
			"iops_total":                         {name: "iops_total"},
		},
	}

	groups, err := s.parseRows(input)
	assert.Nil(t, err)
	assert.Equal(t, len(groups), 1)

	want := map[string]map[string]string{
		"cpu_dblade_background_service_time": {"direct": "1397", "wafl_cp": "0", "background_delete": "0", "readahead_x_stripe": "0"},
		"disk_ssd_background_service_time":   {"direct": "5", "wafl_cp": "0", "background_delete": "0", "readahead_x_stripe": "0"},
	}
	for name, labels := range want {
		got, ok := groups[0][name].(map[string]string)
		assert.True(t, ok)
		for label, value := range labels {
			assert.Equal(t, got[label], value)
		}
	}
	assert.Equal(t, groups[0]["iops_total"], "184097303")
	_, stray := groups[0]["direct"]
	assert.False(t, stray)
}

func TestParseRows_WrappedArrayCounterShortLabels(t *testing.T) {
	input := `Object: wafl
Instance: wafl
Start-time: 10/21/2026 17:44:02
End-time: 10/21/2026 17:44:02
Scope: panaffa400x05
    Counter                                                     Value
    -------------------------------- --------------------------------
    cps_multi_clean_duration_                                       -
    histogram
                                   0                                0
                                   1                                4
                                   2                               24
                                   3                               18
                                   4                                2
                                   5                                0
                                   6                                0
                                   7                                0
                                 <16                                0
                                 <24                                0
                                 <32                                0
                                 <40                                0
                                 <48                                0
                                 <56                                0
                                 <64                                0
                                <128                                0
                                <192                                0
                                <256                                0
                                <320                                0
                                <384                                0
                                <448                                0
                                <512                                0
                               <1024                                0
                               <1536                                0
                               <2048                                0
                               <2560                                0
                               <3072                             1317
                               <3584                              808
                               <4096                             2132
                               <8192                                0
                              <12288                              473
                              <16384                              945
                              <20480                                1
                              <24576                                0
                              <28672                                0
                              <32768                                0
                              <65536                                0
                              <98304                                0
                             <131072                                0
                             <163840                                0
                             <196608                                0
                             <229376                                0
                             <262144                                0
                             >262144                                0`

	s := &StatPerf{
		Rest: &rest2.Rest{AbstractCollector: &collector.AbstractCollector{}},
	}
	s.Logger = slog.Default()
	s.perfProp = &perfProp{
		counterInfo: map[string]*counter{
			"cps_multi_clean_duration_histogram": {name: "cps_multi_clean_duration_histogram", counterType: "array", labelCount: 44},
		},
	}

	groups, err := s.parseRows(input)
	assert.Nil(t, err)
	assert.Equal(t, len(groups), 1)

	got, ok := groups[0]["cps_multi_clean_duration_histogram"].(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, len(got), 44)
	want := map[string]string{"0": "0", "1": "4", "2": "24", "<512": "0", "<3072": "1317", ">262144": "0"}
	for label, value := range want {
		assert.Equal(t, got[label], value)
	}
}

func TestParseRows_EntriesDisplayedLine(t *testing.T) {
	input := `Object: volume
Instance: test_volume
Start-time: 11/7/2025 10:00:00
End-time: 11/7/2025 10:00:00
Scope: cluster-01

    Counter                                                     Value
    -------------------------------- --------------------------------
    instance_name                                        test_volume
    read_data                                                   104KB
2 entries were displayed.`

	s := &StatPerf{
		Rest: &rest2.Rest{AbstractCollector: &collector.AbstractCollector{}},
	}
	s.Logger = slog.Default()
	s.perfProp = &perfProp{
		counterInfo: map[string]*counter{
			"instance_name": {name: "instance_name"},
			"read_data":     {name: "read_data"},
		},
	}

	groups, err := s.parseRows(input)
	assert.Nil(t, err)
	assert.Equal(t, len(groups), 1)

	val, ok := groups[0]["read_data"]
	assert.True(t, ok)
	assert.Equal(t, val, "104KB")
	valStr, isString := val.(string)
	assert.True(t, isString)
	assert.False(t, strings.Contains(valStr, "entries"))
}
