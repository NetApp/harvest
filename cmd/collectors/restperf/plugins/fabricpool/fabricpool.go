package fabricpool

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strings"
)

type FabricPool struct {
	*plugin.AbstractPlugin
	includeConstituents bool
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FabricPool{AbstractPlugin: p}
}

func (f *FabricPool) Init(conf.Remote) error {
	err := f.InitAbc()
	if err != nil {
		return err
	}
	f.includeConstituents = f.LoadParam("include_constituents", f.includeConstituents)
	return nil
}

// Run converts Rest lowercase metric names to uppercase to match ZapiPerf
func (f *FabricPool) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
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

	// GetFlexGroupFabricPoolMetrics looks the ops counter up by matrix key, which is the
	// counter's ONTAP name. RestPerf's name is cloud_bin_op and only its display is renamed
	// to cloud_bin_operation; CmPerf's counter is named cloud_bin_operation.
	opName := "cloud_bin_op"
	if f.IsCmPerfCollector() {
		opName = "cloud_bin_operation"
	}

	cache, err := collectors.GetFlexGroupFabricPoolMetrics(dataMap, f.Object, opName, f.includeConstituents, f.SLogger)
	if err != nil {
		return nil, nil, err
	}
	return []*matrix.Matrix{cache}, nil, nil
}
