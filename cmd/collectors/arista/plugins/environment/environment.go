package environment

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
)

const (
	ambientTemp   = "ambient_temp"
	fanSpeed      = "fan_speed"
	fanUp         = "fan_up"
	powerCapacity = "power_capacity"
	powerIn       = "power_in"
	powerOut      = "power_out"
	powerUp       = "power_up"
	sensorTemp    = "sensor_temp"
)

var metrics = []string{
	ambientTemp,
	fanSpeed,
	fanUp,
	powerCapacity,
	powerIn,
	powerOut,
	powerUp,
	sensorTemp,
}

type Environment struct {
	*plugin.AbstractPlugin
	client         *rest.Client
	matrix         *matrix.Matrix
	templateObject string // object name from the template
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Environment{AbstractPlugin: p}
}

func (e *Environment) Init(remote conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = e.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	if client, err = rest.New(conf.ZapiPoller(e.ParentParams), e.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	if err := client.Init(2, remote); err != nil {
		return err
	}

	e.client = client
	e.templateObject = e.ParentParams.GetChildContentS("object")

	e.matrix = matrix.New(e.Parent+e.templateObject, e.templateObject, e.templateObject)
	if err := e.matrix.NewMetricsFloat64(metrics...); err != nil {
		return fmt.Errorf("error while initializing matrix: %w", err)
	}

	return nil
}

func (e *Environment) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[e.Object]
	e.client.Metadata.Reset()

	e.matrix.PurgeInstances()
	e.matrix.Reset()

	// Set all global labels if they don't already exist
	e.matrix.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := e.ParentParams.GetChildContentS("query")
	output, err := e.client.RunCmds(rest.SplitCommands(command)...)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	temperatureOutput := output.Get("0")
	powerOutput := output.Get("1")
	coolingOutput := output.Get("2")

	e.parseTemperature(temperatureOutput, e.matrix)
	e.parsePower(powerOutput, e.matrix)
	e.parseCooling(coolingOutput, e.matrix)

	e.client.Metadata.NumCalls.Store(1)
	e.client.Metadata.BytesRx.Store(uint64(len(output.Raw)))
	e.client.Metadata.PluginInstances.Store(uint64(len(e.matrix.GetInstances())))

	return []*matrix.Matrix{e.matrix}, e.client.Metadata, nil
}

func (e *Environment) parseTemperature(output gjson.Result, envMat *matrix.Matrix) {
	// Temperature sensors can appear at the top level and nested inside card and
	// power supply slots.
	e.addTempSensors(output.Get("tempSensors"), envMat)

	output.Get("cardSlots").ForEach(func(_, slot gjson.Result) bool {
		e.addTempSensors(slot.Get("tempSensors"), envMat)
		return true
	})
	output.Get("powerSupplySlots").ForEach(func(_, slot gjson.Result) bool {
		e.addTempSensors(slot.Get("tempSensors"), envMat)
		return true
	})
}

func (e *Environment) addTempSensors(sensors gjson.Result, envMat *matrix.Matrix) {
	if !sensors.Exists() {
		return
	}
	metric := envMat.MustGetMetric(sensorTemp)
	sensors.ForEach(func(_, sensor gjson.Result) bool {
		name := sensor.Get("name").ClonedString()
		if name == "" {
			return true
		}
		instanceKey := "temp_" + name
		instance, err := envMat.NewInstance(instanceKey)
		if err != nil {
			e.SLogger.Warn("Failed to create sensor instance", slog.String("key", instanceKey))
			return true
		}
		instance.SetLabel("sensor", name)
		instance.SetLabel("description", sensor.Get("description").ClonedString())
		instance.SetLabel("status", sensor.Get("hwStatus").ClonedString())
		metric.SetValueFloat64(instance, sensor.Get("currentTemperature").Float())
		return true
	})
}

func (e *Environment) parsePower(output gjson.Result, envMat *matrix.Matrix) {
	supplies := output.Get("powerSupplies")
	if !supplies.Exists() {
		return
	}
	powerCapacityMetric := envMat.MustGetMetric(powerCapacity)
	powerInMetric := envMat.MustGetMetric(powerIn)
	powerOutMetric := envMat.MustGetMetric(powerOut)
	powerUpMetric := envMat.MustGetMetric(powerUp)
	supplies.ForEach(func(key, supply gjson.Result) bool {
		psID := key.ClonedString()
		instanceKey := "power_" + psID
		instance, err := envMat.NewInstance(instanceKey)
		if err != nil {
			e.SLogger.Warn("Failed to create power supply instance", slog.String("key", instanceKey))
			return true
		}
		state := supply.Get("state").ClonedString()
		instance.SetLabel("power_supply", psID)
		instance.SetLabel("model", supply.Get("modelName").ClonedString())
		instance.SetLabel("status", state)

		powerCapacityMetric.SetValueFloat64(instance, supply.Get("capacity").Float())

		if inputPower := supply.Get("inputPower"); inputPower.Exists() {
			powerInMetric.SetValueFloat64(instance, inputPower.Float())
		}
		if outputPower := supply.Get("outputPower"); outputPower.Exists() {
			powerOutMetric.SetValueFloat64(instance, outputPower.Float())
		}

		if state == "ok" {
			powerUpMetric.SetValueFloat64(instance, 1)
		} else {
			powerUpMetric.SetValueFloat64(instance, 0)
		}
		return true
	})
}

func (e *Environment) parseCooling(output gjson.Result, envMat *matrix.Matrix) {
	if ambient := output.Get("ambientTemperature"); ambient.Exists() {
		instance, err := envMat.NewInstance("cooling_ambient")
		if err != nil {
			e.SLogger.Error("Failed to create ambient temperature instance", slog.String("key", "cooling_ambient"))
			return
		}
		instance.SetLabel("sensor", "ambient")
		instance.SetLabel("status", output.Get("systemStatus").ClonedString())
		envMat.MustSetValueFloat64(ambientTemp, instance, ambient.Float())
	}

	fanSpeedMetric := envMat.MustGetMetric(fanSpeed)
	fanUpMetric := envMat.MustGetMetric(fanUp)
	output.Get("fanTraySlots").ForEach(func(_, tray gjson.Result) bool {
		trayLabel := tray.Get("label").ClonedString()
		tray.Get("fans").ForEach(func(_, fan gjson.Result) bool {
			fanLabel := fan.Get("label").ClonedString()
			if fanLabel == "" {
				return true
			}
			instanceKey := "fan_" + fanLabel
			instance, err := envMat.NewInstance(instanceKey)
			if err != nil {
				e.SLogger.Warn("Failed to create fan instance", slog.String("key", instanceKey))
				return true
			}
			status := fan.Get("status").ClonedString()
			instance.SetLabel("fan", fanLabel)
			instance.SetLabel("fan_tray", trayLabel)
			instance.SetLabel("status", status)

			fanSpeedMetric.SetValueFloat64(instance, fan.Get("actualSpeed").Float())
			if status == "ok" {
				fanUpMetric.SetValueFloat64(instance, 1)
			} else {
				fanUpMetric.SetValueFloat64(instance, 0)
			}
			return true
		})
		return true
	})
}
