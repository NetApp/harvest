package main

import (
	"strings"
	"goharvest2/poller/collector/plugin"
    "goharvest2/poller/struct/matrix"
	"goharvest2/poller/struct/options"
	"goharvest2/poller/struct/yaml"
)

type Node struct {
	*plugin.AbstractPlugin
}

func New(parent_name string, options *options.Options, params *yaml.Node, pparams *yaml.Node) plugin.Plugin {
	p := plugin.New(parent_name, options, params, pparams)
	return &Node{AbstractPlugin: p}
}

func (p *Node) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	for _, instance := range data.GetInstances() {

		warnings := make([]string, 0)

		if w := instance.Labels.Get("failed_fan_message"); w != "" && w != "There are no failed fans." {
			warnings = append(warnings, w)
		}

		if w := instance.Labels.Get("failed_power_message"); w != "" && w != "There are no failed power supplies." {
			warnings = append(warnings, w)
		}

		instance.Labels.Set("warnings", strings.Join(warnings, " "))
	}

	return nil, nil
}