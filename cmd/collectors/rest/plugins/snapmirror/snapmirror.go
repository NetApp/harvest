/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */
package snapmirror

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"strings"
	"time"
)

const PluginInvocationRate = 10

type SnapMirror struct {
	*plugin.AbstractPlugin
	client         *rest.Client
	query          string
	nodeUpdCounter int
	svmVolToNode   map[string]string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SnapMirror{AbstractPlugin: p}
}

func (my *SnapMirror) Init() error {

	var err error

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

	my.query = "api/private/cli/volume"
	my.svmVolToNode = make(map[string]string)

	// Assigned the value to nodeUpdCounter so that plugin would be invoked first time to populate cache.
	my.nodeUpdCounter = PluginInvocationRate

	return nil
}

func (my *SnapMirror) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	if my.nodeUpdCounter >= PluginInvocationRate {
		my.nodeUpdCounter = 0
		if err := my.updateNodeCache(); err != nil {
			return nil, err
		}
		my.Logger.Debug().Msg("updated node cache")
	}

	// update volume instance labels
	my.updateSMLabels(data)
	my.nodeUpdCounter++

	return nil, nil
}

func (my *SnapMirror) updateNodeCache() error {
	var (
		records []interface{}
		content []byte
		err     error
	)
	href := rest.BuildHref("", "node", nil, "", "", "", "", my.query)

	err = rest.FetchData(my.client, href, &records)
	if err != nil {
		my.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		return err
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
		return errors.New(errors.API_RESPONSE, "Invalid json")
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return errors.New(errors.ERR_NO_INSTANCE, "no "+my.query+" instances on cluster")
	}

	for _, volume := range results[1].Array() {
		volumeName := volume.Get("volume").String()
		vserverName := volume.Get("vserver").String()
		nodeName := volume.Get("node").String()
		my.svmVolToNode[vserverName+volumeName] = nodeName
	}
	return nil
}

func (my *SnapMirror) updateProtectedFields(instance *matrix.Instance) {

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

		// Update protectedBy field
		if isSvmDr || isConstituentVolumeRelationshipWithinSvmDr {
			instance.SetLabel("protectedBy", "storage_vm")
		} else if isCg || isConstituentVolumeRelationshipWithinCG {
			instance.SetLabel("protectedBy", "cg")
		} else {
			instance.SetLabel("protectedBy", "volume")
		}

		// SVM-DR related information is populated
		// Update protectionSourceType field
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

func (my *SnapMirror) updateSMLabels(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		volumeName := instance.GetLabel("source_volume")
		vserverName := instance.GetLabel("source_vserver")

		// Update source_node label in snapmirror
		if node, ok := my.svmVolToNode[vserverName+volumeName]; ok {
			instance.SetLabel("source_node", node)
		}

		// update the protectedBy and protectionSourceType fields and derivedRelationshipType in snapmirror_labels
		my.updateProtectedFields(instance)
	}
}
