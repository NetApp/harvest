package snapshotviolation

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"strconv"
	"strings"
)

const (
	ViolationCount     = "violation_count"
	ViolationTotalSize = "violation_total_size"
	KeyToken           = ","
)

var Metrics = []string{
	ViolationCount,
	ViolationTotalSize,
}

type Stats struct {
	Svm       string
	Volume    string
	Count     int
	TotalSize int64
}

func InitMatrix(parent string) (*matrix.Matrix, error) {
	mat := matrix.New(parent+".SnapshotVolume", "snapshot_volume", "snapshot_volume")

	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "svm")
	instanceKeys.NewChildS("", "volume")
	mat.SetExportOptions(exportOptions)
	for _, k := range Metrics {
		err := matrix.CreateMetric(k, mat)
		if err != nil {
			return mat, err
		}
	}
	return mat, nil
}

func ProcessSnapshotData(svm, volume, snapshot, sizeStr string, prefixMap map[string]*set.Set, filteredSnapshotStats map[string]Stats, logger *slog.Logger) {
	key := svm + KeyToken + volume

	prefixes := prefixMap[svm]

	// Check if snapshot name starts with any prefix
	hasPrefix := false
	if prefixes != nil {
		for _, prefix := range prefixes.Values() {
			if strings.HasPrefix(snapshot, prefix) {
				hasPrefix = true
				break
			}
		}
	}

	// If no SVM-specific prefix found, check cluster-scoped prefixes (empty key)
	if !hasPrefix {
		clusterPrefixes := prefixMap[""]
		if clusterPrefixes != nil {
			for _, prefix := range clusterPrefixes.Values() {
				if strings.HasPrefix(snapshot, prefix) {
					hasPrefix = true
					break
				}
			}
		}
	}

	// Only process snapshots that don't have any prefix
	if !hasPrefix {
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			logger.Warn("Failed to parse size", slog.String("size", sizeStr), slogx.Err(err))
			return
		}

		stats, exists := filteredSnapshotStats[key]
		if !exists {
			stats = Stats{
				Svm:    svm,
				Volume: volume,
			}
		}
		stats.Count++
		stats.TotalSize += size
		filteredSnapshotStats[key] = stats
	}
}
