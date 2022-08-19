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
	data           *matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	client         *rest.Client
	query          string
	quotaType      []string
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
	timeout := rest.DefaultTimeout * time.Second
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil {
		timeout = duration
	} else {
		my.Logger.Info().Str("timeout", timeout.String()).Msg("Using default timeout")
	}
	if my.client, err = rest.New(conf.ZapiPoller(my.ParentParams), timeout); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.query = "api/storage/quota/reports"

	my.data = matrix.New(my.Parent+".Qtree", "qtree", "qtree")
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	exportOptions := node.NewS("export_options")
	exportOptions.NewChildS("include_all_labels", "true")

	quotaType := my.Params.GetChildS("quotaType")
	if quotaType != nil {
		my.quotaType = []string{}
		my.quotaType = append(my.quotaType, quotaType.GetAllChildContentS()...)

	} else {
		my.quotaType = []string{"user", "group", "tree"}
	}

	for _, obj := range quotaMetric {
		metricName, display, _, _ := util.ParseMetric(obj)

		metric, err := my.data.NewMetricFloat64(metricName)
		if err != nil {
			my.Logger.Error().Stack().Err(err).Msg("add metric")
			return err
		}

		metric.SetName(display)
		my.Logger.Trace().Msgf("added metric: (%s) [%s] %s", metricName, display, metric)
	}

	my.Logger.Debug().Msgf("added data with %d metrics", len(my.data.GetMetrics()))
	my.data.SetExportOptions(exportOptions)

	return nil
}

func (my *Qtree) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		result     []gjson.Result
		err        error
		numMetrics int
	)
	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from Rest.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	filter := []string{"return_unmatched_nested_array_objects=true", "show_default_records=false", "type=" + strings.Join(my.quotaType[:], "|")}

	href := rest.BuildHref("", "*", filter, "", "", "", "", my.query)

	if result, err = collectors.InvokeRestCall(my.client, my.query, href, my.Logger); err != nil {
		return nil, err
	}

	quotaCount := 0
	for quotaIndex, quota := range result {
		var tree string

		if !quota.IsObject() {
			my.Logger.Error().Str("type", quota.Type.String()).Msg("Quota is not an object, skipping")
			return nil, errs.New(errs.ErrNoInstance, "quota is not an object")
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
		quotaCount++

		for attribute, m := range my.data.GetMetrics() {
			// set -1 for unlimited
			value := -1.0

			quotaInstanceKey := strconv.Itoa(quotaIndex) + "." + attribute

			quotaInstance, err := my.data.NewInstance(quotaInstanceKey)
			if err != nil {
				my.Logger.Debug().Msgf("add (%s) instance: %v", attribute, err)
				return nil, err
			}
			//set labels
			quotaInstance.SetLabel("type", quotaType)
			quotaInstance.SetLabel("qtree", tree)
			quotaInstance.SetLabel("volume", volume)
			quotaInstance.SetLabel("svm", vserver)
			quotaInstance.SetLabel("index", strconv.Itoa(quotaIndex))

			if quotaType == "user" {
				quotaInstance.SetLabel("user", uName)
				quotaInstance.SetLabel("user_id", uid)
			} else if quotaType == "group" {
				quotaInstance.SetLabel("group", group)
				quotaInstance.SetLabel("group_id", uid)
			}

			if attrValue := quota.Get(attribute); attrValue.Exists() {
				// space limits are in bytes, converted to kilobytes
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
				numMetrics++
				my.Logger.Trace().Str("attribute", attribute).Float64("value", value).Msg("added value")
			}
		}

	}

	my.Logger.Info().
		Int("numQuotas", quotaCount).
		Int("metrics", numMetrics).
		Msg("Collected")

	return []*matrix.Matrix{my.data}, nil
}
