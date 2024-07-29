package quota

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
)

type Quota struct {
	*plugin.AbstractPlugin
	qtreeMetrics bool // supports quota metrics with qtree prefix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Quota{AbstractPlugin: p}
}

func (q *Quota) Init() error {
	if q.Params.HasChildS("qtreeMetrics") {
		q.qtreeMetrics = true
	}
	return nil
}

func (q *Quota) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[q.Object]

	// Purge and reset data
	instanceMap := data.GetInstances()
	metricsMap := data.GetMetrics()
	data.PurgeInstances()
	data.PurgeMetrics()

	for metricName, m := range metricsMap {
		_, err := data.NewMetricFloat64(metricName, m.GetName())
		if err != nil {
			q.Logger.Error().Stack().Err(err).Msg("add metric")
		}
	}

	if err := q.handlingQuotaMetrics(instanceMap, metricsMap, data); err != nil {
		return nil, nil, err
	}

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

func (q *Quota) handlingQuotaMetrics(instanceMap map[string]*matrix.Instance, metricMap map[string]*matrix.Metric, data *matrix.Matrix) error {
	for _, quota := range instanceMap {
		if !quota.IsExportable() {
			continue
		}
		index := quota.GetLabel("index")
		uName := quota.GetLabel("userName")
		uid := quota.GetLabel("userId")
		group := quota.GetLabel("groupName")
		quotaType := quota.GetLabel("type")

		if quotaType == "user" {
			if uName != "" {
				quota.SetLabel("user", uName)
			} else if uid != "" {
				quota.SetLabel("user", uid)
			}
		} else if quotaType == "group" {
			if group != "" {
				quota.SetLabel("group", group)
			} else if uid != "" {
				quota.SetLabel("group", uid)
			}
		}

		for metricName, m := range metricMap {
			// set -1 for unlimited
			value := -1.0
			quotaInstanceKey := index + metricName
			quotaInstance, err := data.NewInstance(quotaInstanceKey)
			if err != nil {
				q.Logger.Debug().Msgf("add (%s) instance: %v", metricName, err)
				return err
			}
			// set labels
			for k, v := range quota.GetLabels() {
				quotaInstance.SetLabel(k, v)
			}

			if v, ok := m.GetValueFloat64(quota); ok {
				// space limits are in bytes, converted to kibibytes to match ZAPI
				if metricName == "space.hard_limit" || metricName == "space.soft_limit" || metricName == "space.used.total" {
					value = v / 1024
					quotaInstance.SetLabel("unit", "kibibytes")
					if metricName == "space.soft_limit" {
						t := data.GetMetric("threshold")
						if err := t.SetValueFloat64(quotaInstance, value); err != nil {
							q.Logger.Error().Err(err).Str("metricName", metricName).Float64("value", value).Msg("Failed to parse value")
						}
					}
				} else {
					value = v
				}
			}

			// populate numeric data
			t := data.GetMetric(metricName)
			if err = t.SetValueFloat64(quotaInstance, value); err != nil {
				q.Logger.Error().Stack().Err(err).Str("metricName", metricName).Float64("value", value).Msg("Failed to parse value")
			}
		}
	}
	return nil
}
