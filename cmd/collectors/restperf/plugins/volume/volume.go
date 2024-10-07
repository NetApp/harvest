package volume

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

type Volume struct {
	*plugin.AbstractPlugin
	styleType           string
	includeConstituents bool
	client              *rest.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) Init() error {
	var err error
	if err := v.InitAbc(); err != nil {
		return err
	}

	v.styleType = "style"

	if v.Params.HasChildS("historicalLabels") {
		v.styleType = "type"
	}

	// Read template to decide inclusion of flexgroup constituents
	v.includeConstituents = collectors.ReadPluginKey(v.Params, "include_constituents")

	if v.Options.IsTest {
		v.client = &rest.Client{Metadata: &util.Metadata{}}
		return nil
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout, v.Auth); err != nil {
		v.SLogger.Error("connecting", slog.Any("err", err))
		return err
	}

	return v.client.Init(5)
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[v.Object]
	style := v.styleType
	opsKeyPrefix := "temp_"
	volumesMap := v.fetchVolumes()

	return collectors.ProcessFlexGroupData(v.SLogger, data, style, v.includeConstituents, opsKeyPrefix, volumesMap)
}

func (v *Volume) fetchVolumes() map[string]string {
	volumesMap := make(map[string]string)
	query := "api/storage/volumes"

	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields([]string{"volume", "volume_style_extended"}).
		Build()

	records, err := rest.FetchAll(v.client, href)
	if err != nil {
		v.SLogger.Error("Failed to fetch data", slog.Any("err", err), slog.String("href", href))
		return nil
	}

	if len(records) == 0 {
		return nil
	}

	for _, volume := range records {
		if !volume.IsObject() {
			v.SLogger.Warn("volume is not object, skipping", slog.String("type", volume.Type.String()))
			continue
		}
		styleExtended := volume.Get("volume_style_extended").String()
		name := volume.Get("volume").String()
		volumesMap[name] = styleExtended
	}

	return volumesMap
}
