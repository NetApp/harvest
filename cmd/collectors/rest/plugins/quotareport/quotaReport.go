package quotareport

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type QuotaReport struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &QuotaReport{AbstractPlugin: p}
}

func (r *QuotaReport) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	for _, instance := range data.GetInstances() {

		uType := instance.GetLabel("type")

		// set user/group
		if uType == "user" {
			uName := instance.GetLabel("user_name")
			instance.SetLabel("user", uName)
		} else if uType == "group" {
			gName := instance.GetLabel("group_name")
			instance.SetLabel("group", gName)
		}
	}
	return nil, nil
}
