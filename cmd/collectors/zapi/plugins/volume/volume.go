package volume

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strconv"
	"strings"
)

const BatchSize = "500"
const DefaultPluginInvocationRate = 10
const DefaultDataPollDuration = 180

type Volume struct {
	*plugin.AbstractPlugin
	data                 *matrix.Matrix
	instanceKeys         map[string]string
	instanceLabels       map[string]*dict.Dict
	batchSize            string
	pluginInvocationRate int
	currentVal           int
	client               *zapi.Client
	query                string
	outgoingSM           map[string][]string
	incomingSM           map[string]string
	isHealthySM          map[string]bool
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

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	if my.currentVal, err = my.setPluginInterval(); err != nil {
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
		smSourceMap, smDestinationMap, err := my.GetSnapMirrors()
		if err != nil {
			my.Logger.Error().Stack().Err(err).Msg("Failed to collect snapmirror data")
		}

		// update internal cache based on volume and SM maps
		my.updateMaps(data, smSourceMap, smDestinationMap)
	}

	// update volume instance labels
	my.updateVolumeLabels(data)

	my.currentVal++
	return nil, nil
}

func (my *Volume) setPluginInterval() (int, error) {

	volumeDataInterval := my.getDataInterval(my.ParentParams, DefaultDataPollDuration)
	pluginDataInterval := my.getDataInterval(my.Params, DefaultPluginInvocationRate)
	my.Logger.Debug().Int("VolumeDataInterval", volumeDataInterval).Int("PluginDataInterval", pluginDataInterval).Msg("Poll interval duration")
	my.pluginInvocationRate = pluginDataInterval / volumeDataInterval

	return my.pluginInvocationRate, nil
}

func (my *Volume) getDataInterval(param *node.Node, defaultInterval int) int {
	var dataIntervalStr = ""
	schedule := param.GetChildS("schedule")
	if schedule != nil {
		dataInterval := schedule.GetChildS("data")
		if dataInterval != nil {
			dataIntervalStr = dataInterval.GetContentS()
		}
	}

	// Convert the interval from str to int
	if dataIntervalStr != "" {
		dataIntervalStr = strings.Split(dataIntervalStr, "s")[0]
		if intervalVal, err := strconv.Atoi(dataIntervalStr); err == nil {
			return intervalVal
		}
	}
	return defaultInterval
}

func (my *Volume) updateProtectedFields(instance *matrix.Instance) {

	// check for group_type
	// Supported group_type are: "none", "vserver", "infinitevol", "consistencygroup", "flexgroup"
	if instance.GetLabel("group_type") != "" {

		groupType := instance.GetLabel("group_type")
		destinationVolume := instance.GetLabel("destination_volume")
		sourceVolume := instance.GetLabel("source_volume")
		destinationLocation := instance.GetLabel("destination_location")

		isSvmDr := groupType == "vserver" && destinationVolume == "" && sourceVolume == ""
		isCg := groupType == "CONSISTENCYGROUP" && strings.Contains(destinationLocation, ":/cg/")
		isConstituentVolumeRelationshipWithinSvmDr := groupType == "vserver" && !strings.HasSuffix(destinationLocation, ":")
		isConstituentVolumeRelationshipWithinCG := groupType == "CONSISTENCYGROUP" && !strings.Contains(destinationLocation, ":/cg/")

		// Update protectedBy label
		if isSvmDr || isConstituentVolumeRelationshipWithinSvmDr {
			instance.SetLabel("protectedBy", "storage_vm")
		} else if isCg || isConstituentVolumeRelationshipWithinCG {
			instance.SetLabel("protectedBy", "cg")
		} else {
			instance.SetLabel("protectedBy", "volume")
		}

		// SVM-DR related information is populated, Update protectionSourceType label
		if isSvmDr {
			instance.SetLabel("protectionSourceType", "storage_vm")
		} else if isCg {
			instance.SetLabel("protectionSourceType", "cg")
		} else if isConstituentVolumeRelationshipWithinSvmDr || isConstituentVolumeRelationshipWithinCG || groupType == "none" || groupType == "flexgroup" {
			instance.SetLabel("protectionSourceType", "volume")
		} else {
			instance.SetLabel("protectionSourceType", "not_mapped")
		}
	}
}

func (my *Volume) GetSnapMirrors() (map[string][]*matrix.Instance, map[string]*matrix.Instance, error) {
	var (
		request, result *node.Node
		snapMirrors     []*node.Node
		tag             string
		err             error
	)

	smSourceMap := make(map[string][]*matrix.Instance)
	smDestinationMap := make(map[string]*matrix.Instance)

	request = node.NewXmlS(my.query)
	if my.client.IsClustered() && my.batchSize != "" {
		request.NewChildS("max-records", my.batchSize)
	}

	tag = "initial"
	snapmirrorData := matrix.New(my.Parent+".SnapMirror", "sm", "sm")

	for {
		result, tag, err = my.client.InvokeBatchRequest(request, tag)

		if err != nil {
			return nil, nil, err
		}

		if result == nil {
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			snapMirrors = x.GetChildren()
		}

		if len(snapMirrors) == 0 {
			my.Logger.Info().Msg("no snapmirror instances found on system")
			break
		}

		my.Logger.Debug().Int("snapmirrors", len(snapMirrors)).Msg("fetching snapmirrors")

		for _, snapMirror := range snapMirrors {
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
			my.updateProtectedFields(instance)

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

		// Update all_sm_healthy label in volume, when all relationships belong to this volume are healthy then true, otherwise false
		if healthy, ok := my.isHealthySM[key]; ok {
			volume.SetLabel("all_sm_healthy", strconv.FormatBool(healthy))
		}

	}
}
