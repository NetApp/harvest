package host

import (
	"github.com/netapp/harvest/v2/cmd/collectors/eseries/hostcluster"
	"time"

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

	// Get arrayID from ParentParams
	arrayID := h.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		h.SLogger.Warn("arrayID not found in ParentParams, skipping host enrichment")
		return nil, nil, nil
	}

	// Build hosts lookup map
	hostClusterNames, err := hostcluster.BuildHostClusterLookup(h.client, arrayID, h.SLogger)
	if err != nil {
		h.SLogger.Warn("Failed to build host lookup", slogx.Err(err))
		return nil, nil, nil
	}

	// update host instances with host cluster names
	for _, instance := range data.GetInstances() {
		hostClusterID := instance.GetLabel("cluster_id")
		if hostClusterID != "" {
			if hostClusterName, ok := hostClusterNames[hostClusterID]; ok {
				instance.SetLabel("host_cluster", hostClusterName)
			}
		}
	}

	return nil, nil, nil
}
