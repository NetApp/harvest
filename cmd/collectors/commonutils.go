package collectors

import (
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/errors"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

func InvokeRestCall(client *rest.Client, query string, href string, logger *logging.Logger) ([]gjson.Result, error) {
	result, err := rest.Fetch(client, href)
	if err != nil {
		logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return []gjson.Result{}, err
	}

	if len(result) == 0 {
		return []gjson.Result{}, errors.New(errors.ErrNoInstance, "no "+query+" instances on cluster")
	}

	return result, nil
}

func InvokeZapiCall(client *zapi.Client, request *node.Node, logger *logging.Logger, tag string) ([]*node.Node, string, error) {

	var (
		result   *node.Node
		response []*node.Node
		newTag   string
		err      error
	)

	if tag != "" {
		if result, newTag, err = client.InvokeBatchRequest(request, tag); err != nil {
			return nil, "", err
		}
	} else {
		if result, err = client.InvokeRequest(request); err != nil {
			return nil, "", err
		}
	}

	if result == nil {
		return nil, "", nil
	}

	if x := result.GetChildS("attributes-list"); x != nil {
		response = x.GetChildren()
	} else if y := result.GetChildS("attributes"); y != nil {
		// Check for non-list response
		response = y.GetChildren()
	}

	if len(response) == 0 {
		return nil, "", nil
	}

	logger.Trace().Int("object", len(response)).Msg("fetching")

	return response, newTag, nil
}

func UpdateProtectedFields(instance *matrix.Instance) {

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

	// Update derived_relationship_type field based on the policyType
	relationshipType := instance.GetLabel("relationship_type")
	if policyType := instance.GetLabel("policy_type"); policyType != "" {
		if policyType == "strict_sync_mirror" {
			instance.SetLabel("derived_relationship_type", "sync_mirror_strict")
		} else if policyType == "sync_mirror" {
			instance.SetLabel("derived_relationship_type", "sync_mirror")
		} else if policyType == "mirror_vault" {
			instance.SetLabel("derived_relationship_type", "mirror_vault")
		} else if policyType == "automated_failover" {
			instance.SetLabel("derived_relationship_type", "sync_mirror")
		} else {
			instance.SetLabel("derived_relationship_type", relationshipType)
		}
	} else {
		instance.SetLabel("derived_relationship_type", relationshipType)
	}
}

func SetNameservice(nsDb, nsSource, nisDomain string, instance *matrix.Instance) {
	requiredNSDb := false
	requiredNSSource := false

	if strings.Contains(nsDb, "passwd") || strings.Contains(nsDb, "group") || strings.Contains(nsDb, "netgroup") {
		requiredNSDb = true
		if strings.Contains(nsSource, "nis") {
			requiredNSSource = true
		}
	}

	if nisDomain != "" && requiredNSDb && requiredNSSource {
		instance.SetLabel("nis_authentication_enabled", "true")
	} else {
		instance.SetLabel("nis_authentication_enabled", "false")
	}
}

func SetPluginInterval(parentParams *node.Node, params *node.Node, logger *logging.Logger, defaultDataPollDuration time.Duration, defaultPluginDuration time.Duration) (int, error) {

	volumeDataInterval := GetDataInterval(parentParams, defaultDataPollDuration)
	pluginDataInterval := GetDataInterval(params, defaultPluginDuration)
	logger.Debug().Float64("VolumeDataInterval", volumeDataInterval).Float64("PluginDataInterval", pluginDataInterval).Msg("Poll interval duration")
	pluginInvocationRate := int(pluginDataInterval / volumeDataInterval)

	return pluginInvocationRate, nil
}

func GetDataInterval(param *node.Node, defaultInterval time.Duration) float64 {
	schedule := param.GetChildS("schedule")
	if schedule != nil {
		dataInterval := schedule.GetChildContentS("data")
		if dataInterval != "" {
			if durationVal, err := time.ParseDuration(dataInterval); err == nil {
				return durationVal.Seconds()
			}
		}
	}
	return defaultInterval.Seconds()
}
