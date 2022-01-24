/*
 * Copyright NetApp Inc, 2022 All rights reserved
 */

package qtree

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/cmd/tools/rest"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strings"
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
		"space.used.hard_limit_percent => disk-used-pct-disk-limit",
		"space.used.soft_limit_percent => disk-used-pct-soft-disk-limit",
		"space.soft_limit => soft-disk-limit",
		//"disk-used-pct-threshold"
		//"files.hard_limit => file-limit",
		//"files.used.total => files-used",
		//"files.used.hard_limit_percent => files-used-pct-file-limit",
		//"files.used.soft_limit_percent => files-used-pct-soft-file-limit",
		//"files.soft_limit => soft-file-limit",
		//"threshold",
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
	my.Logger.Info().Msg("plugin connected!")

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

	//objects := my.Params.GetChildS("objects")
	//if objects == nil {
	//	return errors.New(errors.MISSING_PARAM, "objects")
	//}

	for _, obj := range quotaMetric {
		metricName, display := collector.ParseMetricName(obj)

		metric, err := my.data.NewMetricFloat64(metricName)
		if err != nil {
			my.Logger.Error().Stack().Err(err).Msg("add metric")
			return err
		}

		metric.SetName(display)
		my.Logger.Info().Msgf("added metric: (%s) [%s] %s", metricName, display, metric)
	}

	my.Logger.Info().Msgf("added data with %d metrics", len(my.data.GetMetrics()))
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

	// Set all global labels from zapi.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	//request = node.NewXmlS(my.query)

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
		return nil, fmt.Errorf("json is not valid for: %s", my.query)
	}

	results := gjson.GetManyBytes(content, "num_records", "records")
	numRecords := results[0]
	if numRecords.Int() == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no "+my.query+" instances on cluster")
	}

	//for _, quota := range results[1].Array() {
	results[1].ForEach(func(quotaKey, quota gjson.Result) bool {
		var tree string

		if !quota.IsObject() {
			my.Logger.Warn().Str("type", quota.Type.String()).Msg("Quota is not object, skipping")
			return true
		}

		if quota.Get("qtree").Exists() {
			tree = quota.Get("qtree.name").String()
		}
		volume := quota.Get("volume.name").String()
		vserver := quota.Get("svm.name").String()
		quotaIndex := quota.Get("index").String()

		// If quota-type is not a tree, then skip
		//if quota.Get("type").String() != "tree" {
		//	continue
		//}

		// Ex. InstanceKey: SVMA.vol1Abc.qtree1.5989279
		instanceKey := vserver + "." + volume + "." + tree + "." + quotaIndex

		if quotaInstance = my.data.GetInstance(instanceKey); quotaInstance == nil {
			if quotaInstance, err = my.data.NewInstance(instanceKey); err != nil {
				my.Logger.Info().Msgf("add (%s) instance: %v", instanceKey, err)
				//return nil, err
				return true
			}
			my.Logger.Info().Msgf("add (%s) instance: %s.%s.%s.%s", instanceKey, vserver, volume, tree, quotaIndex)
		}

		for attribute, m := range my.data.GetMetrics() {

			//objectElem := quota.Get(attribute)
			//if !objectElem.Exists() {
			//	my.Logger.Warn().Msgf("no [%s] instances on this %s.%s.%s", attribute, vserver, volume, tree)
			//	continue
			//}

			if attrValue := quota.Get(attribute); attrValue.Exists() {
				qtreeInstance := data.GetInstance(tree + "." + volume + "." + vserver)
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
				// Ex. InstanceKey: SVMA.vol1Abc.qtree1.5.disk-limit
				//instanceKey := vserver + "." + volume + "." + tree + "." + quotaIndex + "." + attribute
				//instance, err := my.data.NewInstance(instanceKey)

				//if err != nil {
				//	my.Logger.Debug().Msgf("add (%s) instance: %v", attribute, err)
				//	return nil, err
				//}

				//my.Logger.Debug().Msgf("add (%s) instance: %s.%s.%s", attribute, vserver, volume, tree)

				for _, label := range my.data.GetExportOptions().GetChildS("instance_keys").GetAllChildContentS() {
					if value := qtreeInstance.GetLabel(label); value != "" {
						quotaInstance.SetLabel(label, value)
					}
				}

				// If the Qtree is the volume itself, than qtree label is empty, so copy the volume name to qtree.
				if tree != "" {
					quotaInstance.SetLabel("qtree", volume)
				}

				// populate numeric data
				if value := strings.Split(attrValue.String(), " ")[0]; value != "" {
					// Few quota metrics would have value '-' which means unlimited (ex: disk-limit)
					if value == "-" {
						value = "0"
					}
					if err := m.SetValueString(quotaInstance, value); err != nil {
						my.Logger.Info().Msgf("(%s) failed to parse value (%s): %v", attribute, value, err)
					} else {
						my.Logger.Info().Msgf("(%s) added value (%s)", attribute, value)
					}
				}

			} else {
				my.Logger.Info().Msgf("instance without [%s], skipping", attribute)
			}

			output = append(output, my.data)
		}
		return true
	})

	return output, nil
}
