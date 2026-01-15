package host

import (
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/cluster"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
)

type Host struct {
	*plugin.AbstractPlugin
	client *rest.Client
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Host{AbstractPlugin: p}
}

func (h *Host) Init(remote conf.Remote) error {
	if err := h.InitAbc(); err != nil {
		return err
	}

	// Initialize REST client
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	poller, err := conf.PollerNamed(h.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, h.SLogger)
	if h.client, err = rest.New(poller, timeout, credentials, ""); err != nil {
		return err
	}

	if err := h.client.Init(1, remote); err != nil {
		return err
	}

	return nil
}

func (h *Host) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[h.Object]

	// Get systemID from ParentParams
	systemID := h.ParentParams.GetChildContentS("system_id")
	if systemID == "" {
		h.SLogger.Warn("systemID not found in ParentParams, skipping host cluster enrichment")
		return nil, nil, nil
	}

	// Build cluster lookup map
	clusterNames, err := cluster.BuildClusterLookup(h.client, systemID, h.SLogger)
	if err != nil {
		h.SLogger.Warn("Failed to build cluster lookup", slogx.Err(err))
		return nil, nil, nil
	}

	// update host instances with cluster names
	enrichedCount := 0
	for _, instance := range data.GetInstances() {
		clusterID := instance.GetLabel("cluster_id")
		if clusterID != "" {
			if clusterName, ok := clusterNames[clusterID]; ok {
				instance.SetLabel("cluster", clusterName)
				enrichedCount++
			}
		}
	}

	return nil, nil, nil
}
