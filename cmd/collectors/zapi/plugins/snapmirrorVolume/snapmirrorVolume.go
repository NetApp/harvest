package snapmirrorVolume

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strconv"
	"strings"
)

const BatchSize = "500"
const DefaultPluginInvocationRate = 10

type SnapmirrorVolume struct {
	*plugin.AbstractPlugin
	data                 *matrix.Matrix
	instanceKeys         map[string]string
	instanceLabels       map[string]*dict.Dict
	batchSize            string
	pluginInvocationRate int
	currentVal           int
	client               *zapi.Client
	query                string
	outgoingSM           map[string]string
	incomingSM           map[string]string
	isHealthySM          map[string]string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SnapmirrorVolume{AbstractPlugin: p}
}

func (my *SnapmirrorVolume) Init() error {

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

	my.outgoingSM = make(map[string]string)
	my.incomingSM = make(map[string]string)
	my.isHealthySM = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	if my.currentVal, err = my.setPluginInterval(); err != nil {
		my.Logger.Error().Err(err).Stack().Msg("Failed while setting the plugin interval")
	}

	if my.client.IsClustered() {
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
	}

	return nil
}

func (my *SnapmirrorVolume) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from zapi.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	my.Logger.Info().Msgf("periodic refresh count %d current %d", my.pluginInvocationRate, my.currentVal)

	if my.currentVal >= my.pluginInvocationRate {
		my.Logger.Info().Int("CurrentValue", my.currentVal).Msg("current count")
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

func (my *SnapmirrorVolume) setPluginInterval() (int, error) {

	schedule := my.Params.GetChildS("schedule")
	if schedule == nil {
		return DefaultPluginInvocationRate, errors.New(errors.MISSING_PARAM, "schedule")
	}

	dataInterval := schedule.GetChildS("data")
	if dataInterval == nil {
		return DefaultPluginInvocationRate, errors.New(errors.MISSING_PARAM, "data")
	}

	// Convert the interval from str to int and converted to periodicity.
	// '360s' convert to 360 and converted to 360/180 = 2
	if intervalStr := dataInterval.GetContentS(); intervalStr != "" {
		intervalStr = strings.Split(intervalStr, "s")[0]
		if intervalVal, err := strconv.Atoi(intervalStr); err == nil {
			my.pluginInvocationRate = intervalVal / 180
			my.Logger.Info().Int("PluginInvocationInterval", intervalVal).Msg("Plugin invoked at every ")
		}
	} else {
		my.pluginInvocationRate = DefaultPluginInvocationRate
		my.Logger.Trace().Int("PluginInvocationInterval", DefaultPluginInvocationRate*180).Msg("Using default interval")
	}

	return my.pluginInvocationRate, nil
}

func (my *SnapmirrorVolume) updateProtectedfields(instance *matrix.Instance) {

	// check for group_type
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
		}
	}
}

func (my *SnapmirrorVolume) GetSnapMirrors() (map[string][]*matrix.Instance, map[string]*matrix.Instance, error) {
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
			my.Logger.Error().Msg("Snapmirror zapi invocation return nil result")
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			snapMirrors = x.GetChildren()
		}

		if len(snapMirrors) == 0 {
			return nil, nil, errors.New(errors.ERR_NO_INSTANCE, "no snapmirror info instances found")
		}

		my.Logger.Info().Int("snapmirrors", len(snapMirrors)).Msg("fetching snapmirrors")

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
			my.updateProtectedfields(instance)

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

func (my *SnapmirrorVolume) updateMaps(data *matrix.Matrix, smSourceMap map[string][]*matrix.Instance, smDestinationMap map[string]*matrix.Instance) {
	// Clean all the snapmirror maps
	my.outgoingSM = make(map[string]string)
	my.incomingSM = make(map[string]string)
	my.isHealthySM = make(map[string]string)

	for _, volume := range data.GetInstances() {
		volumeName := volume.GetLabel("volume")
		svmName := volume.GetLabel("svm")
		volumeType := volume.GetLabel("type")
		key := volumeName + "-" + svmName

		for _, smRelationship := range smSourceMap[key] {
			// Update outgoingSM map based on the source snapmirror info
			if sm, ok := my.outgoingSM[key]; !ok {
				my.outgoingSM[key] = smRelationship.GetLabel("protectedBy")
			} else {
				if sm != smRelationship.GetLabel("protectedBy") {
					my.outgoingSM[key] = sm + "_and_" + smRelationship.GetLabel("protectedBy")
				}
			}

			// Update isHealthySM map based on the source snapmirror info
			if volumeType == "rw" {
				if h, ok := my.isHealthySM[key]; !ok {
					my.isHealthySM[key] = smRelationship.GetLabel("is_healthy")
				} else {
					// any relationship in volume is unhealthy would be treated as volume unhealthy category
					if h != smRelationship.GetLabel("is_healthy") {
						my.isHealthySM[key] = "false"
					}
				}
			}
		}

		// Update incomingSM map based on the destination snapmirror info
		if smDestinationMap[key] != nil {
			my.incomingSM[key] = "destination"
		}
	}
}

func (my *SnapmirrorVolume) updateVolumeLabels(data *matrix.Matrix) {
	for _, volume := range data.GetInstances() {
		volumeName := volume.GetLabel("volume")
		svmName := volume.GetLabel("svm")
		volumeType := volume.GetLabel("type")
		key := volumeName + "-" + svmName

		// Update protectionRole label in volume
		if volumeType == "rw" && my.incomingSM[key] == "" && my.outgoingSM[key] == "" {
			volume.SetLabel("protectionRole", "unprotected")
		} else if volumeType == "rw" && my.outgoingSM[key] != "" {
			volume.SetLabel("protectionRole", "protected")
		} else if volumeType == "dp" || (volumeType == "rw" && my.incomingSM[key] != "") {
			volume.SetLabel("protectionRole", "destination")
		} else {
			volume.SetLabel("protectionRole", "not_applicable")
		}

		// Update protectedBy label in volume
		if outgoing, ok := my.outgoingSM[key]; ok {
			if outgoing == "volume_and_storage_vm" {
				volume.SetLabel("protectedBy", "svmdr_and_snapmirror")
			} else if outgoing == "cg_and_volume" {
				volume.SetLabel("protectedBy", "cg_and_snapmirror")
			} else if outgoing == "cg" {
				volume.SetLabel("protectedBy", "consistency_group")
			} else if outgoing == "storage_vm" {
				volume.SetLabel("protectedBy", "storage_vm_dr")
			} else if outgoing == "volume" {
				volume.SetLabel("protectedBy", "snapmirror")
			}
		} else {
			volume.SetLabel("protectedBy", "not_applicable")
		}

		// Update all_sm_healthy label in volume, when all relationships belongs to this volume are healthy then true, otherwise false
		if healthy, ok := my.isHealthySM[key]; ok {
			volume.SetLabel("all_sm_healthy", healthy)
		}

	}
}
