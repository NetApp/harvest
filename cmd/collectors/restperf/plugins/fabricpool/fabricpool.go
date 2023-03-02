package fabricpool

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strings"
)

type FabricPool struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FabricPool{AbstractPlugin: p}
}

// Run converts Rest lowercase metric names to uppercase to match ZapiPerf
func (f *FabricPool) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	verbs := []string{
		"delete",
		"get",
		"head",
		"list",
		"put",
		"unknown",
	}
	metrics := []string{
		"cloud_bin_op#",
		"cloud_bin_op_latency_average#",
	}

	for _, verb := range verbs {
		for _, metricName := range metrics {
			metric := data.GetMetric(metricName + verb)
			if metric == nil {
				continue
			}
			v := metric.GetLabel("metric")
			metric.SetLabel("metric", strings.ToUpper(v))
		}
	}

	return nil, nil
}
