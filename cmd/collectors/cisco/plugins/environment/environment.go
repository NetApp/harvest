package environment

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"
)

var metrics = []string{
	"fan_up",
	"fan_speed",
	"fan_zone_speed",
	"power_capacity",
	"power_in",
	"power_mode",
	"power_out",
	"power_up",
	"sensor_temp",
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
	e.templateObject = e.ParentParams.GetChildContentS("object")

	e.matrix = matrix.New(e.Parent+".Environment", e.templateObject, e.templateObject)

	return nil
}

func (e *Environment) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
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
	output, err := e.client.CLIShow(command)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	envOutput := output.Get("output.0.body")
	fanDetailsOutput := output.Get("output.1.body")

	e.parseEnvironment(envOutput, envMat)
	e.parseFanDetails(fanDetailsOutput, envMat)

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

func (e *Environment) parseEnvironment(content gjson.Result, envMat *matrix.Matrix) {
	e.parseTemperature(content, envMat)
	e.parsePower(content, envMat)
	e.parseFanZoneSpeed(content, envMat)
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

var hexReplacer = strings.NewReplacer("0x", "", "0X", "")

func (e *Environment) parseFanZoneSpeed(output gjson.Result, envMat *matrix.Matrix) {

	const (
		query3k = "fandetails_3k.TABLE_fan_zone_speed.ROW_fan_zone_speed.speed"
		query9k = "fandetails.TABLE_fan_zone_speed.ROW_fan_zone_speed.zonespeed"
	)

	query := query3k
	fanSpeed := output.Get(query)
	if !fanSpeed.Exists() {
		query = query9k
		fanSpeed = output.Get(query)
	}
	if !fanSpeed.Exists() {
		e.SLogger.Warn("Unable to parse fan speed because rows are missing", slog.String("query", query))
		return
	}

	speed := hexReplacer.Replace(fanSpeed.String())
	zoneSpeed, err := strconv.ParseInt(speed, 16, 64)

	if err != nil {
		e.SLogger.Warn("error parsing fan_zone_speed", slog.Any("err", err), slog.String("zoneSpeed", speed))
		return
	}

	instanceKey := ""
	instance, err := envMat.NewInstance(instanceKey)
	if err != nil {
		e.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
		return
	}

	// Calculate the fan zone speed as a percentage
	fanZonePerc := math.Round(float64(zoneSpeed) / 255 * 100)

	envMat.GetMetric("fan_zone_speed").SetValueFloat64(instance, fanZonePerc)
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
			powerUpMetric.SetValueFloat64(instance, 1)
		} else {
			powerUpMetric.SetValueFloat64(instance, 0)
		}

		// If actualOut is 0, we don't want to export the other metrics
		if ps.ActualOut > 0 {
			envMat.GetMetric("power_out").SetValueFloat64(instance, ps.ActualOut)
			envMat.GetMetric("power_capacity").SetValueFloat64(instance, ps.TotalCapacity)
			envMat.GetMetric("power_in").SetValueFloat64(instance, ps.ActualIn)
		}
	}

	e.setRedundancyMode("configured", model.RedunMode, envMat)
	e.setRedundancyMode("operational", model.OperationMode, envMat)
}

func (e *Environment) setRedundancyMode(key string, mode string, envMat *matrix.Matrix) {
	instanceKey := key
	instance, err := envMat.NewInstance(instanceKey)
	if err != nil {
		e.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
		return
	}
	instance.SetLabel("item", key)
	instance.SetLabel("value", mode)
	envMat.GetMetric("power_mode").SetValueFloat64(instance, 1.0)
}

