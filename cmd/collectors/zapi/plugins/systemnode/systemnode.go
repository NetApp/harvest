package systemnode

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

type SystemNode struct {
	*plugin.AbstractPlugin
	client *zapi.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SystemNode{AbstractPlugin: p}
}

func (s *SystemNode) Init(remote conf.Remote) error {

	var err error

	if err := s.InitAbc(); err != nil {
		return err
	}

	if s.client, err = zapi.New(conf.ZapiPoller(s.ParentParams), s.Auth); err != nil {
		s.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := s.client.Init(5, remote); err != nil {
		return err
	}

	return nil
}

func (s *SystemNode) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[s.Object]
	s.client.Metadata.Reset()
	nodeStateMap := make(map[string]string)

	// invoke service-processor-get-iter zapi and populate the BMC firmware version
	serviceProcessorMap, err := s.getServiceProcessor()
	if err != nil {
		s.SLogger.Error("Failed to collect service processor info", slogx.Err(err))
	}

	// invoke system-get-node-info-iter zapi and populate node info
	partnerNameMap, err := s.getPartnerNodeInfo()
	if err != nil {
		s.SLogger.Error("Failed to collect partner node info", slogx.Err(err))
	}

	for _, aNode := range data.GetInstanceKeys() {
		nodeStateMap[aNode] = data.GetInstance(aNode).GetLabel("healthy")
	}

	// update node instance with partner, partner_healthy, and BMC version info
	for nodeName, inst := range data.GetInstances() {
		inst.SetLabel("ha_partner", partnerNameMap[nodeName])
		inst.SetLabel("partner_healthy", nodeStateMap[inst.GetLabel("ha_partner")])
		inst.SetLabel("bmc_firmware_version", serviceProcessorMap[nodeName])
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
		s.SLogger.Debug("no records found")
		return nodePartnerNodeMap, nil
	}

	for _, objectStore := range result {
		partnerName := objectStore.GetChildContentS("partner-system-name")
		nodeName := objectStore.GetChildContentS("system-name")
		nodePartnerNodeMap[nodeName] = partnerName
	}
	return nodePartnerNodeMap, nil
}

func (s *SystemNode) getServiceProcessor() (map[string]string, error) {
	var (
		result []*node.Node
		err    error
	)

	nodeFirmwareMap := make(map[string]string) // node-name -> BMC firmware version map
	request := node.NewXMLS("service-processor-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)

	desired := node.NewXMLS("desired-attributes")
	spInfo := node.NewXMLS("service-processor-info")
	spInfo.NewChildS("firmware-version", "")
	desired.AddChild(spInfo)
	request.AddChild(desired)

	if result, err = s.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		s.SLogger.Debug("no records found")
		return nodeFirmwareMap, nil
	}

	for _, objectStore := range result {
		firmware := objectStore.GetChildContentS("firmware-version")
		nodeName := objectStore.GetChildContentS("node")
		nodeFirmwareMap[nodeName] = firmware
	}
	return nodeFirmwareMap, nil
}
