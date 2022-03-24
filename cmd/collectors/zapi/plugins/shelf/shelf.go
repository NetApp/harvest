/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package shelf

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
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

type shelfEnvironmentMetric struct {
	key                   string
	ambientTemperature    []float64
	nonAmbientTemperature []float64
	fanSpeed              []float64
	voltageSensor         map[string]float64
	currentSensor         map[string]float64
}

var eMetrics = []string{"power", "ambient_temperature", "max_temperature", "average_temperature", "average_fan_speed", "max_fan_speed", "min_fan_speed"}

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
		result *node.Node
		err    error
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

	var output []*matrix.Matrix

	if my.client.IsClustered() {
		output, err = my.handleCMode(result)
	} else {
		output, err = my.handle7Mode(result)
	}
	if err != nil {
		return output, err
	}

	return my.calculateEnvironmentMetrics(output, data)
}

func (my *Shelf) calculateEnvironmentMetrics(output []*matrix.Matrix, data *matrix.Matrix) ([]*matrix.Matrix, error) {
	var err error
	shelfEnvironmentMetricMap := make(map[string]*shelfEnvironmentMetric, 0)
	for _, o := range my.data {
		for k, instance := range o.GetInstances() {
			lastInd := strings.LastIndex(k, ".")
			iKey := k[:lastInd]
			iKey2 := k[lastInd+1:]
			if _, ok := shelfEnvironmentMetricMap[iKey]; !ok {
				shelfEnvironmentMetricMap[iKey] = &shelfEnvironmentMetric{key: iKey, ambientTemperature: []float64{}, nonAmbientTemperature: []float64{}, fanSpeed: []float64{}}
			}
			for mkey, metric := range o.GetMetrics() {
				if o.Object == "shelf_temperature" {
					if mkey == "temp-sensor-reading" {
						isAmbient := instance.GetLabel("temp_is_ambient")
						if isAmbient == "true" {
							if value, ok := metric.GetValueFloat64(instance); ok {
								shelfEnvironmentMetricMap[iKey].ambientTemperature = append(shelfEnvironmentMetricMap[iKey].ambientTemperature, value)
							}
						}
						if isAmbient == "false" {
							if value, ok := metric.GetValueFloat64(instance); ok {
								shelfEnvironmentMetricMap[iKey].nonAmbientTemperature = append(shelfEnvironmentMetricMap[iKey].nonAmbientTemperature, value)
							}
						}
					}
				} else if o.Object == "shelf_fan" {
					if mkey == "fan-rpm" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							shelfEnvironmentMetricMap[iKey].fanSpeed = append(shelfEnvironmentMetricMap[iKey].fanSpeed, value)
						}
					}
				} else if o.Object == "shelf_voltage" {
					if mkey == "voltage-sensor-reading" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							if shelfEnvironmentMetricMap[iKey].voltageSensor == nil {
								shelfEnvironmentMetricMap[iKey].voltageSensor = make(map[string]float64, 0)
							}
							shelfEnvironmentMetricMap[iKey].voltageSensor[iKey2] = value
						}
					}
				} else if o.Object == "shelf_sensor" {
					if mkey == "current-sensor-reading" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							if shelfEnvironmentMetricMap[iKey].currentSensor == nil {
								shelfEnvironmentMetricMap[iKey].currentSensor = make(map[string]float64, 0)
							}
							shelfEnvironmentMetricMap[iKey].currentSensor[iKey2] = value
						}
					}
				}
			}
		}
	}

	for _, k := range eMetrics {
		my.createEnvironmentMetric(data, k)
	}
	for key, v := range shelfEnvironmentMetricMap {
		for _, k := range eMetrics {
			m := data.GetMetric(k)
			instance := data.GetInstance(key)
			switch k {
			case "power":
				var sumPower float64
				for k1, v1 := range v.voltageSensor {
					if v2, ok := v.currentSensor[k1]; ok {
						// in W
						sumPower += (v1 * v2) / 1000
					} else {
						my.Logger.Warn().Str("voltage sensor id", k1).Msg("missing current sensor")
					}
				}
				// convert to KW
				sumPower = sumPower / 1000

				err = m.SetValueFloat64(instance, sumPower)
				if err != nil {
					my.Logger.Error().Float64("power", sumPower).Err(err).Msg("Unable to set power")
				} else {
					m.SetLabel("unit", "kW")
				}

			case "ambient_temperature":
				if len(v.ambientTemperature) > 0 {
					aT := util.SumNumbers(v.ambientTemperature) / float64(len(v.ambientTemperature))
					err = m.SetValueFloat64(instance, aT)
					if err != nil {
						my.Logger.Error().Float64("ambient_temperature", aT).Err(err).Msg("Unable to set ambient_temperature")
					} else {
						m.SetLabel("unit", "C")
					}
				}
			case "max_temperature":
				mT := util.Max(v.nonAmbientTemperature)
				err = m.SetValueFloat64(instance, util.Max(v.nonAmbientTemperature))
				if err != nil {
					my.Logger.Error().Float64("max_temperature", mT).Err(err).Msg("Unable to set max_temperature")
				} else {
					m.SetLabel("unit", "C")
				}
			case "average_temperature":
				if len(v.nonAmbientTemperature) > 0 {
					nat := util.SumNumbers(v.nonAmbientTemperature) / float64(len(v.nonAmbientTemperature))
					err = m.SetValueFloat64(instance, nat)
					if err != nil {
						my.Logger.Error().Float64("average_temperature", nat).Err(err).Msg("Unable to set average_temperature")
					} else {
						m.SetLabel("unit", "C")
					}
				}
			case "average_fan_speed":
				if len(v.fanSpeed) > 0 {
					afs := util.SumNumbers(v.fanSpeed) / float64(len(v.fanSpeed))
					err = m.SetValueFloat64(instance, afs)
					if err != nil {
						my.Logger.Error().Float64("average_fan_speed", afs).Err(err).Msg("Unable to set average_fan_speed")
					} else {
						m.SetLabel("unit", "rpm")
					}
				}
			case "max_fan_speed":
				mfs := util.Max(v.fanSpeed)
				err = m.SetValueFloat64(instance, mfs)
				if err != nil {
					my.Logger.Error().Float64("max_fan_speed", mfs).Err(err).Msg("Unable to set max_fan_speed")
				} else {
					m.SetLabel("unit", "rpm")
				}
			case "min_fan_speed":
				mfs := util.Min(v.fanSpeed)
				err = m.SetValueFloat64(instance, mfs)
				if err != nil {
					my.Logger.Error().Float64("min_fan_speed", mfs).Err(err).Msg("Unable to set min_fan_speed")
				} else {
					m.SetLabel("unit", "rpm")
				}
			}
		}
	}
	return output, nil
}

func (my *Shelf) createEnvironmentMetric(data *matrix.Matrix, key string) {
	var err error
	at := data.GetMetric(key)
	if at == nil {
		if at, err = data.NewMetricFloat64(key); err != nil {
			my.Logger.Error().Stack().Err(err).Msg("error")
		}
	}
}

func (my *Shelf) handleCMode(result *node.Node) ([]*matrix.Matrix, error) {
	var (
		shelves []*node.Node
	)

	if x := result.GetChildS("attributes-list"); x != nil {
		shelves = x.GetChildren()
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
							my.Logger.Error().Err(err).Str("attribute", attribute).Msg("Failed to add instance")
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

func (my *Shelf) handle7Mode(result *node.Node) ([]*matrix.Matrix, error) {
	var (
		shelves  []*node.Node
		channels []*node.Node
	)
	//fallback to 7mode
	channels = result.SearchChildren([]string{"shelf-environ-channel-info"})

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
