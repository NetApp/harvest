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
	"strings"
	"time"
)

const BatchSize = "500"
const DefaultPluginDuration = 30 * time.Minute
const DefaultDataPollDuration = 3 * time.Minute

type Volume struct {
	*plugin.AbstractPlugin
	data                 *matrix.Matrix
	batchSize            string
	pluginInvocationRate int
	currentVal           int
	client               *zapi.Client
	query                string
	outgoingSM           map[string][]string
	incomingSM           map[string]string
	isHealthySM          map[string]bool
	aggrsMap             map[string]string // aggregate-uuid -> aggregate-name map
}

type aggrData struct {
	aggrUuid string
	aggrName string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (my *Volume) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	if my.client, err = zapi.New(conf.ZapiPoller(my.ParentParams)); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.query = "snapmirror-get-iter"
	my.data = matrix.New(my.Parent+".Volume", "volume", "volume")

	my.outgoingSM = make(map[string][]string)
	my.incomingSM = make(map[string]string)
	my.isHealthySM = make(map[string]bool)
	my.aggrsMap = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	if my.currentVal, err = collectors.SetPluginInterval(my.ParentParams, my.Params, my.Logger, DefaultDataPollDuration, DefaultPluginDuration); err != nil {
		my.Logger.Error().Err(err).Stack().Msg("Failed while setting the plugin interval")
	}

	// batching the request
	if b := my.Params.GetChildContentS("batch_size"); b != "" {
		if _, err := strconv.Atoi(b); err == nil {
			my.batchSize = b
			my.Logger.Info().Str("BatchSize", my.batchSize).Msg("using batch-size")
		}
	} else {
		my.batchSize = BatchSize
		my.Logger.Trace().Str("BatchSize", BatchSize).Msg("Using default batch-size")
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
			my.Logger.Warn().Err(err).Msg("Failed to collect snapmirror data")
		} else {
			// update internal cache based on volume and SM maps
			my.updateMaps(data, smSourceMap, smDestinationMap)
		}

		// invoke disk-encrypt-get-iter zapi and populate disk info
		disks, err1 := my.getEncryptedDisks()
		// invoke aggr-status-get-iter zapi and populate aggr disk mapping info
		aggrDiskMap, err2 := my.getAggrDiskMapping()

		if err1 != nil {
			if errors.Is(err1, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err1).Msg("Failed to collect disk data")
			} else {
				my.Logger.Error().Err(err1).Msg("Failed to collect disk data")
			}
		}
		if err2 != nil {
			if errors.Is(err1, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err2).Msg("Failed to collect aggregate-disk mapping data")
			} else {
				my.Logger.Error().Err(err2).Msg("Failed to collect aggregate-disk mapping data")
			}
		}
		// update aggrsMap based on disk data and addr disk mapping
		my.updateAggrMap(disks, aggrDiskMap)
	}

	// update volume instance labels
	my.updateVolumeLabels(data)

	my.currentVal++
	return nil, nil
}

func (my *Volume) GetSnapMirrors() (map[string][]*matrix.Instance, map[string]*matrix.Instance, error) {
	var (
		request *node.Node
		result  []*node.Node
		tag     string
		err     error
	)

	smSourceMap := make(map[string][]*matrix.Instance)
	smDestinationMap := make(map[string]*matrix.Instance)

	request = node.NewXMLS(my.query)
	if my.client.IsClustered() && my.batchSize != "" {
		request.NewChildS("max-records", my.batchSize)
	}

	tag = "initial"
	snapmirrorData := matrix.New(my.Parent+".SnapMirror", "sm", "sm")

	for {
		if result, tag, err = collectors.InvokeZapiCall(my.client, request, my.Logger, tag); err != nil {
			return nil, nil, err
		}

		if len(result) == 0 || result == nil {
			break
		}

		for _, snapMirror := range result {
			relationshipId := snapMirror.GetChildContentS("relationship-id")
			groupType := snapMirror.GetChildContentS("relationship-group-type")
			destinationVolume := snapMirror.GetChildContentS("destination-volume")
			sourceVolume := snapMirror.GetChildContentS("source-volume")
			destinationLocation := snapMirror.GetChildContentS("destination-location")
			relationshipType := snapMirror.GetChildContentS("relationship-type")
			isHealthy := snapMirror.GetChildContentS("is-healthy")
			sourceSvm := snapMirror.GetChildContentS("source-vserver")
			destinationSvm := snapMirror.GetChildContentS("destination-vserver")

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
			collectors.UpdateProtectedFields(instance)

			// Update source snapmirror and destination snapmirror info in maps
			if relationshipType == "data_protection" || relationshipType == "extended_data_protection" || relationshipType == "vault" {
				sourceKey := sourceVolume + "-" + sourceSvm
				destinationKey := destinationVolume + "-" + destinationSvm
				if instance.GetLabel("protectionSourceType") == "volume" {
					smSourceMap[sourceKey] = append(smSourceMap[sourceKey], instance)
				}
				smDestinationMap[destinationKey] = instance
			}
		}

		// To break the batch zapi call when all the records were fetched.
		if tag == "" {
			break
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
		aggrUuid := volume.GetLabel("aggrUuid")
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

		// Update all_sm_healthy label in volume, when all relationships belong to this volume are healthy then true, otherwise false
		if healthy, ok := my.isHealthySM[key]; ok {
			volume.SetLabel("all_sm_healthy", strconv.FormatBool(healthy))
		}

		_, exist := my.aggrsMap[aggrUuid]
		volume.SetLabel("isHardwareEncrypted", strconv.FormatBool(exist))
	}
}

func (my *Volume) getEncryptedDisks() ([]string, error) {
	var (
		result    []*node.Node
		diskNames []string
		err       error
	)

	request := node.NewXMLS("disk-encrypt-get-iter")
	//algorithm is -- Protection mode needs to be DATA or FULL
	// Fetching rest of them and add as
	query := request.NewChildS("query", "")
	encryptInfoQuery := query.NewChildS("disk-encrypt-info", "")
	encryptInfoQuery.NewChildS("protection-mode", "open|part|miss")

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
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

func (my *Volume) updateAggrMap(disks []string, aggrDiskMap map[string]aggrData) {
	if disks != nil && aggrDiskMap != nil {
		// Clean aggrsMap map
		my.aggrsMap = make(map[string]string)

		for _, disk := range disks {
			aggr := aggrDiskMap[disk]
			my.aggrsMap[aggr.aggrUuid] = aggr.aggrName
		}
	}
}

func (my *Volume) getAggrDiskMapping() (map[string]aggrData, error) {
	var (
		result        []*node.Node
		aggrsDisksMap map[string]aggrData
		diskName      string
		err           error
	)

	request := node.NewXMLS("aggr-status-get-iter")
	aggrsDisksMap = make(map[string]aggrData)

	if result, _, err = collectors.InvokeZapiCall(my.client, request, my.Logger, ""); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, aggrDiskData := range result {
		aggrUuid := aggrDiskData.GetChildContentS("aaggregate-uuid")
		aggrName := aggrDiskData.GetChildContentS("aggregate")
		aggrDiskList := aggrDiskData.GetChildS("aggr-plex-list").GetChildS("aggr-plex-info").GetChildS("aggr-raidgroup-list").GetChildS("aggr-raidgroup-info").GetChildS("aggr-disk-list").GetChildren()
		for _, aggrDisk := range aggrDiskList {
			diskName = aggrDisk.GetChildContentS("disk")
			aggrsDisksMap[diskName] = aggrData{aggrUuid: aggrUuid, aggrName: aggrName}
		}
	}
	return aggrsDisksMap, nil
}
