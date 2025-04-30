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
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
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

func (m *SnapMirror) Init(remote conf.Remote) error {
	var err error
	if err := m.InitAbc(); err != nil {
		return err
	}
	if m.client, err = zapi.New(conf.ZapiPoller(m.ParentParams), m.Auth); err != nil {
		m.SLogger.Error("connecting", slogx.Err(err))
		return err
	}
	if err := m.client.Init(5, remote); err != nil {
		return err
	}
	m.nodeUpdCounter = 0
	m.svmPeerDataMap = make(map[string]Peer)

	m.SLogger.Debug("plugin initialized")
	return nil
}
func (m *SnapMirror) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[m.Object]
	m.client.Metadata.Reset()

	destUpdCount := 0
	srcUpdCount := 0

	if cluster, ok := data.GetGlobalLabels()["cluster"]; ok {
		if err := m.getSVMPeerData(cluster); err != nil {
			return nil, nil, err
		}
		m.SLogger.Debug("updated svm peer detail")
	}

	lastTransferSizeMetric := data.GetMetric("snapmirror-info.last-transfer-size")
	lagTimeMetric := data.GetMetric("snapmirror-info.lag-time")
	if lastTransferSizeMetric == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "last_transfer_size")
	}
	if lagTimeMetric == nil {
		return nil, nil, errs.New(errs.ErrNoMetric, "lag_time")
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

		if m.client.IsClustered() {
			vserverName := instance.GetLabel("source_vserver")
			// Update source_vserver in snapmirror (In case of inter-cluster SM - vserver name may differ)
			if peerDetail, ok := m.svmPeerDataMap[vserverName]; ok {
				instance.SetLabel("source_vserver", peerDetail.svm)
				instance.SetLabel("source_cluster", peerDetail.cluster)
			}

			// It's local relationship, so updating the source_cluster and local labels
			if sourceCluster := instance.GetLabel("source_cluster"); sourceCluster == "" {
				instance.SetLabel("source_cluster", m.client.Name())
				instance.SetLabel("local", "true")
			}

			// update the protectedBy and protectionSourceType fields and derivedRelationshipType in snapmirror_labels
			collectors.UpdateProtectedFields(instance)

			// Update lag time based on checks
			collectors.UpdateLagTime(instance, lastTransferSizeMetric, lagTimeMetric)
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
	m.SLogger.Debug(
		"updated source and destination nodes",
		slog.Int("destUpdCount", destUpdCount),
		slog.Int("srcUpdCount", srcUpdCount),
	)

	return nil, m.client.Metadata, nil
}

func (m *SnapMirror) getSVMPeerData(cluster string) error {
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
	m.svmPeerDataMap = make(map[string]Peer)

	// fetching only remote vserver peer data
	if result, err = m.client.InvokeZapiCall(request); err != nil {
		return err
	}

	if len(result) == 0 || result == nil {
		m.SLogger.Debug("No vserver peer found")
	}

	for _, peerData := range result {
		localSvmName := peerData.GetChildContentS("peer-vserver")
		actualSvmName := peerData.GetChildContentS("remote-vserver-name")
		peerClusterName := peerData.GetChildContentS("peer-cluster")
		m.svmPeerDataMap[localSvmName] = Peer{svm: actualSvmName, cluster: peerClusterName}
	}
	return nil
}
