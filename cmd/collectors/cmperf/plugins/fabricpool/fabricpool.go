package fabricpool

import (
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/fabricpool"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
)

// New This reuses the restperf fabricpool plugin implementation as the functionality is identical
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return fabricpool.New(p)
}
