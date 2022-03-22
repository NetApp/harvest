/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package sensor

import (
	"fmt"
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

var ambientRegex = regexp.MustCompile(`^(Ambient Temp|Ambient Temp \d|PSU\d AmbTemp|PSU\d Inlet|PSU\d Inlet Temp|In Flow Temp|Front Temp|System Inlet)$`)
var powerRegex = regexp.MustCompile(`^PSU\d (InPwr Monitor|InPower|PIN|Power In)$`)
var voltageRegex = regexp.MustCompile(`^PSU\d (\d+V|InVoltage|VIN|AC In Volt)$`)
var currentRegex = regexp.MustCompile(`^PSU\d (\d+V Curr|Curr|InCurrent|Curr IIN|AC In Curr)$`)
var eMetrics = []string{"power", "ambient_temperature", "max_temperature", "average_temperature", "average_fan_speed", "max_fan_speed", "min_fan_speed"}

func (my *Sensor) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	my.data = matrix.New(my.Parent+".Sensor", "environment_sensor", "environment_sensor")
	my.instanceKeys = make(map[string]string)
	my.instanceLabels = make(map[string]*dict.Dict)

	for _, metricName := range eMetrics {
		metric, err := my.data.NewMetricFloat64(metricName)
		if err != nil {
			my.Logger.Error().Stack().Err(err).Msg("add metric")
			return err
		}

		metric.SetName(metricName)
		my.Logger.Debug().Msgf("added metric: (%s) [%s] %s", metricName, metricName, metric)
	}

	my.Logger.Debug().Msgf("added data with %d metrics", len(my.data.GetMetrics()))

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
	sensorEnvironmentMetricMap := make(map[string]*sensorEnvironmentMetric, 0)
	for k, instance := range data.GetInstances() {
		lastInd := strings.Index(k, ".")
		iKey := k[:lastInd]
		iKey2 := k[lastInd+1:]
		if _, ok := sensorEnvironmentMetricMap[iKey]; !ok {
			sensorEnvironmentMetricMap[iKey] = &sensorEnvironmentMetric{key: iKey, ambientTemperature: []float64{}, nonAmbientTemperature: []float64{}, fanSpeed: []float64{}}
		}
		for mkey, metric := range data.GetMetrics() {
			if mkey == "environment-sensors-info.threshold-sensor-value" {
				sensorType := instance.GetLabel("type")
				sensorName := instance.GetLabel("sensor")
				sensorUnit := instance.GetLabel("unit")
				isAmbientMatch := ambientRegex.MatchString(sensorName)
				isPowerMatch := powerRegex.MatchString(sensorName)
				isVoltageMatch := voltageRegex.MatchString(sensorName)
				isCurrentMatch := currentRegex.MatchString(sensorName)

				if sensorType == "thermal" && isAmbientMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						sensorEnvironmentMetricMap[iKey].ambientTemperature = append(sensorEnvironmentMetricMap[iKey].ambientTemperature, value)
					}
				}

				if sensorType == "thermal" && !isAmbientMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						sensorEnvironmentMetricMap[iKey].nonAmbientTemperature = append(sensorEnvironmentMetricMap[iKey].nonAmbientTemperature, value)
					}
				}

				if sensorType == "fan" {
					if value, ok := metric.GetValueFloat64(instance); ok {
						sensorEnvironmentMetricMap[iKey].fanSpeed = append(sensorEnvironmentMetricMap[iKey].fanSpeed, value)
					}
				}

				if isPowerMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						if sensorEnvironmentMetricMap[iKey].powerSensor == nil {
							sensorEnvironmentMetricMap[iKey].powerSensor = make(map[string]*sensorValue, 0)
						}
						sensorEnvironmentMetricMap[iKey].powerSensor[iKey2] = &sensorValue{name: iKey2, value: value, unit: sensorUnit}
					}
				}

				if isVoltageMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						if sensorEnvironmentMetricMap[iKey].voltageSensor == nil {
							sensorEnvironmentMetricMap[iKey].voltageSensor = make(map[string]*sensorValue, 0)
						}
						sensorEnvironmentMetricMap[iKey].voltageSensor[iKey2] = &sensorValue{name: iKey2, value: value, unit: sensorUnit}
					}
				}

				if isCurrentMatch {
					if value, ok := metric.GetValueFloat64(instance); ok {
						if sensorEnvironmentMetricMap[iKey].currentSensor == nil {
							sensorEnvironmentMetricMap[iKey].currentSensor = make(map[string]*sensorValue, 0)
						}
						sensorEnvironmentMetricMap[iKey].currentSensor[iKey2] = &sensorValue{name: iKey2, value: value, unit: sensorUnit}
					}
				}

			}
		}
	}

	for _, k := range eMetrics {
		my.createEnvironmentMetric(k)
	}
	for key, v := range sensorEnvironmentMetricMap {
		instance, err := my.data.NewInstance(key)
		instance.SetLabel("node", key)
		for _, k := range eMetrics {
			m := my.data.GetMetric(k)
			switch k {
			case "power":
				var sumPower float64
				if len(v.powerSensor) > 0 {
					for _, v1 := range v.powerSensor {
						if v1.unit == "mW" {
							sumPower += sumPower + (v1.value / 1000)
						} else if v1.unit == "W" {
							sumPower += sumPower + v1.value
						} else {
							my.Logger.Warn().Str("unit", v1.unit).Str("value", fmt.Sprintf("%f", v1.value)).Msg("unknown power unit")
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

					for i, _ := range currentKeys {
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
							my.Logger.Warn().Str("unit", currentSensorValue.unit).Str("value", fmt.Sprintf("%f", currentSensorValue.value)).Msg("unknown current unit")
						}

						if voltageSensorValue.unit == "mV" {
							voltageSensorValue.value = voltageSensorValue.value / 1000
						} else if voltageSensorValue.unit == "A" {
							// do nothing
						} else {
							my.Logger.Warn().Str("unit", voltageSensorValue.unit).Str("value", fmt.Sprintf("%f", voltageSensorValue.value)).Msg("unknown voltage unit")
						}

						p := currentSensorValue.value * voltageSensorValue.value

						if strings.Contains(voltageSensorValue.name, "in") || strings.Contains(voltageSensorValue.name, "IN") {
							if strings.Contains(currentSensorValue.name, "in") || strings.Contains(currentSensorValue.name, "IN") {
								p = p / 0.93
							}
						}
						sumPower += p
					}
				}
				// convert to KW
				sumPower = sumPower / 1000
				err = m.SetValueFloat64(instance, sumPower)
				if err != nil {
					my.Logger.Error().Stack().Err(err).Msg("error")
				}
				m.SetLabel("unit", "kW")

			case "ambient_temperature":
				if len(v.ambientTemperature) > 0 {
					err = m.SetValueFloat64(instance, util.SumNumbers(v.ambientTemperature)/float64(len(v.ambientTemperature)))
					if err != nil {
						my.Logger.Error().Stack().Err(err).Msg("error")
					}
				}
				m.SetLabel("unit", "C")
			case "max_temperature":
				err = m.SetValueFloat64(instance, util.Max(v.nonAmbientTemperature))
				if err != nil {
					my.Logger.Error().Stack().Err(err).Msg("error")
				}
				m.SetLabel("unit", "C")
			case "average_temperature":
				if len(v.nonAmbientTemperature) > 0 {
					err = m.SetValueFloat64(instance, util.SumNumbers(v.nonAmbientTemperature)/float64(len(v.nonAmbientTemperature)))
					if err != nil {
						my.Logger.Error().Stack().Err(err).Msg("error")
					}
				}
				m.SetLabel("unit", "C")
			case "average_fan_speed":
				if len(v.fanSpeed) > 0 {
					err = m.SetValueFloat64(instance, util.SumNumbers(v.fanSpeed)/float64(len(v.fanSpeed)))
					if err != nil {
						my.Logger.Error().Stack().Err(err).Msg("error")
					}
				}
				m.SetLabel("unit", "rpm")
			case "max_fan_speed":
				err = m.SetValueFloat64(instance, util.Max(v.fanSpeed))
				if err != nil {
					my.Logger.Error().Stack().Err(err).Msg("error")
				}
				m.SetLabel("unit", "rpm")
			case "min_fan_speed":
				err = m.SetValueFloat64(instance, util.Min(v.fanSpeed))
				if err != nil {
					my.Logger.Error().Stack().Err(err).Msg("error")
				}
				m.SetLabel("unit", "rpm")
			}
		}
	}
	return []*matrix.Matrix{my.data}, nil
}

func (my *Sensor) createEnvironmentMetric(key string) {
	var err error
	at := my.data.GetMetric(key)
	if at == nil {
		if at, err = my.data.NewMetricFloat64(key); err != nil {
			my.Logger.Error().Stack().Err(err).Msg("error")
		}
	}
}
