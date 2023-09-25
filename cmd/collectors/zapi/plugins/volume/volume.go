package volume

import (
	"errors"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strconv"
)

type Volume struct {
	*plugin.AbstractPlugin
	currentVal int
	client     *zapi.Client
	aggrsMap   map[string]string // aggregate-uuid -> aggregate-name map
}

type aggrData struct {
	uuid string
	name string
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

func (v *Volume) Init() error {

	var err error

	if err = v.InitAbc(); err != nil {
		return err
	}

	if v.client, err = zapi.New(conf.ZapiPoller(v.ParentParams), v.Auth); err != nil {
		v.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = v.client.Init(5); err != nil {
		return err
	}

	v.aggrsMap = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	v.currentVal = v.SetPluginInterval()

	return nil
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {

	data := dataMap[v.Object]
	if v.currentVal >= v.PluginInvocationRate {
		v.currentVal = 0

		// invoke disk-encrypt-get-iter zapi and populate disk info
		disks, err1 := v.getEncryptedDisks()
		// invoke aggr-status-get-iter zapi and populate aggr disk mapping info
		aggrDiskMap, err2 := v.getAggrDiskMapping()

		if err1 != nil {
			if errors.Is(err1, errs.ErrNoInstance) {
				v.Logger.Debug().Err(err1).Msg("Failed to collect disk data")
			} else {
				v.Logger.Error().Err(err1).Msg("Failed to collect disk data")
			}
		}
		if err2 != nil {
			if errors.Is(err2, errs.ErrNoInstance) {
				v.Logger.Debug().Err(err2).Msg("Failed to collect aggregate-disk mapping data")
			} else {
				v.Logger.Error().Err(err2).Msg("Failed to collect aggregate-disk mapping data")
			}
		}
		// update aggrsMap based on disk data and addr disk mapping
		v.updateAggrMap(disks, aggrDiskMap)
	}

	volumeCloneMap, err := v.getVolumeCloneInfo()
	if err != nil {
		v.Logger.Error().Err(err).Msg("Failed to update clone data")
	}

	volumeFootprintMap, err := v.getVolumeFootprint()
	if err != nil {
		v.Logger.Error().Err(err).Msg("Failed to update footprint data")
	}

	// update volume instance labels
	v.updateVolumeLabels(data, volumeCloneMap, volumeFootprintMap)

	v.currentVal++
	return nil, nil
}

func (v *Volume) updateVolumeLabels(data *matrix.Matrix, volumeCloneMap map[string]volumeClone, volumeFootprintMap map[string]map[string]string) {
	var err error
	for _, volume := range data.GetInstances() {
		if !volume.IsExportable() {
			continue
		}
		aggrUUID := volume.GetLabel("aggrUuid")
		_, exist := v.aggrsMap[aggrUUID]
		volume.SetLabel("isHardwareEncrypted", strconv.FormatBool(exist))

		name := volume.GetLabel("volume")
		svm := volume.GetLabel("svm")
		key := name + svm

		if vc, ok := volumeCloneMap[key]; ok {
			volume.SetLabel("clone_parent_snapshot", vc.parentSnapshot)
			volume.SetLabel("clone_parent_volume", vc.parentVolume)
			volume.SetLabel("clone_parent_svm", vc.parentSvm)
			splitEstimate := data.GetMetric("clone_split_estimate")
			if splitEstimate == nil {
				if splitEstimate, err = data.NewMetricFloat64("clone_split_estimate"); err != nil {
					v.Logger.Error().Err(err).Str("metric", "clone_split_estimate").Msg("add metric")
					continue
				}
			}

			// splitEstimate is 4KB blocks, Convert to bytes as in REST

			var splitEstimateBytes float64
			if splitEstimateBytes, err = strconv.ParseFloat(vc.splitEstimate, 64); err != nil {
				v.Logger.Error().Err(err).Str("clone_split_estimate", vc.splitEstimate).Msg("parse clone_split_estimate")
				continue
			}
			splitEstimateBytes = splitEstimateBytes * 4 * 1024
			if err = splitEstimate.SetValueFloat64(volume, splitEstimateBytes); err != nil {
				v.Logger.Error().Err(err).Str("clone_split_estimate", vc.splitEstimate).Msg("set clone_split_estimate")
				continue
			}
		}

		// Handling volume footprint metrics
		if vf, ok := volumeFootprintMap[key]; ok {
			for vfKey, vfVal := range vf {
				vfMetric := data.GetMetric(vfKey)
				if vfMetric == nil {
					if vfMetric, err = data.NewMetricFloat64(vfKey); err != nil {
						v.Logger.Error().Err(err).Str("metric", vfKey).Msg("add metric")
						continue
					}
				}

				if vfVal != "" {
					vfMetricVal, err := strconv.ParseFloat(vfVal, 64)
					if err != nil {
						v.Logger.Error().Err(err).Str(vfKey, vfVal).Msg("parse")
						continue
					}
					if err = vfMetric.SetValueFloat64(volume, vfMetricVal); err != nil {
						v.Logger.Error().Err(err).Str(vfKey, vfVal).Msg("set")
						continue
					}
				}
			}
		}
	}
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
	footprintInfo.NewChildS("volume-blocks-footprint-bin0", "")
	footprintInfo.NewChildS("volume-blocks-footprint-bin0-percent", "")
	footprintInfo.NewChildS("volume-blocks-footprint-bin1", "")
	footprintInfo.NewChildS("volume-blocks-footprint-bin1-percent", "")
	desired.AddChild(footprintInfo)
	request.AddChild(desired)

	if result, err = v.client.InvokeZapiCall(request); err != nil {
		return volumeFootprintMap, err
	}

	if len(result) == 0 {
		return volumeFootprintMap, nil
	}

	for _, footprint := range result {
		footprintMatrics := make(map[string]string)
		volume := footprint.GetChildContentS("volume")
		svm := footprint.GetChildContentS("vserver")
		performanceTierFootprint := footprint.GetChildContentS("volume-blocks-footprint-bin0")
		performanceTierFootprintPerc := footprint.GetChildContentS("volume-blocks-footprint-bin0-percent")
		capacityTierFootprint := footprint.GetChildContentS("volume-blocks-footprint-bin1")
		capacityTierFootprintPerc := footprint.GetChildContentS("volume-blocks-footprint-bin1-percent")
		footprintMatrics["performance_tier_footprint"] = performanceTierFootprint
		footprintMatrics["performance_tier_footprint_percent"] = performanceTierFootprintPerc
		footprintMatrics["capacity_tier_footprint"] = capacityTierFootprint
		footprintMatrics["capacity_tier_footprint_percent"] = capacityTierFootprintPerc
		volumeFootprintMap[volume+svm] = footprintMatrics
	}

	return volumeFootprintMap, nil
}

func (v *Volume) getEncryptedDisks() ([]string, error) {
	var (
		result    []*node.Node
		diskNames []string
		err       error
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

	for _, disk := range result {
		diskName := disk.GetChildContentS("disk-name")
		diskNames = append(diskNames, diskName)
	}
	return diskNames, nil
}

func (v *Volume) updateAggrMap(disks []string, aggrDiskMap map[string]aggrData) {
	if disks != nil && aggrDiskMap != nil {
		// Clean aggrsMap map
		v.aggrsMap = make(map[string]string)

		for _, disk := range disks {
			aggr := aggrDiskMap[disk]
			v.aggrsMap[aggr.uuid] = aggr.name
		}
	}
}

func (v *Volume) getAggrDiskMapping() (map[string]aggrData, error) {
	var (
		result        []*node.Node
		aggrsDisksMap map[string]aggrData
		diskName      string
		err           error
	)

	request := node.NewXMLS("aggr-status-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	aggrsDisksMap = make(map[string]aggrData)

	if result, err = v.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, aggrDiskData := range result {
		aggrUUID := aggrDiskData.GetChildContentS("aggregate-uuid")
		aggrName := aggrDiskData.GetChildContentS("aggregate")
		aggrDiskList := aggrDiskData.GetChildS("aggr-plex-list").GetChildS("aggr-plex-info").GetChildS("aggr-raidgroup-list").GetChildS("aggr-raidgroup-info").GetChildS("aggr-disk-list").GetChildren()
		for _, aggrDisk := range aggrDiskList {
			diskName = aggrDisk.GetChildContentS("disk")
			aggrsDisksMap[diskName] = aggrData{uuid: aggrUUID, name: aggrName}
		}
	}
	return aggrsDisksMap, nil
}
