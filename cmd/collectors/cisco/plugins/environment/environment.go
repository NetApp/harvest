package environment

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

var metrics = []string{
	"power_capacity",
	"power_in",
	"power_out",
	"power_up",
	"sensor_temp",
}

type Environment struct {
	*plugin.AbstractPlugin
	client *rest.Client
	matrix *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Environment{AbstractPlugin: p}
}

func (e *Environment) Init(_ conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = e.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)

	if client, err = rest.New(conf.ZapiPoller(e.ParentParams), timeout, e.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	e.client = client

	e.matrix = matrix.New(e.Parent+".Environment", "cisco_environment", "cisco_environment")

	return nil
}

func (e *Environment) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[e.Object]
	e.client.Metadata.Reset()

	envMat, err := e.initMatrix("cisco_environment")
	if err != nil {
		return nil, nil, fmt.Errorf("error while initializing matrix: %w", err)
	}

	// Set all global labels if they don't already exist
	envMat.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := e.ParentParams.GetChildContentS("query")
	output, err := e.client.CallAPI(command)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	e.parseEnvironment(output, envMat)

	e.client.Metadata.NumCalls = 1
	e.client.Metadata.BytesRx = uint64(len(output.Raw))
	e.client.Metadata.PluginInstances = uint64(len(envMat.GetInstances()))

	return []*matrix.Matrix{envMat}, e.client.Metadata, nil
}

func (e *Environment) initMatrix(name string) (*matrix.Matrix, error) {

	mat := matrix.New(e.Parent+name, name, name)

	for _, k := range metrics {
		if err := matrix.CreateMetric(k, mat); err != nil {
			return nil, fmt.Errorf("error while creating metric %s: %w", k, err)
		}
	}

	return mat, nil
}

func (e *Environment) parseEnvironment(output gjson.Result, envMat *matrix.Matrix) {
	e.parseTemperature(output, envMat)
	e.parsePower(output, envMat)
}

func (e *Environment) parseTemperature(output gjson.Result, envMat *matrix.Matrix) {

	rowQuery := "TABLE_tempinfo.ROW_tempinfo"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		e.SLogger.Warn("Unable to parse temperature because rows are missing", slog.String("query", rowQuery))
		return
	}

	rows.ForEach(func(_, value gjson.Result) bool {
		sensorName := value.Get("sensor").ClonedString()
		sensorName = strings.ReplaceAll(sensorName, " ", "")
		curTemp := value.Get("curtemp").ClonedString()

		instanceKey := sensorName

		instance, err := envMat.NewInstance(instanceKey)
		if err != nil {
			e.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
			return true
		}

		instance.SetLabel("sensor", sensorName)

		if err := envMat.GetMetric("sensor_temp").SetValueString(instance, curTemp); err != nil {
			e.SLogger.Error(
				"Unable to set value on metric",
				slogx.Err(err),
				slog.String("metric", "sensor_temp"),
			)
		}

		return true
	})
}

func (e *Environment) parsePower(output gjson.Result, envMat *matrix.Matrix) {
	model := NewPowerModel(output, e.SLogger)

	for _, ps := range model.PowerSupplies {
		instanceKey := ps.Num + "_" + ps.Model

		instance, err := envMat.NewInstance(instanceKey)

		if err != nil {
			e.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
			continue
		}

		instance.SetLabel("status", ps.Status)
		instance.SetLabel("ps", ps.Num)
		instance.SetLabel("model", ps.Model)

		powerUpMetric := envMat.GetMetric("power_up")
		if ps.Status == "ok" {
			_ = powerUpMetric.SetValueFloat64(instance, 1)
		} else {
			_ = powerUpMetric.SetValueFloat64(instance, 0)
		}

		// If actualOut is 0, we don't want to export the other metrics
		if ps.ActualOut > 0 {
			_ = envMat.GetMetric("power_out").SetValueFloat64(instance, ps.ActualOut)
			_ = envMat.GetMetric("power_capacity").SetValueFloat64(instance, ps.TotalCapacity)
			_ = envMat.GetMetric("power_in").SetValueFloat64(instance, ps.ActualIn)
		}
	}
}

