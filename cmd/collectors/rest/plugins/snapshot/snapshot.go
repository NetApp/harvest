package snapshot

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strings"
)

type Snapshot struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Snapshot{AbstractPlugin: p}
}

func (s *Snapshot) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[s.Object]
	for _, instance := range data.GetInstances() {
		tags := instance.GetLabel("tags")
		if strings.Contains(tags, "VOPL_owner") {
			tagArray := strings.SplitSeq(tags, ",")
			for tag := range tagArray {
				if strings.Contains(tag, "VOPL_owner") {
					instance.SetLabel("owners", strings.Split(tag, "=")[1])
				}
			}
		}
	}
	return nil, nil, nil
}
