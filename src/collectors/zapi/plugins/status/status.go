package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/dict"
	"goharvest2/share/errors"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
)

type Status struct {
	*plugin.AbstractPlugin
	target_labels *dict.Dict
	target_values *dict.Dict
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Status{AbstractPlugin: p}
}

func (p *Status) Init() error {

	p.target_labels = dict.New()
	p.target_values = dict.New()

	if len(p.Params.GetChildren()) == 0 {
		return errors.New(errors.MISSING_PARAM, "status parameters")
	}
	for _, m := range p.Params.GetChildren() {
		name := m.GetNameS()
		p.target_labels.Set(name, m.GetChildContentS("label"))
		p.target_values.Set(name, m.GetChildContentS("ok_value"))
		logger.Debug(p.Prefix, "adding metric [%s] for label (%s) and OK value (%s)", name, m.GetChildContentS("label"), m.GetChildContentS("ok_value"))
	}
	logger.Debug(p.Prefix, "will cook status metrics from %d instance labels", p.target_labels.Size())
	return nil
}

func (p *Status) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var err error
	metrics := make(map[string]matrix.Metric)

	for key := range p.target_labels.Map() {
		if m := data.GetMetric(key); m != nil {
			metrics[key] = m
			m.Reset(data.SizeInstances())
		} else if m, err = data.AddMetricUint8(key); err == nil {
			metrics[key] = m
			m.Reset(data.SizeInstances())
		} else {
			return nil, err
		}
	}

	for _, instance := range data.GetInstances() {

		for key, metric := range metrics {
			label := p.target_labels.Get(key)
			value := p.target_values.Get(key)

			if x := instance.GetLabel(label); x == value {
				metric.SetValueUint8(instance, 0)
			} else {
				metric.SetValueUint8(instance, 1)
			}
		}
	}
	return nil, nil
}
