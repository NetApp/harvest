package main

import (
	"strings"
	"strconv"
	"goharvest2/poller/collector/plugin"
    "goharvest2/poller/struct/matrix"
	"goharvest2/poller/struct/options"
	"goharvest2/poller/struct/yaml"
	"goharvest2/poller/util/logger"
	"goharvest2/poller/errors"

    client "goharvest2/poller/api/zapi"
)

var Log *logger.Logger = logger.New(1, "")

type Shelf struct {
	*plugin.AbstractPlugin
	data map[string]*matrix.Matrix
	instance_keys map[string]string
	connection *client.Client
}

func New(parent_name string, options *options.Options, params *yaml.Node, pparams *yaml.Node) plugin.Plugin {
	p := plugin.New(parent_name, options, params, pparams)
	return &Shelf{AbstractPlugin: p}
}

func (p *Shelf) Init() error {
	
	var err error
	
	if err = p.InitAbc(); err != nil {
		return err
	}

	Log = logger.New(p.Options.LogLevel, "PLUGIN:"+p.Name)

    if p.connection, err = client.New(p.ParentParams); err != nil {
        Log.Error("connecting: %v", err)
		return err
	}

	system, err := p.connection.GetSystem()
	if err != nil {
        return err
    }

	Log.Debug("plugin connected!")

	p.data = make(map[string]*matrix.Matrix)
	p.instance_keys = make(map[string]string)

	objects := p.Params.GetChild("objects")
	if objects == nil {
		return errors.New(errors.MISSING_PARAM, "objects")
	}

	for _, obj := range objects.GetChildren() {

		attribute := obj.Name
		object_name := strings.ReplaceAll(attribute, "-", "_")

		if x := strings.Split(attribute, "=>"); len(x) == 2 {
			attribute = strings.TrimSpace(x[0])
			object_name = strings.TrimSpace(x[1])
		}

		p.data[attribute] = matrix.New(p.Parent, object_name, "shelf")
		p.data[attribute].SetGlobalLabel("datacenter", p.ParentParams.GetChildValue("datacenter"))
		p.data[attribute].SetGlobalLabel("sytsem", system.Name)

		export_options := yaml.New("export_options", "")
		instance_labels := yaml.New("instance_labels", "")
		instance_keys := yaml.New("instance_keys", "")
		instance_keys.AddValue("shelf")
		instance_keys.AddValue("shelf-id")

		for _, x := range obj.GetChildren() {
			for _, c := range x.Values {

				_, display := parse_display(c)

				if strings.HasPrefix(c, "^") {
					if strings.HasPrefix(c, "^^") {
						p.instance_keys[attribute] = c[2:]
						p.data[attribute].AddLabelKeyName(c[2:], display)
						instance_keys.AddValue(display)
						Log.Debug("Adding as instance key: (%s) (%s) [%s]", attribute, x.Name, display)
					} else {
						p.data[attribute].AddLabelKeyName(c[1:], display)
						instance_labels.AddValue(display)
						Log.Debug("Adding as label: (%s) (%s) [%s]", attribute, x.Name, display)
					}
				} else {
					p.data[attribute].AddMetric(c, display, true)
					Log.Debug("Adding as label: (%s) (%s) [%s]", attribute, x.Name, c)
				}
			}
		}
		Log.Debug("added data for [%s] with %d metrics and %d labels", attribute, len(p.data[attribute].Metrics), p.data[attribute].LabelNames.Size())
		export_options.AddChild(instance_keys)
		export_options.AddChild(instance_labels)
		export_options.CreateChild("include_instance_names", "False")
		p.data[attribute].SetExportOptions(export_options)
	}

	Log.Debug("initialized data with [%d] objects", len(p.data))
	return nil
}

func (p *Shelf) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	Log.Info("Running plugin!!!!")

	p.connection.BuildRequestString("storage-shelf-info-get-iter")
	result, err := p.connection.Invoke()
	if err != nil {
		return nil, err
	}

	shelves, has := result.GetChild("attributes-list")
	if !has {
		return nil, errors.New(errors.NO_INSTANCES, "no shelf instances")
	}

	Log.Debug("fetching %d shelf counters", len(shelves.GetChildren()))


	output := make([]*matrix.Matrix, 0)

	for _, shelf := range shelves.GetChildren() {

		shelf_name := shelf.GetChildContentS("shelf")
		shelf_id := shelf.GetChildContentS("shelf-uid")

		for attribute, data := range p.data {

			data.ResetInstances()

			if p.instance_keys[attribute] == "" {
				Log.Warn("no instances key defined for object [%s], skipping....", attribute)
				continue
			}

			object_elem, has := shelf.GetChild(attribute)
			if !has {
				Log.Warn("no [%s] instances on this system", attribute)
				continue
			}

			Log.Debug("fetching %d [%s] instances.....", len(object_elem.GetChildren(	)), attribute)

			for _, obj := range object_elem.GetChildren() {

				if key := obj.GetChildContentS(p.instance_keys[attribute]); key != "" {

					instance, err := data.AddInstance(shelf_id + "." + key)

					if err != nil {
						Log.Warn("add (%s) instance: %v", attribute, err)
						continue
					}

					Log.Debug("add (%s) instance: %s.%s", attribute, shelf_id, key)

					for label, label_display := range data.LabelNames.Iter() {
						if value := obj.GetChildContentS(label); value != "" {
							instance.Labels.Set(label_display, value)
						}
					}

					instance.Labels.Set("shelf", shelf_name)
					instance.Labels.Set("shelf_id", shelf_id)

				} else  {
					Log.Warn("instance without [%s], skipping", p.instance_keys[attribute])
				}
			}

			output = append(output, data)
		}
	}

	// second loop to populate numeric data
	
	for _, shelf := range shelves.GetChildren() {

		shelf_id := shelf.GetChildContentS("shelf-uid")

		for attribute, data := range p.data {

			if data.InitData() != nil {
				// means no numeric metrics
				continue
			}

			object_elem, has := shelf.GetChild(attribute)

			if !has {
				continue
			}

			for _, obj := range object_elem.GetChildren() {

				key := obj.GetChildContentS(p.instance_keys[attribute])
				
				if key == "" {
					continue
				}

				instance := data.GetInstance(shelf_id + "." + key)
				
				if instance == nil {
					Log.Warn("(%s) instance [%s.%s] not found in cache skipping", attribute, shelf_id, key)
					continue
				}

			
				for mkey, m := range data.Metrics {

					if value := obj.GetChildContentS(mkey); value != "" {
						if num, err := strconv.ParseFloat(value, 32); err == nil {
							data.SetValue(m, instance, num)
							Log.Debug("Added numeric [%s] = [%f]", mkey, num)
						} else {
							Log.Warn("Failed to convert [%s] = [%s]", mkey, value)
						}
					}
				}
			}
		}
	}

	return output, nil
}

func parse_display(raw string) (string, string) {

	var name, display string

	name = strings.ReplaceAll(raw, "^", "")	

	if x := strings.Split(name, "=>"); len(x) == 2 {
		name = strings.TrimSpace(x[0])
		display = strings.TrimSpace(x[1])
	} else {
		display = strings.ReplaceAll(name, "-", "_")
	}

	return name, display
}