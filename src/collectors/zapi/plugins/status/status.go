package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/dict"
	"goharvest2/share/errors"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"strings"
)

type Status struct {
	*plugin.AbstractPlugin
	target_labels *dict.Dict
	//target_values *dict.Dict
	target_values map[string]map[string]uint8
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Status{AbstractPlugin: p}
}

func (me *Status) Init() error {

	if err := me.AbstractPlugin.Init(); err != nil {
		return err
	}

	me.target_labels = dict.New()
	//p.target_values = dict.New()
	me.target_values = make(map[string]map[string]uint8)

	if len(me.Params.GetChildren()) == 0 {
		return errors.New(errors.MISSING_PARAM, "status parameters")
	}
	for _, m := range me.Params.GetChildren() {
		name := m.GetNameS()
		label := m.GetChildContentS("label")
		if name == "" || label == "" {
			logger.Warn(me.Prefix, "skipped (%s) with label (%s)", name, label)
			continue
		}
		me.target_labels.Set(name, label)
		me.target_values[name] = make(map[string]uint8)

		if values := m.GetChildContentS("values"); values != "" {
			for i, v := range strings.Fields(values) {
				me.target_values[name][v] = uint8(i)
			}
		} else {
			me.target_values[name][m.GetChildContentS("ok_value")] = 0
		}
		logger.Debug(me.Prefix, "adding metric [%s] for label (%s) with value mappings (%v)", name, label, me.target_values[name])
	}
	logger.Debug(me.Prefix, "will cook status metrics from %d instance labels", me.target_labels.Size())
	return nil
}

func (me *Status) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var err error
	metrics := make(map[string]matrix.Metric)

	for key := range me.target_labels.Map() {
		if metric := data.GetMetric(key); metric != nil {
			metrics[key] = metric
			metric.Reset(data.SizeInstances())
		} else if metric, err = data.AddMetricUint8(key); err == nil {
			metrics[key] = metric
			metric.Reset(data.SizeInstances())
		} else {
			return nil, err
		}
	}

	for _, instance := range data.GetInstances() {

		for key, metric := range metrics {
			label := me.target_labels.Get(key)
			value := instance.GetLabel(label)
			mapping := me.target_values[key]

			if num, ok := mapping[value]; ok {
				metric.SetValueUint8(instance, num)
			} else {
				metric.SetValueUint8(instance, 1)
			}
		}
	}
	return nil, nil
}
