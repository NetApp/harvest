package workload

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	constant "github.com/netapp/harvest/v2/pkg/const"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
)

const (
	maxThroughputIOPS = "max_throughput_iops"
	maxThroughputMBPS = "max_throughput_mbps"
)

var metrics = []string{
	maxThroughputIOPS,
	maxThroughputMBPS,
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
			w.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
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
		collectors.SetThroughput(data, instance, "max_xput", maxThroughputIOPS, maxThroughputMBPS, w.Logger)
	}

	return nil, nil, nil
}

func (w *Workload) GetGeneratedMetrics() []plugin.CustomMetric {

	return []plugin.CustomMetric{
		{
			Name:         maxThroughputIOPS,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Maximum throughput IOPs allowed for the workload.",
		},
		{
			Name:         maxThroughputMBPS,
			Endpoint:     "NA",
			ONTAPCounter: constant.HarvestGenerated,
			Description:  "Maximum throughput Mbps allowed for the workload.",
		},
	}
}
