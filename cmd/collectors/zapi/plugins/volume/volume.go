package volume

import (
	"errors"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"strconv"
	"strings"
)

type Volume struct {
	*plugin.AbstractPlugin
	currentVal          int
	client              *zapi.Client
	aggrsMap            map[string]bool // aggregate-name -> exist map
	includeConstituents bool
}

type volumeClone struct {
	name           string
	svm            string
	parentSnapshot string
	parentVolume   string
	parentSvm      string
	splitEstimate  string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) Init(remote conf.Remote) error {

	var err error

	if err := v.InitAbc(); err != nil {
		return err
	}

	if v.client, err = zapi.New(conf.ZapiPoller(v.ParentParams), v.Auth); err != nil {
		v.SLogger.Error("connecting", slogx.Err(err))

		return err
	}

	if err := v.client.Init(5, remote); err != nil {
		return err
	}

	v.aggrsMap = make(map[string]bool)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	v.currentVal = v.SetPluginInterval()

	// Read template to decide inclusion of flexgroup constituents
	v.includeConstituents = collectors.ReadPluginKey(v.Params, "include_constituents")
	return nil
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	data := dataMap[v.Object]
	v.client.Metadata.Reset()

	if v.currentVal >= v.PluginInvocationRate {
		v.currentVal = 0

		// invoke disk-encrypt-get-iter zapi and populate disk info
		disks, err1 := v.getEncryptedDisks()
		// invoke aggr-status-get-iter zapi and populate aggr disk mapping info
		aggrDiskMap, err2 := v.getAggrDiskMapping()

		if err1 != nil {
			if errors.Is(err1, errs.ErrNoInstance) {
				v.SLogger.Debug("Failed to collect disk data", slog.Any("err", err1))
			} else {
				v.SLogger.Error("Failed to collect disk data", slog.Any("err", err1))
			}
		}
		if err2 != nil {
			if errors.Is(err2, errs.ErrNoInstance) {
				v.SLogger.Debug("Failed to collect aggregate-disk mapping data", slog.Any("err", err2))
			} else {
				v.SLogger.Error("Failed to collect aggregate-disk mapping data", slog.Any("err", err2))
			}
		}
		// update aggrsMap based on disk data and addr disk mapping
		v.updateAggrMap(disks, aggrDiskMap)
	}

	volumeCloneMap, err := v.getVolumeCloneInfo()
	if err != nil {
		v.SLogger.Error("Failed to update clone data", slogx.Err(err))
	}

	volumeFootprintMap, err := v.getVolumeFootprint()
	if err != nil {
		v.SLogger.Error("Failed to update footprint data", slogx.Err(err))
		// clean the map in case of the error
		clear(volumeFootprintMap)
	}

	flexgroupFootPrintMatrix := v.processAndUpdateVolume(data, volumeFootprintMap, volumeCloneMap)

	v.currentVal++
	return []*matrix.Matrix{flexgroupFootPrintMatrix}, v.client.Metadata, nil
}

