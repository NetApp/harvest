package collectors

import (
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

const DefaultBatchSize = "500"

func InvokeRestCall(client *rest.Client, href string, logger *logging.Logger) ([]gjson.Result, error) {
	result, err := rest.Fetch(client, href)
	if err != nil {
		logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return []gjson.Result{}, err
	}

	if len(result) == 0 {
		return []gjson.Result{}, nil
	}

	return result, nil
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
		isCg := groupType == "consistencygroup" && strings.Contains(destinationLocation, ":/cg/")
		isConstituentVolumeRelationshipWithinSvmDr := groupType == "vserver" && !strings.HasSuffix(destinationLocation, ":")
		isConstituentVolumeRelationshipWithinCG := groupType == "consistencygroup" && !strings.Contains(destinationLocation, ":/cg/")

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

func SetNameservice(nsDB, nsSource, nisDomain string, instance *matrix.Instance) {
	requiredNSDb := false
	requiredNSSource := false

	if strings.Contains(nsDB, "passwd") || strings.Contains(nsDB, "group") || strings.Contains(nsDB, "netgroup") {
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

// IsTimestampOlderThanDuration - timestamp units are micro seconds
func IsTimestampOlderThanDuration(timestamp float64, duration time.Duration) bool {
	return time.Since(time.UnixMicro(int64(timestamp))) > duration
}
