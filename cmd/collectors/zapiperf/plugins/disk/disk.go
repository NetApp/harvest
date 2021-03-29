package main

import (
    "goharvest2/cmd/poller/collector/plugin"
    "goharvest2/pkg/matrix"
    "strings"
)

type Disk struct {
    *plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
    return &Disk{AbstractPlugin: p}
}

func (p *Disk) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
    for _, instance := range data.GetInstances() {
        if x := strings.Split(instance.Labels.Get("raid_name"), "/"); len(x) == 5 {
            instance.Labels.Set("aggr", x[1])
            instance.Labels.Set("plex", x[2])
            instance.Labels.Set("raid", x[3])
            instance.Labels.Set("disk", x[4])
        }
    }
    return nil, nil
}
