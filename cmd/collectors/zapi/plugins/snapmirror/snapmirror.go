/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package snapmirror

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strings"
)

type SnapMirror struct {
	*plugin.AbstractPlugin
	client          *zapi.Client
	nodeCache       *dict.Dict
	destLimitCache  *dict.Dict
	srcLimitCache   *dict.Dict
	nodeUpdCounter  int
	limitUpdCounter int
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SnapMirror{AbstractPlugin: p}
}
func (my *SnapMirror) Init() error {
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
	my.nodeUpdCounter = 0
	my.limitUpdCounter = 0
	my.nodeCache = dict.New()
	my.destLimitCache = dict.New()
	my.srcLimitCache = dict.New()
	my.Logger.Debug().Msg("plugin initialized")
	return nil
}
func (my *SnapMirror) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	// update caches every so while
	if my.nodeUpdCounter == 0 && my.client.IsClustered() {
		if err := my.updateNodeCache(); err != nil {
			return nil, err
		}
		my.Logger.Debug().Msg("updated node cache")
	} else if my.nodeUpdCounter > 10 {
		my.nodeUpdCounter = 0
	} else {
		my.nodeUpdCounter += 1
	}
	if my.limitUpdCounter == 0 {
		if err := my.updateLimitCache(); err != nil {
			return nil, err
		}
		my.Logger.Debug().Msg("updated limit cache")
	} else if my.limitUpdCounter > 100 {
		my.limitUpdCounter = 0
	} else {
		my.limitUpdCounter += 1
	}
	destUpdCount := 0
	srcUpdCount := 0
	limitUpdCount := 0
	for _, instance := range data.GetInstances() {
		if my.client.IsClustered() {
			// check instances where destination node is missing
			if instance.GetLabel("destination_node") == "" {
				key := instance.GetLabel("destination_vserver") + "." + instance.GetLabel("destination_volume")
				if destVol, has := my.nodeCache.GetHas(key); has {
					instance.SetLabel("destination_node", destVol)
					destUpdCount += 1
				}
			}
			// check instances where source node is missing
			if instance.GetLabel("source_node") == "" {
				key := instance.GetLabel("source_vserver") + "." + instance.GetLabel("source_volume")
				if srcVol, has := my.nodeCache.GetHas(key); has {
					instance.SetLabel("source_node", srcVol)
					srcUpdCount += 1
				}
			}

			// update the protectedBy and protectionSourceType fields and derivedRelationshipType in snapmirror_labels
			my.updateProtectedFields(instance)
		} else {
			// 7 Mode
			// source / destination nodes can be something like:
			//		tobago-1:vol_4kb_neu
			//      tobago-1:D
			if src := instance.GetLabel("source_node"); src != "" {
				if x := strings.Split(src, ":"); len(x) == 2 {
					instance.SetLabel("source_node", x[0])
					if len(x[1]) != 1 {
						instance.SetLabel("source_volume", x[1])
						srcUpdCount += 1
					}
				} else {
					break
				}
			}
			if dest := instance.GetLabel("destination_node"); dest != "" {
				if x := strings.Split(dest, ":"); len(x) == 2 {
					instance.SetLabel("destination_node", x[0])
					if len(x[1]) != 1 {
						instance.SetLabel("destination_volume", x[1])
						destUpdCount += 1
					}
				} else {
					break
				}
			}
		}
		// check if destination node limit is missing
		if instance.GetLabel("destination_node_limit") == "" {
			if limit, has := my.srcLimitCache.GetHas(instance.GetLabel("destination_node")); has {
				instance.SetLabel("destination_node_limit", limit)
				limitUpdCount += 1
			}
		}
		// check if destination node limit is missing
		if instance.GetLabel("source_node_limit") == "" {
			if limit, has := my.srcLimitCache.GetHas(instance.GetLabel("source_node")); has {
				instance.SetLabel("source_node_limit", limit)
			}
		}
	}
	my.Logger.Debug().Msgf("updated %d destination and %d source nodes, %d node limits", destUpdCount, srcUpdCount, limitUpdCount)
	return nil, nil
}
func (my *SnapMirror) updateNodeCache() error {
	var (
		request, resp *node.Node
		err           error
	)
	count := 0
	request = node.NewXmlS("perf-object-get-instances")
	request.NewChildS("objectname", "volume")
	//request.CreateChild("max-records", my.batch_size)
	requestInstances := request.NewChildS("instances", "")
	requestInstances.NewChildS("instance", "*")
	requestCounters := request.NewChildS("counters", "")
	requestCounters.NewChildS("counter", "node_name")
	requestCounters.NewChildS("counter", "vserver_name")
	if resp, err = my.client.InvokeRequest(request); err != nil {
		return err
	}
	if instances := resp.GetChildS("instances"); instances != nil {
		for _, i := range instances.GetChildren() {
			vol := i.GetChildContentS("name")
			counters := i.GetChildS("counters")
			if counters != nil {
				var nodeName string
				var svm string
				for _, c := range counters.GetChildren() {
					name := c.GetChildContentS("name")
					value := c.GetChildContentS("value")
					switch name {
					case "node_name":
						nodeName = value
					case "vserver_name":
						svm = value
					}
				}
				if nodeName != "" && svm != "" {
					my.nodeCache.Set(svm+"."+vol, nodeName)
					count += 1
				}
			}
		}
	}
	my.Logger.Debug().Msgf("updated node cache for %d volumes", count)
	return nil
}
func (my *SnapMirror) updateLimitCache() error {
	var (
		request, response *node.Node
		err               error
	)
	request = node.NewXmlS("perf-object-get-instances")
	request.NewChildS("objectname", "smc_em")
	requestInstances := request.NewChildS("instances", "")
	requestInstances.NewChildS("instance", "*")
	requestCounters := request.NewChildS("counters", "")
	requestCounters.NewChildS("counter", "node_name")
	requestCounters.NewChildS("counter", "dest_meter_count")
	requestCounters.NewChildS("counter", "src_meter_count")
	if response, err = my.client.InvokeRequest(request); err != nil {
		return err
	}
	count := 0
	if instances := response.GetChildS("instances"); instances != nil {
		for _, i := range instances.GetChildren() {
			nodeName := i.GetChildContentS("node_name")
			my.destLimitCache.Set(nodeName, i.GetChildContentS("dest_meter_count"))
			my.srcLimitCache.Set(nodeName, i.GetChildContentS("src_meter_count"))
			count += 1
		}
	}
	my.Logger.Debug().Msgf("updated limit cache for %d nodes", count)
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
