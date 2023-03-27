/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package qtree

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/dict"
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
	instanceLabels   map[string]*dict.Dict
	client           *rest.Client
	query            string
	quotaType        []string
	historicalLabels bool // supports labels, metrics for 22.05
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Qtree{AbstractPlugin: p}
}

func (my *Qtree) Init() error {

	var err error
	quotaMetric := []string{
		"space.hard_limit => disk_limit",
		"space.used.total => disk_used",
		"space.used.hard_limit_percent => disk_used_pct_disk_limit",
		"space.used.soft_limit_percent => disk_used_pct_soft_disk_limit",
		"space.soft_limit => soft_disk_limit",
		//"disk-used-pct-threshold" # deprecated and workaround to use same as disk_used_pct_soft_disk_limit
		"files.hard_limit => file_limit",
		"files.used.total => files_used",
		"files.used.hard_limit_percent => files_used_pct_file_limit",
		"files.used.soft_limit_percent => files_used_pct_soft_file_limit",
		"files.soft_limit => soft_file_limit",
		//"threshold",   # deprecated
	}

	if err = my.InitAbc(); err != nil {
		return err
	}

	clientTimeout := my.ParentParams.GetChildContentS("client_timeout")
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		timeout = duration
	} else {
		my.Logger.Info().Str("timeout", timeout.String()).Msg("Using default timeout")
	}
	if my.client, err = rest.New(conf.ZapiPoller(my.ParentParams), timeout, my.Auth); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.query = "api/storage/quota/reports"

	my.data = matrix.New(my.Parent+".Qtree", "quota", "quota")
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)
	my.historicalLabels = false

	if my.Params.HasChildS("historicalLabels") {
		exportOptions := node.NewS("export_options")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")

		// apply all instance keys, instance labels from parent (qtree.yaml) to all quota metrics
		if exportOption := my.ParentParams.GetChildS("export_options"); exportOption != nil {
			//parent instancekeys would be added in plugin metrics
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

		my.data.SetExportOptions(exportOptions)
		my.historicalLabels = true
	}

	quotaType := my.Params.GetChildS("quotaType")
	if quotaType != nil {
		my.quotaType = []string{}
		my.quotaType = append(my.quotaType, quotaType.GetAllChildContentS()...)

	} else {
		my.quotaType = []string{"user", "group", "tree"}
	}

	for _, obj := range quotaMetric {
		metricName, display, _, _ := util.ParseMetric(obj)

		metric, err := my.data.NewMetricFloat64(metricName, display)
		if err != nil {
			my.Logger.Error().Stack().Err(err).Msg("add metric")
			return err
		}

		my.Logger.Trace().Msgf("added metric: (%s) [%s] %v", metricName, display, metric)
	}

	my.Logger.Debug().Msgf("added data with %d metrics", len(my.data.GetMetrics()))

	return nil
}

func (my *Qtree) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		result     []gjson.Result
		err        error
		numMetrics int
	)
	data := dataMap[my.Object]
	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from Rest.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	filter := []string{"return_unmatched_nested_array_objects=true", "show_default_records=false", "type=" + strings.Join(my.quotaType[:], "|")}

	// In 22.05, all qtrees were exported
	if my.historicalLabels {
		for _, qtreeInstance := range data.GetInstances() {
			qtreeInstance.SetExportable(true)
		}
		// In 22.05, we would need default records as well.
		filter = []string{"return_unmatched_nested_array_objects=true", "show_default_records=true", "type=" + strings.Join(my.quotaType[:], "|")}
	}

	href := rest.BuildHref("", "*", filter, "", "", "", "", my.query)

	if result, err = collectors.InvokeRestCall(my.client, href, my.Logger); err != nil {
		return nil, err
	}

	quotaCount := 0
	cluster, _ := data.GetGlobalLabels().GetHas("cluster")

	if my.historicalLabels {
		// In 22.05, populate metrics with qtree prefix and old labels
		err = my.handlingHistoricalMetrics(result, data, cluster, &quotaCount, &numMetrics)
	} else {
		// Populate metrics with quota prefix and current labels
		err = my.handlingQuotaMetrics(result, cluster, &quotaCount, &numMetrics)
	}

	if err != nil {
		return nil, err
	}

	my.Logger.Info().
		Int("numQuotas", quotaCount).
		Int("metrics", numMetrics).
		Msg("Collected")

	// metrics with qtree prefix and quota prefix are available to support backward compatibility
	qtreePluginData := my.data.Clone(true, true, true)
	qtreePluginData.UUID = my.Parent + ".Qtree"
	qtreePluginData.Object = "qtree"
	qtreePluginData.Identifier = "qtree"
	return []*matrix.Matrix{qtreePluginData, my.data}, nil
}

