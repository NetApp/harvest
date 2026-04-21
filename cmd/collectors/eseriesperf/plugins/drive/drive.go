package drive

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
)

type Drive struct {
	*plugin.AbstractPlugin
	client     *rest.Client
	schedule   int
	trayLabels map[string]string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Drive{AbstractPlugin: p}
}

func (d *Drive) Init(remote conf.Remote) error {
	if err := d.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	poller, err := conf.PollerNamed(d.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, d.SLogger)
	if d.client, err = rest.New(poller, timeout, credentials, ""); err != nil {
		return err
	}

	if !d.Options.IsTest {
		if err := d.client.Init(1, remote); err != nil {
			return err
		}
	}

	d.trayLabels = make(map[string]string)
	d.schedule = d.SetPluginInterval()
	return nil
}

func (d *Drive) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[d.Object]

	arrayID := d.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		d.SLogger.Warn("arrayID not found in ParentParams, skipping tray labels")
		return nil, nil, nil
	}

	if d.schedule >= d.PluginInvocationRate {
		d.schedule = 0
		d.refreshTrayLabels(arrayID)
	}
	d.schedule++

	d.applyTrayLabels(data)

	return nil, nil, nil
}

func (d *Drive) refreshTrayLabels(arrayID string) {
	d.trayLabels = make(map[string]string)

	trayLabels, err := d.buildTrayLabelMap(arrayID)
	if err != nil {
		d.SLogger.Warn("Failed to build tray label map", slogx.Err(err))
		return
	}

	d.trayLabels = trayLabels
	d.SLogger.Debug("Refreshed tray labels", slog.Int("count", len(d.trayLabels)))
}

func (d *Drive) buildTrayLabelMap(arrayID string) (map[string]string, error) {
	trayLabels := make(map[string]string)

	apiPath := d.client.APIPath + "/storage-systems/" + arrayID + "/hardware-inventory"
	results, err := d.client.Fetch(apiPath, nil)
	if err != nil {
		return trayLabels, fmt.Errorf("failed to fetch hardware-inventory: %w", err)
	}

	if len(results) == 0 {
		return trayLabels, nil
	}

	trays := results[0].Get("trays")
	if !trays.Exists() || !trays.IsArray() {
		return trayLabels, nil
	}

	for _, tray := range trays.Array() {
		trayRef := tray.Get("trayRef").ClonedString()
		if trayRef == "" {
			trayRef = tray.Get("id").ClonedString()
		}
		if trayRef == "" {
			continue
		}

		trayID := tray.Get("trayId").ClonedString()
		if trayID == "" {
			trayID = tray.Get("physicalLocation.label").ClonedString()
		}

		if trayID != "" {
			trayLabels[trayRef] = trayID
		}
	}

	d.SLogger.Debug("Built tray label map", slog.Int("entries", len(trayLabels)))
	return trayLabels, nil
}

func (d *Drive) applyTrayLabels(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		rawTrayRef := instance.GetLabel("tray_id")
		if rawTrayRef == "" {
			continue
		}

		if label, ok := d.trayLabels[rawTrayRef]; ok {
			instance.SetLabel("tray_id", label)
		} else {
			d.SLogger.Debug("Tray label not found in cache",
				slog.String("tray_ref", rawTrayRef))
		}
	}
}
