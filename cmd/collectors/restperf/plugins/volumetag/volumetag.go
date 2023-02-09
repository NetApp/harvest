package volumetag

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
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
	if err = v.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout); err != nil {
		v.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = v.client.Init(5); err != nil {
		return err
	}
	return nil
}

func (v *VolumeTag) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		err error
	)

	query := "api/storage/volumes"

	href := rest.BuildHref("", "comment", nil, "", "", "", "", query)

	records, err := rest.Fetch(v.client, href)
	if err != nil {
		v.Logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	for _, volume := range records {

		if !volume.IsObject() {
			v.Logger.Warn().Str("type", volume.Type.String()).Msg("volume is not object, skipping")
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

	return nil, nil
}
