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

func (p *Lun) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	for _, instance := range data.GetInstances() {

        if x := strings.Split(instance.Labels.Get("lun"), "/"); len(x) > 3 {
            instance.Labels.Set("volume", x[2])
            instance.Labels.Set("lun", x[3])
        } else {
            break
        }
	}
	return nil, nil
}
