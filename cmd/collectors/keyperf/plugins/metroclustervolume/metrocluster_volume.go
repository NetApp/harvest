package metroclustervolume

import (
	"strings"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type MetroClusterVolume struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &MetroClusterVolume{AbstractPlugin: p}
}

func (m *MetroClusterVolume) Init(conf.Remote) error {
	return m.InitAbc()
}

func (m *MetroClusterVolume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[m.Object]
	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		// SVM names ending with "-mc" are MetroCluster SVMs.
		// Only export volume metrics from MetroCluster SVMs if the volume is online.
		if strings.HasSuffix(instance.GetLabel("svm"), "-mc") {
			instance.SetExportable(instance.GetLabel("state") == "online")
		}
	}
	return nil, nil, nil
}
