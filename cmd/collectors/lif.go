package collectors

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

func NewLif(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Lif{AbstractPlugin: p}
}

type Lif struct {
	*plugin.AbstractPlugin
}

func (l *Lif) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[l.Object]
	clusterName := data.GetGlobalLabels()["cluster"]

	for _, lif := range data.GetInstances() {
		if svm := lif.GetLabel("svm"); svm == "" {
			if ipspace := lif.GetLabel("ipspace"); ipspace == "Cluster" {
				svm = "Cluster"
			} else {
				svm = clusterName
			}
			lif.SetLabel("svm", svm)
		}
	}
	return nil, nil, nil
}
