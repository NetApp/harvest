package cifssession

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"slices"
	"strings"
)

type CIFSSession struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &CIFSSession{AbstractPlugin: p}
}

func (c *CIFSSession) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	for _, instance := range dataMap[c.Object].GetInstances() {
		if volumesData := instance.GetLabel("volumes"); volumesData != "" {
			volumes := strings.Split(volumesData, ",")
			slices.Sort(volumes)
			instance.SetLabel("volumes", strings.Join(volumes, ","))
		}
	}
	return nil, nil, nil
}
