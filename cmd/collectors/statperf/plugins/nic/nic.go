package nic

import (
	"github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/nic"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
)

// New This reuses the restperf NIC plugin implementation as the functionality is identical
func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return nic.New(p)
}
