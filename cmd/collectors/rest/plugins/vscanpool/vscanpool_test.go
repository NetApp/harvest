package vscanpool

import (
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"testing"
)

func createRestVscanServer(params *node.Node, testFile string) plugin.Plugin {
	o := options.Options{IsTest: true}
	v := &VscanPool{AbstractPlugin: plugin.New("vscan_server", &o, params, nil, "vscan_server", nil)}
	v.SLogger = slog.Default()
	v.testFile = "../../testdata/" + testFile
	v.client = &rest.Client{Metadata: &util.Metadata{}}
	return v
}

func TestRunForAllImplementations(t *testing.T) {
	clusterName := "cluster-name-1"
	data := matrix.New("vscan_server", "vscan_server", "vscan_server")
	data.GetGlobalLabels()["cluster"] = clusterName

	type test struct {
		name                        string
		svm                         string
		scannerPools                string
		testFile                    string
		expectedInstances           int
		expectedDisconnectedServers string
	}

	svm := "vs_8272"
	scannerPool1 := `[
        {
          "name": "VscanPool1",
          "vsName": "vs_8272",
          "servers": [
            "10.92.153.246",
            "10.92.153.245"
          ]
        }
      ],`

	tests := []test{
		{name: "TestEmptyScannerPool", svm: svm, scannerPools: "", testFile: "vscan-server-status1.json", expectedInstances: 0, expectedDisconnectedServers: ""},
		{name: "TestValidScannerPoolNoDisconnected", svm: svm, scannerPools: scannerPool1, testFile: "vscan-server-status1.json", expectedInstances: 0, expectedDisconnectedServers: ""},
		{name: "TestValidScannerPoolOneDisconnected", svm: svm, scannerPools: scannerPool1, testFile: "vscan-server-status2.json", expectedInstances: 1, expectedDisconnectedServers: "10.92.153.245"},
		{name: "TestValidScannerPoolTwoDisconnected", svm: svm, scannerPools: scannerPool1, testFile: "vscan-server-status3.json", expectedInstances: 1, expectedDisconnectedServers: "10.92.153.245,10.92.153.246"},
		{name: "TestValidScannerPoolZeroDisconnectedBasedOnUpdatetime", svm: svm, scannerPools: scannerPool1, testFile: "vscan-server-status4.json", expectedInstances: 0, expectedDisconnectedServers: ""},
		{name: "TestValidScannerPoolOneDisconnectedBasedOnUpdatetime", svm: svm, scannerPools: scannerPool1, testFile: "vscan-server-status5.json", expectedInstances: 1, expectedDisconnectedServers: "10.92.153.245"},
		{name: "TestValidScannerPoolTwoDisconnectedBasedOnUpdatetime", svm: svm, scannerPools: scannerPool1, testFile: "vscan-server-status6.json", expectedInstances: 1, expectedDisconnectedServers: "10.92.153.245,10.92.153.246"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scannerPoolInstance, _ := data.NewInstance(tt.name)
			scannerPoolInstance.SetLabel("svm", tt.svm)
			scannerPoolInstance.SetLabel("scanner_pools", tt.scannerPools)

			dataMap := map[string]*matrix.Matrix{
				data.Object: data,
			}

			v := createRestVscanServer(node.NewS("VscanPool"), tt.testFile)
			_ = v.Init(conf.Remote{})
			// run the plugin
			newData, _, err := v.Run(dataMap)
			if err != nil {
				t.Fatal(err)
			}

			if len(newData) != 1 {
				t.Fatalf("expected 1 output matrices, got %d", len(newData))
			}

			if len(newData[0].GetInstances()) != tt.expectedInstances {
				t.Fatalf("expected %d output matrices, got %d", tt.expectedInstances, len(newData[0].GetInstances()))
			}

			for _, instance := range newData[0].GetInstances() {
				actualDisconnectedServers := instance.GetLabel("vscan_server")
				if actualDisconnectedServers != tt.expectedDisconnectedServers {
					t.Fatalf("expected disconnected servers %s, got %s", tt.expectedDisconnectedServers, actualDisconnectedServers)
				}
			}
			data.RemoveInstance(tt.name)
		})
	}
}
