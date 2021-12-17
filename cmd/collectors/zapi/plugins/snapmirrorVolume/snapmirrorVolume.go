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
const RefreshPeriodicity = 10

type SnapmirrorVolume struct {
	*plugin.AbstractPlugin
	data               *matrix.Matrix
	instanceKeys       map[string]string
	instanceLabels     map[string]*dict.Dict
	batchSize          string
	RefreshPeriodicity int // TODO: appropriate names required
	currentVal         int // TODO: appropriate names required
	client             *zapi.Client
	query              string
	outgoingSM         map[string]string
	incomingSM         map[string]string
	isHealthySM        map[string]string
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
	my.Logger.Info().Msg("plugin connected!")
	my.data = matrix.New(my.Parent+".Volume", "volume", "volume")

	my.outgoingSM = make(map[string]string)
	my.incomingSM = make(map[string]string)
	my.isHealthySM = make(map[string]string)

	if my.client.IsClustered() {
		// batching the request
		if b := my.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				my.batchSize = b
				my.Logger.Info().Msgf("using batch-size [%s]", my.batchSize)
			}
		} else {
			my.batchSize = BatchSize
			my.Logger.Trace().Str("BatchSize", BatchSize).Msg("Using default batch-size")
		}

		// refresh interval
		if r := my.Params.GetChildContentS("refresh_periodicity"); r != "" {
			if refreshval, err := strconv.Atoi(r); err == nil {
				my.RefreshPeriodicity = refreshval
				my.Logger.Info().Msgf("using refresh at [%d]", my.RefreshPeriodicity)
			}
		} else {
			my.RefreshPeriodicity = RefreshPeriodicity
			my.Logger.Trace().Int("refreshPeriod", RefreshPeriodicity).Msg("Using default refresh")
		}
		my.currentVal = my.RefreshPeriodicity
	}

	return nil
}

func (my *SnapmirrorVolume) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from zapi.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	my.Logger.Info().Msgf("periodic refresh count %d current %d", my.RefreshPeriodicity, my.currentVal)

	if my.currentVal >= my.RefreshPeriodicity {
		my.Logger.Info().Msgf("current count %d", my.currentVal)
		my.currentVal = 0

		// invoke snapmirror zapi and populate info in source and destination snapmirror maps
		smSourceMap, smDestinationMap, err := my.GetSnapMirrors()
		if err != nil {
			my.Logger.Error().Msgf("add (%s) instance: %v", err)
		}

		// update internal cache based on volume and SM maps
		my.updateMaps(data, smSourceMap, smDestinationMap)
	}

	// update volume instance labels
	my.updateVolumeLabels(data)

	my.currentVal++
	return nil, nil
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
			break
		}

		if x := result.GetChildS("attributes-list"); x != nil {
			snapMirrors = x.GetChildren()
		}

		if len(snapMirrors) == 0 {
			return nil, nil, errors.New(errors.ERR_NO_INSTANCE, "no sm info instances found")
		}

		my.Logger.Info().Msgf("fetching %d snapmirrors", len(snapMirrors))

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
				my.Logger.Debug().Msgf("add (%s) instance: %v", relationshipId, err)
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
				if instance.GetLabel("protectionSourceType") == "volume" {
					smSourceMap[sourceVolume+"-"+sourceSvm] = append(smSourceMap[sourceVolume+"-"+sourceSvm], instance)
				}
				smDestinationMap[destinationVolume+"-"+destinationSvm] = instance
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
			if my.outgoingSM[key] == "" {
				my.outgoingSM[key] = smRelationship.GetLabel("protectedBy")
			} else {
				if my.outgoingSM[key] != smRelationship.GetLabel("protectedBy") {
					my.outgoingSM[key] = my.outgoingSM[key] + "_and_" + smRelationship.GetLabel("protectedBy")
				}
			}

			// Update isHealthySM map based on the source snapmirror info
			if volumeType == "rw" {
				healthy, _ := strconv.ParseBool(smRelationship.GetLabel("is_healthy"))
				if my.isHealthySM[key] == "" {
					my.isHealthySM[key] = smRelationship.GetLabel("is_healthy")
				} else {
					previousVal, _ := strconv.ParseBool(my.isHealthySM[key])
					my.isHealthySM[key] = strconv.FormatBool(previousVal && healthy)
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
		if my.outgoingSM[key] != "" {
			if my.outgoingSM[key] == "volume_and_storage_vm" {
				volume.SetLabel("protectedBy", "svmdr_and_snapmirror")
			} else if my.outgoingSM[key] == "cg_and_volume" {
				volume.SetLabel("protectedBy", "cg_and_snapmirror")
			} else if my.outgoingSM[key] == "cg" {
				volume.SetLabel("protectedBy", "consistency_group")
			} else if my.outgoingSM[key] == "storage_vm" {
				volume.SetLabel("protectedBy", "storage_vm_dr")
			} else if my.outgoingSM[key] == "volume" {
				volume.SetLabel("protectedBy", "snapmirror")
			}
		} else {
			volume.SetLabel("protectedBy", "not_applicable")
		}

		// Update all_sm_healthy label in volume, when all relationships belongs to this volume are healthy then true, otherwise false
		if my.isHealthySM[key] == "false" {
			volume.SetLabel("all_sm_healthy", "false")
		} else if my.isHealthySM[key] == "true" {
			volume.SetLabel("all_sm_healthy", "true")
		}

	}
}
