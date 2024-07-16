// Copyright NetApp Inc, 2021 All rights reserved

package quota

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"strconv"
	"strings"
	"time"
)

const BatchSize = "500"

// Quota plugin is needed to match qtrees with quotas.
type Quota struct {
	*plugin.AbstractPlugin
	batchSize        string
	client           *zapi.Client
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

	if q.client, err = zapi.New(conf.ZapiPoller(q.ParentParams), q.Auth); err != nil {
		q.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err := q.client.Init(5); err != nil {
		return err
	}

	if q.client.IsClustered() {
		q.query = "qtree-list-iter"
	} else {
		q.query = "qtree-list"
	}
	q.Logger.Debug().Msg("plugin connected!")

	q.historicalLabels = false

	if q.Params.HasChildS("qtreeMetrics") {
		q.qtreeMetrics = true
	}

	if q.Params.HasChildS("historicalLabels") {
		// apply all instance keys, instance labels from parent (qtree.yaml) to all quota metrics
		if exportOption := q.ParentParams.GetChildS("export_options"); exportOption != nil {
			// parent instancekeys would be added in plugin metrics
			if parentKeys := exportOption.GetChildS("instance_keys"); parentKeys != nil {
				parentKeys.NewChildS("", "export_policy")
				parentKeys.NewChildS("", "oplocks")
				parentKeys.NewChildS("", "security_style")
				parentKeys.NewChildS("", "status")
				parentKeys.NewChildS("", "index")
				parentKeys.NewChildS("", "unit")
			}
		}

		q.historicalLabels = true
	}

	// setup batchSize for request
	q.batchSize = BatchSize
	if q.client.IsClustered() {
		if b := q.Params.GetChildContentS("batch_size"); b != "" {
			if _, err := strconv.Atoi(b); err == nil {
				q.batchSize = b
			}
		}
	}

	return nil
}

func (q *Quota) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	var (
		request, response *node.Node
		qtrees            []*node.Node
		ad, pd            time.Duration // Request/API time, Parse time, Fetch time
		err               error
		numMetrics        int
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

	apiT := 0 * time.Second
	parseT := 0 * time.Second

	request = node.NewXMLS(q.query)

	if q.client.IsClustered() && q.batchSize != "" {
		request.NewChildS("max-records", q.batchSize)
	}

	tag := "initial"
	qtreeCount := 0

	cluster := data.GetGlobalLabels()["cluster"]

	if q.historicalLabels {
		for {
			response, tag, ad, pd, err = q.client.InvokeBatchWithTimers(request, tag)

			if err != nil {
				return nil, nil, err
			}

			if response == nil {
				break
			}

			apiT += ad
			parseT += pd

			if q.client.IsClustered() {
				if x := response.GetChildS("attributes-list"); x != nil {
					qtrees = append(qtrees, x.GetChildren()...)
				}
			} else {
				qtrees = append(qtrees, response.SearchChildren([]string{"qtree-info"})...)
			}

			if len(qtrees) == 0 {
				q.Logger.Debug().Msg("no qtree instances found")
				return nil, q.client.Metadata, nil
			}

			q.Logger.Debug().Int("qtrees", len(qtrees)).Msg("fetching qtrees")
		}

		// In 22.05, populate metrics with qtree prefix and old labels
		if err := q.handlingHistoricalMetrics(qtrees, instanceMap, metricsMap, data, cluster, &qtreeCount, &numMetrics); err != nil {
			return nil, nil, err
		}
	} else {
		// Populate metrics with quota prefix and current labels
		if err := q.handlingQuotaMetrics(instanceMap, metricsMap, data, &numMetrics); err != nil {
			return nil, nil, err
		}
	}

	q.client.Metadata.PluginInstances = uint64(qtreeCount)

	q.Logger.Info().
		Int("numQtrees", qtreeCount).
		Int("metrics", numMetrics).
		Str("apiD", apiT.Round(time.Millisecond).String()).
		Str("parseD", parseT.Round(time.Millisecond).String()).
		Str("batchSize", q.batchSize).
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

func (q *Quota) handlingHistoricalMetrics(qtrees []*node.Node, instanceMap map[string]*matrix.Instance, metricMap map[string]*matrix.Metric, data *matrix.Matrix, cluster string, qtreeCount *int, numMetrics *int) error {
	qtreeMap := make(map[string]QtreeData)
	for _, qtree := range qtrees {
		var svm, qtreeInstanceKey string
		qtreeName := qtree.GetChildContentS("tree")
		volume := qtree.GetChildContentS("volume")
		if q.client.IsClustered() {
			svm = qtree.GetChildContentS("vserver")
			// Ex. InstanceKey: vserver1vol1qtree31
			qtreeInstanceKey = svm + volume + qtreeName
		} else {
			// Ex. InstanceKey: vol1qtree31
			qtreeInstanceKey = volume + qtreeName
		}

		oplockMode := qtree.GetChildContentS("oplocks")
		status := qtree.GetChildContentS("status")
		exportPolicy := qtree.GetChildContentS("export-policy")
		securityStyle := qtree.GetChildContentS("security-style")
		qtreeMap[qtreeInstanceKey] = QtreeData{oplocks: oplockMode, status: status, exportPolicy: exportPolicy, securityStyle: securityStyle}
		*qtreeCount++
	}

	quotaIndex := 0
	for _, quota := range instanceMap {
		if !quota.IsExportable() {
			continue
		}
		var svm, quotaInstanceKey string
		var qtreeInstance QtreeData
		qtreeName := quota.GetLabel("qtree")
		volume := quota.GetLabel("volume")
		if q.client.IsClustered() {
			svm = quota.GetLabel("svm")
			qtreeInstance = qtreeMap[svm+volume+qtreeName]
		} else {
			qtreeInstance = qtreeMap[volume+qtreeName]
		}

		for metricName, m := range metricMap {
			v, ok := m.GetValueString(quota)
			if !ok {
				continue
			}

			// Ex. InstanceKey: SVMA.vol1Abc.qtree1.5.disk-limit
			if q.client.IsClustered() {
				quotaInstanceKey = svm + "." + volume + "." + qtreeName + "." + strconv.Itoa(quotaIndex) + "." + metricName
			} else {
				quotaInstanceKey = volume + "." + qtreeName + "." + strconv.Itoa(quotaIndex) + "." + metricName
			}
			quotaInstance, err := data.NewInstance(quotaInstanceKey)

			if err != nil {
				q.Logger.Debug().Msgf("add (%s) instance: %v", metricName, err)
				return err
			}

			// set labels
			for k, val := range quota.GetLabels() {
				quotaInstance.SetLabel(k, val)
			}

			// set labels
			quotaInstance.SetLabel("qtree", qtreeName)
			quotaInstance.SetLabel("volume", volume)
			quotaInstance.SetLabel("svm", svm)
			quotaInstance.SetLabel("index", cluster+"_"+strconv.Itoa(quotaIndex))
			quotaIndex++

			// set qtree labels
			quotaInstance.SetLabel("oplocks", qtreeInstance.oplocks)
			quotaInstance.SetLabel("status", qtreeInstance.status)
			quotaInstance.SetLabel("export_policy", qtreeInstance.exportPolicy)
			quotaInstance.SetLabel("security_style", qtreeInstance.securityStyle)

			// If the Qtree is the volume itself, then qtree label is empty, so copy the volume name to qtree.
			if qtreeName == "" {
				quotaInstance.SetLabel("qtree", volume)
			}

			// populate numeric data
			if value := strings.Split(v, " ")[0]; value != "" {
				if metricName == "quota.soft-disk-limit" || metricName == "quota.disk-limit" || metricName == "quota.disk-used" {
					quotaInstance.SetLabel("unit", "Kbyte")
				}
				// populate numeric data
				t := data.GetMetric(metricName)
				if err = t.SetValueString(quotaInstance, value); err != nil {
					q.Logger.Error().Stack().Err(err).Str("metricName", metricName).Str("value", value).Msg("Failed to parse value")
				} else {
					*numMetrics++
				}
			}
		}
	}
	return nil
}

func (q *Quota) handlingQuotaMetrics(instanceMap map[string]*matrix.Instance, metricMap map[string]*matrix.Metric, data *matrix.Matrix, numMetrics *int) error {
	for _, quota := range instanceMap {
		if !quota.IsExportable() {
			continue
		}
		var svm, quotaInstanceKey string
		quotaType := quota.GetLabel("type")
		tree := quota.GetLabel("qtree")
		volume := quota.GetLabel("volume")
		if q.client.IsClustered() {
			svm = quota.GetLabel("svm")
		}
		uName := quota.GetLabel("quota-user-name")
		uid := quota.GetLabel("quota-user-id")

		// ignore default quotas and set user/group
		// Rest uses service side filtering to remove default records
		switch {
		case quotaType == "user":
			if (uName == "*" && uid == "*") || (uName == "" && uid == "") {
				continue
			}
		case quotaType == "group":
			if uName == "*" || uName == "" {
				continue
			}
		case quotaType == "tree":
			if tree == "" {
				continue
			}
		}

		for metricName, m := range metricMap {
			// Ex. InstanceKey: SVMA.vol1Abc.qtree1.5.disk-limit
			if q.client.IsClustered() {
				quotaInstanceKey = svm + "." + volume + "." + tree + "." + uName + "." + metricName + "." + quotaType
			} else {
				quotaInstanceKey = volume + "." + tree + "." + uName + "." + metricName
			}
			quotaInstance, err := data.NewInstance(quotaInstanceKey)
			if err != nil {
				q.Logger.Debug().Msgf("add (%s) instance: %v", metricName, err)
				return err
			}
			// set labels
			quotaInstance.SetLabel("qtree", tree)
			quotaInstance.SetLabel("volume", volume)
			quotaInstance.SetLabel("svm", svm)

			if quotaType == "user" {
				if uName != "" {
					quotaInstance.SetLabel("user", uName)
				} else if uid != "" {
					quotaInstance.SetLabel("user", uid)
				}
			} else if quotaType == "group" {
				if uName != "" {
					quotaInstance.SetLabel("group", uName)
				} else if uid != "" {
					quotaInstance.SetLabel("group", uid)
				}
			}

			if v, ok := m.GetValueString(quota); ok {
				// populate numeric data
				if value := strings.Split(v, " ")[0]; value != "" {
					if metricName == "quota.soft-disk-limit" || metricName == "quota.disk-limit" || metricName == "quota.disk-used" {
						quotaInstance.SetLabel("unit", "Kbyte")
					}
					// populate numeric data
					t := data.GetMetric(metricName)
					if err = t.SetValueString(quotaInstance, value); err != nil {
						q.Logger.Error().Stack().Err(err).Str("metricName", metricName).Str("value", value).Msg("Failed to parse value")
					} else {
						*numMetrics++
					}
				}
			}
		}
	}
	return nil
}
