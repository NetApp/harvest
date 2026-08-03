package grid

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type Grid struct {
	*plugin.AbstractPlugin
	addr string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Grid{AbstractPlugin: p}
}

func (g *Grid) Init(_ conf.Remote) error {
	if err := g.InitAbc(); err != nil {
		return err
	}

	ap, err := conf.PollerNamed(g.Options.Poller)
	if err != nil {
		return err
	}
	g.addr = ap.Addr

	return nil
}

func (g *Grid) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[g.Object]

	for _, instance := range data.GetInstances() {
		instance.SetLabel("addr", g.addr)
		instance.SetLabel("system_id", g.Remote.UUID)
	}

	return nil, nil, nil
}
