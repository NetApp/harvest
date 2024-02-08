/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package qtree

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"time"
)

type Qtree struct {
	*plugin.AbstractPlugin
	data             *matrix.Matrix
	instanceKeys     map[string]string
	instanceLabels   map[string]map[string]string
	client           *rest.Client
	query            string
	quotaType        []string
	historicalLabels bool // supports labels, metrics for 22.05
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Qtree{AbstractPlugin: p}
}

func (q *Qtree) Init() error {

	var err error
	quotaMetric := []string{
		"space.hard_limit => disk_limit",
		"space.used.total => disk_used",
		"space.used.hard_limit_percent => disk_used_pct_disk_limit",
		"space.used.soft_limit_percent => disk_used_pct_soft_disk_limit",
		"space.soft_limit => soft_disk_limit",
		// "disk-used-pct-threshold" # deprecated and workaround to use same as disk_used_pct_soft_disk_limit
		"files.hard_limit => file_limit",
		"files.used.total => files_used",
		"files.used.hard_limit_percent => files_used_pct_file_limit",
		"files.used.soft_limit_percent => files_used_pct_soft_file_limit",
		"files.soft_limit => soft_file_limit",
		// "threshold",   # deprecated
	}

	if err = q.InitAbc(); err != nil {
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

	if err = q.client.Init(5); err != nil {
		return err
	}

	q.query = "api/storage/quota/reports"

	q.data = matrix.New(q.Parent+".Qtree", "quota", "quota")
	q.instanceKeys = make(map[string]string)
	q.instanceLabels = make(map[string]map[string]string)
	q.historicalLabels = false

	if q.Params.HasChildS("historicalLabels") {
		exportOptions := node.NewS("export_options")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")

		// apply all instance keys, instance labels from parent (qtree.yaml) to all quota metrics
		if exportOption := q.ParentParams.GetChildS("export_options"); exportOption != nil {
			// parent instancekeys would be added in plugin metrics
			if parentKeys := exportOption.GetChildS("instance_keys"); parentKeys != nil {
				for _, parentKey := range parentKeys.GetAllChildContentS() {
					instanceKeys.NewChildS("", parentKey)
				}
			}
			// parent instacelabels would be added in plugin metrics
			if parentLabels := exportOption.GetChildS("instance_labels"); parentLabels != nil {
				for _, parentLabel := range parentLabels.GetAllChildContentS() {
					instanceKeys.NewChildS("", parentLabel)
				}
			}
		}

		instanceKeys.NewChildS("", "type")
		instanceKeys.NewChildS("", "index")
		instanceKeys.NewChildS("", "unit")

		q.data.SetExportOptions(exportOptions)
		q.historicalLabels = true
	}

	quotaType := q.Params.GetChildS("quotaType")
	if quotaType != nil {
		q.quotaType = []string{}
		q.quotaType = append(q.quotaType, quotaType.GetAllChildContentS()...)

	} else {
		q.quotaType = []string{"user", "group", "tree"}
	}

	for _, obj := range quotaMetric {
		metricName, display, _, _ := util.ParseMetric(obj)

		metric, err := q.data.NewMetricFloat64(metricName, display)
		if err != nil {
			q.Logger.Error().Stack().Err(err).Msg("add metric")
			return err
		}

		q.Logger.Trace().Msgf("added metric: (%s) [%s] %v", metricName, display, metric)
	}

	q.Logger.Debug().Msgf("added data with %d metrics", len(q.data.GetMetrics()))

	return nil
}

func (q *Qtree) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		result     []gjson.Result
		err        error
		numMetrics int
	)
	data := dataMap[q.Object]
	// Purge and reset data
	q.data.PurgeInstances()
	q.data.Reset()

	// Set all global labels from Rest.go if already not exist
	q.data.SetGlobalLabels(data.GetGlobalLabels())

	filter := []string{"return_unmatched_nested_array_objects=true", "show_default_records=false", "type=" + strings.Join(q.quotaType[:], "|")}

	// In 22.05, all qtrees were exported
	if q.historicalLabels {
		for _, qtreeInstance := range data.GetInstances() {
			qtreeInstance.SetExportable(true)
		}
		// In 22.05, we would need default records as well.
		filter = []string{"return_unmatched_nested_array_objects=true", "show_default_records=true", "type=" + strings.Join(q.quotaType[:], "|")}
	}

	href := rest.NewHrefBuilder().
		APIPath(q.query).
		Fields([]string{"*"}).
		Filter(filter).
		Build()

	if result, err = collectors.InvokeRestCall(q.client, href, q.Logger); err != nil {
		return nil, err
	}

	quotaCount := 0
	cluster := data.GetGlobalLabels()["cluster"]

	if q.historicalLabels {
		// In 22.05, populate metrics with qtree prefix and old labels
		err = q.handlingHistoricalMetrics(result, data, cluster, &quotaCount, &numMetrics)
	} else {
		// Populate metrics with quota prefix and current labels
		err = q.handlingQuotaMetrics(result, cluster, &quotaCount, &numMetrics)
	}

	if err != nil {
		return nil, err
	}

	q.Logger.Info().
		Int("numQuotas", quotaCount).
		Int("metrics", numMetrics).
		Msg("Collected")

	// metrics with qtree prefix and quota prefix are available to support backward compatibility
	qtreePluginData := q.data.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	qtreePluginData.UUID = q.Parent + ".Qtree"
	qtreePluginData.Object = "qtree"
	qtreePluginData.Identifier = "qtree"
	return []*matrix.Matrix{qtreePluginData, q.data}, nil
}

