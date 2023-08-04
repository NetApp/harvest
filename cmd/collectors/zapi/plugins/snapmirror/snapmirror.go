/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package snapmirror

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"regexp"
	"strings"
)

type SnapMirror struct {
	*plugin.AbstractPlugin
	client         *zapi.Client
	nodeUpdCounter int
	svmPeerDataMap map[string]Peer // [peer SVM alias name] -> [peer detail] map
}

type Peer struct {
	svm     string
	cluster string
}

var flexgroupConstituentName = regexp.MustCompile(`^(.*)__(\d{4})$`)

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SnapMirror{AbstractPlugin: p}
}
func (my *SnapMirror) Init() error {
	var err error
	if err = my.InitAbc(); err != nil {
		return err
	}
	if my.client, err = zapi.New(conf.ZapiPoller(my.ParentParams), my.Auth); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}
	if err = my.client.Init(5); err != nil {
		return err
	}
	my.nodeUpdCounter = 0
	my.svmPeerDataMap = make(map[string]Peer)

	my.Logger.Debug().Msg("plugin initialized")
	return nil
}
func (my *SnapMirror) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[my.Object]
	destUpdCount := 0
	srcUpdCount := 0

	if cluster, ok := data.GetGlobalLabels().GetHas("cluster"); ok {
		if err := my.getSVMPeerData(cluster); err != nil {
			return nil, err
		}
		my.Logger.Debug().Msg("updated svm peer detail")
	}

	lastTransferSizeMetric := data.GetMetric("snapmirror-info.last-transfer-size")
	lagTimeMetric := data.GetMetric("snapmirror-info.lag-time")
	if lastTransferSizeMetric == nil {
		return nil, errs.New(errs.ErrNoMetric, "last_transfer_size")
	}
	if lagTimeMetric == nil {
		return nil, errs.New(errs.ErrNoMetric, "lag_time")
	}

	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		// Zapi call with `expand=true` returns all the constituent's relationships. We do not want to export them.
		if match := flexgroupConstituentName.FindStringSubmatch(instance.GetLabel("destination_volume")); len(match) == 3 {
			instance.SetExportable(false)
			continue
		}

		if my.client.IsClustered() {
			vserverName := instance.GetLabel("source_vserver")
			// Update source_vserver in snapmirror (In case of inter-cluster SM - vserver name may differ)
			if peerDetail, ok := my.svmPeerDataMap[vserverName]; ok {
				instance.SetLabel("source_vserver", peerDetail.svm)
				instance.SetLabel("source_cluster", peerDetail.cluster)
			}

			// It's local relationship, so updating the source_cluster and local labels
			if sourceCluster := instance.GetLabel("source_cluster"); sourceCluster == "" {
				instance.SetLabel("source_cluster", my.client.Name())
				instance.SetLabel("local", "true")
			}

			// update the protectedBy and protectionSourceType fields and derivedRelationshipType in snapmirror_labels
			collectors.UpdateProtectedFields(instance)

			// Update lag time based on checks
			collectors.UpdateLagTime(instance, lastTransferSizeMetric, lagTimeMetric, my.Logger)
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
						srcUpdCount++
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
						destUpdCount++
					}
				} else {
					break
				}
			}
		}
	}
	my.Logger.Debug().Msgf("updated %d destination and %d source nodes", destUpdCount, srcUpdCount)
	return nil, nil
}

func (my *SnapMirror) getSVMPeerData(cluster string) error {
	var (
		result []*node.Node
		err    error
	)

	request := node.NewXMLS("vserver-peer-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	// Fetching only remote vserver-peer
	query := request.NewChildS("query", "")
	vserverPeerInfo := query.NewChildS("vserver-peer-info", "")
	vserverPeerInfo.NewChildS("peer-cluster", "!"+cluster)

	// Clean svmPeerMap map
	my.svmPeerDataMap = make(map[string]Peer)

	// fetching only remote vserver peer data
	if result, err = my.client.InvokeZapiCall(request); err != nil {
		return err
	}

	if len(result) == 0 || result == nil {
		my.Logger.Debug().Msg("No vserver peer found")
	}

	for _, peerData := range result {
		localSvmName := peerData.GetChildContentS("peer-vserver")
		actualSvmName := peerData.GetChildContentS("remote-vserver-name")
		peerClusterName := peerData.GetChildContentS("peer-cluster")
		my.svmPeerDataMap[localSvmName] = Peer{svm: actualSvmName, cluster: peerClusterName}
	}
	return nil
}
