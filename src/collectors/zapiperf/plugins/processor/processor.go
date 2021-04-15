package main

import (
	"goharvest2/poller/collector/plugin"
	//"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"strconv"
)

type Processor struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Processor{AbstractPlugin: p}
}

func (me *Processor) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	cpu_count := make(map[string]int)

	sum := data.Clone(false, true, false)
	sum.Object = "node_cpu"
	sum.Plugin = "processor"
	//matrix.New("processor", "node_processor", "processor")
	//summary.GlobalLabels = data.GlobalLabels
	//summary.SetExportOptions(data.ExportOptions.Copy())

	/*
		for key, m := range data.GetMetrics() {
			if m.Enabled {
				if m.Labels != nil && m.GetLabel("metric") == "idle" {
					summary.AddMetric(key, m.Name, false)
				} else {
					nm, _ := summary.AddMetric(key, m.Name, true)
					if m.Labels != nil {
						nm.Labels = m.Labels
					}
				}
			}
		}
	*/

	// create new instance cache
	for _, instance := range data.GetInstances() {
		node := instance.GetLabel("node")
		if sum.GetInstance(node) == nil {
			if node_instance, err := sum.AddInstance(node); err == nil {
				node_instance.SetLabel("node", node)
			} else {
				return nil, err
			}
		}
		cpu_count[node]++
	}

	if err := sum.Reset(); err != nil {
		return nil, err
	}

	for _, instance := range data.GetInstances() {

		node := instance.GetLabel("node")

		if node_instance := sum.GetInstance(node); node_instance != nil {

			count := cpu_count[node]
			//logger.Debug(me.Prefix, "creating summary  instance [%s] with %d CPUs", node, count)

			node_instance.SetLabel("cpus", strconv.Itoa(count))

			for key, node_metric := range sum.GetMetrics() {

				if metric := data.GetMetric(key); metric != nil && metric.GetType() == "float64" {

					if value, ok := metric.GetValueFloat64(instance); ok {

						node_value, _ := node_metric.GetValueFloat64(node_instance)
						node_metric.SetValueFloat64(node_instance, node_value+value)
					}
				}
			}
		}
	}

	// normalize processor_busy by cpu_count
	for _, m := range sum.GetMetrics() {
		if m.GetName() == "busy" || m.GetName() == "domain_busy" {
			for _, i := range sum.GetInstances() {
				if v, ok := m.GetValueFloat64(i); ok {
					count := cpu_count[i.GetLabel("node")]
					m.SetValueFloat64(i, v/float64(count))
				}
			}
		}
		if m.GetLabel("metric") == "idle" {
			m.SetExportable(false)
		}
	}

	return []*matrix.Matrix{sum}, nil
}
