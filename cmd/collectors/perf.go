package collectors

import "github.com/netapp/harvest/v2/pkg/matrix"

// GetMetric retrieves the metric associated with the given key from the current matrix (curMat).
// If the metric does not exist in curMat, it is created with the provided display settings.
// The function also ensures that the same metric exists in the previous matrix (prevMat) to
// allow for subsequent calculations (e.g., prevMetric - curMetric).
// This is particularly important in cases such as ONTAP upgrades, where curMat may contain
// additional metrics that are not present in prevMat. If prevMat does not have the metric,
// it is created to prevent a panic when attempting to perform calculations with non-existent metrics.
//
// This metric creation process within RestPerf/StatPerf is necessary during PollData because the information about whether a metric
// is an array is not available in the RestPerf/StatPerf PollCounter. The determination of whether a metric is an array
// is made by examining the actual data in RestPerf/StatPerf. Therefore, metric creation in RestPerf/StatPerf is performed during
// the poll data phase, and special handling is required for such cases.
//
// The function returns the current metric and any error encountered during its retrieval or creation.
func GetMetric(curMat *matrix.Matrix, prevMat *matrix.Matrix, key string, display ...string) (*matrix.Metric, error) {
	var err error
	curMetric := curMat.GetMetric(key)
	if curMetric == nil {
		curMetric, err = curMat.NewMetricFloat64(key, display...)
		if err != nil {
			return nil, err
		}
	}

	prevMetric := prevMat.GetMetric(key)
	if prevMetric == nil {
		_, err = prevMat.NewMetricFloat64(key, display...)
		if err != nil {
			return nil, err
		}
	}
	return curMetric, nil
}