func (v *Volume) processAndUpdateVolume(data *matrix.Matrix, volumeFootprintMap map[string]map[string]string, volumeCloneMap map[string]volumeClone) *matrix.Matrix {
	var err error
	// Handling volume footprint metrics and updating volume labels
	for _, volume := range data.GetInstances() {
		name := volume.GetLabel("volume")
		svm := volume.GetLabel("svm")
		volState := volume.GetLabel("state")
		key := name + svm

		// Process volume footprint metrics
		if vf, ok := volumeFootprintMap[key]; ok {
			for vfKey, vfVal := range vf {
				vfMetric := data.GetMetric(vfKey)
				if vfMetric == nil {
					if vfMetric, err = data.NewMetricFloat64(vfKey); err != nil {
						v.SLogger.Error("add metric", slogx.Err(err), slog.String("metric", vfKey))
						continue
					}
				}

				if vfVal != "" {
					err := vfMetric.SetValueString(volume, vfVal)
					if err != nil {
						v.SLogger.Error("parse", slogx.Err(err), slog.String(vfKey, vfVal))
						continue
					}
				}
			}
		}

		// Update volume labels
		if !volume.IsExportable() {
			continue
		}

		// ZAPI includes node root and temp volumes, while REST does not. To make ZAPI and REST consistent, Harvest will exclude the node root and temp volumes by not exporting them.
		if volume.GetLabel("node_root") == "true" || volume.GetLabel("type") == "tmp" {
			volume.SetExportable(false)
			continue
		}

		if volState == "offline" && strings.HasSuffix(svm, "-mc") {
			volume.SetExportable(false)
		}

		if volume.GetLabel("style") == "flexgroup_constituent" {
			volume.SetExportable(v.includeConstituents)
		}

		volume.SetLabel("isHardwareEncrypted", strconv.FormatBool(v.aggrsMap[volume.GetLabel("aggr")]))

		if vc, ok := volumeCloneMap[key]; ok {
			volume.SetLabel("clone_parent_snapshot", vc.parentSnapshot)
			volume.SetLabel("clone_parent_volume", vc.parentVolume)
			volume.SetLabel("clone_parent_svm", vc.parentSvm)
			splitEstimate := data.GetMetric("clone_split_estimate")
			if splitEstimate == nil {
				if splitEstimate, err = data.NewMetricFloat64("clone_split_estimate"); err != nil {
					v.SLogger.Error(
						"Failed to add metric",
						slogx.Err(err),
						slog.String("metric", "clone_split_estimate"),
					)
					continue
				}
			}

			if vc.splitEstimate == "" {
				continue
			}
			// splitEstimate is 4KB blocks, Convert to bytes as in REST
			var splitEstimateBytes float64
			if splitEstimateBytes, err = strconv.ParseFloat(vc.splitEstimate, 64); err != nil {
				v.SLogger.Error(
					"Failed to parse clone_split_estimate",
					slogx.Err(err),
					slog.String("clone_split_estimate", vc.splitEstimate),
				)
				continue
			}
			splitEstimateBytes = splitEstimateBytes * 4 * 1024
			splitEstimate.SetValueFloat64(volume, splitEstimateBytes)
		}
	}
	return collectors.ProcessFlexGroupFootPrint(data, v.SLogger)
}

func (v *Volume) getVolumeCloneInfo() (map[string]volumeClone, error) {
	var (
		result         []*node.Node
		volumeCloneMap map[string]volumeClone
		err            error
	)

	volumeCloneMap = make(map[string]volumeClone)
	request := node.NewXMLS("volume-clone-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	if result, err = v.client.InvokeZapiCall(request); err != nil {
		return volumeCloneMap, err
	}

	if len(result) == 0 || result == nil {
		return volumeCloneMap, nil
	}

	for _, clone := range result {
		name := clone.GetChildContentS("volume")
		vserver := clone.GetChildContentS("vserver")
		parentSnapshot := clone.GetChildContentS("parent-snapshot")
		parentVolume := clone.GetChildContentS("parent-volume")
		parentSvm := clone.GetChildContentS("parent-vserver")
		splitEstimate := clone.GetChildContentS("split-estimate")
		volC := volumeClone{
			name:           name,
			svm:            vserver,
			parentSnapshot: parentSnapshot,
			parentVolume:   parentVolume,
			parentSvm:      parentSvm,
			splitEstimate:  splitEstimate,
		}
		key := volC.name + volC.svm
		volumeCloneMap[key] = volC
	}

	return volumeCloneMap, nil
}

