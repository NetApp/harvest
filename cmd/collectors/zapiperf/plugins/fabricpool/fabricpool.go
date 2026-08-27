package fabricpool

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type FabricPool struct {
	*plugin.AbstractPlugin
	includeConstituents bool
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FabricPool{AbstractPlugin: p}
}

func (f *FabricPool) Init(conf.Remote) error {
	err := f.InitAbc()
	if err != nil {
		return err
	}
	f.includeConstituents = f.LoadParam("include_constituents", f.includeConstituents)
	return nil
}

func (f *FabricPool) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	cache, err := collectors.GetFlexGroupFabricPoolMetrics(dataMap, f.Object, "cloud_bin_operation", f.includeConstituents, f.SLogger)
	if err != nil {
		return nil, nil, err
	}
	return []*matrix.Matrix{cache}, nil, nil
}
