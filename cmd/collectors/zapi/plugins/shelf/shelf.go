// Package shelf Copyright NetApp Inc, 2021 All rights reserved
package shelf

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"strings"
)

const BatchSize = "500"

type Shelf struct {
	*plugin.AbstractPlugin
	data           map[string]*matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	batchSize      string
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

	my.query = "storage-shelf-environment-list-info"

	my.Logger.Debug().Msg("plugin connected!")

	my.data = make(map[string]*matrix.Matrix)
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	objects := my.Params.GetChildS("objects")
	if objects == nil {
		return errs.New(errs.ErrMissingParams, "objects")
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
		_, _ = my.data[attribute].NewMetricUint8("status")

		for _, x := range obj.GetChildren() {

			for _, c := range x.GetAllChildContentS() {

				metricName, display, kind, _ := util.ParseMetric(c)

				switch kind {
				case "key":
					my.instanceKeys[attribute] = metricName
					my.instanceLabels[attribute].Set(metricName, display)
					instanceKeys.NewChildS("", display)
					my.Logger.Debug().Msgf("added instance key: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				case "label":
					my.instanceLabels[attribute].Set(metricName, display)
					instanceLabels.NewChildS("", display)
					my.Logger.Debug().Msgf("added instance label: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				case "float":
					_, err := my.data[attribute].NewMetricFloat64(metricName, display)
					if err != nil {
						my.Logger.Error().Stack().Err(err).Msg("add metric")
						return err
					}
					my.Logger.Debug().Msgf("added metric: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				}
			}
		}

		my.Logger.Debug().Msgf("added data for [%s] with %d metrics", attribute, len(my.data[attribute].GetMetrics()))

		my.data[attribute].SetExportOptions(exportOptions)
	}

	my.Logger.Debug().Msgf("initialized with data [%d] objects", len(my.data))

	// setup batchSize for request
	my.batchSize = BatchSize
	return nil
}

func (my *Shelf) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		err    error
		output []*matrix.Matrix
	)

	// Only 7mode is supported through this plugin
	if my.client.IsClustered() {
		return nil, nil
	}

	for _, instance := range data.GetInstances() {
		instance.SetLabel("shelf", instance.GetLabel("shelf_id"))
	}

	// Set all global labels from zapi.go if already not exist
	for a := range my.instanceLabels {
		my.data[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	request := node.NewXMLS(my.query)

	result, err := my.client.InvokeZapiCall(request)
	if err != nil {
		return nil, err
	}

	output, err = my.handle7Mode(result)

	if err != nil {
		return output, err
	}

	return output, nil
}

func (my *Shelf) handle7Mode(result []*node.Node) ([]*matrix.Matrix, error) {
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
							instanceKey := shelfID + "." + key + "." + channelName
							instance, err := data1.NewInstance(instanceKey)

							if err != nil {
								my.Logger.Error().Msgf("add (%s) instance: %v", attribute, err)
								return nil, err
							}
							my.Logger.Debug().Msgf("add (%s) instance: %s.%s", attribute, shelfID, key)

							for label, labelDisplay := range my.instanceLabels[attribute].Map() {
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
	}
	return output, nil
}
