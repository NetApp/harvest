package qospolicyadaptive

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
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

func (p *QosPolicyAdaptive) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[p.Object]

	// create metrics
	for _, k := range metrics {
		err := matrix.CreateMetric(k, data)
		if err != nil {
			p.SLogger.Error("error while creating metric", slog.Any("err", err), slog.String("key", k))
		}
	}

	for _, instance := range data.GetInstances() {
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

	m := data.GetMetric(labelName)
	if m != nil {
		err = m.SetValueString(instance, xput.IOPS)
		if err != nil {
			p.SLogger.Error("Unable to set metric", slog.Any("err", err), slog.String(labelName, xput.IOPS))
		}
	}
}
