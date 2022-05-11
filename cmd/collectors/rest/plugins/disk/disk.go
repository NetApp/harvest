/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package disk

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type Disk struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Disk{AbstractPlugin: p}
}

func (me *Disk) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	for _, instance := range data.GetInstances() {

		containerType := instance.GetLabel("container_type")

		if containerType == "shared" {
			instance.SetLabel("shared", "true")
		} else {
			instance.SetLabel("shared", "false")
		}

		if containerType == "broken" {
			instance.SetLabel("failed", "true")
		} else {
			instance.SetLabel("failed", "false")
		}

	}

	return nil, nil
}
