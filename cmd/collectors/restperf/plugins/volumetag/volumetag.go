package volumetag

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"time"
)

type VolumeTag struct {
	*plugin.AbstractPlugin
	client *rest.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &VolumeTag{AbstractPlugin: p}
}

func (v *VolumeTag) Init() error {
	var err error
	if err := v.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout, v.Auth); err != nil {
		v.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	return v.client.Init(5)
}

func (v *VolumeTag) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	var (
		err error
	)

	data := dataMap[v.Object]
	v.client.Metadata.Reset()

	query := "api/storage/volumes"

	href := rest.NewHrefBuilder().
		APIPath(query).
		MaxRecords(collectors.DefaultBatchSize).
		Fields([]string{"comment"}).
		Build()

	records, err := rest.FetchAll(v.client, href)
	if err != nil {
		v.SLogger.Error("Failed to fetch data", slogx.Err(err), slog.String("href", href))
		return nil, nil, err
	}

	if len(records) == 0 {
		return nil, nil, nil
	}

	for _, volume := range records {

		if !volume.IsObject() {
			v.SLogger.Warn("volume is not object, skipping", slog.String("type", volume.Type.String()))
			continue
		}
		key := volume.Get("uuid").String()
		comment := volume.Get("comment").String()
		instance := data.GetInstance(key)
		if instance != nil && comment != "" {
			instance.SetLabel("comment", comment)
		}
	}

	if exportOption := v.ParentParams.GetChildS("export_options"); exportOption != nil {
		if exportedKeys := exportOption.GetChildS("instance_keys"); exportedKeys != nil {
			if exportedKeys.GetChildByContent("comment") == nil {
				exportedKeys.NewChildS("", "comment")
			}
		}
	}

	return nil, v.client.Metadata, nil
}
