package optic

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"strings"
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

	o.matrix = matrix.New(o.Parent+o.templateObject, o.templateObject, o.templateObject)
	if err := o.matrix.NewMetricsFloat64(metrics...); err != nil {
		return fmt.Errorf("error while initializing matrix: %w", err)
	}

	return nil
}

func (o *Optic) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[o.Object]
	o.client.Metadata.Reset()

	o.matrix.PurgeInstances()
	o.matrix.Reset()

	// Set all global labels if they don't already exist
	o.matrix.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := o.ParentParams.GetChildContentS("query")
	output, err := o.client.CLIShowArray(command)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	o.parseOptic(output, o.matrix)

	o.client.Metadata.NumCalls.Store(1)
	o.client.Metadata.BytesRx.Store(uint64(len(output.Raw)))
	o.client.Metadata.PluginInstances.Store(uint64(len(o.matrix.GetInstances())))

	return []*matrix.Matrix{o.matrix}, o.client.Metadata, nil
}

func (o *Optic) parseOptic(output gjson.Result, opticMat *matrix.Matrix) {

	var models []Model

	rowQuery := "output.body.TABLE_interface.ROW_interface"

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

	rxMetric := opticMat.MustGetMetric(rx)
	txMetric := opticMat.MustGetMetric(tx)

	for _, model := range models {
		instanceKey := model.Name

		instance, err := opticMat.NewInstance(instanceKey)
		if err != nil {
			o.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
			continue
		}

		instance.SetLabel("interface", model.Name)

		rxMetric.SetValueFloat64(instance, model.RxPower)
		txMetric.SetValueFloat64(instance, model.TxPower)
	}
}

type Model struct {
	Name    string
	RxPower float64
	TxPower float64
}

func NewOpticModel(output gjson.Result) Model {

	var m Model

	// NX-API version 7.X returns optics data inline while later versions return a nested structure under TABLE_lane.ROW_lane

	list := output.Get("TABLE_lane.ROW_lane")
	if list.Exists() {
		list.ForEach(func(_, value gjson.Result) bool {
			rxVal := value.Get("rx_pwr")
			if rxVal.Exists() {
				m.Name = strings.TrimSpace(output.Get("interface").ClonedString())
				m.RxPower = rxVal.Float()
			}

			txVal := value.Get("tx_pwr")
			if txVal.Exists() {
				m.Name = strings.TrimSpace(output.Get("interface").ClonedString())
				m.TxPower = txVal.Float()
			}

			return false // Stop iterating after the first element
		})
		return m
	}

	rxVal := output.Get("rx_pwr")
	name := strings.TrimSpace(output.Get("interface").ClonedString())
	if rxVal.Exists() {
		m.Name = name
		m.RxPower = rxVal.Float()
	}

	txVal := output.Get("tx_pwr")
	if txVal.Exists() {
		m.Name = name
		m.TxPower = txVal.Float()
	}

	return m
}
