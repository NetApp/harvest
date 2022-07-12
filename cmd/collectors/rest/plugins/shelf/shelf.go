package shelf

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"sort"
	"strings"
	"time"
)

type Shelf struct {
	*plugin.AbstractPlugin
	data           map[string]*matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	client         *rest.Client
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

var eMetrics = []string{
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
	return &Shelf{AbstractPlugin: p}
}

func (my *Shelf) Init() error {

	var err error
	shelfMetric := make(map[string][]string)

	shelfMetric["fans => fan"] = []string{
		"^^id => fan_id",
		"^location",
		"^state => status",
		"rpm",
	}
	shelfMetric["current_sensors => sensor"] = []string{
		"^^id => sensor_id",
		"^location",
		"^state => status",
		"current => reading",
	}
	shelfMetric["frus => psu"] = []string{
		"^^id => psu_id",
		"^installed => enabled",
		//"^location",
		"^part_number",
		"^serial_number => serial",
		"^psu.model => type",
		"^state => status",
		"psu.power_drawn => power_drawn",
		"psu.power_rating => power_rating",
	}
	shelfMetric["temperature_sensors => temperature"] = []string{
		"^^id => sensor_id",
		"^threshold.high.critical => high_critical",
		"^threshold.high.warning => high_warning",
		"^ambient => temp_is_ambient",
		"^threshold.low.critical => low_critical",
		"^threshold.low.warning => low_warning",
		"^state => status",
		"temperature => reading",
	}
	shelfMetric["voltage_sensors => voltage"] = []string{
		"^^id => sensor_id",
		"^location",
		"^state => status",
		"voltage => reading",
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

	my.query = "api/storage/shelves"

	my.data = make(map[string]*matrix.Matrix)
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	for attribute, childObj := range shelfMetric {

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

		for _, c := range childObj {

			metricName, display, kind, _ := util.ParseMetric(c)

			switch kind {
			case "key":
				my.instanceKeys[attribute] = metricName
				my.instanceLabels[attribute].Set(metricName, display)
				instanceKeys.NewChildS("", display)
				my.Logger.Debug().Msgf("added instance key: (%s) [%s]", attribute, display)
			case "label":
				my.instanceLabels[attribute].Set(metricName, display)
				instanceLabels.NewChildS("", display)
				my.Logger.Debug().Msgf("added instance label: (%s) [%s]", attribute, display)
			case "float":
				metric, err := my.data[attribute].NewMetricFloat64(metricName)
				if err != nil {
					my.Logger.Error().Stack().Err(err).Msg("add metric")
					return err
				}
				metric.SetName(display)
				my.Logger.Debug().Msgf("added metric: (%s) [%s]", attribute, display)
			}
		}

		my.Logger.Debug().Msgf("added data for [%s] with %d metrics", attribute, len(my.data[attribute].GetMetrics()))

		my.data[attribute].SetExportOptions(exportOptions)
	}

	my.Logger.Debug().Msgf("initialized with data [%d] objects", len(my.data))
	return nil
}

func (my *Shelf) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {

	// Set all global labels from rest.go if already not exist
	for a := range my.instanceLabels {
		my.data[a].SetGlobalLabels(data.GetGlobalLabels())
	}

	href := rest.BuildHref("", "*", nil, "", "", "", "", my.query)

	records, err := rest.Fetch(my.client, href)
	if err != nil {
		my.Logger.Error().Err(err).Str("href", href).Msg("Failed to fetch data")
		return nil, err
	}

	if len(records) == 0 {
		return nil, errs.New(errs.ErrNoInstance, "no "+my.query+" instances on cluster")
	}

	var output []*matrix.Matrix
	noSet := make(map[string]any)

	// Purge and reset data
	for _, data1 := range my.data {
		data1.PurgeInstances()
		data1.Reset()
	}
	for _, shelf := range records {

		if !shelf.IsObject() {
			my.Logger.Warn().Str("type", shelf.Type.String()).Msg("Shelf is not object, skipping")
			continue
		}

		shelfName := shelf.Get("name").String()
		shelfSerialNumber := shelf.Get("serial_number").String()

		for attribute, data1 := range my.data {
			if statusMetric := data1.GetMetric("status"); statusMetric != nil {

				if my.instanceKeys[attribute] == "" {
					my.Logger.Warn().Str("attribute", attribute).Msg("no instance keys defined for object, skipping")
					continue
				}

				if childObj := shelf.Get(attribute); childObj.Exists() {
					if childObj.IsArray() {
						for _, obj := range childObj.Array() {

							// This is special condition, because child records can't be filterable in parent REST call
							// frus type can be [module, psu] and we would only need psu for our use-case.
							if attribute == "frus" && obj.Get("type").Exists() && obj.Get("type").String() != "psu" {
								continue
							}

							if key := obj.Get(my.instanceKeys[attribute]); key.Exists() {
								instanceKey := shelfSerialNumber + "#" + attribute + "#" + key.String()
								shelfChildInstance, err2 := data1.NewInstance(instanceKey)

								if err2 != nil {
									my.Logger.Error().Err(err).Str("attribute", attribute).Str("instanceKey", instanceKey).Msg("Failed to add instance")
									break
								}
								my.Logger.Debug().Msgf("add (%s) instance: %s.%s.%s", attribute, shelfSerialNumber, attribute, key)

								for label, labelDisplay := range my.instanceLabels[attribute].Map() {
									if value := obj.Get(label); value.Exists() {
										if value.IsArray() {
											var labelArray []string
											for _, r := range value.Array() {
												labelString := r.String()
												labelArray = append(labelArray, labelString)
											}
											shelfChildInstance.SetLabel(labelDisplay, strings.Join(labelArray, ","))
										} else {
											shelfChildInstance.SetLabel(labelDisplay, value.String())
										}
									} else {
										// spams a lot currently due to missing label mappings. Moved to debug for now till rest gaps are filled
										my.Logger.Debug().Str("Instance key", instanceKey).Str("label", label).Msg("Missing label value")
									}
								}

								shelfChildInstance.SetLabel("shelf", shelfName)

								// Each child would have different possible values which is ugly way to write all of them,
								// so normal value would be mapped to 1 and rest all are mapped to 0.
								if shelfChildInstance.GetLabel("status") == "ok" {
									_ = statusMetric.SetValueInt(shelfChildInstance, 1)
								} else {
									_ = statusMetric.SetValueInt(shelfChildInstance, 0)
								}

								for metricKey, m := range data1.GetMetrics() {

									if value := obj.Get(metricKey); value.Exists() {
										if err = m.SetValueString(shelfChildInstance, value.String()); err != nil { // float
											my.Logger.Error().Err(err).Str("key", metricKey).Str("metric", m.GetName()).Str("value", value.String()).
												Msg("Unable to set float key on metric")
										} else {
											my.Logger.Debug().Str("metricKey", metricKey).Str("value", value.String()).Msg("added")
										}
									}
								}

							} else {
								my.Logger.Debug().Msgf("instance without [%s], skipping", my.instanceKeys[attribute])
							}
						}
					}
				} else {
					noSet[attribute] = struct{}{}
					continue
				}
				output = append(output, data1)
			}
		}
	}

	if len(noSet) > 0 {
		attributes := make([]string, 0)
		for k := range noSet {
			attributes = append(attributes, k)
		}
		sort.Strings(attributes)
		my.Logger.Warn().Strs("attributes", attributes).Msg("No instances")
	}
	err = my.calculateEnvironmentMetrics(data)

	return output, err
}

func (my *Shelf) calculateEnvironmentMetrics(data *matrix.Matrix) error {
	var err error
	shelfEnvironmentMetricMap := make(map[string]*shelfEnvironmentMetric, 0)
	for _, o := range my.data {
		for k, instance := range o.GetInstances() {
			firstInd := strings.Index(k, "#")
			lastInd := strings.LastIndex(k, "#")
			iKey := k[:firstInd]
			iKey2 := k[lastInd+1:]
			if _, ok := shelfEnvironmentMetricMap[iKey]; !ok {
				shelfEnvironmentMetricMap[iKey] = &shelfEnvironmentMetric{key: iKey, ambientTemperature: []float64{}, nonAmbientTemperature: []float64{}, fanSpeed: []float64{}}
			}
			for mkey, metric := range o.GetMetrics() {
				if o.Object == "shelf_temperature" {
					if mkey == "temperature" {
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
					if mkey == "rpm" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							shelfEnvironmentMetricMap[iKey].fanSpeed = append(shelfEnvironmentMetricMap[iKey].fanSpeed, value)
						}
					}
				} else if o.Object == "shelf_voltage" {
					if mkey == "voltage" {
						if value, ok := metric.GetValueFloat64(instance); ok {
							if shelfEnvironmentMetricMap[iKey].voltageSensor == nil {
								shelfEnvironmentMetricMap[iKey].voltageSensor = make(map[string]float64, 0)
							}
							shelfEnvironmentMetricMap[iKey].voltageSensor[iKey2] = value
						}
					}
				} else if o.Object == "shelf_sensor" {
					if mkey == "current" {
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
		err := matrix.CreateMetric(k, data)
		if err != nil {
			my.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
		}
	}
	for key, v := range shelfEnvironmentMetricMap {
		for _, k := range eMetrics {
			m := data.GetMetric(k)
			instance := data.GetInstance(key)
			if instance == nil {
				my.Logger.Warn().Str("key", key).Msg("Instance not found")
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
						my.Logger.Warn().Str("voltage sensor id", k1).Msg("missing current sensor")
					}
				}

				err = m.SetValueFloat64(instance, sumPower)
				if err != nil {
					my.Logger.Error().Float64("power", sumPower).Err(err).Msg("Unable to set power")
				} else {
					m.SetLabel("unit", "W")
				}

			case "average_ambient_temperature":
				if len(v.ambientTemperature) > 0 {
					aaT := util.Avg(v.ambientTemperature)
					err = m.SetValueFloat64(instance, aaT)
					if err != nil {
						my.Logger.Error().Float64("average_ambient_temperature", aaT).Err(err).Msg("Unable to set average_ambient_temperature")
					} else {
						m.SetLabel("unit", "C")
					}
				}
			case "min_ambient_temperature":
				maT := util.Min(v.ambientTemperature)
				err = m.SetValueFloat64(instance, maT)
				if err != nil {
					my.Logger.Error().Float64("min_ambient_temperature", maT).Err(err).Msg("Unable to set min_ambient_temperature")
				} else {
					m.SetLabel("unit", "C")
				}
			case "max_temperature":
				mT := util.Max(v.nonAmbientTemperature)
				err = m.SetValueFloat64(instance, mT)
				if err != nil {
					my.Logger.Error().Float64("max_temperature", mT).Err(err).Msg("Unable to set max_temperature")
				} else {
					m.SetLabel("unit", "C")
				}
			case "average_temperature":
				if len(v.nonAmbientTemperature) > 0 {
					nat := util.Avg(v.nonAmbientTemperature)
					err = m.SetValueFloat64(instance, nat)
					if err != nil {
						my.Logger.Error().Float64("average_temperature", nat).Err(err).Msg("Unable to set average_temperature")
					} else {
						m.SetLabel("unit", "C")
					}
				}
			case "min_temperature":
				mT := util.Min(v.nonAmbientTemperature)
				err = m.SetValueFloat64(instance, mT)
				if err != nil {
					my.Logger.Error().Float64("min_temperature", mT).Err(err).Msg("Unable to set min_temperature")
				} else {
					m.SetLabel("unit", "C")
				}
			case "average_fan_speed":
				if len(v.fanSpeed) > 0 {
					afs := util.Avg(v.fanSpeed)
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
	return nil
}
