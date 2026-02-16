package hostcluster

import (
	"fmt"
	"log/slog"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
)

// BuildHostClusterLookup creates a map of host cluster IDs to host cluster names
func BuildHostClusterLookup(client *rest.Client, arrayID string, logger *slog.Logger) (map[string]string, error) {
	hostClusterNames := make(map[string]string)

	apiPath := client.APIPath + "/storage-systems/" + arrayID + "/host-groups"
	hosts, err := client.Fetch(apiPath, nil)
	if err != nil {
		return hostClusterNames, fmt.Errorf("failed to fetch host groups: %w", err)
	}

	for _, host := range hosts {
		hostRef := host.Get("hostRef").ClonedString()
		if hostRef == "" {
			hostRef = host.Get("id").ClonedString()
		}
		hostName := host.Get("name").ClonedString()
		if hostName == "" {
			hostName = host.Get("label").ClonedString()
		}
		if hostRef != "" && hostName != "" {
			hostClusterNames[hostRef] = hostName
		}
	}

	if logger != nil {
		logger.Debug("built host group lookup", slog.Int("count", len(hostClusterNames)))
	}
	return hostClusterNames, nil
}
