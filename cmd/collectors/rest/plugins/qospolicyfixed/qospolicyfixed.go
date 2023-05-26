package qospolicyfixed

import (
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/qospolicyfixed"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strings"
)

var metrics = []string{
	"max_throughput_iops",
	"max_throughput_mbps",
	"min_throughput_iops",
	"min_throughput_mbps",
}

type QosPolicyFixed struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &QosPolicyFixed{AbstractPlugin: p}
}

func (p *QosPolicyFixed) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
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
		p.setFixed(data, instance)
	}

	return nil, nil
}

func (p *QosPolicyFixed) setFixed(data *matrix.Matrix, instance *matrix.Instance) {
	label := instance.GetLabel("throughput_policy")
	if label == "" {
		return
	}
	before, after, found := strings.Cut(label, "-")
	if !found {
		p.Logger.Warn().Str("label", label).Msg("Unable to parse fixed xput label")
		return
	}
	min, err := qospolicyfixed.ZapiXputToRest(before)
	if err != nil {
		p.Logger.Error().Err(err).Str("label", before).Msg("Failed to parse fixed xput label")
		return
	}
	max, err := qospolicyfixed.ZapiXputToRest(after)
	if err != nil {
		p.Logger.Error().Err(err).Str("label", after).Msg("Failed to parse fixed xput label")
		return
	}
	p.setLabelMetric(data, instance, "min_throughput_iops", min.IOPS)
	p.setLabelMetric(data, instance, "max_throughput_iops", max.IOPS)
	p.setLabelMetric(data, instance, "min_throughput_mbps", min.Mbps)
	p.setLabelMetric(data, instance, "max_throughput_mbps", max.Mbps)
}

func (p *QosPolicyFixed) setLabelMetric(data *matrix.Matrix, instance *matrix.Instance, labelName string, value string) {
	instance.SetLabel(labelName, value)
	m := data.GetMetric(labelName)
	if m != nil {
		err := m.SetValueString(instance, value)
		if err != nil {
			p.Logger.Error().Str(labelName, value).Err(err).Msg("Unable to set metric")
		}
	}
}
