package fabricpool

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"strconv"
)

type FabricPool struct {
	*plugin.AbstractPlugin
	includeConstituents bool
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FabricPool{AbstractPlugin: p}
}

func (f *FabricPool) Init() error {
	err := f.InitAbc()
	if err != nil {
		return err
	}
	if val := f.Params.GetChildContentS("include_constituents"); val != "" {
		if boolValue, err := strconv.ParseBool(val); err == nil {
			f.includeConstituents = boolValue
		}
	}
	return nil
}

func (f *FabricPool) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	cache, err := collectors.GetFlexGroupFabricPoolMetrics(dataMap, f.Object, "cloud_bin_operation", f.includeConstituents, f.SLogger)
	if err != nil {
		return nil, nil, err
	}
	return []*matrix.Matrix{cache}, nil, nil
}
