package quotaReport

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

		// ignore default quotas and set user/group
		if uType == "user" {
			uName := instance.GetLabel("user_name")
			uid := instance.GetLabel("user_id")
			if (uName == "*" && uid == "*") || (uName == "" && uid == "*") {
				instance.SetExportable(false)
				continue
			}
			instance.SetLabel("user", uName)
		} else if uType == "group" {
			uName := instance.GetLabel("user_name")
			if uName == "*" || uName == "" {
				instance.SetExportable(false)
				continue
			}
			instance.SetLabel("group", uName)
		} else if uType == "tree" {
			qtree := instance.GetLabel("qtree")
			if qtree == "" {
				instance.SetExportable(false)
				continue
			}
		}
	}
	return nil, nil
}
