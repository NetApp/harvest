package collectors

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"testing"
)

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
