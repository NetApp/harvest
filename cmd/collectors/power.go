package collectors

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/num"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	zapiValueKey = "environment-sensors-info.threshold-sensor-value"
	restValueKey = "value"
)

// CollectChassisFRU is here because both ZAPI and REST sensor.go plugin call it to collect
// `system chassis fru show`.
// Chassis FRU information is only available via private CLI
func collectChassisFRU(client *rest.Client, logger *slog.Logger) (map[string]int, error) {
	fields := []string{"fru-name", "type", "status", "connected-nodes", "num-nodes"}
	query := "api/private/cli/system/chassis/fru"
	filter := []string{"type=psu"}
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		Filter(filter).
		MaxRecords(DefaultBatchSize).
		Build()

	result, err := rest.FetchAll(client, href)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data href=%s err=%w", href, err)
	}

	// map of PSUs node -> numNode
	nodeToNumNode := make(map[string]int)

	for _, r := range result {
		cn := r.Get("connected_nodes")
		if !cn.Exists() {
			logger.Warn(
				"fru has no connected nodes",
				slog.String("cluster", client.Remote().Name),
				slog.String("fru", r.Get("fru_name").ClonedString()),
			)
			continue
		}
		numNodes := int(r.Get("num_nodes").Int())
		for _, e := range cn.Array() {
			nodeToNumNode[e.ClonedString()] = numNodes
		}
	}
	return nodeToNumNode, nil
}

type sensorValue struct {
	node  string
	name  string
	value float64
	unit  string
}

type environmentMetric struct {
	key                   string
	ambientTemperature    []float64
	nonAmbientTemperature []float64
	fanSpeed              []float64
	powerSensor           map[string]*sensorValue
	voltageSensor         map[string]*sensorValue
	currentSensor         map[string]*sensorValue
}

var ambientRegex = regexp.MustCompile(`^(Ambient Temp|Ambient Temp \d|PSU\d AmbTemp|PSU\d Inlet|PSU\d Inlet Temp|In Flow Temp|Front Temp|Bat_Ambient \d|Riser Inlet Temp)$`)

var powerInRegex = regexp.MustCompile(`^PSU\d (InPwr Monitor|InPower|PIN|Power In|In Pwr)$`)

var voltageRegex = regexp.MustCompile(`^PSU\d (\d+V|InVoltage|VIN|AC In Volt|In Volt)$`)

var currentRegex = regexp.MustCompile(`^PSU\d (\d+V Curr|Curr|InCurrent|Curr IIN|AC In Curr|In Curr)$`)

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

