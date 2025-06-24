package cifssession

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
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
		var volumes []string
		if volumesData := instance.GetLabel("volumes"); volumesData != "" {
			volumesJSON := gjson.Result{Type: gjson.JSON, Raw: "[" + volumesData + "]"}
			for _, volData := range volumesJSON.Array() {
				volName := volData.Get("name").String()
				volumes = append(volumes, volName)
			}
			instance.SetLabel("volumes", strings.Join(volumes, ","))
		}
	}
	return nil, nil, nil
}