// PowerModel represents the power metrics of the device and is needed
// to normalize the differences we see between Cisco 3000 and 9000 switches.
// Cisco 3K switches include total power but do not include power for individual power supplies.
// Cisco 9K switches include both total power and individual power supplies.
// Each family of switches also uses different key names for the same data.
type PowerModel struct {
	PowerSupplies  []PowerSupply
	TotalPowerDraw float64
}

type PowerSupply struct {
	Num           string
	Model         string
	Status        string
	ActualOut     float64
	ActualIn      float64
	TotalCapacity float64
}

func NewPowerModel(output gjson.Result, logger *slog.Logger) PowerModel {
	var powerModel PowerModel

	// Check if the output is from a 3000 or 9000 switch
	is3K := output.Get("powersup.power_summary.ps_redun_mode_3k").Exists()
	if is3K {
		powerModel = newPowerModel3K(output, logger)
	} else {
		powerModel = newPowerModel9K(output, logger)
	}

	return powerModel
}

func newPowerModel9K(output gjson.Result, logger *slog.Logger) PowerModel {

	var powerSupplies []PowerSupply
	powerModel := PowerModel{}

	rowQuery := "powersup.TABLE_psinfo.ROW_psinfo"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		logger.Warn("Unable to parse power because rows are missing", slog.String("query", rowQuery))
		return powerModel
	}

	rows.ForEach(func(_, value gjson.Result) bool {
		psNum := value.Get("psnum").ClonedString()
		psModel := value.Get("psmodel").ClonedString()
		psStatus := strings.ToLower(value.Get("ps_status").ClonedString())
		actualOut := value.Get("actual_out").String()
		actualIn := value.Get("actual_input").String()
		totalCapacity := value.Get("tot_capa").String()

		ps := PowerSupply{
			Num:           psNum,
			Model:         psModel,
			Status:        psStatus,
			ActualIn:      wattsToFloat(actualIn, logger),
			ActualOut:     wattsToFloat(actualOut, logger),
			TotalCapacity: wattsToFloat(totalCapacity, logger),
		}

		powerSupplies = append(powerSupplies, ps)
		return true
	})

	wattsRequested := output.Get("powersup.power_summary.tot_pow_out_actual_draw").String()
	powerModel.TotalPowerDraw = wattsToFloat(wattsRequested, logger)

	powerModel.PowerSupplies = powerSupplies

	return powerModel
}

func wattsToFloat(out string, logger *slog.Logger) float64 {
	trimmed := strings.ReplaceAll(out, " W", "")
	if trimmed != "" {
		actualPowerFloat, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			logger.Warn("Unable to parse actualOut", slogx.Err(err), slog.String("trimmed", trimmed))
		}

		return actualPowerFloat
	}

	return 0
}

func newPowerModel3K(output gjson.Result, logger *slog.Logger) PowerModel {

	var powerSupplies []PowerSupply
	powerModel := PowerModel{}

	rowQuery := "powersup.TABLE_psinfo.ROW_psinfo"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		logger.Warn("Unable to parse power because rows are missing", slog.String("query", rowQuery))
		return powerModel
	}

	rows.ForEach(func(_, value gjson.Result) bool {
		psNum := value.Get("psnum").ClonedString()
		psModel := value.Get("psmodel").ClonedString()
		psStatus := value.Get("ps_status_3k").ClonedString()
		actualIn := value.Get("watts").String()

		ps := PowerSupply{
			Num:      psNum,
			Model:    psModel,
			Status:   psStatus,
			ActualIn: wattsToFloat(actualIn, logger),
		}

		powerSupplies = append(powerSupplies, ps)
		return true
	})

	wattsRequested := output.Get("powersup.TABLE_mod_pow_info.ROW_mod_pow_info.watts_requested").String()
	powerModel.TotalPowerDraw = wattsToFloat(wattsRequested, logger)

	powerModel.PowerSupplies = powerSupplies

	return powerModel
}
