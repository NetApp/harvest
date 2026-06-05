package optic

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
	rx          = "rx"
	tx          = "tx"
	temperature = "temperature"
	voltage     = "voltage"
)

var metrics = []string{
	rx,
	tx,
	temperature,
	voltage,
}

type Optic struct {
	*plugin.AbstractPlugin
	matrix         *matrix.Matrix
	client         *rest.Client
	templateObject string // object name from the template
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Optic{AbstractPlugin: p}
}

func (o *Optic) Init(remote conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = o.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	if client, err = rest.New(conf.ZapiPoller(o.ParentParams), o.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	if err := client.Init(2, remote); err != nil {
		return err
	}

	o.client = client
	o.templateObject = o.ParentParams.GetChildContentS("object")

	o.matrix = matrix.New(o.Parent+".Optic", o.templateObject, o.templateObject)

	return nil
}

func (o *Optic) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[o.Object]
	o.client.Metadata.Reset()

	opticMat, err := o.initMatrix(o.templateObject)
	if err != nil {
		return nil, nil, fmt.Errorf("error while initializing matrix: %w", err)
	}

	// Set all global labels if they don't already exist
	opticMat.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := o.ParentParams.GetChildContentS("query")
	output, err := o.client.RunCmds(rest.SplitCommands(command)...)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	o.parseOptic(output, opticMat)

	o.client.Metadata.NumCalls.Store(1)
	o.client.Metadata.BytesRx.Store(uint64(len(output.Raw)))
	o.client.Metadata.PluginInstances.Store(uint64(len(opticMat.GetInstances())))

	return []*matrix.Matrix{opticMat}, o.client.Metadata, nil
}

func (o *Optic) initMatrix(name string) (*matrix.Matrix, error) {

	mat := matrix.New(o.Parent+name, name, name)

	for _, k := range metrics {
		if err := matrix.CreateMetric(k, mat); err != nil {
			return nil, fmt.Errorf("error while creating metric %s: %w", k, err)
		}
	}

	return mat, nil
}

func (o *Optic) parseOptic(output gjson.Result, opticMat *matrix.Matrix) {

	rowQuery := "0.interfaces"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		o.SLogger.Warn("Unable to parse optics because interfaces are missing", slog.String("query", rowQuery))
		return
	}

	var models []Model

	rows.ForEach(func(key, value gjson.Result) bool {
		opticModel := NewOpticModel(key.ClonedString(), value)
		// Skip interfaces without transceiver DOM data (e.g. empty SFP slots or copper ports)
		if !opticModel.HasData {
			return true
		}
		models = append(models, opticModel)
		return true
	})

	for _, model := range models {
		instanceKey := model.Name

		instance, err := opticMat.NewInstance(instanceKey)
		if err != nil {
			o.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
			continue
		}

		instance.SetLabel("interface", model.Name)

		opticMat.GetMetric(rx).SetValueFloat64(instance, model.RxPower)
		opticMat.GetMetric(tx).SetValueFloat64(instance, model.TxPower)
		opticMat.GetMetric(temperature).SetValueFloat64(instance, model.Temperature)
		opticMat.GetMetric(voltage).SetValueFloat64(instance, model.Voltage)
	}
}

type Model struct {
	Name        string
	RxPower     float64
	TxPower     float64
	Temperature float64
	Voltage     float64
	HasData     bool
}

func NewOpticModel(name string, output gjson.Result) Model {

	var m Model
	m.Name = name

	rxVal := output.Get("rxPower")
	txVal := output.Get("txPower")
	tempVal := output.Get("temperature")
	voltVal := output.Get("voltage")

	if rxVal.Exists() {
		m.RxPower = rxVal.Float()
		m.HasData = true
	}
	if txVal.Exists() {
		m.TxPower = txVal.Float()
		m.HasData = true
	}
	if tempVal.Exists() {
		m.Temperature = tempVal.Float()
		m.HasData = true
	}
	if voltVal.Exists() {
		m.Voltage = voltVal.Float()
		m.HasData = true
	}

	return m
}
