package grid

import (
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

type Grid struct {
	*plugin.AbstractPlugin
	client *rest.Client
	addr   string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Grid{AbstractPlugin: p}
}

func (g *Grid) Init(remote conf.Remote) error {
	var err error

	if err := g.InitAbc(); err != nil {
		return err
	}

	ap, err := conf.PollerNamed(g.Options.Poller)
	if err != nil {
		return err
	}
	g.addr = ap.Addr

	clientTimeout := g.ParentParams.GetChildContentS("client_timeout")
	if g.client, err = rest.NewClient(g.Options.Poller, clientTimeout, g.Auth); err != nil {
		g.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if apiVersion := g.ParentParams.GetChildContentS("api"); apiVersion != "" {
		g.client.APIPath = "/api/" + apiVersion
	}

	if err := g.client.Init(5, remote); err != nil {
		return err
	}

	return nil
}

func (g *Grid) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[g.Object]

	g.client.Metadata.Reset()

	systemID := ""
	body, err := g.client.GetGridRest("grid/license")
	if err != nil {
		g.SLogger.Error("failed to fetch grid/license", slogx.Err(err))
	} else {
		systemID = gjson.GetBytes(body, "data.systemId").ClonedString()
	}

	for _, instance := range data.GetInstances() {
		instance.SetLabel("addr", g.addr)
		if systemID != "" {
			instance.SetLabel("system_id", systemID)
		}
	}

	return nil, g.client.Metadata, nil
}
