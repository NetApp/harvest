package fcvi

import (
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/fcvi"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
)

// New This reuses the restperf FCVI plugin implementation as the functionality is identical
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return fcvi.New(p)
}
