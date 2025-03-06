package collectors

import (
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"log/slog"
	"testing"
)

func createRestLif() plugin.Plugin {
	o := options.Options{IsTest: true}
	l := &Lif{AbstractPlugin: plugin.New("lif", &o, nil, nil, "lif", nil)}
	l.SLogger = slog.Default()
	return l
}

func TestRunForAllImplementations(t *testing.T) {
	clusterName := "cluster-name-1"
	l := createRestLif()
	data := matrix.New("lif", "lif", "lif")
	data.GetGlobalLabels()["cluster"] = clusterName

	type test struct {
		name        string
		svm         string
		ipspace     string
		expectedSvm string
	}

	tests := []test{
		{name: "TestValidSvm", svm: "svm1", ipspace: "Default", expectedSvm: "svm1"},
		{name: "TestEmptySvmWithIpspaceCluster", svm: "", ipspace: "Cluster", expectedSvm: "Cluster"},
		{name: "TestEmptySvmWithIpspaceDefault", svm: "", ipspace: "Default", expectedSvm: clusterName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifInstance, _ := data.NewInstance(tt.name)
			lifInstance.SetLabel("svm", tt.svm)
			lifInstance.SetLabel("ipspace", tt.ipspace)

			dataMap := map[string]*matrix.Matrix{
				data.Object: data,
			}
			// run the plugin
			_, _, err := l.Run(dataMap)
			if err != nil {
				t.Fatal(err)
			}

			actualSvm := data.GetInstance(tt.name).GetLabel("svm")
			if actualSvm != tt.expectedSvm {
				t.Fatalf("expected svm is %s, got %s", tt.expectedSvm, actualSvm)
			}
		})
	}
}
