package collectors

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"slices"
	"strings"
)

type StorageUnit struct {
	*plugin.AbstractPlugin
}

func NewStorageUnit(p *plugin.AbstractPlugin) plugin.Plugin {
	return &StorageUnit{AbstractPlugin: p}
}

func (s *StorageUnit) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[s.Object]
	var hostGroups []string
	for _, instance := range data.GetInstances() {
		hostGroups = make([]string, 0)
		mapsData := gjson.Result{Type: gjson.JSON, Raw: "[" + instance.GetLabel("maps") + "]"}
		if mapsData.Exists() {
			for _, mapData := range mapsData.Array() {
				hostGroup := mapData.Get("host_group.name").ClonedString()
				hostGroups = append(hostGroups, hostGroup)
			}
		}
		slices.Sort(hostGroups)
		instance.SetLabel("host_group", strings.Join(hostGroups, ","))
	}
	return nil, nil, nil
}
