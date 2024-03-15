package systemnode

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
)

type SystemNode struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SystemNode{AbstractPlugin: p}
}

func (s *SystemNode) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[s.Object]
	nodeStateMap := make(map[string]string)

	for _, node := range data.GetInstanceKeys() {
		nodeStateMap[node] = data.GetInstance(node).GetLabel("healthy")
	}

	// update node instance with partner_healthy info
	for _, node := range data.GetInstances() {
		node.SetLabel("partner_healthy", nodeStateMap[node.GetLabel("ha_partner")])
	}

	return nil, nil, nil
}
