package volume

import (
	"strings"

	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type Volume struct {
	*plugin.AbstractPlugin
	includeConstituents bool
	volumesMap          map[string]string // volume-name -> volume-style map
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) Init(conf.Remote) error {
	if err := v.InitAbc(); err != nil {
		return err
	}

	v.volumesMap = make(map[string]string)

	// Read template to decide inclusion of flexgroup constituents
	v.includeConstituents = collectors.ReadPluginKey(v.Params, "include_constituents")

	return nil
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]
	style := "style"
	opsKeyPrefix := "temp_"
	v.volumesMap = v.getVolumeMap(data)
	matrices, metadata, err := collectors.ProcessFlexGroupData(v.SLogger, data, style, v.includeConstituents, opsKeyPrefix, v.volumesMap, false)
	if err != nil {
		return matrices, metadata, err
	}
	v.updateExportable(data)
	return matrices, metadata, err
}

func (v *Volume) updateExportable(data *matrix.Matrix) {
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
}

func (v *Volume) getVolumeMap(data *matrix.Matrix) map[string]string {
	volumesMap := make(map[string]string)
	for _, instance := range data.GetInstances() {
		style := instance.GetLabel("style")
		name := instance.GetLabel("volume")
		svm := instance.GetLabel("svm")
		volumesMap[svm+name] = style
	}
	return volumesMap
}
