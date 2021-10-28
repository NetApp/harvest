/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package disk

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/matrix"
)

type Disk struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Disk{AbstractPlugin: p}
}

func (me *Disk) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	for _, instance := range data.GetInstances() {

		container_type := instance.GetLabel("container_type")

		// example name = DISK_SSD_aggr01_8a700cc6-068b-4a42-9a66-9d97f0e761c1
		// disk_type    = SSD
		// aggr         = aggr01

		if container_type == "shared" {
			instance.SetLabel("shared", "true")
		} else {
			instance.SetLabel("shared", "false")
		}

		if container_type == "broken" {
			instance.SetLabel("failed", "true")
		} else {
			instance.SetLabel("failed", "false")
		}

	}

	return nil, nil
}