func (v *Volume) getVolumeFootprint() (map[string]map[string]string, error) {
	var (
		result             []*node.Node
		volumeFootprintMap map[string]map[string]string
		err                error
	)

	volumeFootprintMap = make(map[string]map[string]string)
	request := node.NewXMLS("volume-footprint-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	desired := node.NewXMLS("desired-attributes")
	footprintInfo := node.NewXMLS("footprint-info")
	footprintInfo.NewChildS("volume", "")
	footprintInfo.NewChildS("vserver", "")
	footprintInfo.NewChildS("delayed-free-footprint", "")
	footprintInfo.NewChildS("flexvol-metadata-footprint", "")
	footprintInfo.NewChildS("total-footprint", "")
	footprintInfo.NewChildS("total-metadata-footprint", "")
	footprintInfo.NewChildS("volume-blocks-footprint-bin0", "")
	footprintInfo.NewChildS("volume-blocks-footprint-bin0-percent", "")
	footprintInfo.NewChildS("volume-blocks-footprint-bin1", "")
	footprintInfo.NewChildS("volume-blocks-footprint-bin1-percent", "")
	footprintInfo.NewChildS("volume-guarantee-footprint", "")
	desired.AddChild(footprintInfo)
	request.AddChild(desired)

	if result, err = v.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return volumeFootprintMap, nil
	}

	for _, footprint := range result {
		footprintMetrics := make(map[string]string)
		volume := footprint.GetChildContentS("volume")
		svm := footprint.GetChildContentS("vserver")
		performanceTierFootprint := footprint.GetChildContentS("volume-blocks-footprint-bin0")
		performanceTierFootprintPerc := footprint.GetChildContentS("volume-blocks-footprint-bin0-percent")
		capacityTierFootprint := footprint.GetChildContentS("volume-blocks-footprint-bin1")
		capacityTierFootprintPerc := footprint.GetChildContentS("volume-blocks-footprint-bin1-percent")
		delayedFreeFootprint := footprint.GetChildContentS("delayed-free-footprint")
		metadataFootprint := footprint.GetChildContentS("flexvol-metadata-footprint")
		totalFootprint := footprint.GetChildContentS("total-footprint")
		totalMetadataFootprint := footprint.GetChildContentS("total-metadata-footprint")
		volumeBlocksFootprint := footprint.GetChildContentS("volume-guarantee-footprint")

		footprintMetrics["performance_tier_footprint"] = performanceTierFootprint
		footprintMetrics["performance_tier_footprint_percent"] = performanceTierFootprintPerc
		footprintMetrics["capacity_tier_footprint"] = capacityTierFootprint
		footprintMetrics["capacity_tier_footprint_percent"] = capacityTierFootprintPerc
		footprintMetrics["delayed_free_footprint"] = delayedFreeFootprint
		footprintMetrics["metadata_footprint"] = metadataFootprint
		footprintMetrics["total_footprint"] = totalFootprint
		footprintMetrics["total_metadata_footprint"] = totalMetadataFootprint
		footprintMetrics["guarantee_footprint"] = volumeBlocksFootprint

		volumeFootprintMap[volume+svm] = footprintMetrics
	}

	return volumeFootprintMap, nil
}

func (v *Volume) getEncryptedDisks() ([]string, error) {
	var (
		result []*node.Node
		err    error
	)

	request := node.NewXMLS("disk-encrypt-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	// algorithm is -- Protection mode needs to be DATA or FULL
	// Fetching rest of them and add as
	query := request.NewChildS("query", "")
	encryptInfoQuery := query.NewChildS("disk-encrypt-info", "")
	encryptInfoQuery.NewChildS("protection-mode", "open|part|miss")

	// fetching only disks whose protection-mode is open/part/miss
	if result, err = v.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	diskNames := make([]string, 0, len(result))
	for _, disk := range result {
		diskName := disk.GetChildContentS("disk-name")
		diskNames = append(diskNames, diskName)
	}
	return diskNames, nil
}

func (v *Volume) updateAggrMap(disks []string, aggrDiskMap map[string][]string) {
	if disks != nil && aggrDiskMap != nil {
		// Clean aggrsMap map
		clear(v.aggrsMap)
		for _, disk := range disks {
			if aggrList, exist := aggrDiskMap[disk]; exist {
				for _, aggr := range aggrList {
					v.aggrsMap[aggr] = true
				}
			}
		}
	}
}

func (v *Volume) getAggrDiskMapping() (map[string][]string, error) {
	var (
		result        []*node.Node
		aggrsDisksMap map[string][]string
		diskName      string
		err           error
	)

	request := node.NewXMLS("aggr-status-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	aggrsDisksMap = make(map[string][]string)

	if result, err = v.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, aggrDiskData := range result {
		aggrName := aggrDiskData.GetChildContentS("aggregate")
		for _, plexList := range aggrDiskData.GetChildS("aggr-plex-list").GetChildren() {
			for _, raidGroupList := range plexList.GetChildS("aggr-raidgroup-list").GetChildren() {
				for _, diskList := range raidGroupList.GetChildS("aggr-disk-list").GetChildren() {
					diskName = diskList.GetChildContentS("disk")
					aggrsDisksMap[diskName] = append(aggrsDisksMap[diskName], aggrName)
				}
			}
		}
	}
	return aggrsDisksMap, nil
}
