package cluster

import (
	"fmt"
	"log/slog"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
)

// BuildClusterLookup creates a map of cluster IDs to cluster names by fetching host groups
// from the E-Series API for the given storage system.
func BuildClusterLookup(client *rest.Client, clusterID string, logger *slog.Logger) (map[string]string, error) {
	clusterNames := make(map[string]string)

	apiPath := client.APIPath + "/storage-systems/" + clusterID + "/host-groups"
	clusters, err := client.Fetch(apiPath, nil)
	if err != nil {
		return clusterNames, fmt.Errorf("failed to fetch host groups: %w", err)
	}

	for _, cluster := range clusters {
		clusterRef := cluster.Get("clusterRef").String()
		if clusterRef == "" {
			clusterRef = cluster.Get("id").String()
		}
		clusterName := cluster.Get("name").String()
		if clusterName == "" {
			clusterName = cluster.Get("label").String()
		}
		if clusterRef != "" && clusterName != "" {
			clusterNames[clusterRef] = clusterName
		}
	}

	if logger != nil {
		logger.Debug("built cluster lookup", slog.Int("count", len(clusterNames)))
	}
	return clusterNames, nil
}
