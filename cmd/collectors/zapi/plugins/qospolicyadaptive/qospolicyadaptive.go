package qospolicyadaptive

import (
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/qospolicyfixed"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
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

func (p *QosPolicyAdaptive) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[p.Object]
	// create metrics
	for _, k := range metrics {
		err := matrix.CreateMetric(k, data)
		if err != nil {
			p.Logger.Error().Err(err).Str("key", k).Msg("error while creating metric")
			return nil, err
		}
	}

	for _, instance := range data.GetInstances() {
		p.setIOPs(data, instance, "absolute_min_iops")
		p.setIOPs(data, instance, "expected_iops")
		p.setIOPs(data, instance, "peak_iops")
	}

	return nil, nil
}

func (p *QosPolicyAdaptive) setIOPs(data *matrix.Matrix, instance *matrix.Instance, labelName string) {
	val := instance.GetLabel(labelName)
	xput, err := qospolicyfixed.ZapiXputToRest(val)
	if err != nil {
		p.Logger.Warn().Str("label", labelName).Str("val", val).Msg("Unable to convert label, skipping")
		return
	}
	instance.SetLabel(labelName, xput.IOPS)

	m := data.GetMetric(labelName)
	if m != nil {
		err = m.SetValueString(instance, xput.IOPS)
		if err != nil {
			p.Logger.Error().Str(labelName, xput.IOPS).Err(err).Msg("Unable to set metric")
		}
	}
}
