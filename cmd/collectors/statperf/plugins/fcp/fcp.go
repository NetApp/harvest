package fcp

import (
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/fcp"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
)

// New This reuses the restperf FCP plugin implementation as the functionality is identical
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return fcp.New(p)
}
