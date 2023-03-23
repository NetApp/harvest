package qospolicyadaptive

import (
	"github.com/netapp/harvest/v2/cmd/collectors/zapi/plugins/qospolicyfixed"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type QosPolicyAdaptive struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &QosPolicyAdaptive{AbstractPlugin: p}
}

func (p *QosPolicyAdaptive) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[p.Object]
	for _, instance := range data.GetInstances() {
		p.setIOPs(instance, "absolute_min_iops")
		p.setIOPs(instance, "expected_iops")
		p.setIOPs(instance, "peak_iops")
	}

	return nil, nil
}

func (p *QosPolicyAdaptive) setIOPs(instance *matrix.Instance, labelName string) {
	val := instance.GetLabel(labelName)
	xput, err := qospolicyfixed.ZapiXputToRest(val)
	if err != nil {
		p.Logger.Warn().Str("label", labelName).Str("val", val).Msg("Unable to convert label, skipping")
		return
	}
	instance.SetLabel(labelName, xput.IOPS)
}
