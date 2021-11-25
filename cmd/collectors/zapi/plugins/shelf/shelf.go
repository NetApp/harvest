/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package shelf

import (
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"strings"
)

type Shelf struct {
	*plugin.AbstractPlugin
	data           map[string]*matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	client         *zapi.Client
	query          string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Shelf{AbstractPlugin: p}
}

func (my *Shelf) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	if my.client, err = zapi.New(conf.ZapiPoller(my.ParentParams)); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	if my.client.IsClustered() {
		my.query = "storage-shelf-info-get-iter"
	} else {
		my.query = "storage-shelf-environment-list-info"
	}

	my.Logger.Debug().Msg("plugin connected!")

	my.data = make(map[string]*matrix.Matrix)
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	objects := my.Params.GetChildS("objects")
	if objects == nil {
		return errors.New(errors.MISSING_PARAM, "objects")
	}

	for _, obj := range objects.GetChildren() {

		attribute := obj.GetNameS()
		objectName := strings.ReplaceAll(attribute, "-", "_")

		if x := strings.Split(attribute, "=>"); len(x) == 2 {
			attribute = strings.TrimSpace(x[0])
			objectName = strings.TrimSpace(x[1])
		}

		my.instanceLabels[attribute] = dict.New()

		my.data[attribute] = matrix.New(my.Parent+".Shelf", "shelf_"+objectName, "shelf_"+objectName)
		my.data[attribute].SetGlobalLabel("datacenter", my.ParentParams.GetChildContentS("datacenter"))

		exportOptions := node.NewS("export_options")
		instanceLabels := exportOptions.NewChildS("instance_labels", "")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")
		instanceKeys.NewChildS("", "shelf")
		instanceKeys.NewChildS("", "channel")

		// artificial metric for status of child object of shelf
		my.data[attribute].NewMetricUint8("status")

		for _, x := range obj.GetChildren() {

			for _, c := range x.GetAllChildContentS() {

				metricName, display := collector.ParseMetricName(c)

				if strings.HasPrefix(c, "^") {
					if strings.HasPrefix(c, "^^") {
						my.instanceKeys[attribute] = metricName
						my.instanceLabels[attribute].Set(metricName, display)
						instanceKeys.NewChildS("", display)
						my.Logger.Debug().Msgf("added instance key: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
					} else {
						my.instanceLabels[attribute].Set(metricName, display)
						instanceLabels.NewChildS("", display)
						my.Logger.Debug().Msgf("added instance label: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
					}
				} else {
					metric, err := my.data[attribute].NewMetricFloat64(metricName)
					if err != nil {
						my.Logger.Error().Stack().Err(err).Msg("add metric")
						return err
					}
					metric.SetName(display)
					my.Logger.Debug().Msgf("added metric: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				}
			}
		}

		my.Logger.Debug().Msgf("added data for [%s] with %d metrics", attribute, len(my.data[attribute].GetMetrics()))

		my.data[attribute].SetExportOptions(exportOptions)
	}

	my.Logger.Debug().Msgf("initialized with data [%d] objects", len(my.data))
	return nil
}

func (my *Shelf) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		result  *node.Node
		shelves []*node.Node
		err     error
	)

	if !my.client.IsClustered() {
		for _, instance := range data.GetInstances() {
			instance.SetLabel("shelf", instance.GetLabel("shelf_id"))
		}
	}

	if result, err = my.client.InvokeRequestString(my.query); err != nil {
		return nil, err
	}

	// Set all global labels from zapi.go if already not exist
	for a := range my.instanceLabels {
		my.data[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	if x := result.GetChildS("attributes-list"); x != nil {
		shelves = x.GetChildren()
	} else if !my.client.IsClustered() {
		return my.handle7Mode(result)
	}

	if len(shelves) == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no shelf instances found")
	}

	my.Logger.Debug().Msgf("fetching %d shelf counters", len(shelves))

	var output []*matrix.Matrix

	// Purge and reset data
	for _, data1 := range my.data {
		data1.PurgeInstances()
		data1.Reset()
	}

	for _, shelf := range shelves {

		shelfName := shelf.GetChildContentS("shelf")
		shelfId := shelf.GetChildContentS("shelf-uid")

		if !my.client.IsClustered() {
			uid := shelf.GetChildContentS("shelf-id")
			shelfName = uid // no shelf name in 7mode
			shelfId = uid
		}

		for attribute, data1 := range my.data {
			if statusMetric := data1.GetMetric("status"); statusMetric != nil {

				if my.instanceKeys[attribute] == "" {
					my.Logger.Warn().Msgf("no instance keys defined for object [%s], skipping", attribute)
					continue
				}

				objectElem := shelf.GetChildS(attribute)
				if objectElem == nil {
					my.Logger.Warn().Msgf("no [%s] instances on this system", attribute)
					continue
				}

				my.Logger.Debug().Msgf("fetching %d [%s] instances", len(objectElem.GetChildren()), attribute)

				for _, obj := range objectElem.GetChildren() {

					if key := obj.GetChildContentS(my.instanceKeys[attribute]); key != "" {
						instanceKey := shelfId + "." + key
						instance, err := data1.NewInstance(instanceKey)

						if err != nil {
							my.Logger.Error().Msgf("add (%s) instance: %v", attribute, err)
							return nil, err
						}
						my.Logger.Debug().Msgf("add (%s) instance: %s.%s", attribute, shelfId, key)

						for label, labelDisplay := range my.instanceLabels[attribute].Map() {
							if value := obj.GetChildContentS(label); value != "" {
								instance.SetLabel(labelDisplay, value)
							}
						}

						instance.SetLabel("shelf", shelfName)
						instance.SetLabel("shelf_id", shelfId)

						// Each child would have different possible values which is ugly way to write all of them,
						// so normal value would be mapped to 1 and rest all are mapped to 0.
						if instance.GetLabel("status") == "normal" {
							statusMetric.SetValueInt(instance, 1)
						} else {
							statusMetric.SetValueInt(instance, 0)
						}

					} else {
						my.Logger.Debug().Msgf("instance without [%s], skipping", my.instanceKeys[attribute])
					}
				}

				output = append(output, data1)
			}
		}
	}

	// second loop to populate numeric data

	for _, shelf := range shelves {

		shelfId := shelf.GetChildContentS("shelf-uid")
		if !my.client.IsClustered() {
			shelfId = shelf.GetChildContentS("shelf-id")
		}

		for attribute, data1 := range my.data {

			objectElem := shelf.GetChildS(attribute)
			if objectElem == nil {
				continue
			}

			for _, obj := range objectElem.GetChildren() {

				key := obj.GetChildContentS(my.instanceKeys[attribute])

				if key == "" {
					continue
				}

				instance := data1.GetInstance(shelfId + "." + key)

				if instance == nil {
					my.Logger.Debug().Msgf("(%s) instance [%s.%s] not found in cache skipping", attribute, shelfId, key)
					continue
				}

				for metricKey, m := range data1.GetMetrics() {

					if value := strings.Split(obj.GetChildContentS(metricKey), " ")[0]; value != "" {
						if err := m.SetValueString(instance, value); err != nil {
							my.Logger.Debug().Msgf("(%s) failed to parse value (%s): %v", metricKey, value, err)
						} else {
							my.Logger.Debug().Msgf("(%s) added value (%s)", metricKey, value)
						}
					}
				}
			}
		}
	}

	return output, nil
}

func (my *Shelf) handle7Mode(result *node.Node) ([]*matrix.Matrix, error) {
	var (
		shelves  []*node.Node
		channels []*node.Node
	)
	//fallback to 7mode
	channels = result.SearchChildren([]string{"shelf-environ-channel-info"})
	//shelves = result.SearchChildren([]string{"shelf-environ-channel-info", "shelf-environ-shelf-list", "shelf-environ-shelf-info"})

	if len(channels) == 0 {
		return nil, errors.New(errors.ERR_NO_INSTANCE, "no channels found")
	}

	var output []*matrix.Matrix

	// Purge and reset data
	for _, data1 := range my.data {
		data1.PurgeInstances()
		data1.Reset()
	}

	for _, channel := range channels {
		channelName := channel.GetChildContentS("channel-name")
		shelves = channel.SearchChildren([]string{"shelf-environ-shelf-list", "shelf-environ-shelf-info"})

		if len(shelves) == 0 {
			my.Logger.Warn().Str("channel", channelName).Msg("no shelves found")
			continue
		}

		for _, shelf := range shelves {

			uid := shelf.GetChildContentS("shelf-id")
			shelfName := uid // no shelf name in 7mode
			shelfId := uid

			for attribute, data1 := range my.data {
				if statusMetric := data1.GetMetric("status"); statusMetric != nil {

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
							instanceKey := shelfId + "." + key + "." + channelName
							instance, err := data1.NewInstance(instanceKey)

							if err != nil {
								my.Logger.Error().Msgf("add (%s) instance: %v", attribute, err)
								return nil, err
							}
							my.Logger.Debug().Msgf("add (%s) instance: %s.%s", attribute, shelfId, key)

							for label, labelDisplay := range my.instanceLabels[attribute].Map() {
								if value := obj.GetChildContentS(label); value != "" {
									instance.SetLabel(labelDisplay, value)
								}
							}

							instance.SetLabel("shelf", shelfName)
							instance.SetLabel("shelf_id", shelfId)
							instance.SetLabel("channel", channelName)

							// Each child would have different possible values which is ugly way to write all of them,
							// so normal value would be mapped to 1 and rest all are mapped to 0.
							if instance.GetLabel("status") == "normal" {
								statusMetric.SetValueInt(instance, 1)
							} else {
								statusMetric.SetValueInt(instance, 0)
							}

						} else {
							my.Logger.Debug().Msgf("instance without [%s], skipping", my.instanceKeys[attribute])
						}
					}

					output = append(output, data1)
				}
			}
		}
	}

	// second loop to populate numeric data

	for _, channel := range channels {
		channelName := channel.GetChildContentS("channel-name")
		shelves = channel.SearchChildren([]string{"shelf-environ-shelf-list", "shelf-environ-shelf-info"})

		if len(shelves) == 0 {
			continue
		}
		for _, shelf := range shelves {

			shelfId := shelf.GetChildContentS("shelf-id")

			for attribute, data1 := range my.data {

				objectElem := shelf.GetChildS(attribute)
				if objectElem == nil {
					continue
				}

				for _, obj := range objectElem.GetChildren() {

					key := obj.GetChildContentS(my.instanceKeys[attribute])

					if key == "" {
						continue
					}

					instance := data1.GetInstance(shelfId + "." + key + "." + channelName)

					if instance == nil {
						my.Logger.Debug().Msgf("(%s) instance [%s.%s] not found in cache skipping", attribute, shelfId, key)
						continue
					}

					for metricKey, m := range data1.GetMetrics() {

						if value := strings.Split(obj.GetChildContentS(metricKey), " ")[0]; value != "" {
							if err := m.SetValueString(instance, value); err != nil {
								my.Logger.Debug().Msgf("(%s) failed to parse value (%s): %v", metricKey, value, err)
							} else {
								my.Logger.Debug().Msgf("(%s) added value (%s)", metricKey, value)
							}
						}
					}
				}
			}
		}
	}

	return output, nil
}
