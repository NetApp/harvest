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
	"github.com/tidwall/gjson"
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
	svmPeerDataMap map[string]string // [peer SVM alias name] -> [peer SVM actual name] - [peer cluster name] map
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
	my.svmPeerDataMap = make(map[string]string)

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

		cluster, _ := data.GetGlobalLabels().GetHas("cluster")
		if err := my.getSVMPeerData(cluster); err != nil {
			return nil, err
		}
		my.Logger.Debug().Msg("updated svm peer")
	}

	// update volume instance labels
	my.updateSMLabels(data)
	my.nodeUpdCounter++

	return nil, nil
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
	my.svmPeerDataMap = make(map[string]string)
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
		my.svmPeerDataMap[localSvmName] = actualSvmName + ":" + peerClusterName
	}
	return nil
}

func (my *SnapMirror) updateSMLabels(data *matrix.Matrix) {
	cluster, _ := data.GetGlobalLabels().GetHas("cluster")
	for _, instance := range data.GetInstances() {
		volumeName := instance.GetLabel("source_volume")
		vserverName := instance.GetLabel("source_vserver")

		// Update source_node label in snapmirror
		if node, ok := my.svmVolToNode[vserverName+volumeName]; ok {
			instance.SetLabel("source_node", node)
		}

		// Update source_vserver in snapmirror (In case of inter-cluster SM- vserver name may differ)
		if peerDetail, ok := my.svmPeerDataMap[vserverName]; ok {
			peerData := strings.Split(peerDetail, ":")
			instance.SetLabel("source_vserver", peerData[0])
			instance.SetLabel("source_cluster", peerData[1])
		}

		if sourceCluster := instance.GetLabel("source_cluster"); sourceCluster == "" {
			instance.SetLabel("source_cluster", cluster)
		}

		// update the protectedBy and protectionSourceType fields and derivedRelationshipType in snapmirror_labels
		collectors.UpdateProtectedFields(instance)
	}
}
