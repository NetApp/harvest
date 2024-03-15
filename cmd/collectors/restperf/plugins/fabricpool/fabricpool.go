package fabricpool

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"strconv"
	"strings"
)

type FabricPool struct {
	*plugin.AbstractPlugin
	includeConstituents bool
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FabricPool{AbstractPlugin: p}
}

func (f *FabricPool) Init() error {
	err := f.InitAbc()
	if err != nil {
		return err
	}
	if val := f.Params.GetChildContentS("include_constituents"); val != "" {
		if boolValue, err := strconv.ParseBool(val); err == nil {
			f.includeConstituents = boolValue
		}
	}
	return nil
}

// Run converts Rest lowercase metric names to uppercase to match ZapiPerf
func (f *FabricPool) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
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

	cache, err := collectors.GetFlexGroupFabricPoolMetrics(dataMap, f.Object, "cloud_bin_op", f.includeConstituents, f.Logger)
	if err != nil {
		return nil, nil, err
	}
	return []*matrix.Matrix{cache}, nil, nil
}
