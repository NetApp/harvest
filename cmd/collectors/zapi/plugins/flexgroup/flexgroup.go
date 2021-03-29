package main

import (
    "goharvest2/cmd/poller/collector/plugin"
    "goharvest2/pkg/logger"
    "goharvest2/pkg/matrix"
    "strconv"
    "strings"
)

type FlexGroup struct {
    *plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
    return &FlexGroup{AbstractPlugin: p}
}

func fetch_names(instance *matrix.Instance) (string, string) {
    var key, name, vol string

    if instance.Labels.Get("style") == "flexgroup_constituent" {
        if vol = instance.Labels.Get("volume"); len(vol) > 6 {
            name = vol[:len(vol)-6]
            key = instance.Labels.Get("svm") + "." + instance.Labels.Get("node") + "." + name
        }
    }

    return key, name
}

func (p *FlexGroup) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

    n := data.Clone(false)
    n.Plugin = "zapi.volume.flexgroup"
    n.ResetInstances()

    counts := make(map[string]int)

    // create new instance cache
    for _, i := range data.GetInstances() {

        if key, name := fetch_names(i); key != "" {

            i.Enabled = false

            if n.GetInstance(key) == nil {

                instance, err := n.AddInstance(key)

                if err != nil {
                    logger.Error(p.Prefix, err.Error())
                    return nil, err
                }
                instance.Labels = i.Labels.Copy()
                instance.Labels.Set("volume", name)

                counts[key] = 1
            } else {
                counts[key] += 1
            }
        }
    }

    logger.Debug(p.Prefix, "extracted %d flexgroup instances", len(counts))

    if err := n.InitData(); err != nil {
        logger.Error(p.Prefix, err.Error())
        return nil, err
    }

    // create summaries
    for _, i := range data.GetInstances() {
        if key, _ := fetch_names(i); key != "" {
            if instance := n.GetInstance(key); instance != nil {
                n.InstanceWiseAddition(instance, i, data)
            }
        }
    }

    // normalize percentage counters
    for key, instance := range n.GetInstances() {

        // set count as label
        count, _ := counts[key]
        instance.Labels.Set("count", strconv.Itoa(count))

        for _, metric := range n.GetMetrics() {
            if strings.Contains(metric.Name, "percent") {
                if value, has := n.GetValue(metric, instance); has {
                    n.SetValue(metric, instance, value/float64(count))
                }
            } else if metric.Name == "status" {
                if instance.Labels.Get("state") == "online" {
                    n.SetValue(metric, instance, float64(0.0))
                } else {
                    n.SetValue(metric, instance, float64(1.0))
                }
            }
        }
    }
    return []*matrix.Matrix{n}, nil
}
