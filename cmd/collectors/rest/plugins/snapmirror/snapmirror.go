/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */
package snapmirror

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"strings"
	"time"
)

const PluginInvocationRate = 10

type SnapMirror struct {
	*plugin.AbstractPlugin
	data           *matrix.Matrix
	client         *rest.Client
	query          string
	nodeUpdCounter int
	svmVolToNode   map[string]string
	svmPeerDataMap map[string]Peer // [peer SVM alias name] -> [peer detail] map
}

type Peer struct {
	svm     string
	cluster string
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
	my.svmPeerDataMap = make(map[string]Peer)

	my.data = matrix.New(my.Parent+".SnapMirror", "snapmirror", "snapmirror")

	exportOptions := node.NewS("export_options")
	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")

	if exportOption := my.ParentParams.GetChildS("export_options"); exportOption != nil {
		if exportedLabels := exportOption.GetChildS("instance_labels"); exportedLabels != nil {
			for _, label := range exportedLabels.GetAllChildContentS() {
				instanceLabels.NewChildS("", label)
			}
		}
		if exportedKeys := exportOption.GetChildS("instance_keys"); exportedKeys != nil {
			for _, key := range exportedKeys.GetAllChildContentS() {
				instanceKeys.NewChildS("", key)
			}
		}
	}
	my.data.SetExportOptions(exportOptions)

	// Assigned the value to nodeUpdCounter so that plugin would be invoked first time to populate cache.
	my.nodeUpdCounter = PluginInvocationRate
	return nil
}

func (my *SnapMirror) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from Rest.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	if my.nodeUpdCounter >= PluginInvocationRate {
		my.nodeUpdCounter = 0
		if err := my.updateNodeCache(); err != nil {
			return nil, err
		}
		my.Logger.Debug().Msg("updated node cache")

		if cluster, ok := data.GetGlobalLabels().GetHas("cluster"); ok {
			if err := my.getSVMPeerData(cluster); err != nil {
				return nil, err
			}
			my.Logger.Debug().Msg("updated svm peer detail")
		}
	}

	// update volume instance labels
	my.updateSMLabels(data)
	my.nodeUpdCounter++

	return []*matrix.Matrix{my.data}, nil
}

func (my *SnapMirror) updateNodeCache() error {
	var (
		result []gjson.Result
		err    error
	)

	// Clean svmVolToNode map
	my.svmVolToNode = make(map[string]string)
	href := rest.BuildHref("", "node", nil, "", "", "", "", my.query)

	if result, err = collectors.InvokeRestCall(my.client, my.query, href, my.Logger); err != nil {
		return err
	}

	for _, volume := range result {
		volumeName := volume.Get("volume").String()
		vserverName := volume.Get("vserver").String()
		nodeName := volume.Get("node").String()
		key := vserverName + volumeName

		if _, ok := my.svmVolToNode[key]; ok {
			my.Logger.Warn().Str("key", key).Msg("Duplicate key found")
		}
		my.svmVolToNode[key] = nodeName
	}
	return nil
}

func (my *SnapMirror) getSVMPeerData(cluster string) error {
	// Clean svmPeerMap map
	my.svmPeerDataMap = make(map[string]Peer)
	fields := []string{"name", "peer.svm.name", "peer.cluster.name"}
	query := "api/svm/peers"
	href := rest.BuildHref("", strings.Join(fields, ","), []string{"peer.cluster.name=!" + cluster}, "", "", "", "", query)

	result, err := rest.Fetch(my.client, href)
	if err != nil {
		my.Logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return err
	}

	if len(result) == 0 {
		my.Logger.Debug().Msg("No svm peer found")
		return nil
	}

	for _, peerData := range result {
		localSvmName := peerData.Get("name").String()
		actualSvmName := peerData.Get("peer.svm.name").String()
		peerClusterName := peerData.Get("peer.cluster.name").String()
		my.svmPeerDataMap[localSvmName] = Peer{svm: actualSvmName, cluster: peerClusterName}
	}
	return nil
}

func (my *SnapMirror) updateSMLabels(data *matrix.Matrix) {
	var keys []string
	cluster, _ := data.GetGlobalLabels().GetHas("cluster")

	for key, instance := range data.GetInstances() {
		if instance.GetLabel("group_type") == "consistencygroup" {
			keys = append(keys, key)
		}
		volumeName := instance.GetLabel("source_volume")
		vserverName := instance.GetLabel("source_vserver")

		// Update source_node label in snapmirror
		if nodeName, ok := my.svmVolToNode[vserverName+volumeName]; ok {
			instance.SetLabel("source_node", nodeName)
		}

		// Update source_vserver in snapmirror (In case of inter-cluster SM - vserver name may differ)
		if peerDetail, ok := my.svmPeerDataMap[vserverName]; ok {
			instance.SetLabel("source_vserver", peerDetail.svm)
			instance.SetLabel("source_cluster", peerDetail.cluster)
		}

		if sourceCluster := instance.GetLabel("source_cluster"); sourceCluster == "" {
			instance.SetLabel("source_cluster", cluster)
		}

		// update the protectedBy and protectionSourceType fields and derivedRelationshipType in snapmirror_labels
		collectors.UpdateProtectedFields(instance)
	}

	// handle CG relationships
	my.handleCGRelationships(data, keys)

}

func (my *SnapMirror) handleCGRelationships(data *matrix.Matrix, keys []string) {

	for _, key := range keys {
		cgInstance := data.GetInstance(key)
		cgItemMappings := cgInstance.GetLabel("cg_item_mappings")
		// cg_item_mappings would be array of cgMapping. Example: vols1:@vold1,vols2:@vold2
		cgMappingData := strings.Split(cgItemMappings, ",")
		for _, cgMapping := range cgMappingData {
			var (
				cgVolumeInstance *matrix.Instance
				err              error
			)
			// cgMapping would be {source_volume}:@{destination volume}. Example: vols1:@vold1
			if volumes := strings.Split(cgMapping, ":@"); len(volumes) == 2 {
				sourceVol := volumes[0]
				destinationVol := volumes[1]
				/*
				 * cgVolumeInstanceKey: cgInstance's relationshipId + sourceVol + destinationVol
				 * Example:
				 * cgInstance's relationshipId: 958805a8-302a-11ed-a6ad-005056a79f6e, sourceVol: vols1, destinationVol: vold1.
				 * cgVolumeInstanceKey would be 958805a8-302a-11ed-a6ad-005056a79f6evols1vold1.
				 */
				cgVolumeInstanceKey := key + sourceVol + destinationVol

				if cgVolumeInstance, err = my.data.NewInstance(cgVolumeInstanceKey); err != nil {
					my.Logger.Error().Err(err).Str("Instance key", cgVolumeInstanceKey).Msg("")
					continue
				}

				for k, v := range cgInstance.GetLabels().Map() {
					cgVolumeInstance.SetLabel(k, v)
				}
				cgVolumeInstance.SetLabel("relationship_id", cgVolumeInstanceKey)
				cgVolumeInstance.SetLabel("source_volume", sourceVol)
				cgVolumeInstance.SetLabel("destination_volume", destinationVol)
			}
		}
	}

}
