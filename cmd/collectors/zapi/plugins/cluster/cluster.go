package cluster

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type Cluster struct {
	*plugin.AbstractPlugin
	addr string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Cluster{AbstractPlugin: p}
}

func (c *Cluster) Init(_ conf.Remote) error {
	var err error

	if err := c.InitAbc(); err != nil {
		return err
	}

	ap, err := conf.PollerNamed(c.Options.Poller)
	if err != nil {
		return err
	}
	c.addr = ap.Addr
	return nil
}

func (c *Cluster) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[c.Object]

	data.SetGlobalLabel("addr", c.addr)
	data.SetGlobalLabel("poller", c.Options.Poller)

	return []*matrix.Matrix{data}, nil, nil
}
