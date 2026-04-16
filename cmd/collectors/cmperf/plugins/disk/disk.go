package disk

import (
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/disk"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
)

// New This reuses the restperf Disk plugin implementation as the functionality is identical
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return disk.New(p)
}
