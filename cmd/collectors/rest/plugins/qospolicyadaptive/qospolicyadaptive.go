package qospolicyadaptive

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	constant "github.com/netapp/harvest/v2/pkg/const"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
)

const (
	absoluteMinIOPS = "absolute_min_iops"
	expectedIOPS    = "expected_iops"
	peakIOPS        = "peak_iops"
)

var metrics = []string{
	absoluteMinIOPS,
	expectedIOPS,
	peakIOPS,
}

type QosPolicyAdaptive struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &QosPolicyAdaptive{AbstractPlugin: p}
}

func (q *QosPolicyAdaptive) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[q.Object]

	// create metrics
	for _, k := range metrics {
		err := matrix.CreateMetric(k, data)
		if err != nil {
			q.Logger.Error().Err(err).Str("key", k).Msg("error while creating metric")
		}
	}

	for _, instance := range data.GetInstances() {
		q.setIOPs(data, instance, absoluteMinIOPS)
		q.setIOPs(data, instance, expectedIOPS)
		q.setIOPs(data, instance, peakIOPS)
	}

	return nil, nil, nil
}

func (q *QosPolicyAdaptive) setIOPs(data *matrix.Matrix, instance *matrix.Instance, labelName string) {
	val := instance.GetLabel(labelName)
	xput, err := collectors.ZapiXputToRest(val)
	if err != nil {
		q.Logger.Warn().Str("label", labelName).Str("val", val).Msg("Unable to convert label, skipping")
		return
	}
	instance.SetLabel(labelName, xput.IOPS)

	m := data.GetMetric(labelName)
	if m != nil {
		err = m.SetValueString(instance, xput.IOPS)
		if err != nil {
			q.Logger.Error().Str(labelName, xput.IOPS).Err(err).Msg("Unable to set metric")
		}
	}
}

func (q *QosPolicyAdaptive) GetGeneratedMetrics() []plugin.CustomMetric {
	return []plugin.CustomMetric{
		{
			Name:         absoluteMinIOPS,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Specifies the absolute minimum IOPS that is used as an override when the expected_iops is less than this value.",
		},
		{
			Name:         expectedIOPS,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Specifies the size to be used to calculate expected IOPS per TB.",
		},
		{
			Name:         peakIOPS,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Specifies the maximum possible IOPS per TB allocated based on the storage object allocated size or the storage object used size.",
		},
	}
}
