/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package qtree

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"time"
)

type Qtree struct {
	*plugin.AbstractPlugin
	data           *matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	batchSize      string
	client         *rest.Client
	query          string
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

	timeout := rest.DefaultTimeout * time.Second
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
	instanceKeys := exportOptions.NewChildS("instance_keys", "")

	// apply all instance keys, instance labels from parent (qtree.yaml) to all quota metrics
	//parent instancekeys would be added in plugin metrics
	for _, parentKeys := range my.ParentParams.GetChildS("export_options").GetChildS("instance_keys").GetAllChildContentS() {
		instanceKeys.NewChildS("", parentKeys)
	}
	// parent instacelabels would be added in plugin metrics
	for _, parentLabels := range my.ParentParams.GetChildS("export_options").GetChildS("instance_labels").GetAllChildContentS() {
		instanceKeys.NewChildS("", parentLabels)
	}

	for _, obj := range quotaMetric {
		metricName, display, _ := util.ParseMetric(obj)

		metric, err := my.data.NewMetricFloat64(metricName)
		if err != nil {
			my.Logger.Error().Stack().Err(err).Msg("add metric")
			return err
		}

		metric.SetName(display)
		my.Logger.Debug().Msgf("added metric: (%s) [%s] %s", metricName, display, metric)
	}

	my.Logger.Debug().Msgf("added data with %d metrics", len(my.data.GetMetrics()))
	my.data.SetExportOptions(exportOptions)

	return nil
}

func (my *Qtree) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	var (
		records       []interface{}
		content       []byte
		quotaInstance *matrix.Instance
		output        []*matrix.Matrix
		err           error
	)

	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from Rest.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	href := rest.BuildHref("", "*", nil, "", "", "", "", my.query)

	err = rest.FetchData(my.client, href, &records)
	if err != nil {
		my.Logger.Error().Stack().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, err
	}

	all := rest.Pagination{
		Records:    records,
		NumRecords: len(records),
	}

	content, err = json.Marshal(all)
	if err != nil {
		my.Logger.Error().Err(err).Str("ApiPath", my.query).Msg("Unable to marshal rest pagination")
	}

	if !gjson.ValidBytes(content) {
		my.Logger.Error().Err(err).Str("Api", my.query).Msg("Invalid json")
		return nil, errors.New(errors.API_RESPONSE, "Invalid json")
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no "+my.query+" instances on cluster")
	}

	for _, quota := range results[1].Array() {
		var tree string

		if !quota.IsObject() {
			my.Logger.Error().Str("type", quota.Type.String()).Msg("Quota is not an object, skipping")
			return nil, errors.New(errors.ERR_NO_INSTANCE, "quota is not an object")
		}

		if quota.Get("qtree.name").Exists() {
			tree = quota.Get("qtree.name").String()
		}
		volume := quota.Get("volume.name").String()
		vserver := quota.Get("svm.name").String()
		quotaIndex := quota.Get("index").String()

		// If quota-type is not a tree, then skip
		if quotaType := quota.Get("type").String(); quotaType != "tree" {
			my.Logger.Trace().Str("quotaType", quotaType).Msg("Quota is not tree type, skipping")
			continue
		}

		// Ex. InstanceKey: vserver1vol1qtree15989279
		quotaInstanceKey := vserver + volume + tree + quotaIndex

		if quotaInstance = my.data.GetInstance(quotaInstanceKey); quotaInstance == nil {
			if quotaInstance, err = my.data.NewInstance(quotaInstanceKey); err != nil {
				my.Logger.Error().Stack().Err(err).Str("quotaInstanceKey", quotaInstanceKey).Msg("Failed to create quota instance")
				return nil, err
			}
			my.Logger.Debug().Msgf("add (%s) quota instance: %s.%s.%s.%s", quotaInstanceKey, vserver, volume, tree, quotaIndex)
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

		// If the Qtree is the volume itself, than qtree label is empty, so copy the volume name to qtree.
		if tree == "" {
			quotaInstance.SetLabel("qtree", volume)
		}

		for attribute, m := range my.data.GetMetrics() {
			value := 0.0

			if attrValue := quota.Get(attribute); attrValue.Exists() {
				// space limits are in bytes, converted to kilobytes
				if attribute == "space.hard_limit" || attribute == "space.soft_limit" {
					value = attrValue.Float() / 1024
				} else {
					value = attrValue.Float()
				}
			}

			// populate numeric data
			if err = m.SetValueFloat64(quotaInstance, value); err != nil {
				my.Logger.Error().Stack().Err(err).Str("attribute", attribute).Float64("value", value).Msg("Failed to parse value")
			} else {
				my.Logger.Debug().Str("attribute", attribute).Float64("value", value).Msg("added value")
			}

			output = append(output, my.data)
		}

	}

	return output, nil
}
