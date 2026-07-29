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

func (i *Interface) Init(remote conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = i.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	if client, err = rest.New(conf.ZapiPoller(i.ParentParams), i.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	if err := client.Init(2, remote); err != nil {
		return err
	}

	i.client = client
	i.templateObject = i.ParentParams.GetChildContentS("object")

	i.matrix = matrix.New(i.Parent+i.templateObject, i.templateObject, i.templateObject)
	if err := i.matrix.NewMetricsFloat64(metrics...); err != nil {
		return fmt.Errorf("error while initializing matrix: %w", err)
	}

	return nil
}

func (i *Interface) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[i.Object]
	i.client.Metadata.Reset()

	i.matrix.PurgeInstances()
	i.matrix.Reset()

	// Set all global labels if they don't already exist
	i.matrix.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := i.ParentParams.GetChildContentS("query")
	output, err := i.client.CLIShowArray(command)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	i.parseInterface(output, i.matrix)

	i.client.Metadata.NumCalls.Store(1)
	i.client.Metadata.BytesRx.Store(uint64(len(output.Raw)))
	i.client.Metadata.PluginInstances.Store(uint64(len(i.matrix.GetInstances())))

	return []*matrix.Matrix{i.matrix}, i.client.Metadata, nil
}

func (i *Interface) parseInterface(output gjson.Result, envMat *matrix.Matrix) {

	rowQuery := "output.body.TABLE_interface.ROW_interface"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		i.SLogger.Warn("Unable to parse interfaces because rows are missing", slog.String("query", rowQuery))
		return
	}

	crcErrorsMetric := envMat.MustGetMetric(crcErrors)
	ethOutDiscardMetric := envMat.MustGetMetric(ethOutDiscard)
	receiveBytesMetric := envMat.MustGetMetric(receiveBytes)
	receiveErrorsMetric := envMat.MustGetMetric(receiveErrors)
	transmitBytesMetric := envMat.MustGetMetric(transmitBytes)
	transmitErrorsMetric := envMat.MustGetMetric(transmitErrors)
	receiveMulticastMetric := envMat.MustGetMetric(receiveMulticast)
	receiveBroadcastMetric := envMat.MustGetMetric(receiveBroadcast)
	receiveDropsMetric := envMat.MustGetMetric(receiveDrops)
	transmitDropsMetric := envMat.MustGetMetric(transmitDrops)
	adminUpMetric := envMat.MustGetMetric(adminUp)
	upMetric := envMat.MustGetMetric(up)
	errorStatusMetric := envMat.MustGetMetric(errorStatus)

	rows.ForEach(func(_, value gjson.Result) bool {
		interfaceName := value.Get("interface").ClonedString()
		macAddr := value.Get("eth_hw_addr").ClonedString()
		desc := value.Get("desc").ClonedString()
		if desc == "" {
			// Cisco 5K
			desc = value.Get("eth_hw_desc").ClonedString()
		}
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
		ethClearCounters := value.Get("eth_clear_counters").String()

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

		crcErrorsMetric.SetValueFloat64(instance, ethCrcErrors)
		ethOutDiscardMetric.SetValueFloat64(instance, ethOutDiscards)
		receiveBytesMetric.SetValueFloat64(instance, ethInBytes)
		receiveErrorsMetric.SetValueFloat64(instance, ethInErrors)
		transmitBytesMetric.SetValueFloat64(instance, ethOutBytes)
		transmitErrorsMetric.SetValueFloat64(instance, ethOutErrors)
		receiveMulticastMetric.SetValueFloat64(instance, ethInMcast)
		receiveBroadcastMetric.SetValueFloat64(instance, ethInBcast)
		receiveDropsMetric.SetValueFloat64(instance, ethInDrops)
		transmitDropsMetric.SetValueFloat64(instance, ethOutDrops)

		if adminState == "up" {
			adminUpMetric.SetValueFloat64(instance, 1)
		} else {
			adminUpMetric.SetValueFloat64(instance, 0)
		}

		if state == "up" {
			upMetric.SetValueFloat64(instance, 1)
		} else {
			upMetric.SetValueFloat64(instance, 0)
		}

		if adminState != state {
			errorStatusMetric.SetValueFloat64(instance, 1)
		} else {
			errorStatusMetric.SetValueFloat64(instance, 0)
		}

		spuriousZero := ethClearCounters == "never" && (ethInBytes == 0 || ethOutBytes == 0)

		if spuriousZero {
			instance.SetExportable(false)
			i.SLogger.Warn(
				"Skipping invalid zero samples",
				slog.String("interface", interfaceName),
				slog.Float64("eth_inbytes", ethInBytes),
				slog.Float64("eth_outbytes", ethOutBytes),
				slog.String("eth_clear_counters", ethClearCounters),
			)
		}

		return true
	})
}
