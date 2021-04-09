package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/matrix"
	"strings"
)

type Lun struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Lun{AbstractPlugin: p}
}

func (me *Lun) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	for _, instance := range data.GetInstances() {

		if x := strings.Split(instance.GetLabel("lun"), "/"); len(x) > 3 {
			instance.SetLabel("volume", x[2])
			instance.SetLabel("lun", x[3])
		} else {
			break
		}
	}
	return nil, nil
}
