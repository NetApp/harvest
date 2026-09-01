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

func TestSplitMetricRename(t *testing.T) {
	tests := []struct {
		rawName string
		name    string
		display string
	}{
		{rawName: "cpu => cpu_percent", name: "cpu", display: "cpu_percent"},
		// No surrounding whitespace parses the same as the spaced form, matching
		// ParseMetric. It previously fell through unparsed, yielding a metric
		// literally named "cpu=>cpu_percent".
		{rawName: "cpu=>cpu_percent", name: "cpu", display: "cpu_percent"},
		{rawName: "cpu\t=>\tcpu_percent", name: "cpu", display: "cpu_percent"},
		{rawName: "cpu  =>  cpu_percent", name: "cpu", display: "cpu_percent"},
		// No rename: the name is its own display, with "." and "-" preserved.
		{rawName: "cpu", name: "cpu", display: "cpu"},
		{rawName: "io.read-bytes", name: "io.read-bytes", display: "io.read-bytes"},
		{rawName: "", name: "", display: ""},
		// Degenerate renames fall back to the raw string rather than producing an
		// empty name or display.
		{rawName: "cpu =>", name: "cpu =>", display: "cpu =>"},
		{rawName: "=> cpu_percent", name: "=> cpu_percent", display: "=> cpu_percent"},
	}

	for _, test := range tests {
		t.Run(test.rawName, func(t *testing.T) {
			name, display := SplitMetricRename(test.rawName)
			if name != test.name || display != test.display {
				t.Fatalf("SplitMetricRename(%q) = (%q, %q), want (%q, %q)", test.rawName, name, display, test.name, test.display)
			}
		})
	}
}

// TestParseMetric locks in ParseMetric's behavior now that it shares its rename
// splitting with SplitMetricRename.
func TestParseMetric(t *testing.T) {
	tests := []struct {
		rawName    string
		name       string
		display    string
		kind       string
		metricType string
	}{
		{rawName: "cpu", name: "cpu", display: "cpu", kind: "float"},
		{rawName: "io.read-bytes", name: "io.read-bytes", display: "io_read_bytes", kind: "float"},
		{rawName: "some_long_name => short", name: "some_long_name", display: "short", kind: "float"},
		{rawName: "some_long_name=>short", name: "some_long_name", display: "short", kind: "float"},
		{
			rawName: "last_transfer_duration(duration) => ltd", name: "last_transfer_duration",
			display: "ltd", kind: "float", metricType: "duration",
		},
		{rawName: "^^name", name: "name", display: "name", kind: "key"},
		{rawName: "^^name => renamed", name: "name", display: "renamed", kind: "key"},
		{rawName: "^label", name: "label", display: "label", kind: "label"},
		{rawName: "^label => renamed", name: "label", display: "renamed", kind: "label"},
	}

	for _, test := range tests {
		t.Run(test.rawName, func(t *testing.T) {
			name, display, kind, metricType := ParseMetric(test.rawName)
			if name != test.name || display != test.display || kind != test.kind || metricType != test.metricType {
				t.Fatalf("ParseMetric(%q) = (%q, %q, %q, %q), want (%q, %q, %q, %q)",
					test.rawName, name, display, kind, metricType,
					test.name, test.display, test.kind, test.metricType)
			}
		})
	}
}
