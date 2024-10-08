package qospolicyfixed

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
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

func (q *QosPolicyFixed) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[q.Object]

	// create metrics
	for _, k := range metrics {
		err := matrix.CreateMetric(k, data)
		if err != nil {
			q.SLogger.Error("error while creating metric", slogx.Err(err), slog.String("key", k))
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
		q.SLogger.Warn("Unable to parse fixed xput label", slog.String("label", label))
		return
	}
	minV, err := collectors.ZapiXputToRest(before)
	if err != nil {
		q.SLogger.Error("Failed to parse fixed xput label", slogx.Err(err), slog.String("label", before))
		return
	}
	maxV, err := collectors.ZapiXputToRest(after)
	if err != nil {
		q.SLogger.Error("Failed to parse fixed xput label", slogx.Err(err), slog.String("label", after))
		return
	}
	collectors.QosSetLabel("min_throughput_iops", data, instance, minV.IOPS, q.SLogger)
	collectors.QosSetLabel("max_throughput_iops", data, instance, maxV.IOPS, q.SLogger)
	collectors.QosSetLabel("min_throughput_mbps", data, instance, minV.Mbps, q.SLogger)
	collectors.QosSetLabel("max_throughput_mbps", data, instance, maxV.Mbps, q.SLogger)
}
