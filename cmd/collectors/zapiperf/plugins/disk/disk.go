// Package shelf Copyright NetApp Inc, 2021 All rights reserved
package disk

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"strconv"
	"strings"
)

const BatchSize = "500"

type Disk struct {
	*plugin.AbstractPlugin
	shelfData      map[string]*matrix.Matrix
	powerData      map[string]*matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	batchSize      string
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

var shelfMetrics = []string{
	"average_ambient_temperature",
	"average_fan_speed",
	"average_temperature",
	"max_fan_speed",
	"max_temperature",
	"min_ambient_temperature",
	"min_fan_speed",
	"min_temperature",
	"power",
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Disk{AbstractPlugin: p}
}

func (d *Disk) Init() error {

	var err error

	if err = d.InitAbc(); err != nil {
		return err
	}

	if d.client, err = zapi.New(conf.ZapiPoller(d.ParentParams)); err != nil {
		d.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = d.client.Init(5); err != nil {
		return err
	}

	d.query = "storage-shelf-info-get-iter"

	d.Logger.Debug().Msg("plugin connected!")

	d.shelfData = make(map[string]*matrix.Matrix)
	d.instanceKeys = make(map[string]string)
	d.instanceLabels = make(map[string]*dict.Dict)

	objects := d.Params.GetChildS("objects")
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

		d.instanceLabels[attribute] = dict.New()

		d.shelfData[attribute] = matrix.New(d.Parent+".Shelf", "shelf_"+objectName, "shelf_"+objectName)
		d.shelfData[attribute].SetGlobalLabel("datacenter", d.ParentParams.GetChildContentS("datacenter"))

		exportOptions := node.NewS("export_options")
		instanceLabels := exportOptions.NewChildS("instance_labels", "")
		instanceKeys := exportOptions.NewChildS("instance_keys", "")
		instanceKeys.NewChildS("", "shelf")
		instanceKeys.NewChildS("", "channel")

		// artificial metric for status of child object of shelf
		_, _ = d.shelfData[attribute].NewMetricUint8("status")

		for _, x := range obj.GetChildren() {

			for _, c := range x.GetAllChildContentS() {

				metricName, display, kind, _ := util.ParseMetric(c)

				switch kind {
				case "key":
					d.instanceKeys[attribute] = metricName
					d.instanceLabels[attribute].Set(metricName, display)
					instanceKeys.NewChildS("", display)
					d.Logger.Debug().Msgf("added instance key: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				case "label":
					d.instanceLabels[attribute].Set(metricName, display)
					instanceLabels.NewChildS("", display)
					d.Logger.Debug().Msgf("added instance label: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				case "float":
					_, err := d.shelfData[attribute].NewMetricFloat64(metricName, display)
					if err != nil {
						d.Logger.Error().Stack().Err(err).Msg("add metric")
						return err
					}
					d.Logger.Debug().Msgf("added metric: (%s) (%s) [%s]", attribute, x.GetNameS(), display)
				}
			}
		}

		d.Logger.Debug().Msgf("added shelfData for [%s] with %d metrics", attribute, len(d.shelfData[attribute].GetMetrics()))

		d.shelfData[attribute].SetExportOptions(exportOptions)
	}

	d.Logger.Debug().Msgf("initialized with shelfData [%d] objects", len(d.shelfData))

	// setup batchSize for request
	d.batchSize = BatchSize
	if b := d.Params.GetChildContentS("batch_size"); b != "" {
		if _, err := strconv.Atoi(b); err == nil {
			d.batchSize = b
		}
	}

	d.initShelfPowerMatrix()
	return nil
}

func (d *Disk) initShelfPowerMatrix() {
	d.powerData = make(map[string]*matrix.Matrix)
	d.powerData["shelf"] = matrix.New(d.Parent+".Shelf", "shelf", "shelf")

	for _, k := range shelfMetrics {
		err := matrix.CreateMetric(k, d.powerData["shelf"])
		if err != nil {
			d.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
		}
	}
}

func (d *Disk) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	var (
		err    error
		output []*matrix.Matrix
	)

	// Set all global labels from zapi.go if already not exist
	for a := range d.instanceLabels {
		d.shelfData[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	for a := range d.powerData {
		d.powerData[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	request := node.NewXMLS(d.query)
	if d.client.IsClustered() {
		request.NewChildS("max-records", d.batchSize)
	}

	result, err := d.client.InvokeZapiCall(request)
	if err != nil {
		return nil, err
	}

	output, err = d.handleCMode(result)
	if err != nil {
		return output, err
	}

	return d.handleShelfPower(result, output)
}

func (d *Disk) handleShelfPower(shelves []*node.Node, output []*matrix.Matrix) ([]*matrix.Matrix, error) {
	// Purge and reset data
	data := d.powerData["shelf"]
	data.PurgeInstances()
	data.Reset()

	for _, shelf := range shelves {
		shelfName := shelf.GetChildContentS("shelf")
		shelfID := shelf.GetChildContentS("shelf-uid")
		instanceKey := shelfID
		instance, err := data.NewInstance(instanceKey)
		if err != nil {
			d.Logger.Error().Err(err).Str("key", instanceKey).Msg("Failed to add instance")
			return output, err
		}
		instance.SetLabel("shelf", shelfName)

	}
	err := d.calculateEnvironmentMetrics(data)
	if err != nil {
		return output, err
	}

	output = append(output, data)
	return output, nil
}

func (d *Disk) calculateEnvironmentMetrics(data *matrix.Matrix) error {
	var err error
	shelfEnvironmentMetricMap := make(map[string]*shelfEnvironmentMetric, 0)
	for _, o := range d.shelfData {
		for k, instance := range o.GetInstances() {
			lastInd := strings.LastIndex(k, "#")
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

	for key, v := range shelfEnvironmentMetricMap {
		for _, k := range shelfMetrics {
			m := data.GetMetric(k)
			instance := data.GetInstance(key)
			if instance == nil {
				d.Logger.Warn().Str("key", key).Msg("Instance not found")
				continue
			}
			switch k {
			case "power":
				var sumPower float64
				for k1, v1 := range v.voltageSensor {
					if v2, ok := v.currentSensor[k1]; ok {
						// in W
						sumPower += (v1 * v2) / 1000
					} else {
						d.Logger.Warn().Str("voltage sensor id", k1).Msg("missing current sensor")
					}
				}

				err = m.SetValueFloat64(instance, sumPower)
				if err != nil {
					d.Logger.Error().Float64("power", sumPower).Err(err).Msg("Unable to set power")
				}

			case "average_ambient_temperature":
				if len(v.ambientTemperature) > 0 {
					aaT := util.Avg(v.ambientTemperature)
					err = m.SetValueFloat64(instance, aaT)
					if err != nil {
						d.Logger.Error().Float64("average_ambient_temperature", aaT).Err(err).Msg("Unable to set average_ambient_temperature")
					}
				}
			case "min_ambient_temperature":
				maT := util.Min(v.ambientTemperature)
				err = m.SetValueFloat64(instance, maT)
				if err != nil {
					d.Logger.Error().Float64("min_ambient_temperature", maT).Err(err).Msg("Unable to set min_ambient_temperature")
				}
			case "max_temperature":
				mT := util.Max(v.nonAmbientTemperature)
				err = m.SetValueFloat64(instance, mT)
				if err != nil {
					d.Logger.Error().Float64("max_temperature", mT).Err(err).Msg("Unable to set max_temperature")
				}
			case "average_temperature":
				if len(v.nonAmbientTemperature) > 0 {
					nat := util.Avg(v.nonAmbientTemperature)
					err = m.SetValueFloat64(instance, nat)
					if err != nil {
						d.Logger.Error().Float64("average_temperature", nat).Err(err).Msg("Unable to set average_temperature")
					}
				}
			case "min_temperature":
				mT := util.Min(v.nonAmbientTemperature)
				err = m.SetValueFloat64(instance, mT)
				if err != nil {
					d.Logger.Error().Float64("min_temperature", mT).Err(err).Msg("Unable to set min_temperature")
				}
			case "average_fan_speed":
				if len(v.fanSpeed) > 0 {
					afs := util.Avg(v.fanSpeed)
					err = m.SetValueFloat64(instance, afs)
					if err != nil {
						d.Logger.Error().Float64("average_fan_speed", afs).Err(err).Msg("Unable to set average_fan_speed")
					}
				}
			case "max_fan_speed":
				mfs := util.Max(v.fanSpeed)
				err = m.SetValueFloat64(instance, mfs)
				if err != nil {
					d.Logger.Error().Float64("max_fan_speed", mfs).Err(err).Msg("Unable to set max_fan_speed")
				}
			case "min_fan_speed":
				mfs := util.Min(v.fanSpeed)
				err = m.SetValueFloat64(instance, mfs)
				if err != nil {
					d.Logger.Error().Float64("min_fan_speed", mfs).Err(err).Msg("Unable to set min_fan_speed")
				}
			}
		}
	}
	return nil
}

func (d *Disk) handleCMode(shelves []*node.Node) ([]*matrix.Matrix, error) {
	var (
		output []*matrix.Matrix
	)

	d.Logger.Debug().Msgf("fetching %d shelf counters", len(shelves))

	// Purge and reset data
	for _, data1 := range d.shelfData {
		data1.PurgeInstances()
		data1.Reset()
	}

	for _, shelf := range shelves {

		shelfName := shelf.GetChildContentS("shelf")
		shelfID := shelf.GetChildContentS("shelf-uid")

		if !d.client.IsClustered() {
			uid := shelf.GetChildContentS("shelf-id")
			shelfName = uid // no shelf name in 7mode
			shelfID = uid
		}

		for attribute, data1 := range d.shelfData {
			if statusMetric := data1.GetMetric("status"); statusMetric != nil {

				if d.instanceKeys[attribute] == "" {
					d.Logger.Warn().Msgf("no instance keys defined for object [%s], skipping", attribute)
					continue
				}

				objectElem := shelf.GetChildS(attribute)
				if objectElem == nil {
					d.Logger.Warn().Msgf("no [%s] instances on this system", attribute)
					continue
				}

				d.Logger.Debug().Msgf("fetching %d [%s] instances", len(objectElem.GetChildren()), attribute)

				for _, obj := range objectElem.GetChildren() {

					if key := obj.GetChildContentS(d.instanceKeys[attribute]); key != "" {
						instanceKey := shelfID + "#" + key
						instance, err := data1.NewInstance(instanceKey)

						if err != nil {
							d.Logger.Error().Err(err).Str("attribute", attribute).Msg("Failed to add instance")
							return nil, err
						}
						d.Logger.Debug().Msgf("add (%s) instance: %s.%s", attribute, shelfID, key)

						for label, labelDisplay := range d.instanceLabels[attribute].Map() {
							if value := obj.GetChildContentS(label); value != "" {
								instance.SetLabel(labelDisplay, value)
							}
						}

						instance.SetLabel("shelf", shelfName)
						instance.SetLabel("shelf_id", shelfID)

						// Each child would have different possible values which is an ugly way to write all of them,
						// so normal value would be mapped to 1 and rest all are mapped to 0.
						if instance.GetLabel("status") == "normal" {
							_ = statusMetric.SetValueInt64(instance, 1)
						} else {
							_ = statusMetric.SetValueInt64(instance, 0)
						}

						for metricKey, m := range data1.GetMetrics() {

							if value := strings.Split(obj.GetChildContentS(metricKey), " ")[0]; value != "" {
								if err := m.SetValueString(instance, value); err != nil {
									d.Logger.Debug().Msgf("(%s) failed to parse value (%s): %v", metricKey, value, err)
								} else {
									d.Logger.Debug().Msgf("(%s) added value (%s)", metricKey, value)
								}
							}
						}

					} else {
						d.Logger.Debug().Msgf("instance without [%s], skipping", d.instanceKeys[attribute])
					}
				}

				output = append(output, data1)
			}
		}
	}

	return output, nil
}
