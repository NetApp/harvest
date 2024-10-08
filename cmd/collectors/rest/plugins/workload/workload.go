package workload

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
)

var metrics = []string{
	"max_throughput_iops",
	"max_throughput_mbps",
}

type Workload struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Workload{AbstractPlugin: p}
}

func (w *Workload) Init() error {
	if err := w.InitAbc(); err != nil {
		return err
	}
	return nil
}

func (w *Workload) createMetrics(data *matrix.Matrix) error {
	for _, k := range metrics {
		err := matrix.CreateMetric(k, data)
		if err != nil {
			w.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", k))
			return err
		}
	}
	return nil
}

func (w *Workload) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[w.Object]

	// create metrics
	err := w.createMetrics(data)
	if err != nil {
		return nil, nil, err
	}

	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		collectors.SetThroughput(data, instance, "max_xput", "max_throughput_iops", "max_throughput_mbps", w.SLogger)
	}

	return nil, nil, nil
}
