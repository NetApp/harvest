/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package quota

import (
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"math"
	"strconv"
	"strings"
)

type Quota struct {
	*plugin.AbstractPlugin
	data           map[string]*matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	client         *zapi.Client
	query          string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Quota{AbstractPlugin: p}
}

func (my *Quota) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	if my.client, err = zapi.New(my.ParentParams); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.query = "quota-report-iter"
	my.Logger.Debug().Msg("plugin connected!")

	my.data = make(map[string]*matrix.Matrix)
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	objects := my.Params.GetChildS("objects")
	if objects == nil {
		return errors.New(errors.MISSING_PARAM, "objects")
	}

	for _, obj := range objects.GetAllChildContentS() {

		metricName, display := collector.ParseMetricName(obj)

		my.instanceLabels[metricName] = dict.New()

		my.data[metricName] = matrix.New(my.Parent+".Qtree", "qtree", "qtree_"+display)
		my.data[metricName].SetGlobalLabel("datacenter", my.ParentParams.GetChildContentS("datacenter"))
		my.data[metricName].SetGlobalLabel("cluster", my.client.Name())

		exportOptions := node.NewS("export_options")
		instanceLabels := exportOptions.NewChildS("instance_labels", "")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")

		// apply all instance keys, instance labels from parent (qtree.yaml) to all quota metrics
		//parent instancekeys would be added in plugin metrics
		for _, parentKeys := range my.ParentParams.GetChildS("export_options").GetChildS("instance_keys").GetAllChildContentS() {
			instanceKeys.NewChildS("", parentKeys)
		}
		// parent instacelabels would be added in plugin metrics
		for _, parentLabels := range my.ParentParams.GetChildS("export_options").GetChildS("instance_labels").GetAllChildContentS() {
			instanceKeys.NewChildS("", parentLabels) //strings.ReplaceAll(parentLabels, "_", "-"))
		}

		if strings.HasPrefix(obj, "^") {
			if strings.HasPrefix(obj, "^^") {
				my.instanceKeys[metricName] = metricName
				my.instanceLabels[metricName].Set(metricName, display)
				instanceKeys.NewChildS("", display)
				my.Logger.Debug().Msgf("added instance key: (%s) (%s) [%s]", metricName, objects.GetNameS(), display)
			} else {
				my.instanceLabels[metricName].Set(metricName, display)
				instanceLabels.NewChildS("", display)
				my.Logger.Debug().Msgf("added instance label: (%s) (%s) [%s]", metricName, objects.GetNameS(), display)
			}
		} else {
			metric, err := my.data[metricName].NewMetricFloat64(metricName)
			if err != nil {
				my.Logger.Error().Stack().Err(err).Msg("add metric")
				return err
			}
			metric.SetName(display)
			my.Logger.Debug().Msgf("added metric: (%s) (%s) [%s] %s", metricName, objects.GetNameS(), display, metric)
		}

		my.Logger.Debug().Msgf("added data for [%s] with %d metrics", metricName, len(my.data[metricName].GetMetrics()))
		my.data[metricName].SetExportOptions(exportOptions)
	}

	my.Logger.Debug().Msgf("initialized with data [%d] objects", len(my.data))
	return nil
}

func (my *Quota) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		result *node.Node
		quotas []*node.Node
		err    error
	)

	if result, err = my.client.InvokeRequestString(my.query); err != nil {
		return nil, err
	}

	if x := result.GetChildS("attributes-list"); x != nil {
		quotas = x.GetChildren()
	}

	if len(quotas) == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no quota instances found")
	}

	my.Logger.Debug().Msgf("fetching %d quota counters", len(quotas))

	var output []*matrix.Matrix

	// Purge and reset data
	for _, data1 := range my.data {
		data1.PurgeInstances()
		data1.Reset()
	}

	for key, quota := range quotas {

		tree := quota.GetChildContentS("tree")
		volume := quota.GetChildContentS("volume")
		vserver := quota.GetChildContentS("vserver")

		for attribute, data1 := range my.data {

			objectElem := quota.GetChildS(attribute)
			if objectElem == nil {
				my.Logger.Warn().Msgf("no [%s] instances on this system", attribute)
				continue
			}

			if ok := quota.GetChildContentS(attribute); ok != "" {
				instanceKey := vserver + "." + volume + "." + tree + "." + strconv.Itoa(key)
				my.Logger.Info().Msgf("instance: %s", instanceKey)
				instance, err := data1.NewInstance(instanceKey)

				if err != nil {
					my.Logger.Debug().Msgf("add (%s) instance: %v", attribute, err)
					return nil, err
				}

				my.Logger.Debug().Msgf("add (%s) instance: %s.%s.%s", attribute, vserver, volume, tree)

				qtreeInstance := data.GetInstance(tree + "." + volume + "." + vserver)
				for _, label := range my.data[attribute].GetExportOptions().GetChildS("instance_keys").GetAllChildContentS() {
					if value := qtreeInstance.GetLabel(label); value != "" {
						instance.SetLabel(label, value)
					}
				}

			} else {
				my.Logger.Debug().Msgf("instance without [%s], skipping", attribute)
			}

			output = append(output, data1)
		}
	}

	// second loop to populate numeric data
	for key, quota := range quotas {

		tree := quota.GetChildContentS("tree")
		volume := quota.GetChildContentS("volume")
		vserver := quota.GetChildContentS("vserver")

		for attribute, data1 := range my.data {

			objectElem := quota.GetChildS(attribute)
			if objectElem == nil {
				continue
			}

			instance := data1.GetInstance(vserver + "." + volume + "." + tree + "." + strconv.Itoa(key))

			if instance == nil {
				my.Logger.Debug().Msgf("(%s) instance [%s.%s] not found in cache skipping", attribute, volume, vserver)
				continue
			}

			for metricKey, m := range data1.GetMetrics() {

				if value := strings.Split(quota.GetChildContentS(metricKey), " ")[0]; value != "" {
					// Few quota metrics would have value '-' which means unlimited (ex: disk-limit)
					if value == "-" {
						value = strconv.FormatFloat(math.MaxFloat64, 'E', -1, 64)
					}
					my.Logger.Info().Msgf("vals %s (%s) ", metricKey, value)
					if err := m.SetValueString(instance, value); err != nil {
						my.Logger.Debug().Msgf("(%s) failed to parse value (%s): %v", metricKey, value, err)
					} else {
						my.Logger.Debug().Msgf("(%s) added value (%s)", metricKey, value)
					}
				}
			}
		}
	}
	return output, nil
}
