package volumeunused

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"slices"
	"time"
)

type VolumeUnused struct {
	*plugin.AbstractPlugin
	currentVal    int
	client        *rest.Client
	volHistoryMap map[string]volHistory // volume-key -> volHistory map
	unused        *matrix.Matrix
}

type volHistory struct {
	svm      string
	node     string
	volume   string
	totalOps []float64
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &VolumeUnused{AbstractPlugin: p}
}

func (v *VolumeUnused) Init(remote conf.Remote) error {

	var err error

	if err := v.InitAbc(); err != nil {
		return err
	}

	v.volHistoryMap = make(map[string]volHistory)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	v.currentVal = v.SetPluginInterval()

	if v.Options.IsTest {
		return nil
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout, v.Auth); err != nil {
		v.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := v.client.Init(5, remote); err != nil {
		return err
	}

	v.unused = matrix.New(v.Parent+".VolumeUnused", "volume_unused", "volume_unused")
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "svm")
	instanceKeys.NewChildS("", "node")
	instanceKeys.NewChildS("", "volume")
	v.unused.SetExportOptions(exportOptions)
	_, err = v.unused.NewMetricFloat64("labels", "labels")
	if err != nil {
		v.SLogger.Error("add metric", slogx.Err(err))
		return err
	}
	return nil
}

func (v *VolumeUnused) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]
	v.client.Metadata.Reset()

	if v.currentVal >= v.PluginInvocationRate {
		v.currentVal = 0

		// invoke volume history data
		v.getHistoryData(data)
	}
	// Based on the unused volumes in volHistoryMap, volume_unused instances/metrics would be created
	v.handleUnusedVolumes(data.GetGlobalLabels())

	return []*matrix.Matrix{v.unused}, v.client.Metadata, nil
}

func (v *VolumeUnused) getHistoryData(data *matrix.Matrix) {
	var (
		result         []gjson.Result
		totalIopsSlice []float64
		href           string
		uuid           string
		svm            string
		nodeName       string
		vol            string
		err            error
	)

	// clear volHistoryMap
	clear(v.volHistoryMap)

	for key, volume := range data.GetInstances() {
		uuid = volume.GetLabel("uuid")
		svm = volume.GetLabel("svm")
		nodeName = volume.GetLabel("node")
		vol = volume.GetLabel("volume")
		fields := []string{"timestamp", "duration", "iops.total"}
		query := "api/storage/volumes/" + uuid + "/metrics"
		href = rest.NewHrefBuilder().
			APIPath(query).
			Fields(fields).
			MaxRecords(collectors.DefaultBatchSize).
			Build()

		if result, err = collectors.InvokeRestCall(v.client, href); err != nil {
			continue
		}

		totalIopsSlice = make([]float64, 0)
		for _, volumeHistory := range result {
			totalIops := volumeHistory.Get("iops.total").Float()
			totalIopsSlice = append(totalIopsSlice, totalIops)
		}
		v.volHistoryMap[key] = volHistory{svm: svm, node: nodeName, volume: vol, totalOps: totalIopsSlice}
	}
}

func (v *VolumeUnused) handleUnusedVolumes(globalLabels map[string]string) {
	var (
		unusedInstance *matrix.Instance
		err            error
	)

	// Purge and reset data
	v.unused.PurgeInstances()
	v.unused.Reset()

	// Set all global labels
	v.unused.SetGlobalLabels(globalLabels)

	for key, volumeHistories := range v.volHistoryMap {
		maxValue := slices.Max(volumeHistories.totalOps)
		minValue := slices.Min(volumeHistories.totalOps)
		if (maxValue - minValue) > 10 {
			continue
		}
		if unusedInstance, err = v.unused.NewInstance(key); err != nil {
			v.SLogger.Error(
				"Failed to create unused volume instance",
				slogx.Err(err),
				slog.String("key", key),
			)
			break
		}

		unusedInstance.SetLabel("volume", volumeHistories.volume)
		unusedInstance.SetLabel("svm", volumeHistories.svm)
		unusedInstance.SetLabel("node", volumeHistories.node)
		m := v.unused.GetMetric("labels")
		// populate numeric data
		m.SetValueFloat64(unusedInstance, 1.0)
	}
}
