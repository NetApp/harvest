package fcvi

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strings"
	"time"
)

type FCVI struct {
	*plugin.AbstractPlugin
	client *rest.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &FCVI{AbstractPlugin: p}
}

func (f *FCVI) Init() error {
	var err error
	if err = f.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if f.client, err = rest.New(conf.ZapiPoller(f.ParentParams), timeout, f.Auth); err != nil {
		f.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = f.client.Init(5); err != nil {
		return err
	}
	return nil
}

func (f *FCVI) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	data := dataMap[f.Object]
	query := "api/private/cli/metrocluster/interconnect/adapter"
	fields := []string{"node", "adapter", "port_name"}
	href := rest.BuildHref("", strings.Join(fields, ","), nil, "", "", "", "", query)

	records, err := rest.Fetch(f.client, href)
	if err != nil {
		f.Logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	for _, adapterData := range records {
		if !adapterData.IsObject() {
			f.Logger.Warn().Str("type", adapterData.Type.String()).Msg("adapter is not object, skipping")
			continue
		}
		node := adapterData.Get("node").String()
		adapter := adapterData.Get("adapter").String()
		port := adapterData.Get("port_name").String()

		// Fetch instance and add port label
		if instance := data.GetInstance(node + ":" + adapter); instance != nil {
			instance.SetLabel("port", port)
		}
	}

	return nil, nil
}
