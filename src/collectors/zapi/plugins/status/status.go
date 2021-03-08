package main

import (
    "goharvest2/poller/collector/plugin"
    "goharvest2/share/matrix"
    "goharvest2/share/dict"
    "goharvest2/share/errors"
    "goharvest2/share/logger"
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

    if metrics := p.Params.GetChildS("metrics"); metrics == nil {
        return errors.New(errors.MISSING_PARAM, "metrics")
    } else {
        for _, m := range metrics.GetChildren() {
            name := m.GetNameS()
            p.target_labels.Set(name, m.GetChildContentS("label"))
            p.target_values.Set(name, m.GetChildContentS("ok_value"))
        }
    }
    logger.Debug(p.Prefix, "initialized plugin, will cook status metrics from %d instance labels", p.target_labels.Size())
    return nil
}

func (p *Status) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

    var err error
    metrics := make(map[string]*matrix.Metric)

    for key, _ := range p.target_labels.Iter() {
        if m := data.GetMetric(key); m != nil {
            metrics[key] = m
        } else if m, err = data.AddMetric(key, key, true); err == nil {
            metrics[key] = m
        } else {
            return nil, err
        }
    }

    for _, instance := range data.GetInstances() {

        for key, metric := range metrics {
            label := p.target_labels.Get(key)
            value := p.target_values.Get(key)

            if x := instance.Labels.Get(label); x != "" {
                if x == value {
                    data.SetValue(metric, instance, float64(0))
                } else {
                    data.SetValue(metric, instance, float64(1))
                }
            }
        }
    }
    return nil, nil
}
