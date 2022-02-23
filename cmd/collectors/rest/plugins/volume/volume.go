/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package volume

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/collectors/rest/plugins"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strconv"
	"strings"
	"time"
)

const DefaultPluginDuration = 1800 * time.Second
const DefaultDataPollDuration = 180 * time.Second

type Volume struct {
	*plugin.AbstractPlugin
	data                 *matrix.Matrix
	instanceKeys         map[string]string
	instanceLabels       map[string]*dict.Dict
	pluginInvocationRate int
	currentVal           int
	client               *rest.Client
	query                string
	snapmirrorFields     []string
	outgoingSM           map[string][]string
	incomingSM           map[string]string
	isHealthySM          map[string]bool
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (my *Volume) Init() error {

	var err error
	my.snapmirrorFields = []string{"relationship_id", "relationship_group_type", "destination_volume",
		"source_volume", "destination_path", "type", "healthy", "source_vserver", "destination_vserver"}

	if err = my.InitAbc(); err != nil {
		return err
	}

	timeout := rest.DefaultTimeout * time.Second
	if my.client, err = rest.New(conf.ZapiPoller(my.ParentParams), timeout); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.query = "api/private/cli/snapmirror"
	my.data = matrix.New(my.Parent+".Volume", "volume", "volume")

	my.outgoingSM = make(map[string][]string)
	my.incomingSM = make(map[string]string)
	my.isHealthySM = make(map[string]bool)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	if my.currentVal, err = my.setPluginInterval(); err != nil {
		my.Logger.Error().Err(err).Stack().Msg("Failed while setting the plugin interval")
		return err
	}

	return nil
}

func (my *Volume) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from zapi.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	if my.currentVal >= my.pluginInvocationRate {
		my.currentVal = 0

		// invoke snapmirror zapi and populate info in source and destination snapmirror maps
		if smSourceMap, smDestinationMap, err := my.GetSnapMirrors(); err != nil {
			my.Logger.Warn().Stack().Err(err).Msg("Failed to collect snapmirror data")
		} else {
			// update internal cache based on volume and SM maps
			my.updateMaps(data, smSourceMap, smDestinationMap)
		}
	}

	// update volume instance labels
	my.updateVolumeLabels(data)

	my.updateHardwareEncrypted(data)

	my.currentVal++
	return nil, nil
}

func (my *Volume) setPluginInterval() (int, error) {

	volumeDataInterval := my.getDataInterval(my.ParentParams, DefaultDataPollDuration)
	pluginDataInterval := my.getDataInterval(my.Params, DefaultPluginDuration)
	my.Logger.Debug().Float64("VolumeDataInterval", volumeDataInterval).Float64("PluginDataInterval", pluginDataInterval).Msg("Poll interval duration")
	my.pluginInvocationRate = int(pluginDataInterval / volumeDataInterval)

	return my.pluginInvocationRate, nil
}

func (my *Volume) getDataInterval(param *node.Node, defaultInterval time.Duration) float64 {
	var dataIntervalStr = ""
	schedule := param.GetChildS("schedule")
	if schedule != nil {
		dataInterval := schedule.GetChildS("data")
		if dataInterval != nil {
			dataIntervalStr = dataInterval.GetContentS()
			if durationVal, err := time.ParseDuration(dataIntervalStr); err == nil {
				return durationVal.Seconds()
			}
		}
	}
	return defaultInterval.Seconds()
}

