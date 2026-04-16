package vscan

import (
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/vscan"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
)

// New This uses the zapiperf vscan plugin implementation as the functionality is identical
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return vscan.New(p)
}
