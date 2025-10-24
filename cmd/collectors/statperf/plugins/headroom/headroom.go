package headroom

import (
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/headroom"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
)

type Headroom struct {
	*plugin.AbstractPlugin
}

// New This reuses the restperf Headroom plugin implementation as the functionality is identical
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return headroom.New(p)
}
