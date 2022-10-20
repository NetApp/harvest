package fabricpool

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"strings"
)

type Fabricpool struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Fabricpool{AbstractPlugin: p}
}

func (f *Fabricpool) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	for _, instance := range data.GetInstances() {
		vol := instance.GetLabel("volume")
		for mkey, m := range data.GetMetrics() {
			// mkey: cloud_bin_operation.PUT
			if strings.HasPrefix(mkey, "cloud_bin_operation") {
				m.SetLabel("volRequest", vol+"-"+m.GetLabel("metric"))
			}
		}
	}
	return nil, nil
}
