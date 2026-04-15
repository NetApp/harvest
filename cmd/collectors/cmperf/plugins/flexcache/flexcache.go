package flexcache

import (
	"github.com/netapp/harvest/v2/cmd/collectors/statperf/plugins/flexcache"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
)

// New This uses the statperf flexcahe plugin implementation as the functionality is identical
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return flexcache.New(p)
}
