/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package volume

import (
	"github.com/hashicorp/go-version"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"strconv"
	"time"
)

const HoursInMonth = 24 * 30
const ARWSupportedVersion = "9.10.0"

type Volume struct {
	*plugin.AbstractPlugin
	currentVal            int
	client                *rest.Client
	aggrsMap              map[string]bool // aggregate-name -> exist map
	arw                   *matrix.Matrix
	includeConstituents   bool
	isArwSupportedVersion bool
}

type volumeInfo struct {
	arwStartTime             string
	arwState                 string
	snapshotAutodelete       string
	cloneSnapshotName        string
	cloneSplitEstimateMetric float64
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) Init() error {

	var err error

	if err = v.InitAbc(); err != nil {
		return err
	}

	v.aggrsMap = make(map[string]bool)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	v.currentVal = v.SetPluginInterval()

	if v.Options.IsTest {
		return nil
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout, v.Auth); err != nil {
		v.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = v.client.Init(5); err != nil {
		return err
	}

	v.arw = matrix.New(v.Parent+".Volume", "volume_arw", "volume_arw")
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "ArwStatus")
	v.arw.SetExportOptions(exportOptions)
	_, err = v.arw.NewMetricFloat64("status", "status")
	if err != nil {
		v.Logger.Error().Stack().Err(err).Msg("add metric")
		return err
	}

	// Read template to decide inclusion of flexgroup constituents
	v.includeConstituents = collectors.ReadPluginKey(v.Params, "include_constituents")
	// ARW feature is supported from 9.10 onwards, If we ask this field in Rest call in plugin, then it will be failed.
	v.isArwSupportedVersion = v.versionHigherThan(ARWSupportedVersion)
	return nil
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[v.Object]
	v.client.Metadata.Reset()

	if v.currentVal >= v.PluginInvocationRate {
		v.currentVal = 0

		// invoke disk rest and populate info in aggrsMap
		if disks, err := v.getEncryptedDisks(); err != nil {
			if errs.IsRestErr(err, errs.APINotFound) {
				v.Logger.Debug().Err(err).Msg("Failed to collect disk data")
			} else {
				v.Logger.Error().Err(err).Msg("Failed to collect disk data")
			}
		} else {
			// update aggrsMap based on disk data
			v.updateAggrMap(disks)
		}
	}

	volumeMap, err := v.getVolumeInfo()
	if err != nil {
		v.Logger.Error().Err(err).Msg("Failed to collect volume info data")
	}

	// update volume instance labels
	v.updateVolumeLabels(data, volumeMap)

	// parse anti_ransomware_start_time, antiRansomwareState for all volumes and export at cluster level
	v.handleARWProtection(data)

	v.currentVal++
	return []*matrix.Matrix{v.arw}, v.client.Metadata, nil
}

func (v *Volume) updateVolumeLabels(data *matrix.Matrix, volumeMap map[string]volumeInfo) {
	var err error
	cloneSplitEstimateMetric := data.GetMetric("clone_split_estimate")
	if cloneSplitEstimateMetric == nil {
		if cloneSplitEstimateMetric, err = data.NewMetricFloat64("clone_split_estimate"); err != nil {
			v.Logger.Error().Err(err).Msg("error while creating clone split estimate metric")
			return
		}
	}
	for _, volume := range data.GetInstances() {
		if !volume.IsExportable() {
			continue
		}

		if volume.GetLabel("style") == "flexgroup_constituent" {
			volume.SetExportable(v.includeConstituents)
		}

		volume.SetLabel("isHardwareEncrypted", strconv.FormatBool(v.aggrsMap[volume.GetLabel("aggr")]))

		if vInfo, ok := volumeMap[volume.GetLabel("volume")+volume.GetLabel("svm")]; ok {
			volume.SetLabel("anti_ransomware_start_time", vInfo.arwStartTime)
			volume.SetLabel("antiRansomwareState", vInfo.arwState)
			volume.SetLabel("snapshot_autodelete", vInfo.snapshotAutodelete)
			if volume.GetLabel("is_flexclone") == "true" {
				volume.SetLabel("clone_parent_snapshot", vInfo.cloneSnapshotName)
				if err = cloneSplitEstimateMetric.SetValueFloat64(volume, vInfo.cloneSplitEstimateMetric); err != nil {
					v.Logger.Error().Err(err).Str("metric", "cloneSplitEstimateMetric").Msg("Unable to set value on metric")
				}
			}
		}
	}
}

