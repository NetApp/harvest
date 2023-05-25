package fcvi

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

const batchSize = "500"

type FCVI struct {
	*plugin.AbstractPlugin
	client *zapi.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FCVI{AbstractPlugin: p}
}

func (f *FCVI) Init() error {
	var err error
	if err = f.InitAbc(); err != nil {
		return err
	}

	if f.client, err = zapi.New(conf.ZapiPoller(f.ParentParams), f.Auth); err != nil {
		f.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}
	if err = f.client.Init(5); err != nil {
		return err
	}
	return nil
}

func (f *FCVI) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		result []*node.Node
		err    error
	)

	adapterPortMap := make(map[string]string)
	data := dataMap[f.Object]
	query := "metrocluster-interconnect-adapter-get-iter"
	request := node.NewXMLS(query)
	request.NewChildS("max-records", batchSize)
	desired := node.NewXMLS("desired-attributes")
	metroclusterInterconnectAdapterAttributes := node.NewXMLS("metrocluster-interconnect-adapter")
	metroclusterInterconnectAdapterAttributes.NewChildS("adapter-name", "")
	metroclusterInterconnectAdapterAttributes.NewChildS("node-name", "")
	metroclusterInterconnectAdapterAttributes.NewChildS("port-name", "")
	desired.AddChild(metroclusterInterconnectAdapterAttributes)
	request.AddChild(desired)

	if result, err = f.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}
	f.Logger.Info().Msgf("%d", len(result))

	for _, adapterData := range result {
		adapter := adapterData.GetChildContentS("adapter-name")
		node := adapterData.GetChildContentS("node-name")
		port := adapterData.GetChildContentS("port-name")
		adapterPortMap[node+adapter] = port
	}

	// we would not use getInstance() as key would be `sti8300mcc-215:kernel:fcvi_device_1`
	for _, instance := range data.GetInstances() {
		if port, ok := adapterPortMap[instance.GetLabel("node")+instance.GetLabel("fcvi")]; ok {
			instance.SetLabel("port", port)
		}
	}
	return nil, nil
}
