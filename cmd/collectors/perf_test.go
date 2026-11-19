package collectors

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"testing"
)

func TestDefaultLatencyIOReqd(t *testing.T) {
	assert.Equal(t, DefaultLatencyIOReqd, 5)
}

func TestCounterOverride(t *testing.T) {
	assert.Equal(t, CounterOverride(nil, "ops"), "")

	params := node.NewS("params")
	assert.Equal(t, CounterOverride(params, "ops"), "")

	override := params.NewChildS("override", "")
	override.NewChildS("ops", "rate")
	assert.Equal(t, CounterOverride(params, "ops"), "rate")
	assert.Equal(t, CounterOverride(params, "latency"), "")
}

func TestSetupPerfMatrix(t *testing.T) {
	mat := matrix.New("data", "old", "data")
	metadata := matrix.New("metadata", "metadata", "metadata")
	params := node.NewS("params")
	labels := params.NewChildS("labels", "")
	labels.NewChildS("site", "east")

	SetupPerfMatrix(mat, metadata, params, "cluster-1", "volume")

	assert.Equal(t, mat.Object, "volume")
	assert.Equal(t, mat.GetGlobalLabels()["cluster"], "cluster-1")
	assert.Equal(t, mat.GetGlobalLabels()["site"], "east")
	assert.NotNil(t, metadata.GetMetric("skips"))
	assert.NotNil(t, metadata.GetMetric("numPartials"))
}

func TestEnsureTimestampMetric(t *testing.T) {
	mat := matrix.New("data", "volume", "data")

	EnsureTimestampMetric(mat, slog.Default())
	metric := mat.GetMetric(TimestampMetricName)
	assert.NotNil(t, metric)
	assert.Equal(t, metric.GetProperty(), "raw")
	assert.Equal(t, metric.IsExportable(), false)

	EnsureTimestampMetric(mat, slog.Default())
	assert.Equal(t, mat.GetMetric(TimestampMetricName), metric)
}

func TestEnsureTimestampMetricPreservesExistingMetric(t *testing.T) {
	mat := matrix.New("data", "volume", "data")
	metric, err := mat.NewMetricFloat64(TimestampMetricName)
	assert.Nil(t, err)
	metric.SetProperty("delta")
	metric.SetExportable(true)

	EnsureTimestampMetric(mat, slog.Default())

	assert.Equal(t, mat.GetMetric(TimestampMetricName).GetProperty(), "delta")
	assert.Equal(t, mat.GetMetric(TimestampMetricName).IsExportable(), true)
}

func TestUsesLatencyThreshold(t *testing.T) {
	tests := []struct {
		query, name, denominator string
		want                     bool
	}{
		{query: "volume", name: "read_latency", denominator: "read_ops", want: true},
		{query: "volume", name: "avg_latency", denominator: "total_ops", want: true},
		{query: "workload_queue_combined", name: "cpu_dblade_residence_time", denominator: "iops_total", want: true},
		{query: "workload_queue_combined", name: "disk_ssd_background_service_time#direct", denominator: "iops_total", want: true},
		{query: "workload_queue_combined", name: "delay_cop_wait_time", denominator: "delay_cop_num_wait", want: false},
		{query: "workload_queue_combined", name: "cache_miss_rate", denominator: "cache_miss_rate_base", want: false},
		{query: "wafl", name: "cp_phase_times", denominator: "iops_total", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.query+"/"+tt.name, func(t *testing.T) {
			assert.Equal(t, UsesLatencyThreshold(tt.query, tt.name, tt.denominator), tt.want)
		})
	}
}
