/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package sensor

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/dict"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"regexp"
	"sort"
	"strings"
)

type Sensor struct {
	*plugin.AbstractPlugin
	data           *matrix.Matrix
	instanceKeys   map[string]string
	instanceLabels map[string]*dict.Dict
}

type sensorEnvironmentMetric struct {
	key                   string
	ambientTemperature    []float64
	nonAmbientTemperature []float64
	fanSpeed              []float64
	powerSensor           map[string]*sensorValue
	voltageSensor         map[string]*sensorValue
	currentSensor         map[string]*sensorValue
}

type sensorValue struct {
	name  string
	value float64
	unit  string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Sensor{AbstractPlugin: p}
}

var ambientRegex = regexp.MustCompile(`^(Ambient Temp|Ambient Temp \d|PSU\d AmbTemp|PSU\d Inlet|PSU\d Inlet Temp|In Flow Temp|Front Temp|System Inlet|Bat Ambient \d|Riser Inlet Temp)$`)
var powerInRegex = regexp.MustCompile(`^PSU\d (InPwr Monitor|InPower|PIN|Power In)$`)
var voltageRegex = regexp.MustCompile(`^PSU\d (\d+V|InVoltage|VIN|AC In Volt)$`)
var currentRegex = regexp.MustCompile(`^PSU\d (\d+V Curr|Curr|InCurrent|Curr IIN|AC In Curr)$`)
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
	sensorEnvironmentMetricMap := make(map[string]*sensorEnvironmentMetric)

	for k, instance := range data.GetInstances() {
		iKey := instance.GetLabel("node")
		if iKey == "" {
			my.Logger.Warn().Str("key", k).Msg("missing node label for instance")
			continue
		}
		_, iKey2, found := strings.Cut(k, iKey+".")
		if !found {
			my.Logger.Warn().Str("key", iKey+".").Msg("missing instance key")
			continue
		}
		if _, ok := sensorEnvironmentMetricMap[iKey]; !ok {
			sensorEnvironmentMetricMap[iKey] = &sensorEnvironmentMetric{key: iKey, ambientTemperature: []float64{}, nonAmbientTemperature: []float64{}, fanSpeed: []float64{}}
		}
		for mKey, metric := range data.GetMetrics() {
			if mKey == "environment-sensors-info.threshold-sensor-value" {
				sensorType := instance.GetLabel("type")
				sensorName := instance.GetLabel("sensor")
				sensorUnit := instance.GetLabel("unit")
				warningLowThr := instance.GetLabel("warning_low")
				criticalLowThr, _, _ := data.GetMetric("environment-sensors-info.critical-low-threshold").GetValueFloat64(instance)
				isAmbientMatch := ambientRegex.MatchString(sensorName)
				isPowerMatch := powerInRegex.MatchString(sensorName)
				isVoltageMatch := voltageRegex.MatchString(sensorName)
				isCurrentMatch := currentRegex.MatchString(sensorName)

				my.Logger.Debug().Bool("isAmbientMatch", isAmbientMatch).
					Bool("isPowerMatch", isPowerMatch).
					Bool("isVoltageMatch", isVoltageMatch).
					Bool("isCurrentMatch", isCurrentMatch).
					Str("warningLowThreshold", warningLowThr).
					Float64("criticalLowThreshold", criticalLowThr).
					Str("sensorType", sensorType).
					Str("sensorUnit", sensorUnit).
					Str("sensorName", sensorName).
					Msg("")

				if sensorType == "thermal" && isAmbientMatch {
					if value, ok, _ := metric.GetValueFloat64(instance); ok {
						sensorEnvironmentMetricMap[iKey].ambientTemperature = append(sensorEnvironmentMetricMap[iKey].ambientTemperature, value)
					}
				}

				if sensorType == "thermal" && !isAmbientMatch {
					// Exclude temperature sensors that have crit_low=0 and warn_low is missing
					if !(criticalLowThr == 0.0 && warningLowThr == "") {
						if value, ok, _ := metric.GetValueFloat64(instance); ok {
							sensorEnvironmentMetricMap[iKey].nonAmbientTemperature = append(sensorEnvironmentMetricMap[iKey].nonAmbientTemperature, value)
						}
					} else {
						my.Logger.Debug().Str("warningLowThreshold", warningLowThr).
							Float64("criticalLowThreshold", criticalLowThr).
							Str("sensorName", sensorName).
							Msg("sensor excluded")
					}
				}

				if sensorType == "fan" {
					if value, ok, _ := metric.GetValueFloat64(instance); ok {
						sensorEnvironmentMetricMap[iKey].fanSpeed = append(sensorEnvironmentMetricMap[iKey].fanSpeed, value)
					}
				}

				if isPowerMatch {
					if value, ok, _ := metric.GetValueFloat64(instance); ok {
						if sensorEnvironmentMetricMap[iKey].powerSensor == nil {
							sensorEnvironmentMetricMap[iKey].powerSensor = make(map[string]*sensorValue)
						}
						sensorEnvironmentMetricMap[iKey].powerSensor[iKey2] = &sensorValue{name: iKey2, value: value, unit: sensorUnit}
					}
				}

				if isVoltageMatch {
					if value, ok, _ := metric.GetValueFloat64(instance); ok {
						if sensorEnvironmentMetricMap[iKey].voltageSensor == nil {
							sensorEnvironmentMetricMap[iKey].voltageSensor = make(map[string]*sensorValue)
						}
						sensorEnvironmentMetricMap[iKey].voltageSensor[iKey2] = &sensorValue{name: iKey2, value: value, unit: sensorUnit}
					}
				}

				if isCurrentMatch {
					if value, ok, _ := metric.GetValueFloat64(instance); ok {
						if sensorEnvironmentMetricMap[iKey].currentSensor == nil {
							sensorEnvironmentMetricMap[iKey].currentSensor = make(map[string]*sensorValue)
						}
						sensorEnvironmentMetricMap[iKey].currentSensor[iKey2] = &sensorValue{name: iKey2, value: value, unit: sensorUnit}
					}
				}
			}
		}
	}

	for key, v := range sensorEnvironmentMetricMap {
		instance, err := my.data.NewInstance(key)
		if err != nil {
			my.Logger.Warn().Str("key", key).Msg("instance not found")
			continue
		}
		// set node label
		instance.SetLabel("node", key)
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
							my.Logger.Warn().Str("node", key).Str("unit", v1.unit).Float64("value", v1.value).Msg("unknown power unit")
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
							my.Logger.Warn().Str("node", key).Str("unit", currentSensorValue.unit).Float64("value", currentSensorValue.value).Msg("unknown current unit")
						}

						if voltageSensorValue.unit == "mV" {
							voltageSensorValue.value = voltageSensorValue.value / 1000
						} else if voltageSensorValue.unit == "V" {
							// do nothing
						} else {
							my.Logger.Warn().Str("node", key).Str("unit", voltageSensorValue.unit).Float64("value", voltageSensorValue.value).Msg("unknown voltage unit")
						}

						p := currentSensorValue.value * voltageSensorValue.value

						if !strings.EqualFold(voltageSensorValue.name, "in") && !strings.EqualFold(currentSensorValue.name, "in") {
							p = p / 0.93 //If the sensor names to do NOT contain "IN" or "in", then we need to adjust the power to account for loss in the power supply. We will use 0.93 as the power supply efficiency factor for all systems.
						}

						sumPower += p
					}
				} else {
					my.Logger.Warn().Str("node", key).Int("current size", len(v.currentSensor)).Int("voltage size", len(v.voltageSensor)).Msg("current and voltage sensor are ignored")
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
	return []*matrix.Matrix{my.data}, nil
}
