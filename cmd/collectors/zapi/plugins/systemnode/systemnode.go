package systemnode

import (
	"errors"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

type SystemNode struct {
	*plugin.AbstractPlugin
	client         *zapi.Client
	partnerNameMap map[string]string // node-name -> partner name map
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SystemNode{AbstractPlugin: p}
}

func (a *SystemNode) Init() error {

	var err error

	if err = a.InitAbc(); err != nil {
		return err
	}

	if a.client, err = zapi.New(conf.ZapiPoller(a.ParentParams), a.Auth); err != nil {
		a.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = a.client.Init(5); err != nil {
		return err
	}

	a.partnerNameMap = make(map[string]string)
	return nil
}

func (a *SystemNode) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[a.Object]

	// invoke system-get-node-info-iter zapi and populate node info
	if err := a.getPartnerNodeInfo(); err != nil {
		if errors.Is(err, errs.ErrNoInstance) {
			a.Logger.Debug().Err(err).Msg("Failed to collect cloud store data")
		}
	}

	// update node instance label with partner info
	for nodeName, node := range data.GetInstances() {
		node.SetLabel("ha_partner", a.partnerNameMap[nodeName])
	}
	return nil, nil
}

func (a *SystemNode) getPartnerNodeInfo() error {
	var (
		result []*node.Node
		err    error
	)

	// system-get-node-info-iter zapi
	a.partnerNameMap = make(map[string]string)
	request := node.NewXMLS("system-get-node-info-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)

	desired := node.NewXMLS("desired-attributes")
	systemInfo := node.NewXMLS("system-info")
	systemInfo.NewChildS("partner-system-name", "")
	systemInfo.NewChildS("system-name", "")
	desired.AddChild(systemInfo)
	request.AddChild(desired)

	if result, err = a.client.InvokeZapiCall(request); err != nil {
		return err
	}

	if len(result) == 0 || result == nil {
		return errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, objectStore := range result {
		partnerName := objectStore.GetChildContentS("partner-system-name")
		nodeName := objectStore.GetChildContentS("system-name")
		a.partnerNameMap[nodeName] = partnerName
	}
	return nil
}
