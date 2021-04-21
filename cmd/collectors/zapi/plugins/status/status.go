//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package main

import (
    "goharvest2/cmd/poller/collector/plugin"
    "goharvest2/pkg/dict"
    "goharvest2/pkg/errors"
    "goharvest2/pkg/logger"
    "goharvest2/pkg/matrix"
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
    }
    logger.Debug(p.Prefix, "initialized plugin, will cook status metrics from %d instance labels", p.target_labels.Size())
    return nil
}

func (p *Status) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

    var err error
    metrics := make(map[string]*matrix.Metric)

    for key := range p.target_labels.Iter() {
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
