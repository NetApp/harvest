package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"strconv"
	"strings"
)

type FlexGroup struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FlexGroup{AbstractPlugin: p}
}

func (my *FlexGroup) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	mydata := data.Clone(false, true, false)
	mydata.Plugin = "zapi.volume.flexgroup"

	counts := make(map[string]int)

	// create new instance cache for FlexGroup volumes
	for _, i := range data.GetInstances() {

		if key, name := fetch_flexgroup_name(i); key != "" {
			// mark flexgroup constituents as non-exportable
			i.SetExportable(false)

			if mydata.GetInstance(key) == nil {

				if instance, err := mydata.AddInstance(key); err == nil {
					instance.SetLabels(i.GetLabels().Copy())
					instance.SetLabel("volume", name)
					counts[key] = 1
				} else {
					logger.Error(my.Prefix, err.Error())
					return nil, err
				}

			} else {
				counts[key] += 1
			}
		}
	}

	// change dtype of percentage metrics
	for key, m := range mydata.GetMetrics() {
		name := m.GetName()
		if strings.HasSuffix(name, "_percent") {
			newm, err := mydata.ChangeMetricType(key, "float64")
			if err != nil {
				return nil, err
			}
			newm.SetName(name)
		}
	}

	logger.Debug(my.Prefix, "extracted %d flexgroup instances", len(counts))

	if err := mydata.Reset(); err != nil {
		logger.Error(my.Prefix, err.Error())
		return nil, err
	}

	// create summaries
	for _, instance := range data.GetInstances() {
		if key, _ := fetch_flexgroup_name(instance); key != "" {
			if myinstance := mydata.GetInstance(key); myinstance != nil {
				mydata.InstanceWiseAdditionUint64(myinstance, instance, data)
			}
		}
	}

	// normalize percentage counters
	for key, instance := range mydata.GetInstances() {

		// set count as label
		count, _ := counts[key]
		instance.SetLabel("count", strconv.Itoa(count))

		for _, metric := range mydata.GetMetrics() {
			if strings.HasSuffix(metric.GetName(), "_percent") {
				if value, has := metric.GetValueFloat64(instance); has {
					metric.SetValueFloat64(instance, value/float64(count))
				}
			} else if metric.GetName() == "status" {
				if instance.GetLabel("state") == "online" {
					metric.SetValueUint8(instance, 0)
				} else {
					metric.SetValueUint8(instance, 1)
				}
			}
		}
	}
	return []*matrix.Matrix{mydata}, nil
}

func fetch_flexgroup_name(instance *matrix.Instance) (string, string) {
	var key, name, vol string

	if instance.GetLabel("style") == "flexgroup_constituent" {
		if vol = instance.GetLabel("volume"); len(vol) > 6 {
			name = vol[:len(vol)-6]
			key = instance.GetLabel("svm") + "." + instance.GetLabel("node") + "." + name
		}
	}

	return key, name
}
