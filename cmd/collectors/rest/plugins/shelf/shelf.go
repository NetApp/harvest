// Package shelf Copyright NetApp Inc, 2021 All rights reserved
package shelf

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type Shelf struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Shelf{AbstractPlugin: p}
}

func (my *Shelf) Init(conf.Remote) error {
	return my.InitAbc()
}

func (my *Shelf) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {

	data := dataMap[my.Object]
	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}

		model := instance.GetLabel("model")
		moduleType := instance.GetLabel("module_type")

		isEmbed := collectors.IsEmbedShelf(model, moduleType)
		if isEmbed {
			instance.SetLabel("isEmbedded", "Yes")
		} else {
			instance.SetLabel("isEmbedded", "No")
		}
	}
	return nil, nil, nil

}
