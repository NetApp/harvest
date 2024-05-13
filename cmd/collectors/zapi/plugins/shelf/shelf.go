// Package shelf Copyright NetApp Inc, 2021 All rights reserved
package shelf

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"strings"
)

const BatchSize = "500"

type Shelf struct {
	*plugin.AbstractPlugin
	data                map[string]*matrix.Matrix
	shelfMetrics        *matrix.Matrix
	instanceKeys        map[string]string
	instanceLabels      map[string]map[string]string
	shelfInstanceKeys   []string
	shelfInstanceLabels []shelfInstanceLabel
	batchSize           string
	client              *zapi.Client
	query               string
}

type shelfInstanceLabel struct {
	label        string
	labelDisplay string
	parent       string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Shelf{AbstractPlugin: p}
}

func (my *Shelf) Init() error {

	var err error

	if err := my.InitAbc(); err != nil {
		return err
	}

	if my.client, err = zapi.New(conf.ZapiPoller(my.ParentParams), my.Auth); err != nil {
		my.Logger.Error().Err(err).Msg("connecting")
		return err
	}

	if err := my.client.Init(5); err != nil {
		return err
	}

	if my.client.IsClustered() {
		return nil
	}

	my.query = "storage-shelf-environment-list-info"

	my.Logger.Debug().Msg("plugin connected!")

	// populating shelfMetrics metric shape from template parsing
	my.create7ModeShelfMetrics()

	my.data = make(map[string]*matrix.Matrix)
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]map[string]string)

	objects := my.Params.GetChildS("objects")
	if objects == nil {
		return errs.New(errs.ErrMissingParam, "objects")
	}

	for _, obj := range objects.GetChildren() {

		attribute := obj.GetNameS()
		objectName := strings.ReplaceAll(attribute, "-", "_")

		if x := strings.Split(attribute, "=>"); len(x) == 2 {
			attribute = strings.TrimSpace(x[0])
			objectName = strings.TrimSpace(x[1])
		}

		my.instanceLabels[attribute] = make(map[string]string)

		my.data[attribute] = matrix.New(my.Parent+".Shelf", "shelf_"+objectName, "shelf_"+objectName)
		my.data[attribute].SetGlobalLabel("datacenter", my.ParentParams.GetChildContentS("datacenter"))

		exportOptions := node.NewS("export_options")
		instanceLabels := exportOptions.NewChildS("instance_labels", "")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")
		instanceKeys.NewChildS("", "shelf")
		instanceKeys.NewChildS("", "channel")

		// artificial metric for status of child object of shelf
		_, _ = my.data[attribute].NewMetricUint8("status")

		for _, x := range obj.GetChildren() {

			for _, c := range x.GetAllChildContentS() {

				metricName, display, kind, _ := util.ParseMetric(c)

				switch kind {
				case "key":
					my.instanceKeys[attribute] = metricName
					my.instanceLabels[attribute][metricName] = display
					instanceKeys.NewChildS("", display)
					my.Logger.Debug().Msgf("added instance key: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				case "label":
					my.instanceLabels[attribute][metricName] = display
					instanceLabels.NewChildS("", display)
					my.Logger.Debug().Msgf("added instance label: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				case "float":
					_, err := my.data[attribute].NewMetricFloat64(metricName, display)
					if err != nil {
						my.Logger.Error().Err(err).Msg("add metric")
						return err
					}
					my.Logger.Debug().Msgf("added metric: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				}
			}
		}

		my.Logger.Debug().Str("attribute", attribute).Int("metrics count", len(my.data[attribute].GetMetrics())).Msg("added")

		my.data[attribute].SetExportOptions(exportOptions)
	}

	my.Logger.Debug().Int("objects count", len(my.data)).Msg("initialized")

	// setup batchSize for request
	my.batchSize = BatchSize

	return nil
}

func (my *Shelf) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {

	var (
		err    error
		output []*matrix.Matrix
	)

	data := dataMap[my.Object]
	my.client.Metadata.Reset()

	if my.client.IsClustered() {
		for _, instance := range data.GetInstances() {
			if !instance.IsExportable() {
				continue
			}

			model := instance.GetLabel("model")
			moduleType := instance.GetLabel("module_type")

			isEmbed := collectors.IsEmbedShelf(model, moduleType)
			if isEmbed {
				instance.SetLabel("isEmbedded", "Yes")
			} else {
				instance.SetLabel("isEmbedded", "No")
			}
		}
		return nil, nil, nil
	}

	// 7 mode handling
	for _, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		instance.SetLabel("shelf", instance.GetLabel("shelf_id"))
	}

	// Set all global labels from zapi.go if already not exist
	for a := range my.instanceLabels {
		my.data[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	request := node.NewXMLS(my.query)

	result, err := my.client.InvokeZapiCall(request)
	if err != nil {
		return nil, nil, err
	}

	output, err = my.handle7Mode(data, result)

	if err != nil {
		return output, nil, err
	}

	my.Logger.Debug().Int("Shelves instance count", len(data.GetInstances())).Send()
	output = append(output, data)
	return output, my.client.Metadata, nil
}

func (my *Shelf) handle7Mode(data *matrix.Matrix, result []*node.Node) ([]*matrix.Matrix, error) {
	var (
		shelves  []*node.Node
		channels []*node.Node
		output   []*matrix.Matrix
	)

	// Result would be the zapi response itself with only one record.
	if len(result) != 1 {
		my.Logger.Debug().Msg("no shelves found")
		return output, nil
	}
	// fallback to 7mode
	channels = result[0].SearchChildren([]string{"shelf-environ-channel-info"})

	if len(channels) == 0 {
		my.Logger.Debug().Msg("no channels found")
		return output, nil
	}

	// Purge and reset data
	for _, data1 := range my.data {
		data1.PurgeInstances()
		data1.Reset()
	}

	// reset instance and matrix of shelfMetrics
	my.shelfMetrics.PurgeInstances()
	my.shelfMetrics.Reset()

	// Purge instances and metrics generated from template and updated data metrics and instances from shelfMetrics
	data.PurgeInstances()
	data.PurgeMetrics()
	for metricName, m := range my.shelfMetrics.GetMetrics() {
		_, err := data.NewMetricFloat64(metricName, m.GetName())
		if err != nil {
			my.Logger.Error().Err(err).Msg("add metric")
		}
		my.Logger.Debug().Str("metric", m.GetName()).Msg("added")
	}

	for _, channel := range channels {
		channelName := channel.GetChildContentS("channel-name")
		shelves = channel.SearchChildren([]string{"shelf-environ-shelf-list", "shelf-environ-shelf-info"})

		if len(shelves) == 0 {
			my.Logger.Debug().Str("channel", channelName).Msg("no shelves found")
			continue
		}

		for _, shelf := range shelves {
			uid := shelf.GetChildContentS("shelf-id")
			shelfName := uid // no shelf name in 7mode
			shelfID := uid

			shelfInstanceKey := shelfID + "." + channelName
			// generating new instances from plugin and adding into data
			newShelfInstance, err := data.NewInstance(shelfInstanceKey)
			if err != nil {
				my.Logger.Error().Err(err).Msg("Error while creating shelf instance")
				return nil, err
			}

			for _, key := range my.shelfInstanceKeys {
				newShelfInstance.SetLabel(key, shelf.GetChildContentS(key))
			}
			for _, shelfLabelData := range my.shelfInstanceLabels {
				if shelfLabelData.parent == "" {
					newShelfInstance.SetLabel(shelfLabelData.labelDisplay, shelf.GetChildContentS(shelfLabelData.label))
				} else {
					child := shelf.GetChildS(shelfLabelData.parent)
					newShelfInstance.SetLabel(shelfLabelData.labelDisplay, child.GetChildContentS(shelfLabelData.label))
				}
			}

			newShelfInstance.SetLabel("channel", channelName)
			newShelfInstance.SetLabel("shelf", newShelfInstance.GetLabel("shelf_id"))

			// populate numeric data
			for metricKey, m := range data.GetMetrics() {
				if value := strings.Split(shelf.GetChildContentS(metricKey), " ")[0]; value != "" {
					if err := m.SetValueString(newShelfInstance, value); err != nil {
						my.Logger.Debug().Str("metricKey", metricKey).Str("value", value).Err(err).Msg("failed to parse")
					} else {
						my.Logger.Debug().Str("metricKey", metricKey).Str("value", value).Msg("added")
					}
				}
			}

			for attribute, data1 := range my.data {
				statusMetric := data1.GetMetric("status")
				if statusMetric == nil {
					continue
				}

				if my.instanceKeys[attribute] == "" {
					my.Logger.Warn().Str("attribute", attribute).Msg("no instance keys defined")
					continue
				}

				objectElem := shelf.GetChildS(attribute)
				if objectElem == nil {
					my.Logger.Warn().Str("attribute", attribute).Msg("no instances on this system")
					continue
				}

				my.Logger.Debug().Msgf("fetching %d [%s] instances", len(objectElem.GetChildren()), attribute)

				for _, obj := range objectElem.GetChildren() {

					if key := obj.GetChildContentS(my.instanceKeys[attribute]); key != "" {
						instanceKey := shelfID + "." + key + "." + channelName
						instance, err := data1.NewInstance(instanceKey)

						if err != nil {
							my.Logger.Error().Msgf("add (%s) instance: %v", attribute, err)
							return nil, err
						}
						my.Logger.Debug().Msgf("add (%s) instance: %s.%s", attribute, shelfID, key)

						for label, labelDisplay := range my.instanceLabels[attribute] {
							if value := obj.GetChildContentS(label); value != "" {
								instance.SetLabel(labelDisplay, value)
							}
						}

						instance.SetLabel("shelf", shelfName)
						instance.SetLabel("shelf_id", shelfID)
						instance.SetLabel("channel", channelName)

						// Each child would have different possible values which is an ugly way to write all of them,
						// so normal value would be mapped to 1 and rest all are mapped to 0.
						if instance.GetLabel("status") == "normal" {
							_ = statusMetric.SetValueInt64(instance, 1)
						} else {
							_ = statusMetric.SetValueInt64(instance, 0)
						}

						// populate numeric data
						for metricKey, m := range data1.GetMetrics() {
							if value := strings.Split(obj.GetChildContentS(metricKey), " ")[0]; value != "" {
								if err := m.SetValueString(instance, value); err != nil {
									my.Logger.Debug().Msgf("(%s) failed to parse value (%s): %v", metricKey, value, err)
								} else {
									my.Logger.Debug().Msgf("(%s) added value (%s)", metricKey, value)
								}
							}
						}
					} else {
						my.Logger.Debug().Msgf("instance without [%s], skipping", my.instanceKeys[attribute])
					}
				}

				output = append(output, data1)
			}
		}
	}
	return output, nil
}

func (my *Shelf) create7ModeShelfMetrics() {
	my.shelfMetrics = matrix.New(my.Parent+".Shelf", "shelf", "shelf")
	my.shelfInstanceKeys = make([]string, 0)
	my.shelfInstanceLabels = []shelfInstanceLabel{}
	shelfExportOptions := node.NewS("export_options")
	shelfInstanceKeys := shelfExportOptions.NewChildS("instance_keys", "")
	shelfInstanceLabels := shelfExportOptions.NewChildS("instance_labels", "")

	if counters := my.ParentParams.GetChildS("counters"); counters != nil {
		if channelInfo := counters.GetChildS("shelf-environ-channel-info"); channelInfo != nil {
			if shelfList := channelInfo.GetChildS("shelf-environ-shelf-list"); shelfList != nil {
				if shelfInfo := shelfList.GetChildS("shelf-environ-shelf-info"); shelfInfo != nil {
					my.parse7ModeTemplate(shelfInfo, shelfInstanceKeys, shelfInstanceLabels, "")
				}
			}
		}
	}

	shelfInstanceKeys.NewChildS("", "channel")
	shelfInstanceKeys.NewChildS("", "shelf")
}

func (my *Shelf) parse7ModeTemplate(shelfInfo *node.Node, shelfInstanceKeys, shelfInstanceLabels *node.Node, parent string) {
	for _, shelfProp := range shelfInfo.GetChildren() {
		if len(shelfProp.GetChildren()) > 0 {
			my.parse7ModeTemplate(shelfInfo.GetChildS(shelfProp.GetNameS()), shelfInstanceKeys, shelfInstanceLabels, shelfProp.GetNameS())
		} else {
			metricName, display, kind, _ := util.ParseMetric(shelfProp.GetContentS())
			switch kind {
			case "key":
				my.shelfInstanceKeys = append(my.shelfInstanceKeys, metricName)
				my.shelfInstanceLabels = append(my.shelfInstanceLabels, shelfInstanceLabel{label: metricName, labelDisplay: display, parent: parent})
				shelfInstanceKeys.NewChildS("", display)
				my.Logger.Debug().Str("instance key", display).Msg("added")
			case "label":
				my.shelfInstanceLabels = append(my.shelfInstanceLabels, shelfInstanceLabel{label: metricName, labelDisplay: display, parent: parent})
				shelfInstanceLabels.NewChildS("", display)
				my.Logger.Debug().Str("instance label", display).Msg("added")
			case "float":
				_, err := my.shelfMetrics.NewMetricFloat64(metricName, display)
				if err != nil {
					my.Logger.Error().Err(err).Msg("add metric")
				}
				my.Logger.Debug().Str("metric", display).Msg("added")
			}
		}
	}
}
