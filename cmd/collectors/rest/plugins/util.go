package plugins

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logging"
	"goharvest2/pkg/matrix"
	"strings"
)

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

func InvokeRestCall(client *rest.Client, query string, href string, logger *logging.Logger) ([]gjson.Result, error) {
	var (
		records []interface{}
		content []byte
		err     error
	)

	err = rest.FetchData(client, href, &records)
	if err != nil {
		logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		return []gjson.Result{}, err
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}

	content, err = json.Marshal(all)
	if err != nil {
		logger.Error().Err(err).Str("ApiPath", query).Msg("Unable to marshal rest pagination")
		return []gjson.Result{}, err
	}

	if !gjson.ValidBytes(content) {
		logger.Error().Err(err).Str("Api", query).Msg("Invalid json")
		return []gjson.Result{}, errors.New(errors.API_RESPONSE, "Invalid json")
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return []gjson.Result{}, errors.New(errors.ERR_NO_INSTANCE, "no "+query+" instances on cluster")
	}

	return results[1].Array(), nil
}
