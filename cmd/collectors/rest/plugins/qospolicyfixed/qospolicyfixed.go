package qospolicyfixed

import (
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/qospolicyfixed"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strings"
)

type QosPolicyFixed struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &QosPolicyFixed{AbstractPlugin: p}
}

func (p *QosPolicyFixed) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[p.Object]

	for _, instance := range data.GetInstances() {
		p.setFixed(instance)
	}

	return nil, nil
}

func (p *QosPolicyFixed) setFixed(instance *matrix.Instance) {
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
	instance.SetLabel("min_throughput_iops", min.IOPS)
	instance.SetLabel("max_throughput_iops", max.IOPS)
	instance.SetLabel("min_throughput_mbps", min.Mbps)
	instance.SetLabel("max_throughput_mbps", max.Mbps)
}
