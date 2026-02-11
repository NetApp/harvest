package version

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
)

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
		if tc.expected {
			assert.Nil(t, err)
		}
		assert.Equal(t, result, tc.expected)
	}
}
