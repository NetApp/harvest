package workload

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
)

var metrics = []string{
	"max_throughput_iops",
	"max_throughput_mbps",
	"min_throughput_iops",
	"min_throughput_mbps",
}

type Workload struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Workload{AbstractPlugin: p}
}

func (w *Workload) Init(conf.Remote) error {
	return w.InitAbc()
}

func (w *Workload) createMetrics(data *matrix.Matrix) error {
	if err := data.NewMetricsFloat64(metrics...); err != nil {
		w.SLogger.Warn("error while creating metric", slogx.Err(err))
		return err
	}
	return nil
}

func (w *Workload) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[w.Object]
	if err := w.createMetrics(data); err != nil {
		return nil, nil, err
	}

	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		collectors.SetThroughput(data, instance, "max_xput", "max_throughput_iops", "max_throughput_mbps", w.SLogger)
		collectors.SetThroughput(data, instance, "min_xput", "min_throughput_iops", "min_throughput_mbps", w.SLogger)
	}
	return nil, nil, nil
}
