package volume

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
)

type Volume struct {
	*plugin.AbstractPlugin
	includeConstituents bool
	volumesMap          map[string]string // volume-name -> volume-style map
	zombieVolumeMatrix  *matrix.Matrix
	volumePastOpsMap    map[string]collectors.OpsData
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) Init(conf.Remote) error {
	if err := v.InitAbc(); err != nil {
		return err
	}

	v.volumesMap = make(map[string]string)

	// Read template to decide inclusion of flexgroup constituents
	v.includeConstituents = collectors.ReadPluginKey(v.Params, "include_constituents")

	v.zombieVolumeMatrix = matrix.New(".Volume", "volume_zombie", "volume_zombie")
	metricName := "exist"
	_, err := v.zombieVolumeMatrix.NewMetricFloat64(metricName)
	if err != nil {
		v.SLogger.Error("add metric", slogx.Err(err), slog.String("key", metricName))
		return err
	}

	return nil
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]
	style := "style"
	opsKeyPrefix := "temp_"
	v.volumesMap = v.getVolumeMap(data)
	return collectors.ProcessFlexGroupData(v.SLogger, data, style, v.includeConstituents, opsKeyPrefix, v.volumesMap, false, v.zombieVolumeMatrix, v.volumePastOpsMap, "exist")
}

func (v *Volume) getVolumeMap(data *matrix.Matrix) map[string]string {
	volumesMap := make(map[string]string)
	for _, instance := range data.GetInstances() {
		style := instance.GetLabel("style")
		name := instance.GetLabel("volume")
		svm := instance.GetLabel("svm")
		volumesMap[svm+name] = style
	}
	return volumesMap
}
