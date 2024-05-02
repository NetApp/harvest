package qospolicyfixed

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
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
	collectors.QosCommon
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &QosPolicyFixed{AbstractPlugin: p}
}

func (q *QosPolicyFixed) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[q.Object]

	// create metrics
	for _, k := range metrics {
		err := matrix.CreateMetric(k, data)
		if err != nil {
			q.Logger.Error().Err(err).Str("key", k).Msg("error while creating metric")
			return nil, nil, err
		}
	}

	for _, instance := range data.GetInstances() {
		q.setFixed(data, instance)
	}

	return nil, nil, nil
}

func (q *QosPolicyFixed) setFixed(data *matrix.Matrix, instance *matrix.Instance) {
	label := instance.GetLabel("throughput_policy")
	if label == "" {
		return
	}
	before, after, found := strings.Cut(label, "-")
	if !found {
		q.Logger.Warn().Str("label", label).Msg("Unable to parse fixed xput label")
		return
	}
	minV, err := collectors.ZapiXputToRest(before)
	if err != nil {
		q.Logger.Error().Err(err).Str("label", before).Msg("Failed to parse fixed xput label")
		return
	}
	maxV, err := collectors.ZapiXputToRest(after)
	if err != nil {
		q.Logger.Error().Err(err).Str("label", after).Msg("Failed to parse fixed xput label")
		return
	}
	q.SetLabel("min_throughput_iops", data, instance, minV.IOPS, q.Logger)
	q.SetLabel("max_throughput_iops", data, instance, maxV.IOPS, q.Logger)
	q.SetLabel("min_throughput_mbps", data, instance, minV.Mbps, q.Logger)
	q.SetLabel("max_throughput_mbps", data, instance, maxV.Mbps, q.Logger)
}
