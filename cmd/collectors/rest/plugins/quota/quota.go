package quota

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
)

type Quota struct {
	*plugin.AbstractPlugin
	qtreeMetrics bool // supports quota metrics with qtree prefix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Quota{AbstractPlugin: p}
}

func (q *Quota) Init(conf.Remote) error {
	if err := q.InitAbc(); err != nil {
		return err
	}

	if q.Params.HasChildS("qtreeMetrics") {
		q.qtreeMetrics = true
	}
	return nil
}

func (q *Quota) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[q.Object]

	// The threshold metric does not exist in REST quota template, we are adding it to maintain parity with exported ZAPI metrics
	if data.GetMetric("threshold") == nil {
		_, err := data.NewMetricFloat64("threshold", "threshold")
		if err != nil {
			q.SLogger.Error("add metric", slogx.Err(err))
		}
	}

	q.handlingQuotaMetrics(data)

	if q.qtreeMetrics {
		// metrics with qtree prefix and quota prefix are available to support backward compatibility
		qtreePluginData := data.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
		qtreePluginData.UUID = q.Parent + ".Qtree"
		qtreePluginData.Object = "qtree"
		qtreePluginData.Identifier = "qtree"
		return []*matrix.Matrix{qtreePluginData}, nil, nil
	}
	return nil, nil, nil
}

func (q *Quota) handlingQuotaMetrics(data *matrix.Matrix) {
	for _, quota := range data.GetInstances() {
		if !quota.IsExportable() {
			continue
		}
		uName := quota.GetLabel("userName")
		uid := quota.GetLabel("userId")
		group := quota.GetLabel("groupName")
		quotaType := quota.GetLabel("type")

		switch quotaType {
		case "user":
			quota.SetLabel("user", uName)
			quota.SetLabel("userId", uid)
		case "group":
			quota.SetLabel("group", group)
			quota.SetLabel("groupId", uid)
		}

		for metricName, m := range data.GetMetrics() {
			// set -1 for unlimited
			value := -1.0

			if v, ok := m.GetValueFloat64(quota); ok {
				// space limits are in bytes, converted to kibibytes to match ZAPI
				if metricName == "space.hard_limit" || metricName == "space.soft_limit" || metricName == "space.used.total" {
					value = v / 1024
					m.SetLabel("unit", "kibibytes")
					if metricName == "space.soft_limit" {
						t := data.GetMetric("threshold")
						t.SetValueFloat64(quota, value)
						t.SetLabel("unit", "kibibytes")
					}
				} else {
					value = v
				}
			}

			// populate numeric data
			t := data.GetMetric(metricName)
			t.SetValueFloat64(quota, value)
		}
	}
}
