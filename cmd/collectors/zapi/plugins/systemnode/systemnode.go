package systemnode

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
)

type SystemNode struct {
	*plugin.AbstractPlugin
	client *zapi.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SystemNode{AbstractPlugin: p}
}

func (s *SystemNode) Init() error {

	var err error

	if err = s.InitAbc(); err != nil {
		return err
	}

	if s.client, err = zapi.New(conf.ZapiPoller(s.ParentParams), s.Auth); err != nil {
		s.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = s.client.Init(5); err != nil {
		return err
	}

	return nil
}

func (s *SystemNode) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[s.Object]
	s.client.Metadata.Reset()
	nodeStateMap := make(map[string]string)

	// invoke system-get-node-info-iter zapi and populate node info
	partnerNameMap, err := s.getPartnerNodeInfo()
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to collect partner node detail")
	}

	for _, node := range data.GetInstanceKeys() {
		nodeStateMap[node] = data.GetInstance(node).GetLabel("healthy")
	}

	// update node instance with partner and partner_healthy info
	for nodeName, node := range data.GetInstances() {
		node.SetLabel("ha_partner", partnerNameMap[nodeName])
		node.SetLabel("partner_healthy", nodeStateMap[node.GetLabel("ha_partner")])
	}

	// update node metrics with partner info
	for _, metric := range data.GetMetrics() {
		metric.SetLabel("ha_partner", partnerNameMap[metric.GetLabel("node")])
	}
	return nil, s.client.Metadata, nil
}

func (s *SystemNode) getPartnerNodeInfo() (map[string]string, error) {
	var (
		result             []*node.Node
		nodePartnerNodeMap map[string]string // node-name -> partner-node-name map
		err                error
	)

	// system-get-node-info-iter zapi
	nodePartnerNodeMap = make(map[string]string)
	request := node.NewXMLS("system-get-node-info-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)

	desired := node.NewXMLS("desired-attributes")
	systemInfo := node.NewXMLS("system-info")
	systemInfo.NewChildS("partner-system-name", "")
	systemInfo.NewChildS("system-name", "")
	desired.AddChild(systemInfo)
	request.AddChild(desired)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		s.Logger.Debug().Err(err).Msg("no records found")
		return nodePartnerNodeMap, nil
	}

	for _, objectStore := range result {
		partnerName := objectStore.GetChildContentS("partner-system-name")
		nodeName := objectStore.GetChildContentS("system-name")
		nodePartnerNodeMap[nodeName] = partnerName
	}
	return nodePartnerNodeMap, nil
}
