package version

import (
	"fmt"
	version2 "github.com/netapp/harvest/v2/third_party/go-version"
)

// AtLeast checks if the provided currentVersion of the cluster
// is greater than or equal to the provided minimum version (minVersion).
func AtLeast(currentVersion string, minVersion string) (bool, error) {
	parsedClusterVersion, err := version2.NewVersion(currentVersion)
	if err != nil {
		return false, fmt.Errorf("invalid current version: %w", err)
	}

	minSupportedVersion, err := version2.NewVersion(minVersion)
	if err != nil {
		return false, fmt.Errorf("invalid minimum version: %w", err)
	}

	// Check if the current version is greater than or equal to the minimum version
	// and return the result
	return parsedClusterVersion.GreaterThanOrEqual(minSupportedVersion), nil
}
