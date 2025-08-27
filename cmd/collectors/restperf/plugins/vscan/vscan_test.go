package vscan

import (
	"encoding/json"
	"github.com/netapp/harvest/v2/assert"
	"log/slog"
	"os"
	"testing"

	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

type VscanRecords struct {
	VscanRecords []VscanRecord `json:"records"`
}

type VscanRecord struct {
	ID string `json:"id"`
}

func runTest(t *testing.T, createRestVscan func(params *node.Node) plugin.Plugin, testFile string, metricsPerScanner string, cacheExportedInstanceCount int, dataExportedInstanceCount int) {
	params := node.NewS("Vscan")
	params.NewChildS("metricsPerScanner", metricsPerScanner)
	v := createRestVscan(params)
	data := readTestFile(testFile)
	if data == nil {
		t.Fatalf("failed to read test file %s", testFile)
	}

	dataMap := map[string]*matrix.Matrix{
		data.Object: data,
	}
	// run the plugin
	results, _, err := v.Run(dataMap)
	assert.Nil(t, err)

	if metricsPerScanner == "false" {
		assert.Nil(t, results)
		return
	}

	// Verify the cacheData
	cacheData := results[0]
	assert.Equal(t, len(cacheData.GetInstances()), cacheExportedInstanceCount)

	exportedInstance := 0
	for _, instance := range data.GetInstances() {
		if instance.IsExportable() {
			exportedInstance++
		}
	}
	assert.Equal(t, exportedInstance, dataExportedInstanceCount)
}

func TestRunForAllImplementations(t *testing.T) {
	type test struct {
		name                       string
		testFilePath               string
		metricsPerScanner          string
		cacheExportedInstanceCount int
		dataExportedInstanceCount  int
	}

	tests := []test{
		{name: "TestMetricScannerTrueValidData", metricsPerScanner: "true", testFilePath: "test_valid_vscan.json", cacheExportedInstanceCount: 2, dataExportedInstanceCount: 0},
		{name: "TestMetricScannerTrueNonValidData", metricsPerScanner: "true", testFilePath: "test_nonvalid_vscan.json", cacheExportedInstanceCount: 0, dataExportedInstanceCount: 0},
		{name: "TestMetricScannerFalseValidData", metricsPerScanner: "false", testFilePath: "test_valid_vscan.json", cacheExportedInstanceCount: 0, dataExportedInstanceCount: 4},
		{name: "TestMetricScannerFalseNonValidData", metricsPerScanner: "false", testFilePath: "test_nonvalid_vscan.json", cacheExportedInstanceCount: 0, dataExportedInstanceCount: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTest(t, createRestVscan, tt.testFilePath, tt.metricsPerScanner, tt.cacheExportedInstanceCount, tt.dataExportedInstanceCount)
		})
	}
}

func createRestVscan(params *node.Node) plugin.Plugin {
	o := options.Options{IsTest: true}
	v := &Vscan{AbstractPlugin: plugin.New("vscan", &o, params, nil, "vscan", nil)}
	v.SLogger = slog.Default()
	return v
}

func readTestFile(testFilePath string) *matrix.Matrix {
	data := matrix.New("vscan", "vscan", "vscan")
	var vscanRecords VscanRecords

	file, _ := os.Open(testFilePath)
	err := json.NewDecoder(file).Decode(&vscanRecords)
	if err != nil {
		return nil
	}

	for i := range len(vscanRecords.VscanRecords) {
		id := vscanRecords.VscanRecords[i].ID
		instance, _ := data.NewInstance(id)
		instance.SetLabel("id", id)
	}

	return data
}
