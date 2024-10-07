package fcvi

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
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
	if err := f.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if f.client, err = rest.New(conf.ZapiPoller(f.ParentParams), timeout, f.Auth); err != nil {
		f.SLogger.Error("connecting", slog.Any("err", err))
		return err
	}

	return f.client.Init(5)
}

func (f *FCVI) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[f.Object]
	f.client.Metadata.Reset()

	query := "api/private/cli/metrocluster/interconnect/adapter"
	fields := []string{"node", "adapter", "port_name"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()
	records, err := rest.FetchAll(f.client, href)
	if err != nil {
		f.SLogger.Error("Failed to fetch data", slog.Any("err", err), slog.String("href", href))
		return nil, nil, err
	}

	if len(records) == 0 {
		return nil, nil, nil
	}

	for _, adapterData := range records {
		if !adapterData.IsObject() {
			f.SLogger.Warn("adapter is not object, skipping", slog.String("type", adapterData.Type.String()))
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

	return nil, f.client.Metadata, nil
}
