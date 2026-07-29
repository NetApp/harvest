package workload

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

type Workload struct {
	*plugin.AbstractPlugin
	client         *rest.Client
	schedule       int
	workloadLabels map[string]string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Workload{AbstractPlugin: p}
}

func (w *Workload) Init(remote conf.Remote) error {
	if err := w.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	poller, err := conf.PollerNamed(w.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, w.SLogger)
	if w.client, err = rest.New(poller, timeout, credentials, ""); err != nil {
		return err
	}

	if !w.Options.IsTest {
		if err := w.client.Init(1, remote); err != nil {
			return err
		}
	}

	w.workloadLabels = make(map[string]string)
	w.schedule = w.SetPluginInterval()
	return nil
}

func (w *Workload) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[w.Object]

	arrayID := w.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		w.SLogger.Warn("arrayID not found in ParentParams, skipping workload labels")
		return nil, nil, nil
	}

	if w.schedule >= w.PluginInvocationRate {
		w.schedule = 0
		w.refreshWorkloadLabels(arrayID)
	}
	w.schedule++

	w.applyWorkloadLabels(data)

	return nil, nil, nil
}

func (w *Workload) refreshWorkloadLabels(arrayID string) {
	w.workloadLabels = make(map[string]string)

	workloadLabels, err := w.buildWorkloadLabelMap(arrayID)
	if err != nil {
		w.SLogger.Warn("Failed to build workload label map", slogx.Err(err))
		return
	}

	w.workloadLabels = workloadLabels
	w.SLogger.Debug("Refreshed workload labels", slog.Int("count", len(w.workloadLabels)))
}

func (w *Workload) buildWorkloadLabelMap(arrayID string) (map[string]string, error) {
	workloadLabels := make(map[string]string)

	apiPath := w.client.APIPath + "/storage-systems/" + arrayID + "/workloads"
	workloads, err := w.client.Fetch(apiPath, nil)
	if err != nil {
		return workloadLabels, fmt.Errorf("failed to fetch workloads: %w", err)
	}

	for _, wl := range workloads {
		id := wl.Get("id").ClonedString()
		name := wl.Get("name").ClonedString()

		if id != "" && name != "" {
			workloadLabels[id] = name
		}
	}

	w.SLogger.Debug("Built workload label map", slog.Int("count", len(workloadLabels)))
	return workloadLabels, nil
}

func (w *Workload) applyWorkloadLabels(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		id := instance.GetLabel("id")
		if id == "" {
			continue
		}

		if name, ok := w.workloadLabels[id]; ok {
			instance.SetLabel("workload", name)
		} else {
			w.SLogger.Debug("Workload label not found in cache", slog.String("id", id))
		}
	}
}
