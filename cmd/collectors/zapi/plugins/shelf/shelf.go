/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package main

import (
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strings"
)

type Shelf struct {
	*plugin.AbstractPlugin
	data            map[string]*matrix.Matrix
	instance_keys   map[string]string
	instance_labels map[string]*dict.Dict
	connection      *zapi.Client
	query           string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Shelf{AbstractPlugin: p}
}

func (my *Shelf) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	if my.connection, err = zapi.New(my.ParentParams); err != nil {
		logger.Error(my.Prefix, "connecting: %v", err)
		return err
	}

	system, err := my.connection.GetSystem()
	if err != nil {
		return err
	}

	if system.Clustered {
		my.query = "storage-shelf-info-get-iter"
	} else {
		my.query = "storage-shelf-environment-list-info"
	}

	logger.Debug(my.Prefix, "plugin connected!")

	my.data = make(map[string]*matrix.Matrix)
	my.instance_keys = make(map[string]string)
	my.instance_labels = make(map[string]*dict.Dict)

	objects := my.Params.GetChildS("objects")
	if objects == nil {
		return errors.New(errors.MISSING_PARAM, "objects")
	}

	for _, obj := range objects.GetChildren() {

		attribute := obj.GetNameS()
		object_name := strings.ReplaceAll(attribute, "-", "_")

		if x := strings.Split(attribute, "=>"); len(x) == 2 {
			attribute = strings.TrimSpace(x[0])
			object_name = strings.TrimSpace(x[1])
		}

		my.instance_labels[attribute] = dict.New()

		my.data[attribute] = matrix.New(my.Parent+".Shelf", "shelf_"+object_name)
		my.data[attribute].SetGlobalLabel("datacenter", my.ParentParams.GetChildContentS("datacenter"))
		my.data[attribute].SetGlobalLabel("cluster", system.Name)

		export_options := node.NewS("export_options")
		instance_labels := export_options.NewChildS("instance_labels", "")
		instance_keys := export_options.NewChildS("instance_keys", "")
		instance_keys.NewChildS("", "shelf")

		for _, x := range obj.GetChildren() {

			for _, c := range x.GetAllChildContentS() {

				metric_name, display := collector.ParseMetricName(c)

				if strings.HasPrefix(c, "^") {
					if strings.HasPrefix(c, "^^") {
						my.instance_keys[attribute] = metric_name
						my.instance_labels[attribute].Set(metric_name, display)
						instance_keys.NewChildS("", display)
						logger.Debug(my.Prefix, "added instance key: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
					} else {
						my.instance_labels[attribute].Set(metric_name, display)
						instance_labels.NewChildS("", display)
						logger.Debug(my.Prefix, "added instance label: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
					}
				} else {
					metric, err := my.data[attribute].NewMetricFloat64(metric_name)
					if err != nil {
						logger.Error(my.Prefix, "add metric: %v", err)
						return err
					}
					metric.SetName(display)
					logger.Debug(my.Prefix, "added metric: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				}
			}
		}
		logger.Debug(my.Prefix, "added data for [%s] with %d metrics", attribute, len(my.data[attribute].GetMetrics()))

		my.data[attribute].SetExportOptions(export_options)
	}

	logger.Debug(my.Prefix, "initialized with data [%d] objects", len(my.data))
	return nil
}

func (my *Shelf) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	if !my.connection.IsClustered() {
		for _, instance := range data.GetInstances() {
			instance.SetLabel("shelf", instance.GetLabel("shelf_id"))
		}
	}

	err := my.connection.BuildRequestString(my.query)
	if err != nil {
		logger.Error(my.Prefix, "error: %v", err)
	}

	result, err := my.connection.Invoke()
	if err != nil {
		return nil, err
	}

	var shelves []*node.Node

	if x := result.GetChildS("attributes-list"); x != nil {
		shelves = x.GetChildren()
	} else if !my.connection.IsClustered() {
		logger.Debug(my.Prefix, "fallback to 7mode")
		shelves = result.SearchChildren([]string{"shelf-environ-channel-info", "shelf-environ-shelf-list", "shelf-environ-shelf-info"})
	}

	if len(shelves) == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no shelf instances found")
	}

	logger.Debug(my.Prefix, "fetching %d shelf counters", len(shelves))

	output := make([]*matrix.Matrix, 0)

	for _, shelf := range shelves {

		shelf_name := shelf.GetChildContentS("shelf")
		shelf_id := shelf.GetChildContentS("shelf-uid")

		if !my.connection.IsClustered() {
			uid := shelf.GetChildContentS("shelf-id")
			shelf_id = uid
			shelf_name = uid
		}

		for attribute, data := range my.data {

			data.PurgeInstances()

			if my.instance_keys[attribute] == "" {
				logger.Warn(my.Prefix, "no instance keys defined for object [%s], skipping....", attribute)
				continue
			}

			object_elem := shelf.GetChildS(attribute)
			if object_elem == nil {
				logger.Warn(my.Prefix, "no [%s] instances on this system", attribute)
				continue
			}

			logger.Debug(my.Prefix, "fetching %d [%s] instances", len(object_elem.GetChildren()), attribute)

			for _, obj := range object_elem.GetChildren() {

				if key := obj.GetChildContentS(my.instance_keys[attribute]); key != "" {

					instance, err := data.NewInstance(shelf_id + "." + key)

					if err != nil {
						logger.Debug(my.Prefix, "add (%s) instance: %v", attribute, err)
						return nil, err
					}

					logger.Debug(my.Prefix, "add (%s) instance: %s.%s", attribute, shelf_id, key)

					for label, label_display := range my.instance_labels[attribute].Map() {
						if value := obj.GetChildContentS(label); value != "" {
							instance.SetLabel(label_display, value)
						}
					}

					instance.SetLabel("shelf", shelf_name)
					instance.SetLabel("shelf_id", shelf_id)

				} else {
					logger.Debug(my.Prefix, "instance without [%s], skipping", my.instance_keys[attribute])
				}
			}

			output = append(output, data)
		}
	}

	// second loop to populate numeric data

	for _, shelf := range shelves {

		shelf_id := shelf.GetChildContentS("shelf-uid")
		if !my.connection.IsClustered() {
			shelf_id = shelf.GetChildContentS("shelf-id")
		}

		for attribute, data := range my.data {

			data.Reset()

			object_elem := shelf.GetChildS(attribute)
			if object_elem == nil {
				continue
			}

			for _, obj := range object_elem.GetChildren() {

				key := obj.GetChildContentS(my.instance_keys[attribute])

				if key == "" {
					continue
				}

				instance := data.GetInstance(shelf_id + "." + key)

				if instance == nil {
					logger.Debug(my.Prefix, "(%s) instance [%s.%s] not found in cache skipping", attribute, shelf_id, key)
					continue
				}

				for mkey, m := range data.GetMetrics() {

					if value := strings.Split(obj.GetChildContentS(mkey), " ")[0]; value != "" {
						if err := m.SetValueString(instance, value); err != nil {
							logger.Debug(my.Prefix, "(%s) failed to parse value (%s): %v", mkey, value, err)
						} else {
							logger.Debug(my.Prefix, "(%s) added value (%s)", mkey, value)
						}
					}
				}
			}
		}
	}

	return output, nil
}
