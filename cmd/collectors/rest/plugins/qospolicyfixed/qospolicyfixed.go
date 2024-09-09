package qospolicyfixed

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	constant "github.com/netapp/harvest/v2/pkg/const"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"strings"
)

const (
	maxThroughputIOPS = "max_throughput_iops"
	maxThroughputMBPS = "max_throughput_mbps"
	minThroughputIOPS = "min_throughput_iops"
	minThroughputMBPS = "min_throughput_mbps"
)

var metrics = []string{
	maxThroughputIOPS,
	maxThroughputMBPS,
	minThroughputIOPS,
	minThroughputMBPS,
}

type QosPolicyFixed struct {
	*plugin.AbstractPlugin
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
	collectors.QosSetLabel(minThroughputIOPS, data, instance, minV.IOPS, q.Logger)
	collectors.QosSetLabel(maxThroughputIOPS, data, instance, maxV.IOPS, q.Logger)
	collectors.QosSetLabel(minThroughputMBPS, data, instance, minV.Mbps, q.Logger)
	collectors.QosSetLabel(maxThroughputMBPS, data, instance, maxV.Mbps, q.Logger)
}

func (q *QosPolicyFixed) GetGeneratedMetrics() []plugin.CustomMetric {

	return []plugin.CustomMetric{
		{
			Name:         maxThroughputIOPS,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Maximum throughput defined by this policy. It is specified in terms of IOPS. 0 means no maximum throughput is enforced.",
		},
		{
			Name:         maxThroughputMBPS,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Maximum throughput defined by this policy. It is specified in terms of Mbps. 0 means no maximum throughput is enforced.",
		},
		{
			Name:         minThroughputIOPS,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Minimum throughput defined by this policy. It is specified in terms of IOPS. 0 means no minimum throughput is enforced. These floors are not guaranteed on non-AFF platforms or when FabricPool tiering policies are set.",
		},
		{
			Name:         minThroughputMBPS,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Minimum throughput defined by this policy. It is specified in terms of Mbps. 0 means no minimum throughput is enforced.",
		},
	}
}
