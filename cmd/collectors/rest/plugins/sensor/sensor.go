/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package sensor

import (
	"goharvest2/cmd/poller/plugin"
	"goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/dict"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/util"
	"regexp"
	"sort"
	"strings"
)

type Sensor struct {
	*plugin.AbstractPlugin
	data           *matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
	client         *zapi.Client
	query          string
}

type sensorEnvironmentMetric struct {
	key                   sensorKey
	ambientTemperature    []float64
	nonAmbientTemperature []float64
	fanSpeed              []float64
	powerSensor           map[string]*sensorValue
	voltageSensor         map[string]*sensorValue
	currentSensor         map[string]*sensorValue
}

type sensorKey struct {
	node   string
	sensor string
}

type sensorValue struct {
	name  string
	value float64
	unit  string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Sensor{AbstractPlugin: p}
}

var ambientRegex = regexp.MustCompile(`^(Ambient Temp|Ambient Temp \d|PSU\d AmbTemp|PSU\d Inlet|PSU\d Inlet Temp|In Flow Temp|Front Temp|System Inlet)$`)
var powerInRegex = regexp.MustCompile(`^PSU\d (InPwr Monitor|InPower|PIN|Power In)$`)
var voltageRegex = regexp.MustCompile(`^PSU\d (\d+V|InVoltage|VIN|AC In Volt)$`)
var currentRegex = regexp.MustCompile(`^PSU\d (\d+V Curr|Curr|InCurrent|Curr IIN|AC In Curr)$`)
var eMetrics = []string{"power", "ambient_temperature", "max_temperature", "average_temperature", "average_fan_speed", "max_fan_speed", "min_fan_speed"}

func (my *Sensor) Init() error {
	if err := my.InitAbc(); err != nil {
		return err
	}

	my.data = matrix.New(my.Parent+".Sensor", "environment_sensor", "environment_sensor")
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	// init environment metrics in plugin matrix
	// create environment metric if not exists
	for _, k := range eMetrics {
		err := matrix.CreateMetric(k, my.data)
		if err != nil {
			my.Logger.Warn().Err(err).Str("key", k).Msg("error while creating metric")
		}
	}
	return nil
}

func (my *Sensor) Run(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	// Purge and reset data
	my.data.PurgeInstances()
	my.data.Reset()

	// Set all global labels from zapi.go if already not exist
	my.data.SetGlobalLabels(data.GetGlobalLabels())

	return my.calculateEnvironmentMetrics(data)
}

