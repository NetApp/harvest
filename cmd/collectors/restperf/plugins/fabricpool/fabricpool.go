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
func (f *FabricPool) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[f.Object]
	for _, metric := range data.GetMetrics() {
		if !metric.IsArray() {
			continue
		}
		v := metric.GetLabel("metric")
		if v != "" {
			metric.SetLabel("metric", strings.ToUpper(v))
		}
	}

	return nil, nil
}