func (v *Volume) handleARWProtection(data *matrix.Matrix) {
	var (
		arwInstance       *matrix.Instance
		arwStartTimeValue time.Time
		err               error
	)

	// Purge and reset data
	v.arw.PurgeInstances()
	v.arw.Reset()

	// Set all global labels
	v.arw.SetGlobalLabels(data.GetGlobalLabels())
	arwStatusValue := "Active Mode"
	// Case where cluster doesn't have any volumes, arwStatus show as 'Not Monitoring'
	if len(data.GetInstances()) == 0 {
		arwStatusValue = "Not Monitoring"
	}

	// This is how cluster level arwStatusValue has been calculated based on each volume
	// If any one volume arwStatus is disabled --> "Not Monitoring"
	// If any one volume has been completed learning mode --> "Switch to Active Mode"
	// If all volumes are in learning mode --> "Learning Mode"
	// Else indicates arwStatus for all volumes are enabled --> "Active Mode"
	for _, volume := range data.GetInstances() {
		arwState := volume.GetLabel("antiRansomwareState")
		if arwState == "" {
			// Case where REST calls don't return `antiRansomwareState` field, arwStatus show as 'Not Monitoring'
			arwStatusValue = "Not Monitoring"
			break
		}
		if arwState == "disabled" {
			arwStatusValue = "Not Monitoring"
			break
		} else if arwState == "dry_run" || arwState == "enable_paused" {
			arwStartTime := volume.GetLabel("anti_ransomware_start_time")
			if arwStartTime == "" || arwStatusValue == "Switch to Active Mode" {
				continue
			}
			// If ARW startTime is more than 30 days old, which indicates that learning mode has been finished.
			if arwStartTimeValue, err = time.Parse(time.RFC3339, arwStartTime); err != nil {
				v.Logger.Error().Err(err).Msg("Failed to parse arw start time")
				arwStartTimeValue = time.Now()
			}
			if time.Since(arwStartTimeValue).Hours() > HoursInMonth {
				arwStatusValue = "Switch to Active Mode"
			} else {
				arwStatusValue = "Learning Mode"
			}
		}
	}

	arwInstanceKey := data.GetGlobalLabels()["cluster"] + data.GetGlobalLabels()["datacenter"]
	if arwInstance, err = v.arw.NewInstance(arwInstanceKey); err != nil {
		v.Logger.Error().Err(err).Str("arwInstanceKey", arwInstanceKey).Msg("Failed to create arw instance")
		return
	}

	arwInstance.SetLabel("ArwStatus", arwStatusValue)
	m := v.arw.GetMetric("status")
	// populate numeric data
	value := 1.0
	if err = m.SetValueFloat64(arwInstance, value); err != nil {
		v.Logger.Error().Stack().Err(err).Float64("value", value).Msg("Failed to parse value")
	} else {
		v.Logger.Debug().Float64("value", value).Msg("added value")
	}
}

func (v *Volume) getEncryptedDisks() ([]gjson.Result, error) {
	var (
		result []gjson.Result
		err    error
	)
	fields := []string{"aggregates.name"}
	query := "api/storage/disks"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"protection_mode=!data|full"}).
		Build()

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}
	return result, nil
}

func (v *Volume) getVolumeInfo() (map[string]volumeInfo, error) {
	volumeMap := make(map[string]volumeInfo)
	fields := []string{"name", "svm.name", "space.snapshot.autodelete_enabled", "clone.parent_snapshot.name", "clone.split_estimate"}
	if !v.isArwSupportedVersion {
		return v.getVolume("", fields, volumeMap)
	}

	// Only ask this field when ARW would be supported, is_constituent is supported from 9.10 onwards in public api same as ARW
	fields = append(fields, "anti_ransomware.dry_run_start_time", "anti_ransomware.state")
	if _, err := v.getVolume("is_constituent=false", fields, volumeMap); err != nil {
		return nil, err
	}
	return v.getVolume("is_constituent=true", fields, volumeMap)
}

func (v *Volume) getVolume(field string, fields []string, volumeMap map[string]volumeInfo) (map[string]volumeInfo, error) {
	var (
		result []gjson.Result
		err    error
	)
	query := "api/storage/volumes"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{field}).
		Build()

	if result, err = collectors.InvokeRestCall(v.client, href, v.Logger); err != nil {
		return nil, err
	}

	for _, volume := range result {
		volName := volume.Get("name").String()
		svmName := volume.Get("svm.name").String()
		arwStartTime := volume.Get("anti_ransomware.dry_run_start_time").String()
		arwState := volume.Get("anti_ransomware.state").String()
		snapshotAutodelete := volume.Get("space.snapshot.autodelete_enabled").String()
		cloneSnapshotName := volume.Get("clone.parent_snapshot.name").String()
		cloneSplitEstimate := volume.Get("clone.split_estimate").Float()
		volumeMap[volName+svmName] = volumeInfo{arwStartTime: arwStartTime, arwState: arwState, snapshotAutodelete: snapshotAutodelete, cloneSnapshotName: cloneSnapshotName, cloneSplitEstimateMetric: cloneSplitEstimate}
	}
	return volumeMap, nil
}

func (v *Volume) updateAggrMap(disks []gjson.Result) {
	if disks != nil {
		// Clean aggrsMap map
		clear(v.aggrsMap)
		for _, disk := range disks {
			aggrName := disk.Get("aggregates.#.name").Array()
			for _, aggr := range aggrName {
				v.aggrsMap[aggr.String()] = true
			}
		}
	}
}

func (v *Volume) versionHigherThan(minVersion string) bool {
	currentVersion, err := version.NewVersion(v.client.Cluster().GetVersion())
	if err != nil {
		return false
	}
	minSupportedVersion, err := version.NewVersion(minVersion)
	if err != nil {
		return false
	}
	return currentVersion.GreaterThanOrEqual(minSupportedVersion)
}