func calculateEnvironmentMetrics(data *matrix.Matrix, logger *slog.Logger, valueKey string, myData *matrix.Matrix, nodeToNumNode map[string]int) []*matrix.Matrix {
	sensorEnvironmentMetricMap := make(map[string]*environmentMetric)
	excludedSensors := make(map[string][]sensorValue)

	for k, instance := range data.GetInstances() {
		if !instance.IsExportable() {
			continue
		}
		iKey := instance.GetLabel("node")
		if iKey == "" {
			logger.Warn("missing node label for instance", slog.String("key", k))
			continue
		}
		sensorName := instance.GetLabel("sensor")
		if sensorName == "" {
			logger.Warn("missing sensor name for instance", slog.String("key", k))
			continue
		}
		if _, ok := sensorEnvironmentMetricMap[iKey]; !ok {
			sensorEnvironmentMetricMap[iKey] = &environmentMetric{key: iKey, ambientTemperature: []float64{}, nonAmbientTemperature: []float64{}, fanSpeed: []float64{}}
		}
		for mKey, metric := range data.GetMetrics() {
			if mKey != valueKey {
				continue
			}
			sensorType := instance.GetLabel("type")
			sensorUnit := instance.GetLabel("unit")

			isAmbientMatch := ambientRegex.MatchString(sensorName)
			isPowerMatch := powerInRegex.MatchString(sensorName)
			isVoltageMatch := voltageRegex.MatchString(sensorName)
			isCurrentMatch := currentRegex.MatchString(sensorName)

			if sensorType == "thermal" && isAmbientMatch {
				if value, ok := metric.GetValueFloat64(instance); ok {
					sensorEnvironmentMetricMap[iKey].ambientTemperature = append(sensorEnvironmentMetricMap[iKey].ambientTemperature, value)
				}
			}

			if sensorType == "thermal" && !isAmbientMatch {
				// Exclude temperature sensors that contains sensor name `Margin` and value < 0
				value, ok := metric.GetValueFloat64(instance)
				if value > 0 && !strings.Contains(sensorName, "Margin") {
					if ok {
						sensorEnvironmentMetricMap[iKey].nonAmbientTemperature = append(sensorEnvironmentMetricMap[iKey].nonAmbientTemperature, value)
					}
				} else {
					excludedSensors[iKey] = append(excludedSensors[iKey], sensorValue{
						node:  iKey,
						name:  sensorName,
						value: value,
					})
				}
			}

			if sensorType == "fan" {
				if value, ok := metric.GetValueFloat64(instance); ok {
					sensorEnvironmentMetricMap[iKey].fanSpeed = append(sensorEnvironmentMetricMap[iKey].fanSpeed, value)
				}
			}

			if isPowerMatch {
				if value, ok := metric.GetValueFloat64(instance); ok {
					if !IsValidUnit(sensorUnit) {
						logger.Warn("unknown power unit", slog.String("unit", sensorUnit), slog.Float64("value", value))
					} else {
						if sensorEnvironmentMetricMap[iKey].powerSensor == nil {
							sensorEnvironmentMetricMap[iKey].powerSensor = make(map[string]*sensorValue)
						}
						sensorEnvironmentMetricMap[iKey].powerSensor[k] = &sensorValue{
							node:  iKey,
							name:  sensorName,
							value: value,
							unit:  sensorUnit,
						}
					}
				}
			}

			if isVoltageMatch {
				if value, ok := metric.GetValueFloat64(instance); ok {
					if sensorEnvironmentMetricMap[iKey].voltageSensor == nil {
						sensorEnvironmentMetricMap[iKey].voltageSensor = make(map[string]*sensorValue)
					}
					sensorEnvironmentMetricMap[iKey].voltageSensor[k] = &sensorValue{
						node:  iKey,
						name:  sensorName,
						value: value,
						unit:  sensorUnit,
					}
				}
			}

			if isCurrentMatch {
				if value, ok := metric.GetValueFloat64(instance); ok {
					if sensorEnvironmentMetricMap[iKey].currentSensor == nil {
						sensorEnvironmentMetricMap[iKey].currentSensor = make(map[string]*sensorValue)
					}
					sensorEnvironmentMetricMap[iKey].currentSensor[k] = &sensorValue{
						node:  iKey,
						name:  sensorName,
						value: value,
						unit:  sensorUnit,
					}
				}
			}
		}
	}

	if len(excludedSensors) > 0 {
		var excludedSensorStr strings.Builder
		for k, v := range excludedSensors {
			excludedSensorStr.WriteString(" node:" + k + " sensor:" + fmt.Sprintf("%v", v))
		}
		logger.Info("sensor excluded", slog.String("sensor", excludedSensorStr.String()))
	}

	whrSensors := make(map[string]*sensorValue)

	for key, v := range sensorEnvironmentMetricMap {
		instance, err2 := myData.NewInstance(key)
		if err2 != nil {
			logger.Warn("instance not found", slog.String("key", key))
			continue
		}
		// set node label
		instance.SetLabel("node", key)
		for _, k := range eMetrics {
			m := myData.GetMetric(k)
			switch k {
			case "power":
				var sumPower float64
				switch {
				case len(v.powerSensor) > 0:
					for _, v1 := range v.powerSensor {
						switch v1.unit {
						case "mW", "mW*hr":
							sumPower += v1.value / 1000
						case "W", "W*hr":
							sumPower += v1.value
						default:
							logger.Warn(
								"unknown power unit",
								slog.String("node", key),
								slog.String("name", v1.name),
								slog.String("unit", v1.unit),
								slog.Float64("value", v1.value),
							)
						}
						if v1.unit == "mW*hr" || v1.unit == "W*hr" {
							whrSensors[v1.name] = v1
						}
					}
				case len(v.voltageSensor) > 0 && len(v.voltageSensor) == len(v.currentSensor):
					voltageKeys := make([]string, 0, len(v.voltageSensor))
					for k := range v.voltageSensor {
						voltageKeys = append(voltageKeys, k)
					}
					sort.Strings(voltageKeys)
					currentKeys := make([]string, 0, len(v.currentSensor))
					for k := range v.currentSensor {
						currentKeys = append(currentKeys, k)
					}
					sort.Strings(currentKeys)
					for i := range currentKeys {
						currentKey := currentKeys[i]
						voltageKey := voltageKeys[i]

						// get values
						currentSensorValue := v.currentSensor[currentKey]
						voltageSensorValue := v.voltageSensor[voltageKey]

						// convert units
						if currentSensorValue.unit == "mA" {
							currentSensorValue.value /= 1000
						} else if currentSensorValue.unit != "A" {
							logger.Warn(
								"unknown current unit",
								slog.String("node", key),
								slog.String("unit", currentSensorValue.unit),
								slog.Float64("value", currentSensorValue.value),
							)
						}

						if voltageSensorValue.unit == "mV" {
							voltageSensorValue.value /= 1000
						} else if voltageSensorValue.unit != "V" {
							logger.Warn(
								"unknown voltage unit",
								slog.String("node", key),
								slog.String("unit", voltageSensorValue.unit),
								slog.Float64("value", voltageSensorValue.value),
							)
						}

						p := currentSensorValue.value * voltageSensorValue.value

						if !strings.EqualFold(voltageSensorValue.name, "in") && !strings.EqualFold(currentSensorValue.name, "in") {
							p /= 0.93 // If the sensor names to do NOT contain "IN" or "in", then we need to adjust the power to account for loss in the power supply. We will use 0.93 as the power supply efficiency factor for all systems.
						}

						sumPower += p
					}
				default:
					logger.Warn(
						"current and voltage sensor are ignored",
						slog.String("node", key),
						slog.Int("current size", len(v.currentSensor)),
						slog.Int("voltage size", len(v.voltageSensor)),
					)
				}

				numNode, ok := nodeToNumNode[key]
				if !ok {
					logger.Warn("node not found in nodeToNumNode map", slog.String("node", key))
					numNode = 1
				}
				sumPower /= float64(numNode)
				m.SetValueFloat64(instance, sumPower)
			case "average_ambient_temperature":
				if len(v.ambientTemperature) > 0 {
					aaT := num.Avg(v.ambientTemperature)
					m.SetValueFloat64(instance, aaT)
				}
			case "min_ambient_temperature":
				maT := num.Min(v.ambientTemperature)
				m.SetValueFloat64(instance, maT)
			case "max_temperature":
				mT := num.Max(v.nonAmbientTemperature)
				m.SetValueFloat64(instance, mT)
			case "average_temperature":
				if len(v.nonAmbientTemperature) > 0 {
					nat := num.Avg(v.nonAmbientTemperature)
					m.SetValueFloat64(instance, nat)
				}
			case "min_temperature":
				mT := num.Min(v.nonAmbientTemperature)
				m.SetValueFloat64(instance, mT)
			case "average_fan_speed":
				if len(v.fanSpeed) > 0 {
					afs := num.Avg(v.fanSpeed)
					m.SetValueFloat64(instance, afs)
				}
			case "max_fan_speed":
				mfs := num.Max(v.fanSpeed)
				m.SetValueFloat64(instance, mfs)
			case "min_fan_speed":
				mfs := num.Min(v.fanSpeed)
				m.SetValueFloat64(instance, mfs)
			}
		}
	}

	if len(whrSensors) > 0 {
		var whrSensor strings.Builder
		for _, v := range whrSensors {
			whrSensor.WriteString(" sensor:" + fmt.Sprintf("%v", *v))
		}
		logger.Info("sensor with *hr units", slog.String("sensor", whrSensor.String()))
	}

	return []*matrix.Matrix{myData}
}

