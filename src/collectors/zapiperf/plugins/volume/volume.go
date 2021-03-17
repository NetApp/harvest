package main

import (
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"strings"
    "regexp"
)

type Volume struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (p *Volume) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

    re := regexp.MustCompile(`^(.*)__(\d{4})$`)

    cache := data.Clone(false)
    cache.ResetInstances()

    // create flexgroup instance cache
    for _, i := range data.GetInstances() {
        if match := re.FindStringSubmatch(i.Labels.Get("volume")); match != nil && len(match) == 3 {
            key := i.Labels.Get("node") +"."+ i.Labels.Get("svm") +"."+ match[1]
            if cache.GetInstance(key) == nil {
                fg, _ := cache.AddInstance(key)
                fg.Labels = i.Labels.Copy()
                fg.Labels.Set("volume", match[1])
                fg.Labels.Set("type", "flexgroup")
            }
        } else {
            i.Labels.Set("type", "flexvol")
        }
    }

    logger.Debug(p.Prefix, "extracted %d flexgroup volumes", len(cache.GetInstances()))

    if err := cache.InitData(); err != nil {
        return nil, err
    }

    // create summary
    for _, i := range data.GetInstances() {
        if match := re.FindStringSubmatch(i.Labels.Get("volume")); match != nil && len(match) == 3 {
            key := i.Labels.Get("node") +"."+ i.Labels.Get("svm") +"."+ match[1]
            fg := cache.GetInstance(key)
            if fg == nil {
                continue // error handling
            }

            for _, m := range data.GetMetrics() {
                value, ok := data.GetValue(m, i)
                if ! ok {
                    continue
                }
                if strings.Contains(m.Name, "_latency") {
                    cache.IncrementValue(m, fg, value)
                    continue
                }
                key := strings.Replace(m.Name, "_latency", "_ops", 1)
                if m.Name == "avg_latency" {
                    key = "total_ops"
                }
                if ops := data.GetMetric(key); ops != nil {
                    if ops_value, ok := data.GetValue(ops, i); ok {
                        cache.IncrementValue(m, fg, value * ops_value)
                    }
                }
            }
        }
    }

    // normalize latency values
    for _, i := range cache.GetInstances() {
        for _, m := range cache.GetMetrics() {
            if strings.Contains(m.Name, "_latency") {

                value, ok := cache.GetValue(m, i)
                if ! ok {
                    continue
                }
                key := strings.Replace(m.Name, "_latency", "_ops", 1)
                if m.Name == "avg_latency" {
                    key = "total_ops"
                }
                if ops := cache.GetMetric(key); ops != nil {
                    if ops_value, ok := cache.GetValue(ops, i); ok {
                        cache.SetValue(m, i, value / ops_value)
                    }
                }
            }
        }
    }

	return []*matrix.Matrix{cache}, nil
}
