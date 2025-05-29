package version

import "testing"

func TestAtLeast(t *testing.T) {

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
		result, err := AtLeast(tc.clusterVersion, tc.minVersion)
		if err != nil && tc.expected {
			t.Errorf("versionAtLeast(%q, %q) returned error: %v", tc.clusterVersion, tc.minVersion, err)
		}
		if result != tc.expected {
			t.Errorf("versionAtLeast(%q, %q) = %v; want %v", tc.clusterVersion, tc.minVersion, result, tc.expected)
		}
	}
}
