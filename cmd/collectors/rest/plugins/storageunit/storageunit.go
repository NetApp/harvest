package storageunit

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

type StorageUnit struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &StorageUnit{AbstractPlugin: p}
}

func (s *StorageUnit) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[s.Object]
	for _, instance := range data.GetInstances() {
		mapsData := gjson.Result{Type: gjson.JSON, Raw: instance.GetLabel("maps")}
		if mapsData.Exists() && mapsData.Get("host_group").Exists() {
			hostGroup := mapsData.Get("host_group").Get("name").ClonedString()
			instance.SetLabel("host_group", hostGroup)
		}
	}
	return nil, nil, nil
}
