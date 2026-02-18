package networkinterface

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
	"time"
)

const (
	adminUp          = "admin_up"
	crcErrors        = "crc_errors"
	errorStatus      = "error_status"
	ethOutDiscard    = "eth_out_discards"
	receiveBroadcast = "receive_broadcast"
	receiveBytes     = "receive_bytes"
	receiveErrors    = "receive_errors"
	receiveMulticast = "receive_multicast"
	receiveDrops     = "receive_drops"
	transmitBytes    = "transmit_bytes"
	transmitErrors   = "transmit_errors"
	transmitDrops    = "transmit_drops"
	up               = "up"
)

var metrics = []string{
	adminUp,
	crcErrors,
	errorStatus,
	ethOutDiscard,
	receiveBroadcast,
	receiveBytes,
	receiveErrors,
	receiveDrops,
	receiveMulticast,
	transmitBytes,
	transmitErrors,
	transmitDrops,
	up,
}

type Interface struct {
	*plugin.AbstractPlugin
	matrix         *matrix.Matrix
	client         *rest.Client
	templateObject string // object name from the template
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Interface{AbstractPlugin: p}
}

func (i *Interface) Init(_ conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = i.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)

	if client, err = rest.New(conf.ZapiPoller(i.ParentParams), timeout, i.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	i.client = client
	i.templateObject = i.ParentParams.GetChildContentS("object")

	i.matrix = matrix.New(i.Parent+".Interface", i.templateObject, i.templateObject)

	return nil
}

func (i *Interface) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[i.Object]
	i.client.Metadata.Reset()

	interfaceMat, err := i.initMatrix("cisco_interface")
	if err != nil {
		return nil, nil, fmt.Errorf("error while initializing matrix: %w", err)
	}

	// Set all global labels if they don't already exist
	interfaceMat.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := i.ParentParams.GetChildContentS("query")
	output, err := i.client.CLIShowArray(command)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	i.parseInterface(output, interfaceMat)

	i.client.Metadata.NumCalls = 1
	i.client.Metadata.BytesRx = uint64(len(output.Raw))
	i.client.Metadata.PluginInstances = uint64(len(interfaceMat.GetInstances()))

	return []*matrix.Matrix{interfaceMat}, i.client.Metadata, nil
}

func (i *Interface) initMatrix(name string) (*matrix.Matrix, error) {

	mat := matrix.New(i.Parent+name, name, name)

	for _, k := range metrics {
		if err := matrix.CreateMetric(k, mat); err != nil {
			return nil, fmt.Errorf("error while creating metric %s: %w", k, err)
		}
	}

	return mat, nil
}

func (i *Interface) parseInterface(output gjson.Result, envMat *matrix.Matrix) {

	rowQuery := "output.body.TABLE_interface.ROW_interface"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		i.SLogger.Warn("Unable to parse interfaces because rows are missing", slog.String("query", rowQuery))
		return
	}

	rows.ForEach(func(_, value gjson.Result) bool {
		interfaceName := value.Get("interface").ClonedString()
		macAddr := value.Get("eth_hw_addr").ClonedString()
		desc := value.Get("desc").ClonedString()
		ethSpeed := value.Get("eth_speed").ClonedString()
		adminState := value.Get("admin_state").ClonedString()
		state := value.Get("state").ClonedString()

		ethInBytes := value.Get("eth_inbytes").Float()
		ethOutBytes := value.Get("eth_outbytes").Float()
		ethInErrors := value.Get("eth_inerr").Float()
		ethOutErrors := value.Get("eth_outerr").Float()
		ethInMcast := value.Get("eth_inmcast").Float()
		ethInBcast := value.Get("eth_inbcast").Float()
		ethCrcErrors := value.Get("eth_crc").Float()
		ethInDrops := value.Get("eth_in_ifdown_drops").Float()
		ethOutDrops := value.Get("eth_out_drops").Float()
		ethOutDiscards := value.Get("eth_outdiscard").Float()

		instanceKey := interfaceName + "_" + macAddr

		instance, err := envMat.NewInstance(instanceKey)
		if err != nil {
			i.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
			return true
		}

		instance.SetLabel("interface", interfaceName)
		instance.SetLabel("mac", macAddr)

		desc = strings.TrimPrefix(desc, `"`)
		desc = strings.TrimSuffix(desc, `"`)
		desc = strings.TrimSpace(desc)

		instance.SetLabel("description", desc)
		instance.SetLabel("speed", ethSpeed)

		envMat.GetMetric(crcErrors).SetValueFloat64(instance, ethCrcErrors)
		envMat.GetMetric(ethOutDiscard).SetValueFloat64(instance, ethOutDiscards)
		envMat.GetMetric(receiveBytes).SetValueFloat64(instance, ethInBytes)
		envMat.GetMetric(receiveErrors).SetValueFloat64(instance, ethInErrors)
		envMat.GetMetric(transmitBytes).SetValueFloat64(instance, ethOutBytes)
		envMat.GetMetric(transmitErrors).SetValueFloat64(instance, ethOutErrors)
		envMat.GetMetric(receiveMulticast).SetValueFloat64(instance, ethInMcast)
		envMat.GetMetric(receiveBroadcast).SetValueFloat64(instance, ethInBcast)
		envMat.GetMetric(receiveDrops).SetValueFloat64(instance, ethInDrops)
		envMat.GetMetric(transmitDrops).SetValueFloat64(instance, ethOutDrops)

		if adminState == "up" {
			envMat.GetMetric(adminUp).SetValueFloat64(instance, 1)
		} else {
			envMat.GetMetric(adminUp).SetValueFloat64(instance, 0)
		}

		if state == "up" {
			envMat.GetMetric(up).SetValueFloat64(instance, 1)
		} else {
			envMat.GetMetric(up).SetValueFloat64(instance, 0)
		}

		if adminState != state {
			envMat.GetMetric(errorStatus).SetValueFloat64(instance, 1)
		} else {
			envMat.GetMetric(errorStatus).SetValueFloat64(instance, 0)
		}

		return true
	})
}
