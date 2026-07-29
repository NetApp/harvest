package snapshotviolation

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"strings"
)

const (
	ViolationCount     = "violation_count"
	ViolationTotalSize = "violation_total_size"
	OldestCreateTime   = "oldest_create_time"
	NewestCreateTime   = "newest_create_time"
	KeyToken           = ","
)

var Metrics = []string{
	ViolationCount,
	ViolationTotalSize,
	OldestCreateTime,
	NewestCreateTime,
}

type Stats struct {
	Svm              string
	Volume           string
	Count            int
	TotalSize        int64
	OldestCreateTime int64
	NewestCreateTime int64
}

func InitMatrix(parent string) (*matrix.Matrix, error) {
	mat := matrix.New(parent+".SnapshotVolume", "snapshot_volume", "snapshot_volume")

	mat.SetExportOptions(matrix.NewExportOptions("svm", "volume"))
	if err := mat.NewMetricsFloat64(Metrics...); err != nil {
		return mat, err
	}
	return mat, nil
}

func ProcessSnapshotData(svm, volume, snapshot string, size int64, createTime int64, prefixMap map[string]*set.Set, filteredSnapshotStats map[string]Stats) {
	key := svm + KeyToken + volume

	if createTime > 0 {
		stats, exists := filteredSnapshotStats[key]
		if !exists {
			stats = Stats{
				Svm:    svm,
				Volume: volume,
			}
		}
		if stats.OldestCreateTime == 0 || createTime < stats.OldestCreateTime {
			stats.OldestCreateTime = createTime
		}
		if createTime > stats.NewestCreateTime {
			stats.NewestCreateTime = createTime
		}
		filteredSnapshotStats[key] = stats
	}

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
