package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"strings"
)

type Disk struct {
	*plugin.AbstractPlugin
	node *matrix.Matrix
	aggr *matrix.Matrix
	plex *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Disk{AbstractPlugin: p}
}

func (me *Disk) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	/*
		node := data.Clone(false, true, false)
		node.Object = "node_disk"
		node.Plugin = "Disk"
		node_count := make(map[string]int)

		/*
		aggr := data.Clone(false, true, false)
		aggr.Object = "aggr_disk"
		aggr.Plugin = "Disk"
		aggr_count := 0

		plex := data.Clone(false, true, false)
		plex.Object = "plex_disk"
		plex.Plugin = "Disk"
		plex_count := 0
	*/

	/*
		if err := node.Reset(); err != nil {
			return nil, err
		}
	*/

	for _, instance := range data.GetInstances() {

		//var node_name, aggr_name, plex_name string
		/*
			var node_name string
			var n *matrix.Instance
			var err error
		*/

		if x := strings.Split(instance.GetLabel("raid_group"), "/"); len(x) == 4 {
			instance.SetLabel("aggr", x[1])
			instance.SetLabel("plex", x[2])
			instance.SetLabel("raid", x[3])

			//node_name = instance.GetLabel("node")
			//aggr_name := x[1]
			//plex_name := x[2]
		} else {
			logger.Warn(me.Prefix, "raid_group (%s) not expected format", instance.GetLabel("raid_group"))
			continue
		}
	}

	/*
			if n = node.GetInstance(node_name); n == nil {
				if n, err = node.AddInstance(node_name); err != nil {
					return nil, err
				}
			}

			for key, metric := range data.GetMetrics() {

				if value, ok := metric.GetValueFloat64(instance); ok {

					if node_metric := node.GetMetric(key); node_metric != nil {

						if err = node_metric.AddValueFloat64(n, value); err != nil {
							logger.Error(me.Prefix, err.Error())
						} else {
							node_count[node_name]++
						}
					}
				}
			}
		}

		// normalize values

		for _, metric := range node.GetMetrics() {
			if p := metric.GetProperty(); p == "average" || p == "percent" {
				for key, instance := range node.GetInstances() {
					if v, ok := metric.GetValueFloat64(instance); ok {
						metric.SetValueFloat64(instance, v/float64(node_count[key]))
					}
				}
			}
		}

		return []*matrix.Matrix{node}, nil
	*/
	return nil, nil
}