func (e *Environment) parseFanDetails(output gjson.Result, envMat *matrix.Matrix) {
	model := NewFanModel(output, e.SLogger)

	for _, f := range model.Fans {
		instanceKey := f.Name + "-" + f.TrayFanNum
		instance, err := envMat.NewInstance(instanceKey)
		if err != nil {
			e.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
			continue
		}

		instance.SetLabel("name", f.Name)
		instance.SetLabel("model", f.Model)
		instance.SetLabel("status", strings.ToLower(f.Status))
		instance.SetLabel("fan_num", f.TrayFanNum)

		fanUpMetric := envMat.GetMetric("fan_up")
		if f.Status == "ok" {
			fanUpMetric.SetValueFloat64(instance, 1)
		} else {
			fanUpMetric.SetValueFloat64(instance, 0)
		}

		if f.Speed < 0 {
			continue
		}

		fanSpeedMetric := envMat.GetMetric("fan_speed")
		fanSpeedMetric.SetValueFloat64(instance, float64(f.Speed))
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
	RedunMode      string
	OperationMode  string
}

type FanModel struct {
	Fans []*FanData
}

type FanData struct {
	Name       string
	Model      string
	Speed      int    // Speed is a percentage value from 0 to 100. Default to -1 to indicate not set
	TrayFanNum string // TrayFanNum is the fan number in the tray, e.g. "fan1" or "fan2"
	Status     string
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

	redunMode := output.Get("powersup.power_summary.ps_redun_mode").String()
	powerModel.RedunMode = redunMode
	operationMode := output.Get("powersup.power_summary.ps_oper_mode").String()
	powerModel.OperationMode = operationMode

	powerModel.PowerSupplies = powerSupplies

	return powerModel
}

func wattsToFloat(out string, logger *slog.Logger) float64 {
	trimmed := strings.TrimSpace(strings.ReplaceAll(out, " W", ""))
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

	redunMode := output.Get("powersup.power_summary.ps_redun_mode_3k").String()
	powerModel.RedunMode = redunMode
	operationMode := output.Get("powersup.power_summary.ps_redun_op_mode").String()
	powerModel.OperationMode = operationMode

	powerModel.PowerSupplies = powerSupplies

	return powerModel
}

// NewFanModel creates a new FanModel from the provided output and logger.
// It parses the fan information and fan tray information to populate the FanModel.
// It supports both Cisco 3000 and 9000 series switches by checking the output structure
// and using the appropriate queries for each series.
// The FanModel contains a slice of FanData, which includes the fan name, status, model, and speed.
// If the output does not contain the expected data, it logs a warning and returns an empty FanModel.
// A quirk about the NX-API response is that the number of returned fan_info and fan_tray objects do not match.
// This is because the fan_info objects show a summary, while the fan_tray objects are for individual fans.
// e.g.
// show environment fan detail
// Fan:
// ---------------------------------------------------------------------------
// Fan             Model                Hw     Direction       Status
// ---------------------------------------------------------------------------
// Fan1(sys_fan1)  NXA-FAN-30CFM-F      --     back-to-front   Ok
// Fan2(sys_fan2)  NXA-FAN-30CFM-F      --     back-to-front   Ok
// Fan3(sys_fan3)  NXA-FAN-30CFM-F      --     back-to-front   Ok
// Fan4(sys_fan4)  NXA-FAN-30CFM-F      --     back-to-front   Ok
// Fan_in_PS1      --                   --     back-to-front   Ok
// Fan_in_PS2      --                   --     back-to-front   Ok
// Fan Zone Speed: Zone 1: 0x9a
// Fan Air Filter : NotSupported
// Fan:
// ------------------------------------------------------------------
// Fan Name          Fan Num   Fan Direction   Speed(%)  Speed(RPM)
// ------------------------------------------------------------------
// Fan1(sys_fan1)      fan1    back-to-front    85        10305
// Fan1(sys_fan1)      fan2    back-to-front    64        7219
// Fan2(sys_fan2)      fan1    back-to-front    86        10364
// Fan2(sys_fan2)      fan2    back-to-front    64        7277
// Fan3(sys_fan3)      fan1    back-to-front    84        10188
// Fan3(sys_fan3)      fan2    back-to-front    64        7200
// Fan4(sys_fan4)      fan1    back-to-front    83        10055
// Fan4(sys_fan4)      fan2    back-to-front    64        7190

func NewFanModel(output gjson.Result, logger *slog.Logger) FanModel {
	var (
		fanModel        FanModel
		fanInfoToModel  = make(map[string]*FanData) // Map to hold fan data by name, e.g. Fan1(sys_fan1) => FanData
		fanInfoCount    = make(map[string]int)
		fanAddedToModel = make(map[string]bool)
		infoQuery       string
		trayQuery       string
	)
	var fans []*FanData //nolint:prealloc

	// Check if the output is from a 3000 or 9000 switch
	is3K := output.Get("fandetails_3k").Exists()

	const (
		fanInfoQuery3K = "fandetails_3k.TABLE_faninfo.ROW_faninfo"
		fanTrayQuery3K = "fandetails_3k.TABLE_fantray.ROW_fantray"
		fanInfoQuery9K = "fandetails.TABLE_faninfo.ROW_faninfo"
		fanTrayQuery9K = "fandetails.TABLE_fantray.ROW_fantray"
	)

	if is3K {
		infoQuery = fanInfoQuery3K
		trayQuery = fanTrayQuery3K
	} else {
		infoQuery = fanInfoQuery9K
		trayQuery = fanTrayQuery9K
	}

	fanModel = FanModel{}

	// Parse fan info rows first to build map of FanData

	rows := output.Get(infoQuery)
	if !rows.Exists() {
		logger.Warn("Unable to parse fan because rows are missing", slog.String("query", infoQuery))
		return fanModel
	}

	rows.ForEach(func(_, value gjson.Result) bool {
		fanName := value.Get("fanname").ClonedString()
		fanStatus := strings.ToLower(value.Get("fanstatus").ClonedString())
		model := value.Get("fanmodel").ClonedString()

		newFan := FanData{
			Name:   fanName,
			Status: fanStatus,
			Model:  model,
			Speed:  -1, // Default speed to -1 to indicate not set
		}

		fanInfoToModel[fanName] = &newFan
		fanAddedToModel[fanName] = false

		return true
	})

	// Parse fan tray rows and add each fan to the fans slice.
	rows = output.Get(trayQuery)
	if !rows.Exists() {
		logger.Warn("Unable to parse fan tray because rows are missing", slog.String("query", trayQuery))
		return fanModel
	}

	rows.ForEach(func(_, value gjson.Result) bool {
		fanName := value.Get("fanname").ClonedString()
		fanSpeed := value.Get("fanperc").Int()

		fanInfoCount[fanName]++

		// If the fan already exists in the fanInfoToModel map, update its speed
		if existingFan, exists := fanInfoToModel[fanName]; exists {
			fanAddedToModel[fanName] = true
			clone := *existingFan
			clone.Speed = int(fanSpeed)
			clone.TrayFanNum = "fan" + strconv.Itoa(fanInfoCount[fanName]) // e.g. "fan1", "fan2"
			fans = append(fans, &clone)
		} else {
			// If the fan does not exist, we will create a new FanData instance below
			logger.Warn("Fan not found in fanInfoToModel", slog.String("fanName", fanName))
			return true
		}

		return true
	})

	// Include all the fans from the fanInfoToModel map that were not added yet
	for fanName, fanData := range fanInfoToModel {
		if fanAddedToModel[fanName] {
			continue
		}

		fans = append(fans, fanData)
		fanAddedToModel[fanName] = true
	}

	fanModel.Fans = fans

	return fanModel
}
