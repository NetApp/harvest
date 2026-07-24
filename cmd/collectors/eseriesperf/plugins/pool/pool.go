package pool

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

type Pool struct {
	*plugin.AbstractPlugin
	client     *rest.Client
	schedule   int
	poolLabels map[string]string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Pool{AbstractPlugin: p}
}

func (p *Pool) Init(remote conf.Remote) error {
	if err := p.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	poller, err := conf.PollerNamed(p.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, p.SLogger)
	if p.client, err = rest.New(poller, timeout, credentials, ""); err != nil {
		return err
	}

	if !p.Options.IsTest {
		if err := p.client.Init(1, remote); err != nil {
			return err
		}
	}

	p.poolLabels = make(map[string]string)
	p.schedule = p.SetPluginInterval()
	return nil
}

func (p *Pool) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[p.Object]

	arrayID := p.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		p.SLogger.Warn("arrayID not found in ParentParams, skipping pool labels")
		return nil, nil, nil
	}

	if p.schedule >= p.PluginInvocationRate {
		p.schedule = 0
		p.refreshPoolLabels(arrayID)
	}
	p.schedule++

	p.applyPoolLabels(data)

	return nil, nil, nil
}

func (p *Pool) refreshPoolLabels(arrayID string) {
	p.poolLabels = make(map[string]string)

	poolLabels, err := p.buildPoolLabelMap(arrayID)
	if err != nil {
		p.SLogger.Warn("Failed to build pool label map", slogx.Err(err))
		return
	}

	p.poolLabels = poolLabels
	p.SLogger.Debug("Refreshed pool labels", slog.Int("count", len(p.poolLabels)))
}

func (p *Pool) buildPoolLabelMap(arrayID string) (map[string]string, error) {
	poolLabels := make(map[string]string)

	apiPath := p.client.APIPath + "/storage-systems/" + arrayID + "/storage-pools"
	pools, err := p.client.Fetch(apiPath, nil)
	if err != nil {
		return poolLabels, fmt.Errorf("failed to fetch storage-pools: %w", err)
	}

	for _, pool := range pools {
		id := pool.Get("id").ClonedString()
		name := pool.Get("name").ClonedString()

		if id != "" && name != "" {
			poolLabels[id] = name
		}
	}

	p.SLogger.Debug("Built pool label map", slog.Int("count", len(poolLabels)))
	return poolLabels, nil
}

func (p *Pool) applyPoolLabels(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		id := instance.GetLabel("id")
		if id == "" {
			continue
		}

		if name, ok := p.poolLabels[id]; ok {
			instance.SetLabel("pool", name)
		} else {
			p.SLogger.Debug("Pool label not found in cache", slog.String("id", id))
		}
	}
}
