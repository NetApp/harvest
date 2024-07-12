package quota

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"time"
)

type Quota struct {
	*plugin.AbstractPlugin
	client           *rest.Client
	query            string
	historicalLabels bool // supports labels, metrics for 22.05
	qtreeMetrics     bool // supports quota metrics with qtree prefix
}

type QtreeData struct {
	exportPolicy  string
	securityStyle string
	oplocks       string
	status        string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Quota{AbstractPlugin: p}
}

func (q *Quota) Init() error {

	var err error

	if err := q.InitAbc(); err != nil {
		return err
	}

	clientTimeout := q.ParentParams.GetChildContentS("client_timeout")
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		timeout = duration
	} else {
		q.Logger.Info().Str("timeout", timeout.String()).Msg("Using default timeout")
	}
	if q.client, err = rest.New(conf.ZapiPoller(q.ParentParams), timeout, q.Auth); err != nil {
		q.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err := q.client.Init(5); err != nil {
		return err
	}

	q.query = "api/private/cli/qtree"
	q.historicalLabels = false

	if q.Params.HasChildS("qtreeMetrics") {
		q.qtreeMetrics = true
	}

	if q.Params.HasChildS("historicalLabels") {
		if exportOption := q.ParentParams.GetChildS("export_options"); exportOption != nil {
			// qtree labels would be added in plugin metrics
			if parentKeys := exportOption.GetChildS("instance_keys"); parentKeys != nil {
				parentKeys.NewChildS("", "export_policy")
				parentKeys.NewChildS("", "oplocks")
				parentKeys.NewChildS("", "security_style")
				parentKeys.NewChildS("", "status")
			}
		}

		q.historicalLabels = true
	}
	return nil
}

func (q *Quota) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	var (
		result     []gjson.Result
		err        error
		numMetrics int
	)
	data := dataMap[q.Object]
	q.client.Metadata.Reset()

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

	quotaCount := 0
	if q.historicalLabels {
		// In 22.05, populate metrics with qtree prefix and old labels
		filter := []string{"qtree=!\"\""}

		href := rest.NewHrefBuilder().
			APIPath(q.query).
			Fields([]string{"vserver", "volume", "qtree", "oplock_mode", "status", "export_policy", "security_style"}).
			Filter(filter).
			Build()

		if result, err = collectors.InvokeRestCall(q.client, href, q.Logger); err != nil {
			return nil, nil, err
		}
		err = q.handlingHistoricalMetrics(result, instanceMap, metricsMap, data, &quotaCount, &numMetrics)
	} else {
		// Populate metrics with quota prefix and current labels
		err = q.handlingQuotaMetrics(instanceMap, metricsMap, data, &quotaCount, &numMetrics)
	}

	if err != nil {
		return nil, nil, err
	}

	q.client.Metadata.PluginInstances = uint64(quotaCount)

	q.Logger.Info().
		Int("numQuotas", quotaCount).
		Int("metrics", numMetrics).
		Msg("Collected")

	if q.qtreeMetrics || q.historicalLabels {
		// metrics with qtree prefix and quota prefix are available to support backward compatibility
		qtreePluginData := data.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
		qtreePluginData.UUID = q.Parent + ".Qtree"
		qtreePluginData.Object = "qtree"
		qtreePluginData.Identifier = "qtree"
		return []*matrix.Matrix{qtreePluginData}, q.client.Metadata, nil
	}
	return nil, q.client.Metadata, nil
}

func (q *Quota) handlingHistoricalMetrics(result []gjson.Result, instanceMap map[string]*matrix.Instance, metricMap map[string]*matrix.Metric, data *matrix.Matrix, quotaCount *int, numMetrics *int) error {
	qtreeMap := make(map[string]QtreeData)
	for _, qtree := range result {
		if !qtree.IsObject() {
			q.Logger.Error().Str("type", qtree.Type.String()).Msg("Qtree is not an object, skipping")
			return errs.New(errs.ErrNoInstance, "qtree is not an object")
		}

		svm := qtree.Get("vserver").String()
		volume := qtree.Get("volume").String()
		qtreeName := qtree.Get("qtree").String()
		oplockMode := qtree.Get("oplock_mode").String()
		status := qtree.Get("status").String()
		exportPolicy := qtree.Get("export_policy").String()
		securityStyle := qtree.Get("security_style").String()

		// Ex. InstanceKey: vserver1vol1qtree31
		qtreeInstanceKey := svm + volume + qtreeName
		qtreeMap[qtreeInstanceKey] = QtreeData{oplocks: oplockMode, status: status, exportPolicy: exportPolicy, securityStyle: securityStyle}
	}

	for _, quota := range instanceMap {
		index := quota.GetLabel("index")
		qtreeName := quota.GetLabel("qtree")
		svm := quota.GetLabel("svm")
		volume := quota.GetLabel("volume")
		qtreeInstanceKey := svm + volume + qtreeName
		qtreeInstance := qtreeMap[qtreeInstanceKey]
		*quotaCount++

		for metricName, m := range metricMap {
			// set 0 for unlimited
			value := 0.0
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

			// set qtree labels
			quotaInstance.SetLabel("oplocks", qtreeInstance.oplocks)
			quotaInstance.SetLabel("status", qtreeInstance.status)
			quotaInstance.SetLabel("export_policy", qtreeInstance.exportPolicy)
			quotaInstance.SetLabel("security_style", qtreeInstance.securityStyle)

			if v, ok := m.GetValueFloat64(quota); ok {
				// space limits are in bytes, converted to kilobytes to match ZAPI
				if metricName == "space.hard_limit" || metricName == "space.soft_limit" {
					value = v / 1024
					quotaInstance.SetLabel("unit", "Kbyte")
					if metricName == "space.soft_limit" {
						t := data.GetMetric("threshold")
						if err := t.SetValueFloat64(quotaInstance, value); err != nil {
							q.Logger.Error().Err(err).Str("metricName", metricName).Float64("value", value).Msg("Failed to parse value")
						} else {
							*numMetrics++
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
			} else {
				*numMetrics++
			}
		}
	}
	return nil
}

func (q *Quota) handlingQuotaMetrics(instanceMap map[string]*matrix.Instance, metricMap map[string]*matrix.Metric, data *matrix.Matrix, quotaCount *int, numMetrics *int) error {
	for _, quota := range instanceMap {
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
		*quotaCount++

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
				// space limits are in bytes, converted to kilobytes to match ZAPI
				if metricName == "space.hard_limit" || metricName == "space.soft_limit" || metricName == "space.used.total" {
					value = v / 1024
					quotaInstance.SetLabel("unit", "Kbyte")
					if metricName == "space.soft_limit" {
						t := data.GetMetric("threshold")
						if err := t.SetValueFloat64(quotaInstance, value); err != nil {
							q.Logger.Error().Err(err).Str("metricName", metricName).Float64("value", value).Msg("Failed to parse value")
						} else {
							*numMetrics++
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
			} else {
				*numMetrics++
			}
		}
	}
	return nil
}