func (my Qtree) handlingHistoricalMetrics(result []gjson.Result, data *matrix.Matrix, cluster string, quotaIndex *int, numMetrics *int) error {
	for _, quota := range result {
		var tree string
		var quotaInstance *matrix.Instance
		var err error

		if !quota.IsObject() {
			my.Logger.Error().Str("type", quota.Type.String()).Msg("Quota is not an object, skipping")
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
			my.Logger.Trace().Str("quotaType", quotaType).Msg("Quota is not tree type, skipping")
			continue
		}

		// Ex. InstanceKey: vserver1vol1qtree15989279
		quotaInstanceKey := vserver + volume + tree + qIndex

		if quotaInstance = my.data.GetInstance(quotaInstanceKey); quotaInstance == nil {
			if quotaInstance, err = my.data.NewInstance(quotaInstanceKey); err != nil {
				my.Logger.Error().Stack().Err(err).Str("quotaInstanceKey", quotaInstanceKey).Msg("Failed to create quota instance")
				return err
			}
			my.Logger.Debug().Msgf("add (%s) quota instance: %s.%s.%s.%s", quotaInstanceKey, vserver, volume, tree, qIndex)
		}

		qtreeInstance := data.GetInstance(vserver + volume + tree)
		if qtreeInstance == nil {
			my.Logger.Warn().
				Str("tree", tree).
				Str("volume", volume).
				Str("vserver", vserver).
				Msg("No instance matching tree.volume.vserver")
			continue
		}

		if !qtreeInstance.IsExportable() {
			continue
		}

		for _, label := range my.data.GetExportOptions().GetChildS("instance_keys").GetAllChildContentS() {
			if value := qtreeInstance.GetLabel(label); value != "" {
				quotaInstance.SetLabel(label, value)
			}
		}

		//set labels
		quotaInstance.SetLabel("type", quotaType)
		quotaInstance.SetLabel("qtree", tree)
		quotaInstance.SetLabel("volume", volume)
		quotaInstance.SetLabel("svm", vserver)
		quotaInstance.SetLabel("index", cluster+"_"+strconv.Itoa(*quotaIndex))

		// If the Qtree is the volume itself, than qtree label is empty, so copy the volume name to qtree.
		if tree == "" {
			quotaInstance.SetLabel("qtree", volume)
		}

		*quotaIndex++
		for attribute, m := range my.data.GetMetrics() {
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
				my.Logger.Error().Stack().Err(err).Str("attribute", attribute).Float64("value", value).Msg("Failed to parse value")
			} else {
				*numMetrics++
				my.Logger.Debug().Str("attribute", attribute).Float64("value", value).Msg("added value")
			}
		}
	}
	return nil
}

func (my Qtree) handlingQuotaMetrics(result []gjson.Result, cluster string, quotaCount *int, numMetrics *int) error {
	for quotaIndex, quota := range result {
		var tree string

		if !quota.IsObject() {
			my.Logger.Error().Str("type", quota.Type.String()).Msg("Quota is not an object, skipping")
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

		for attribute, m := range my.data.GetMetrics() {
			// set -1 for unlimited
			value := -1.0

			quotaInstanceKey := strconv.Itoa(quotaIndex) + "." + attribute

			quotaInstance, err := my.data.NewInstance(quotaInstanceKey)
			if err != nil {
				my.Logger.Debug().Msgf("add (%s) instance: %v", attribute, err)
				return err
			}
			//set labels
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
				my.Logger.Error().Stack().Err(err).Str("attribute", attribute).Float64("value", value).Msg("Failed to parse value")
			} else {
				*numMetrics++
				my.Logger.Trace().Str("attribute", attribute).Float64("value", value).Msg("added value")
			}
		}
	}
	return nil
}
