package qospolicyadaptive

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
)

var metrics = []string{
	"absolute_min_iops",
	"expected_iops",
	"peak_iops",
}

type QosPolicyAdaptive struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &QosPolicyAdaptive{AbstractPlugin: p}
}

func (p *QosPolicyAdaptive) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[p.Object]
	if err := data.NewMetricsFloat64(metrics...); err != nil {
		p.SLogger.Error("error while creating metric", slogx.Err(err))
		return nil, nil, err
	}

	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		p.setIOPs(data, instance, "absolute_min_iops")
		p.setIOPs(data, instance, "expected_iops")
		p.setIOPs(data, instance, "peak_iops")
	}
	return nil, nil, nil
}

func (p *QosPolicyAdaptive) setIOPs(data *matrix.Matrix, instance *matrix.Instance, labelName string) {
	val := instance.GetLabel(labelName)
	xput, err := collectors.ZapiXputToRest(val)
	if err != nil {
		p.SLogger.Warn("Unable to convert label, skipping", slog.String("label", labelName), slog.String("val", val))
		return
	}
	instance.SetLabel(labelName, xput.IOPS)

	if metric := data.GetMetric(labelName); metric != nil {
		if err = metric.SetValueString(instance, xput.IOPS); err != nil {
			p.SLogger.Error("Unable to set metric", slogx.Err(err), slog.String(labelName, xput.IOPS))
		}
	}
}
