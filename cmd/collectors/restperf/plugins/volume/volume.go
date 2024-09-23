package volume

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
)

type Volume struct {
	*plugin.AbstractPlugin
	styleType           string
	includeConstituents bool
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) Init() error {

	if err := v.InitAbc(); err != nil {
		return err
	}

	v.styleType = "style"

	if v.Params.HasChildS("historicalLabels") {
		v.styleType = "type"
	}

	// Read template to decide inclusion of flexgroup constituents
	v.includeConstituents = collectors.ReadPluginKey(v.Params, "include_constituents")
	return nil
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[v.Object]
	style := v.styleType
	opsKeyPrefix := "temp_"

	return collectors.ProcessFlexGroupData(v.SLogger, data, style, v.includeConstituents, opsKeyPrefix)
}
