package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/matrix"
	"strings"
)

type Node struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Node{AbstractPlugin: p}
}

func (my *Node) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	for _, instance := range data.GetInstances() {

		warnings := make([]string, 0)

		if w := instance.GetLabel("failed_fan_message"); w != "" && ! strings.HasPrefix(w, "There are no failed ") {
			warnings = append(warnings, w)
		}

		if w := instance.GetLabel("failed_power_message"); w != "" && ! strings.HasPrefix(w, "There are no failed ") {
			warnings = append(warnings, w)
		}

		instance.SetLabel("warnings", strings.Join(warnings, " "))
	}

	return nil, nil
}
