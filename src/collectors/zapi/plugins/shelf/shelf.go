package main

import (
	"goharvest2/poller/collector"
	"goharvest2/poller/collector/plugin"
	"goharvest2/share/errors"
	"goharvest2/share/logger"
	"goharvest2/share/matrix"
	"goharvest2/share/tree/node"
	"strconv"
	"strings"

	client "goharvest2/apis/zapi"
)

type Shelf struct {
	*plugin.AbstractPlugin
	data          map[string]*matrix.Matrix
	instance_keys map[string]string
	connection    *client.Client
	query         string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Shelf{AbstractPlugin: p}
}

func (p *Shelf) Init() error {

	var err error

	if err = p.InitAbc(); err != nil {
		return err
	}

	if p.connection, err = client.New(p.ParentParams); err != nil {
		logger.Error(p.Prefix, "connecting: %v", err)
		return err
	}

	system, err := p.connection.GetSystem()
	if err != nil {
		return err
	}

	if system.Clustered {
		p.query = "storage-shelf-info-get-iter"
	} else {
		p.query = "storage-shelf-environment-list-info"
	}

	logger.Debug(p.Prefix, "plugin connected!")

	p.data = make(map[string]*matrix.Matrix)
	p.instance_keys = make(map[string]string)

	objects := p.Params.GetChildS("objects")
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

		p.data[attribute] = matrix.New(p.Parent, object_name, "shelf")
		p.data[attribute].SetGlobalLabel("datacenter", p.ParentParams.GetChildContentS("datacenter"))
		p.data[attribute].SetGlobalLabel("cluster", system.Name)

		export_options := node.NewS("export_options")
		export_options.NewChildS("include_instance_names", "False") //@TODO remove!
		instance_labels := export_options.NewChildS("instance_labels", "")
		instance_keys := export_options.NewChildS("instance_keys", "")
		instance_keys.NewChildS("", "shelf")
		instance_keys.NewChildS("", "shelf_id")

		for _, x := range obj.GetChildren() {
			for _, c := range x.GetAllChildContentS() {

				metric_name, display := collector.ParseMetricName(c)

				if strings.HasPrefix(c, "^") {
					if strings.HasPrefix(c, "^^") {
						p.instance_keys[attribute] = metric_name
						p.data[attribute].AddLabel(metric_name, display)
						instance_keys.NewChildS("", display)
						logger.Debug(p.Prefix, "Adding as instance key: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
					} else {
						p.data[attribute].AddLabel(metric_name, display)
						instance_labels.NewChildS("", display)
						logger.Debug(p.Prefix, "Adding as label: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
					}
				} else {
					p.data[attribute].AddMetric(metric_name, display, true)
					logger.Debug(p.Prefix, "Adding as label: (%s) (%s) [%s]", attribute, x.GetNameS(), c)
				}
			}
		}
		logger.Debug(p.Prefix, "added data for [%s] with %d metrics and %d labels", attribute, p.data[attribute].SizeMetrics(), p.data[attribute].SizeLabels())

		p.data[attribute].SetExportOptions(export_options)
	}

	logger.Debug(p.Prefix, "initialized data with [%d] objects", len(p.data))
	return nil
}

func (p *Shelf) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	if !p.connection.IsClustered() {
		for _, instance := range data.GetInstances() {
			instance.Labels.Set("shelf", instance.Labels.Get("shelf_id"))
		}
	}

	p.connection.BuildRequestString(p.query)
	result, err := p.connection.Invoke()
	if err != nil {
		return nil, err
	}

	var shelves []*node.Node
	if x := result.GetChildS("attributes-list"); x != nil {
		shelves = x.GetChildren()
	} else if !p.connection.IsClustered() {
		logger.Debug(p.Prefix, "fallback to 7mode")
		shelves = result.SearchChildren([]string{"shelf-environ-channel-info", "shelf-environ-shelf-list", "shelf-environ-shelf-info"})
	}

	if shelves == nil || len(shelves) == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no shelf instances")
	}

	logger.Debug(p.Prefix, "fetching %d shelf counters", len(shelves))

	output := make([]*matrix.Matrix, 0)

	for _, shelf := range shelves {

		shelf_name := shelf.GetChildContentS("shelf")
		shelf_id := shelf.GetChildContentS("shelf-uid")

		if !p.connection.IsClustered() {
			uid := shelf.GetChildContentS("shelf-id")
			shelf_id = uid
			shelf_name = uid
		}

		for attribute, data := range p.data {

			data.ResetInstances()

			if p.instance_keys[attribute] == "" {
				logger.Warn(p.Prefix, "no instance keys defined for object [%s], skipping....", attribute)
				continue
			}

			object_elem := shelf.GetChildS(attribute)
			if object_elem == nil {
				logger.Warn(p.Prefix, "no [%s] instances on this system", attribute)
				continue
			}

			logger.Debug(p.Prefix, "fetching %d [%s] instances.....", len(object_elem.GetChildren()), attribute)

			for _, obj := range object_elem.GetChildren() {

				if key := obj.GetChildContentS(p.instance_keys[attribute]); key != "" {

					instance, err := data.AddInstance(shelf_id + "." + key)

					if err != nil {
						logger.Debug(p.Prefix, "add (%s) instance: %v", attribute, err)
						continue
					}

					logger.Debug(p.Prefix, "add (%s) instance: %s.%s", attribute, shelf_id, key)

					for label, label_display := range data.GetLabels() {
						if value := obj.GetChildContentS(label); value != "" {
							instance.Labels.Set(label_display, value)
						}
					}

					instance.Labels.Set("shelf", shelf_name)
					instance.Labels.Set("shelf_id", shelf_id)

				} else {
					logger.Debug(p.Prefix, "instance without [%s], skipping", p.instance_keys[attribute])
				}
			}

			output = append(output, data)
		}
	}

	// second loop to populate numeric data

	for _, shelf := range shelves {

		shelf_id := shelf.GetChildContentS("shelf-uid")
		if !p.connection.IsClustered() {
			shelf_id = shelf.GetChildContentS("shelf-id")
		}

		for attribute, data := range p.data {

			if data.InitData() != nil {
				// means no numeric metrics
				continue
			}

			object_elem := shelf.GetChildS(attribute)
			if object_elem == nil {
				continue
			}

			for _, obj := range object_elem.GetChildren() {

				key := obj.GetChildContentS(p.instance_keys[attribute])

				if key == "" {
					continue
				}

				instance := data.GetInstance(shelf_id + "." + key)

				if instance == nil {
					logger.Debug(p.Prefix, "(%s) instance [%s.%s] not found in cache skipping", attribute, shelf_id, key)
					continue
				}

				for mkey, m := range data.Metrics {

					if value := strings.Split(obj.GetChildContentS(mkey), " ")[0]; value != "" {
						if num, err := strconv.ParseFloat(value, 32); err == nil {
							data.SetValue(m, instance, float64(num))
							logger.Debug(p.Prefix, "Added numeric [%s] = [%f]", mkey, num)
						} else {
							logger.Debug(p.Prefix, "Failed to convert [%s] = [%s]", mkey, value)
						}
					}
				}
			}
		}
	}

	return output, nil
}