func (my *Sensor) calculateEnvironmentMetrics(data *matrix.Matrix) ([]*matrix.Matrix, error) {
	sensorEnvironmentMetricMap := make(map[sensorKey]*sensorEnvironmentMetric)

	for k, instance := range data.GetInstances() {
		iKey := instance.GetLabel("node")
		if iKey == "" {
			my.Logger.Warn().Str("key", k).Msg("missing node label for instance")
			continue
		}
		// fetching sensor from instance key
		iKey2 := strings.ReplaceAll(k, iKey, "")
		if iKey2 == "" {
			my.Logger.Warn().Str("key", iKey+".").Msg("missing instance key")
			continue
		}
		nodeSensorKey := sensorKey{node: iKey, sensor: iKey2}
		if _, ok := sensorEnvironmentMetricMap[nodeSensorKey]; !ok {
			sensorEnvironmentMetricMap[nodeSensorKey] = &sensorEnvironmentMetric{key: nodeSensorKey, ambientTemperature: []float64{}, nonAmbientTemperature: []float64{}, fanSpeed: []float64{}}
		}
		for mKey, metric := range data.GetMetrics() {
			if mKey == "value" {
				sensorType := instance.GetLabel("type")
				sensorName := instance.GetLabel("sensor")
				sensorUnit := instance.GetLabel("unit")
				isAmbientMatch := ambientRegex.MatchString(sensorName)
				isPowerMatch := powerInRegex.MatchString(sensorName)
				isVoltageMatch := voltageRegex.MatchString(sensorName)
				isCurrentMatch := currentRegex.MatchString(sensorName)

				my.Logger.Debug().Bool("isAmbientMatch", isAmbientMatch).
					Bool("isPowerMatch", isPowerMatch).
					Bool("isVoltageMatch", isVoltageMatch).
					Bool("isCurrentMatch", isCurrentMatch).
					Str("sensorType", sensorType).
					Str("sensorUnit", sensorUnit).
					Str("sensorName", sensorName).
					Msg("")

				if sensorType == "thermal" && isAmbientMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						sensorEnvironmentMetricMap[nodeSensorKey].ambientTemperature = append(sensorEnvironmentMetricMap[nodeSensorKey].ambientTemperature, value)
					}
				}

				if sensorType == "thermal" && !isAmbientMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						sensorEnvironmentMetricMap[nodeSensorKey].nonAmbientTemperature = append(sensorEnvironmentMetricMap[nodeSensorKey].nonAmbientTemperature, value)
					}
				}

				if sensorType == "fan" {
					if value, ok := metric.GetValueFloat64(instance); ok {
						sensorEnvironmentMetricMap[nodeSensorKey].fanSpeed = append(sensorEnvironmentMetricMap[nodeSensorKey].fanSpeed, value)
					}
				}

				if isPowerMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						if sensorEnvironmentMetricMap[nodeSensorKey].powerSensor == nil {
							sensorEnvironmentMetricMap[nodeSensorKey].powerSensor = make(map[string]*sensorValue)
						}
						sensorEnvironmentMetricMap[nodeSensorKey].powerSensor[iKey2] = &sensorValue{name: iKey2, value: value, unit: sensorUnit}
					}
				}

				if isVoltageMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						if sensorEnvironmentMetricMap[nodeSensorKey].voltageSensor == nil {
							sensorEnvironmentMetricMap[nodeSensorKey].voltageSensor = make(map[string]*sensorValue)
						}
						sensorEnvironmentMetricMap[nodeSensorKey].voltageSensor[iKey2] = &sensorValue{name: iKey2, value: value, unit: sensorUnit}
					}
				}

				if isCurrentMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						if sensorEnvironmentMetricMap[nodeSensorKey].currentSensor == nil {
							sensorEnvironmentMetricMap[nodeSensorKey].currentSensor = make(map[string]*sensorValue)
						}
						sensorEnvironmentMetricMap[nodeSensorKey].currentSensor[iKey2] = &sensorValue{name: iKey2, value: value, unit: sensorUnit}
					}
				}
			}
		}
	}

	for key, v := range sensorEnvironmentMetricMap {
		nodeSensorKey := key.node + key.sensor
		instance, err := my.data.NewInstance(nodeSensorKey)
		if err != nil {
			my.Logger.Warn().Str("key", nodeSensorKey).Msg("instance not found")
			continue
		}
		// set node label
		instance.SetLabel("node", key.node)
		// set sensor label
		instance.SetLabel("sensor", key.sensor)
		for _, k := range eMetrics {
			m := my.data.GetMetric(k)
			switch k {
			case "power":
				var sumPower float64
				if len(v.powerSensor) > 0 {
					for _, v1 := range v.powerSensor {
						if v1.unit == "mW" {
							sumPower += v1.value / 1000
						} else if v1.unit == "W" {
							sumPower += v1.value
						} else {
							my.Logger.Warn().Str("unit", v1.unit).Float64("value", v1.value).Msg("unknown power unit")
						}
					}
				} else if len(v.voltageSensor) > 0 && len(v.voltageSensor) == len(v.currentSensor) {
					// sort voltage keys
					voltageKeys := make([]string, 0, len(v.voltageSensor))
					for k := range v.voltageSensor {
						voltageKeys = append(voltageKeys, k)
					}
					sort.Strings(voltageKeys)

					// sort current keys
					currentKeys := make([]string, 0, len(v.currentSensor))
					for k := range v.currentSensor {
						currentKeys = append(currentKeys, k)
					}
					sort.Strings(currentKeys)

					for i := range currentKeys {
						currentKey := currentKeys[i]
						voltageKey := voltageKeys[i]

						//get values
						currentSensorValue := v.currentSensor[currentKey]
						voltageSensorValue := v.voltageSensor[voltageKey]

						// convert units
						if currentSensorValue.unit == "mA" {
							currentSensorValue.value = currentSensorValue.value / 1000
						} else if currentSensorValue.unit == "A" {
							// do nothing
						} else {
							my.Logger.Warn().Str("unit", currentSensorValue.unit).Float64("value", currentSensorValue.value).Msg("unknown current unit")
						}

						if voltageSensorValue.unit == "mV" {
							voltageSensorValue.value = voltageSensorValue.value / 1000
						} else if voltageSensorValue.unit == "V" {
							// do nothing
						} else {
							my.Logger.Warn().Str("unit", voltageSensorValue.unit).Float64("value", voltageSensorValue.value).Msg("unknown voltage unit")
						}

						p := currentSensorValue.value * voltageSensorValue.value

						if !strings.EqualFold(voltageSensorValue.name, "in") && !strings.EqualFold(currentSensorValue.name, "in") {
							p = p / 0.93 //If the sensor names to do NOT contain "IN" or "in", then we need to adjust the power to account for loss in the power supply. We will use 0.93 as the power supply efficiency factor for all systems.
						}

						sumPower += p
					}
				} else {
					my.Logger.Warn().Int("current size", len(v.currentSensor)).Int("voltage size", len(v.voltageSensor)).Msg("current and voltage sensor are ignored")
				}

				err = m.SetValueFloat64(instance, sumPower)
				if err != nil {
					my.Logger.Error().Float64("power", sumPower).Err(err).Msg("Unable to set power")
				} else {
					m.SetLabel("unit", "W")
				}

			case "ambient_temperature":
				if len(v.ambientTemperature) > 0 {
					aT := util.Avg(v.ambientTemperature)
					err = m.SetValueFloat64(instance, aT)
					if err != nil {
						my.Logger.Error().Float64("ambient_temperature", aT).Err(err).Msg("Unable to set ambient_temperature")
					} else {
						m.SetLabel("unit", "C")
					}
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
	return []*matrix.Matrix{my.data}, nil
}