func NewSensor(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Sensor{AbstractPlugin: p}
}

type Sensor struct {
	*plugin.AbstractPlugin
	data           *matrix.Matrix
	client         *rest.Client
	instanceKeys   map[string]string
	instanceLabels map[string]map[string]string
	hasREST        bool
}

func (s *Sensor) Init(remote conf.Remote) error {

	var err error
	if err := s.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if s.client, err = rest.New(conf.ZapiPoller(s.ParentParams), timeout, s.Auth); err != nil {
		s.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	s.hasREST = true

	if err := s.client.Init(5, remote); err != nil {
		if re, ok := errors.AsType[*errs.RestError](err); ok && re.StatusCode == http.StatusNotFound {
			s.SLogger.Warn("Cluster does not support REST. Power plugin disabled")
			s.hasREST = false
			return nil
		}
		return err
	}

	s.data = matrix.New(s.Parent+".Sensor", "environment_sensor", "environment_sensor")
	s.instanceKeys = make(map[string]string)
	s.instanceLabels = make(map[string]map[string]string)

	// init environment metrics in plugin matrix
	// create environment metric if not exists
	for _, k := range eMetrics {
		err := matrix.CreateMetric(k, s.data)
		if err != nil {
			s.SLogger.Warn("error while creating metric", slogx.Err(err), slog.String("key", k))
		}
	}
	return nil
}

func (s *Sensor) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	if !s.hasREST {
		return nil, nil, nil
	}
	data := dataMap[s.Object]
	// Purge and reset data
	s.data.PurgeInstances()
	s.data.Reset()
	s.client.Metadata.Reset()

	// Set all global labels if they don't already exist
	s.data.SetGlobalLabels(data.GetGlobalLabels())

	// Collect chassis fru show, so we can determine if a controller's PSUs are shared or not
	nodeToNumNode, err := collectChassisFRU(s.client, s.SLogger)
	if err != nil {
		return nil, nil, err
	}
	if len(nodeToNumNode) == 0 {
		s.SLogger.Debug("No chassis field replaceable units found")
	}

	valueKey := zapiValueKey
	if s.Parent == "Rest" {
		valueKey = restValueKey
	}
	metrics := calculateEnvironmentMetrics(data, s.SLogger, valueKey, s.data, nodeToNumNode)

	return metrics, s.client.Metadata, nil
}
