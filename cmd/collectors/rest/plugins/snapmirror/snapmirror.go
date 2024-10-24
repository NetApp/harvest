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
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"path/filepath"
	"slices"
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

func (m *SnapMirror) Init() error {

	var err error

	if err := m.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if m.client, err = rest.New(conf.ZapiPoller(m.ParentParams), timeout, m.Auth); err != nil {
		m.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := m.client.Init(5); err != nil {
		return err
	}

	m.svmPeerDataMap = make(map[string]Peer)
	m.clusterPeerDataMap = make(map[string]string)

	m.data = matrix.New(m.Parent+".SnapMirror", "snapmirror", "snapmirror")

	exportOptions := node.NewS("export_options")
	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")

	if exportOption := m.ParentParams.GetChildS("export_options"); exportOption != nil {
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
	m.data.SetExportOptions(exportOptions)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	m.currentVal = PluginInvocationRate
	return nil
}

func (m *SnapMirror) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	// Purge and reset data
	data := dataMap[m.Object]
	m.data.PurgeInstances()
	m.data.Reset()
	m.client.Metadata.Reset()

	// Set all global labels from Rest.go if already not exist
	m.data.SetGlobalLabels(data.GetGlobalLabels())

	if m.currentVal >= PluginInvocationRate {
		m.currentVal = 0

		if cluster, ok := data.GetGlobalLabels()["cluster"]; ok {
			if err := m.getSVMPeerData(cluster); err != nil {
				return nil, nil, err
			}
			m.SLogger.Debug("updated svm peer map detail")
			if err := m.getClusterPeerData(); err != nil {
				return nil, nil, err
			}
			m.SLogger.Debug("updated cluster peer map detail")
		}
	}

	// update volume instance labels
	m.updateSMLabels(data)
	m.currentVal++

	return []*matrix.Matrix{m.data}, m.client.Metadata, nil
}

func (m *SnapMirror) getSVMPeerData(cluster string) error {
	// Clean svmPeerMap map
	m.svmPeerDataMap = make(map[string]Peer)
	fields := []string{"name", "peer.svm.name", "peer.cluster.name"}
	query := "api/svm/peers"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Filter([]string{"peer.cluster.name=!" + cluster}).
		Build()

	result, err := rest.FetchAll(m.client, href)
	if err != nil {
		m.SLogger.Error("Failed to fetch data", slogx.Err(err), slog.String("href", href))
		return err
	}

	if len(result) == 0 {
		m.SLogger.Debug("No svm peer found")
		return nil
	}

	for _, peerData := range result {
		localSvmName := peerData.Get("name").String()
		actualSvmName := peerData.Get("peer.svm.name").String()
		peerClusterName := peerData.Get("peer.cluster.name").String()
		m.svmPeerDataMap[localSvmName] = Peer{svm: actualSvmName, cluster: peerClusterName}
	}
	return nil
}

func (m *SnapMirror) getClusterPeerData() error {
	// Clean clusterPeerDataMap map
	m.clusterPeerDataMap = make(map[string]string)
	fields := []string{"name", "remote.name"}
	query := "api/cluster/peers"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	result, err := rest.FetchAll(m.client, href)
	if err != nil {
		m.SLogger.Error("Failed to fetch data", slogx.Err(err), slog.String("href", href))
		return err
	}

	if len(result) == 0 {
		m.SLogger.Debug("No cluster peer found")
		return nil
	}

	for _, peerData := range result {
		localClusterName := peerData.Get("name").String()
		actualClusterName := peerData.Get("remote.name").String()
		m.clusterPeerDataMap[localClusterName] = actualClusterName
	}
	return nil
}

func (m *SnapMirror) updateSMLabels(data *matrix.Matrix) {
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
		if peerDetail, ok := m.svmPeerDataMap[vserverName]; ok {
			instance.SetLabel("source_vserver", peerDetail.svm)
			// Update source_cluster in snapmirror (In case of inter-cluster SM - cluster name may differ)
			if peerClusterName, exist := m.clusterPeerDataMap[peerDetail.cluster]; exist {
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
		collectors.UpdateLagTime(instance, lastTransferSizeMetric, lagTimeMetric, m.SLogger)
	}

	// handle CG relationships
	m.handleCGRelationships(data, cgKeys)

	// handle SVM-DR relationships
	m.handleSVMDRRelationships(data, svmDrKeys)
}

func (m *SnapMirror) handleCGRelationships(data *matrix.Matrix, keys []string) {
	for _, key := range keys {
		var cgSourceVolumes, cgDestinationVolumes []string
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

				if cgVolumeInstance, err = m.data.NewInstance(cgVolumeInstanceKey); err != nil {
					m.SLogger.Error("", slogx.Err(err), slog.String("key", cgVolumeInstanceKey))
					continue
				}

				for k, v := range cgInstance.GetLabels() {
					cgVolumeInstance.SetLabel(k, v)
				}
				cgVolumeInstance.SetLabel("relationship_id", "")
				cgVolumeInstance.SetLabel("source_volume", sourceVol)
				cgVolumeInstance.SetLabel("destination_volume", destinationVol)
				cgSourceVolumes = append(cgSourceVolumes, sourceVol)
				cgDestinationVolumes = append(cgDestinationVolumes, destinationVol)
			}
		}
		// Update parent CG source and destination volumes
		slices.Sort(cgSourceVolumes)
		slices.Sort(cgDestinationVolumes)
		cgInstance.SetLabel("source_volume", strings.Join(cgSourceVolumes, ","))
		cgInstance.SetLabel("destination_volume", strings.Join(cgDestinationVolumes, ","))
	}
}

func (m *SnapMirror) handleSVMDRRelationships(data *matrix.Matrix, keys []string) {
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
