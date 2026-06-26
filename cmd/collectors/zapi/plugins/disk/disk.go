// Copyright NetApp Inc, 2021 All rights reserved

package disk

import (
	"strings"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
)

type Disk struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Disk{AbstractPlugin: p}
}

func (d *Disk) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[d.Object]

	pathCountMetric := data.GetMetric("path_count")
	if pathCountMetric == nil {
		var err error
		if pathCountMetric, err = data.NewMetricFloat64("path_count"); err != nil {
			d.SLogger.Error("add metric path_count", slogx.Err(err))
			return nil, nil, err
		}
	}

	for _, instance := range data.GetInstances() {
		pathList := instance.GetLabel("path_list")
		var pathCount float64
		if pathList != "" {
			pathCount = float64(strings.Count(pathList, ",") + 1)
		}
		pathCountMetric.SetValueFloat64(instance, pathCount)

		// Normalize pool label: ZAPI returns "0"/"1", REST returns "pool0"/"pool1"
		if rawPool := instance.GetLabel("pool"); rawPool != "" {
			instance.SetLabel("pool", "pool"+rawPool)
		}
	}

	return nil, nil, nil
}
