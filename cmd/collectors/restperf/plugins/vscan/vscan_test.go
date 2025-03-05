package vscan

import (
	"encoding/json"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"os"
	"testing"
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

	dataMap := map[string]*matrix.Matrix{
		data.Object: data,
	}
	// run the plugin
	results, _, err := v.Run(dataMap)
	if err != nil {
		t.Fatal(err)
	}

	if metricsPerScanner == "false" {
		if results != nil {
			t.Fatalf("result should be nil")
		} else {
			return
		}
	}

	// Verify the cacheData
	cacheData := results[0]
	if len(cacheData.GetInstances()) != cacheExportedInstanceCount {
		t.Fatalf("expected %d cacheExportedInstance count, got %d", cacheExportedInstanceCount, len(cacheData.GetInstances()))
	}

	exportedInstance := 0
	for _, instance := range data.GetInstances() {
		if instance.IsExportable() {
			exportedInstance++
		}
	}
	if exportedInstance != dataExportedInstanceCount {
		t.Fatalf("expected %d dataExportedInstance count, got %d", dataExportedInstanceCount, len(data.GetInstances()))
	}
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