func (my *Volume) GetSnapMirrors() (map[string][]*matrix.Instance, map[string]*matrix.Instance, error) {
	var (
		records []interface{}
		content []byte
		err     error
	)

	smSourceMap := make(map[string][]*matrix.Instance)
	smDestinationMap := make(map[string]*matrix.Instance)

	snapmirrorData := matrix.New(my.Parent+".SnapMirror", "sm", "sm")

	href := rest.BuildHref("", strings.Join(my.snapmirrorFields, ","), nil, "", "", "", "", my.query)

	err = rest.FetchData(my.client, href, &records)
	if err != nil {
		my.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, nil, err
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}

	content, err = json.Marshal(all)
	if err != nil {
		my.Logger.Error().Err(err).Str("ApiPath", my.query).Msg("Unable to marshal rest pagination")
	}

	if !gjson.ValidBytes(content) {
		my.Logger.Error().Err(err).Str("Api", my.query).Msg("Invalid json")
		return nil, nil, errors.New(errors.API_RESPONSE, "Invalid json")
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, nil, errors.New(errors.ERR_NO_INSTANCE, "no "+my.query+" instances on cluster")
	}

	for _, snapMirror := range results[1].Array() {

		relationshipId := snapMirror.Get("relationship_id").String()
		groupType := snapMirror.Get("relationship_group_type").String()
		destinationVolume := snapMirror.Get("destination_volume").String()
		sourceVolume := snapMirror.Get("source_volume").String()
		destinationLocation := snapMirror.Get("destination_path").String()
		relationshipType := snapMirror.Get("type").String()
		isHealthy := snapMirror.Get("healthy").String()
		sourceSvm := snapMirror.Get("source_vserver").String()
		destinationSvm := snapMirror.Get("destination_vserver").String()

		instanceKey := relationshipId
		instance, err := snapmirrorData.NewInstance(instanceKey)

		if err != nil {
			my.Logger.Error().Err(err).Stack().Str("relationshipId", relationshipId).Msg("Failed to create snapmirror cache instance")
			return nil, nil, err
		}

		instance.SetLabel("relationship_id", relationshipId)
		instance.SetLabel("group_type", groupType)
		instance.SetLabel("destination_volume", destinationVolume)
		instance.SetLabel("source_volume", sourceVolume)
		instance.SetLabel("destination_location", destinationLocation)
		instance.SetLabel("relationship_type", relationshipType)
		instance.SetLabel("is_healthy", isHealthy)
		instance.SetLabel("source_svm", sourceSvm)
		instance.SetLabel("destination_svm", destinationSvm)

		// Update the protectedBy and protectionSourceType fields in snapmirror
		plugins.UpdateProtectedFields(instance)

		// Update source snapmirror and destination snapmirror info in maps
		if relationshipType == "data_protection" || relationshipType == "extended_data_protection" || relationshipType == "vault" || relationshipType == "xdp" {
			sourceKey := sourceVolume + "-" + sourceSvm
			destinationKey := destinationVolume + "-" + destinationSvm
			if instance.GetLabel("protectionSourceType") == "volume" {
				smSourceMap[sourceKey] = append(smSourceMap[sourceKey], instance)
			}
			smDestinationMap[destinationKey] = instance
		}
	}

	return smSourceMap, smDestinationMap, nil
}

func (my *Volume) updateMaps(data *matrix.Matrix, smSourceMap map[string][]*matrix.Instance, smDestinationMap map[string]*matrix.Instance) {
	// Clean all the snapmirror maps
	my.outgoingSM = make(map[string][]string)
	my.incomingSM = make(map[string]string)
	my.isHealthySM = make(map[string]bool)

	for _, volume := range data.GetInstances() {
		volumeName := volume.GetLabel("volume")
		svmName := volume.GetLabel("svm")
		volumeType := volume.GetLabel("type")
		key := volumeName + "-" + svmName

		protectedByMap := make(map[string]string)
		var protectedByValue []string
		healthStatus := true
		for _, smRelationship := range smSourceMap[key] {
			/* Example: If 3 relationships belongs to a volume, and out of 3, 2 are at snapmirror and one is svmdr,
			   So, protectedByMap map has 2 records, and outgoingSM map value would be snapmirror, svmdr
			*/
			protectedByMap[smRelationship.GetLabel("protectedBy")] = ""

			// Update isHealthySM map based on the source snapmirror info
			if volumeType == "rw" {
				/* Example: If 3 relationships belongs to a volume, and out of 3, 2 are healthy and one is not,
				   So, isHealthySM map value would be unhealthy - false
				*/
				currentVal, _ := strconv.ParseBool(smRelationship.GetLabel("is_healthy"))
				healthStatus = healthStatus && currentVal
				my.isHealthySM[key] = healthStatus
			}
		}

		// Update outgoingSM map based on the protectedByMap
		protectedByValue = nil
		for protectedByKey := range protectedByMap {
			protectedByValue = append(protectedByValue, protectedByKey)
		}
		if protectedByValue != nil {
			my.outgoingSM[key] = protectedByValue
		}

		// Update incomingSM map based on the destination snapmirror info
		if smDestinationMap[key] != nil {
			my.incomingSM[key] = "destination"
		}
	}
}