func (q *Qtree) handlingHistoricalMetrics(result []gjson.Result, data *matrix.Matrix, cluster string, quotaIndex *int, numMetrics *int) error {
	for _, quota := range result {
		var tree string
		var quotaInstance *matrix.Instance
		var err error

		if !quota.IsObject() {
			q.Logger.Error().Str("type", quota.Type.String()).Msg("Quota is not an object, skipping")
			return errs.New(errs.ErrNoInstance, "quota is not an object")
		}

		if quota.Get("qtree.name").Exists() {
			tree = quota.Get("qtree.name").String()
		}
		volume := quota.Get("volume.name").String()
		vserver := quota.Get("svm.name").String()
		qIndex := quota.Get("index").String()
		quotaType := quota.Get("type").String()

		// If quota-type is not a tree, then skip
		if quotaType != "tree" {
			q.Logger.Trace().Str("quotaType", quotaType).Msg("Quota is not tree type, skipping")
			continue
		}

		// Ex. InstanceKey: vserver1vol1qtree15989279
		quotaInstanceKey := vserver + volume + tree + qIndex

		if quotaInstance = q.data.GetInstance(quotaInstanceKey); quotaInstance == nil {
			if quotaInstance, err = q.data.NewInstance(quotaInstanceKey); err != nil {
				q.Logger.Error().Stack().Err(err).Str("quotaInstanceKey", quotaInstanceKey).Msg("Failed to create quota instance")
				return err
			}
			q.Logger.Trace().
				Str("quotaInstanceKey", quotaInstanceKey).
				Str("svm", vserver).
				Str("volume", volume).
				Str("tree", tree).
				Str("qIndex", qIndex).
				Msg("add quota instance")
		}

		// qtree instancekey would be qtree, svm and volume(sorted keys)
		qtreeInstance := data.GetInstance(tree + vserver + volume)
		if qtreeInstance == nil {
			q.Logger.Warn().
				Str("tree", tree).
				Str("volume", volume).
				Str("vserver", vserver).
				Msg("No instance matching tree.vserver.volume")
			continue
		}

		if !qtreeInstance.IsExportable() {
			continue
		}

		for _, label := range q.data.GetExportOptions().GetChildS("instance_keys").GetAllChildContentS() {
			if value := qtreeInstance.GetLabel(label); value != "" {
				quotaInstance.SetLabel(label, value)
			}
		}

		// set labels
		quotaInstance.SetLabel("type", quotaType)
		quotaInstance.SetLabel("qtree", tree)
		quotaInstance.SetLabel("volume", volume)
		quotaInstance.SetLabel("svm", vserver)
		quotaInstance.SetLabel("index", cluster+"_"+strconv.Itoa(*quotaIndex))

		// If the Qtree is the volume itself, then qtree label is empty, so copy the volume name to qtree.
		if tree == "" {
			quotaInstance.SetLabel("qtree", volume)
		}

		*quotaIndex++
		for attribute, m := range q.data.GetMetrics() {
			value := 0.0

			if attrValue := quota.Get(attribute); attrValue.Exists() {
				// space limits are in bytes, converted to kilobytes
				if attribute == "space.hard_limit" || attribute == "space.soft_limit" {
					value = attrValue.Float() / 1024
					quotaInstance.SetLabel("unit", "Kbyte")
				} else {
					value = attrValue.Float()
				}
			}

			// populate numeric data
			if err = m.SetValueFloat64(quotaInstance, value); err != nil {
				q.Logger.Error().Stack().Err(err).Str("attribute", attribute).Float64("value", value).Msg("Failed to parse value")
			} else {
				*numMetrics++
				q.Logger.Trace().Str("attribute", attribute).Float64("value", value).Msg("added value")
			}
		}
	}
	return nil
}

