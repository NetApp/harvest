package pool

import (
	"strings"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type Pool struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Pool{AbstractPlugin: p}
}

func (p *Pool) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[p.Object]

	for _, instance := range data.GetInstances() {
		if sizes := instance.GetLabel("block_sizes_supported"); sizes != "" {
			instance.SetLabel("block_sizes_supported", stripBrackets(sizes))
		}
	}

	return nil, nil, nil
}

// stripBrackets converts a raw JSON array string like "[512,4096]" into "512,4096"
func stripBrackets(s string) string {
	return strings.Trim(s, "[]")
}