func (my *Volume) updateVolumeLabels(data *matrix.Matrix) {
	for _, volume := range data.GetInstances() {
		volumeName := volume.GetLabel("volume")
		svmName := volume.GetLabel("svm")
		volumeType := volume.GetLabel("type")
		key := volumeName + "-" + svmName

		// Update protectionRole label in volume
		if volumeType == "rw" && my.incomingSM[key] == "" && my.outgoingSM[key] == nil {
			volume.SetLabel("protectionRole", "unprotected")
		} else if volumeType == "rw" && my.outgoingSM[key] != nil {
			volume.SetLabel("protectionRole", "protected")
		} else if volumeType == "dp" || (volumeType == "rw" && my.incomingSM[key] != "") {
			volume.SetLabel("protectionRole", "destination")
		} else {
			volume.SetLabel("protectionRole", "not_applicable")
		}

		// Update protectedBy label in volume
		if outgoing, ok := my.outgoingSM[key]; ok {
			outgoingJoinStr := strings.Join(outgoing, ",")
			if outgoingJoinStr == "volume,storage_vm" || outgoingJoinStr == "storage_vm,volume" {
				volume.SetLabel("protectedBy", "svmdr_and_snapmirror")
			} else if outgoingJoinStr == "cg,volume" || outgoingJoinStr == "volume,cg" {
				volume.SetLabel("protectedBy", "cg_and_snapmirror")
			} else if outgoingJoinStr == "cg" {
				volume.SetLabel("protectedBy", "consistency_group")
			} else if outgoingJoinStr == "storage_vm" {
				volume.SetLabel("protectedBy", "storage_vm_dr")
			} else if outgoingJoinStr == "volume" {
				volume.SetLabel("protectedBy", "snapmirror")
			}
		} else {
			volume.SetLabel("protectedBy", "not_applicable")
		}

		// Update all_sm_healthy label in volume, when all relationships belongs to this volume are healthy then true, otherwise false
		if healthy, ok := my.isHealthySM[key]; ok {
			volume.SetLabel("all_sm_healthy", strconv.FormatBool(healthy))
		}

	}
}

func (my *Volume) updateHardwareEncrypted(data *matrix.Matrix) {

	var (
		records  []interface{}
		content  []byte
		err      error
		aggrsMap map[string]string
	)

	aggrsMap = make(map[string]string)
	diskFields := []string{"aggregates.name", "aggregates.uuid"}
	query := "api/storage/disks"

	href := rest.BuildHref("", strings.Join(diskFields, ","), []string{"protection_mode=!data|full"}, "", "", "", "", query)

	err = rest.FetchData(my.client, href, &records)
	if err != nil {
		my.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		//return nil, nil, err
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}

	content, err = json.Marshal(all)
	if err != nil {
		my.Logger.Error().Err(err).Str("ApiPath", query).Msg("Unable to marshal rest pagination")
	}

	if !gjson.ValidBytes(content) {
		my.Logger.Error().Err(err).Str("Api", query).Msg("Invalid json")
		//return nil, nil, errors.New(errors.API_RESPONSE, "Invalid json")
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		//return nil, nil, errors.New(errors.ERR_NO_INSTANCE, "no "+my.query+" instances on cluster")
	}

	for _, disk := range results[1].Array() {
		aggrName := disk.Get("aggregates.name").String()
		aggrUuid := disk.Get("aggregates.uuid").String()
		aggrsMap[aggrUuid] = aggrName
	}

	for _, volume := range data.GetInstances() {
		_, ok := aggrsMap[volume.GetLabel("aggrUuid")]
		volume.SetLabel("isHardwareEncrypted", strconv.FormatBool(ok))
	}
}
