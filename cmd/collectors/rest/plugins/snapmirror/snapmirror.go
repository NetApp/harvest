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
	"github.com/netapp/harvest/v2/pkg/util"
	"path/filepath"
	"strings"
	"time"
)

const PluginInvocationRate = 10

type SnapMirror struct {
	*plugin.AbstractPlugin
	data               *matrix.Matrix
	client             *rest.Client
	currentVal         int
	svmPeerDataMap     map[string]Peer   // [peer SVM alias name] -> [peer detail] map
	clusterPeerDataMap map[string]string // [peer Cluster alias name] -> [peer Cluster actual name] map
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

	if err := my.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if my.client, err = rest.New(conf.ZapiPoller(my.ParentParams), timeout, my.Auth); err != nil {
		my.Logger.Error().Err(err).Msg("connecting")
		return err
	}

	if err := my.client.Init(5); err != nil {
		return err
	}

	my.svmPeerDataMap = make(map[string]Peer)
	my.clusterPeerDataMap = make(map[string]string)

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

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	my.currentVal = PluginInvocationRate
	return nil
}

func (my *SnapMirror) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	// Purge and reset data
	data := dataMap[my.Object]
	my.data.PurgeInstances()
	my.data.Reset()
	my.client.Metadata.Reset()

	// Set all global labels from Rest.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	if my.currentVal >= PluginInvocationRate {
		my.currentVal = 0

		if cluster, ok := data.GetGlobalLabels()["cluster"]; ok {
			if err := my.getSVMPeerData(cluster); err != nil {
				return nil, nil, err
			}
			my.Logger.Debug().Msg("updated svm peer map detail")
			if err := my.getClusterPeerData(); err != nil {
				return nil, nil, err
			}
			my.Logger.Debug().Msg("updated cluster peer map detail")
		}
	}

	// update volume instance labels
	my.updateSMLabels(data)
	my.currentVal++

	return []*matrix.Matrix{my.data}, my.client.Metadata, nil
}

func (my *SnapMirror) getSVMPeerData(cluster string) error {
	// Clean svmPeerMap map
	my.svmPeerDataMap = make(map[string]Peer)
	fields := []string{"name", "peer.svm.name", "peer.cluster.name"}
	query := "api/svm/peers"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter([]string{"peer.cluster.name=!" + cluster}).
		Build()

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

func (my *SnapMirror) getClusterPeerData() error {
	// Clean clusterPeerDataMap map
	my.clusterPeerDataMap = make(map[string]string)
	fields := []string{"name", "remote.name"}
	query := "api/cluster/peers"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Build()

	result, err := rest.Fetch(my.client, href)
	if err != nil {
		my.Logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return err
	}

	if len(result) == 0 {
		my.Logger.Debug().Msg("No cluster peer found")
		return nil
	}

	for _, peerData := range result {
		localClusterName := peerData.Get("name").String()
		actualClusterName := peerData.Get("remote.name").String()
		my.clusterPeerDataMap[localClusterName] = actualClusterName
	}
	return nil
}

func (my *SnapMirror) updateSMLabels(data *matrix.Matrix) {
	var cgKeys []string
	var svmDrKeys []string
	cluster := data.GetGlobalLabels()["cluster"]

	lastTransferSizeMetric := data.GetMetric("last_transfer_size")
	lagTimeMetric := data.GetMetric("lag_time")
	if lastTransferSizeMetric == nil {
		return
	}
	if lagTimeMetric == nil {
		return
	}

	for key, instance := range data.GetInstances() {
		if instance.GetLabel("group_type") == "consistencygroup" {
			cgKeys = append(cgKeys, key)
		} else if instance.GetLabel("group_type") == "vserver" {
			svmDrKeys = append(svmDrKeys, key)
		}
		vserverName := instance.GetLabel("source_vserver")

		// Update source_vserver in snapmirror (In case of inter-cluster SM - vserver name may differ)
		if peerDetail, ok := my.svmPeerDataMap[vserverName]; ok {
			instance.SetLabel("source_vserver", peerDetail.svm)
			// Update source_cluster in snapmirror (In case of inter-cluster SM - cluster name may differ)
			if peerClusterName, exist := my.clusterPeerDataMap[peerDetail.cluster]; exist {
				instance.SetLabel("source_cluster", peerClusterName)
			}
		}

		if sourceCluster := instance.GetLabel("source_cluster"); sourceCluster == "" {
			instance.SetLabel("source_cluster", cluster)
			instance.SetLabel("local", "true")
		}

		// update the protectedBy and protectionSourceType fields and derivedRelationshipType in snapmirror_labels
		collectors.UpdateProtectedFields(instance)

		// Update lag time based on checks
		collectors.UpdateLagTime(instance, lastTransferSizeMetric, lagTimeMetric, my.Logger)
	}

	// handle CG relationships
	my.handleCGRelationships(data, cgKeys)

	// handle SVM-DR relationships
	my.handleSVMDRRelationships(data, svmDrKeys)
}

func (my *SnapMirror) handleCGRelationships(data *matrix.Matrix, keys []string) {
	for _, key := range keys {
		cgInstance := data.GetInstance(key)
		// find cgName from the destination_location, source_location
		cgInstance.SetLabel("destination_cg_name", filepath.Base(cgInstance.GetLabel("destination_location")))
		cgInstance.SetLabel("source_cg_name", filepath.Base(cgInstance.GetLabel("source_location")))

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
					my.Logger.Error().Err(err).Str("Instance key", cgVolumeInstanceKey).Send()
					continue
				}

				for k, v := range cgInstance.GetLabels() {
					cgVolumeInstance.SetLabel(k, v)
				}
				cgVolumeInstance.SetLabel("relationship_id", "")
				cgVolumeInstance.SetLabel("source_volume", sourceVol)
				cgVolumeInstance.SetLabel("destination_volume", destinationVol)
			}
		}
	}
}

func (my *SnapMirror) handleSVMDRRelationships(data *matrix.Matrix, keys []string) {
	for _, key := range keys {
		svmDrInstance := data.GetInstance(key)
		// check source_volume and destination_volume in svm-dr relationships to identify volumes in svm
		sourceVolume := svmDrInstance.GetLabel("source_volume")
		destinationVolume := svmDrInstance.GetLabel("destination_volume")

		if sourceVolume != "" && destinationVolume != "" {
			svmDrInstance.SetLabel("relationship_id", "")
		}
	}
}
