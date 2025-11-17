package victoriametrics

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/exporter"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"testing"
)

func setupVictoriaMetrics(t *testing.T, exporterName string) *VictoriaMetrics {
	opts := options.New()
	opts.IsTest = true

	_, err := conf.LoadHarvestConfig("../../tools/doctor/testdata/testConfig.yml")
	assert.Nil(t, err)
	e, ok := conf.Config.Exporters[exporterName]
	assert.True(t, ok)

	victoriametrics := &VictoriaMetrics{AbstractExporter: exporter.New("VictoriaMetrics", exporterName, opts, e, nil)}
	err = victoriametrics.Init()
	assert.Nil(t, err)

	return victoriametrics
}

func TestAddrParameter(t *testing.T) {
	expectedURL := "http://localhost:8428/api/v1/import/prometheus"
	exporterName := "victoriametrics-test-addr"
	victoriametrics := setupVictoriaMetrics(t, exporterName)

	assert.Equal(t, victoriametrics.url, expectedURL)
}

func TestUrlParameter(t *testing.T) {
	expectedURL := "http://localhost:8428/api/v1/import/prometheus"
	exporterName := "victoriametrics-test-url"
	victoriametrics := setupVictoriaMetrics(t, exporterName)

	assert.Equal(t, victoriametrics.url, expectedURL)
}

// test that the addr, port and version parameters are handled properly to construct server URL
func TestVersionParameter(t *testing.T) {
	expectedURL := "http://localhost:8400/api/v4/import/prometheus"
	exporterName := "victoriametrics-test-version"
	victoriametrics := setupVictoriaMetrics(t, exporterName)

	assert.Equal(t, victoriametrics.url, expectedURL)
}

// test that `addr` field is ignored when using the `url` field
func TestUrlIgnores(t *testing.T) {
	expectedURL := "https://example.com:8428/api/v1/import/prometheus"
	exporterName := "victoriametrics-with-url"
	victoriametrics := setupVictoriaMetrics(t, exporterName)

	assert.Equal(t, victoriametrics.url, expectedURL)
}

// test rendering
func TestExportDebug(t *testing.T) {
	exporterName := "victoriametrics-test-url"
	victoriametrics := setupVictoriaMetrics(t, exporterName)

	// matrix with fake data
	data := matrix.New("test_exporter", "vm_test_data", "vm_test_data")
	data.SetExportOptions(matrix.DefaultExportOptions())

	// add metric
	m, err := data.NewMetricInt64("test_metric")
	assert.Nil(t, err)

	// add instance
	i, err := data.NewInstance("test_instance")
	assert.Nil(t, err)
	i.SetLabel("test_label", "test_label_value")

	// add numeric data
	m.SetValueInt64(i, 42)

	// render data
	_, err = victoriametrics.Export(data)
	assert.Nil(t, err)
}