func (q *Qtree) handlingQuotaMetrics(result []gjson.Result, cluster string, quotaCount *int, numMetrics *int) error {
	for quotaIndex, quota := range result {
		var tree string

		if !quota.IsObject() {
			q.Logger.Error().Str("type", quota.Type.String()).Msg("Quota is not an object, skipping")
			return errs.New(errs.ErrNoInstance, "quota is not an object")
		}

		if quota.Get("qtree.name").Exists() {
			tree = quota.Get("qtree.name").String()
		}
		quotaType := quota.Get("type").String()
		volume := quota.Get("volume.name").String()
		vserver := quota.Get("svm.name").String()
		uName := quota.Get("users.0.name").String()
		uid := quota.Get("users.0.id").String()
		group := quota.Get("group.name").String()
		*quotaCount++

		for attribute, m := range q.data.GetMetrics() {
			// set -1 for unlimited
			value := -1.0

			quotaInstanceKey := strconv.Itoa(quotaIndex) + "." + attribute

			quotaInstance, err := q.data.NewInstance(quotaInstanceKey)
			if err != nil {
				q.Logger.Debug().Msgf("add (%s) instance: %v", attribute, err)
				return err
			}
			// set labels
			quotaInstance.SetLabel("type", quotaType)
			quotaInstance.SetLabel("qtree", tree)
			quotaInstance.SetLabel("volume", volume)
			quotaInstance.SetLabel("svm", vserver)
			quotaInstance.SetLabel("index", cluster+"_"+strconv.Itoa(quotaIndex))

			if quotaType == "user" {
				if uName != "" {
					quotaInstance.SetLabel("user", uName)
				} else if uid != "" {
					quotaInstance.SetLabel("user", uid)
				}
			} else if quotaType == "group" {
				if group != "" {
					quotaInstance.SetLabel("group", group)
				} else if uid != "" {
					quotaInstance.SetLabel("group", uid)
				}
			}
			if attrValue := quota.Get(attribute); attrValue.Exists() {
				// space limits are in bytes, converted to kilobytes to match ZAPI
				if attribute == "space.hard_limit" || attribute == "space.soft_limit" || attribute == "space.used.total" {
					value = attrValue.Float() / 1024
					quotaInstance.SetLabel("unit", "Kbyte")
				} else {
					value = attrValue.Float()
				}
			}

			// populate numeric data
			if err = m.SetValueFloat64(quotaInstance, value); err != nil {
				q.Logger.Error().Stack().Err(err).Str("attribute", attribute).Float64("value", value).Msg("Failed to parse value")
			} else {
				*numMetrics++
				q.Logger.Trace().Str("attribute", attribute).Float64("value", value).Msg("added value")
			}
		}
	}
	return nil
}
