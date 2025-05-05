package optic

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"time"
)

const (
	rx = "rx"
	tx = "tx"
)

var metrics = []string{
	rx,
	tx,
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

func (o *Optic) Init(_ conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = o.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)

	if client, err = rest.New(conf.ZapiPoller(o.ParentParams), timeout, o.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	o.client = client
	o.templateObject = o.ParentParams.GetChildContentS("object")

	o.matrix = matrix.New(o.Parent+".Optic", o.templateObject, o.templateObject)

	return nil
}

func (o *Optic) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
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
	output, err := o.client.CLIShowArray(command, "")

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	o.parseOptic(output, opticMat)

	o.client.Metadata.NumCalls = 1
	o.client.Metadata.BytesRx = uint64(len(output.Raw))
	o.client.Metadata.PluginInstances = uint64(len(opticMat.GetInstances()))

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

	var models []Model

	rowQuery := "TABLE_interface.ROW_interface"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		o.SLogger.Warn("Unable to parse optics because rows are missing", slog.String("query", rowQuery))
		return
	}

	rows.ForEach(func(_, value gjson.Result) bool {
		opticModel := NewOpticModel(value)
		// Skip empty models
		if opticModel.Name == "" {
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
	}
}

type Model struct {
	Name    string
	RxPower float64
	TxPower float64
}

func NewOpticModel(output gjson.Result) Model {

	var m Model

	output.Get("TABLE_lane.ROW_lane").ForEach(func(_, value gjson.Result) bool {
		rxVal := value.Get("rx_pwr")
		if rxVal.Exists() {
			m.Name = output.Get("interface").ClonedString()
			m.RxPower = rxVal.Float()
		}

		txVal := value.Get("tx_pwr")
		if txVal.Exists() {
			m.Name = output.Get("interface").ClonedString()
			m.TxPower = txVal.Float()
		}

		return false // Stop iterating after the first element
	})

	return m
}
