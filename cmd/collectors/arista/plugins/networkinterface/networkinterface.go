package networkinterface

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/rest"
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
	output, err := i.client.RunCmds(rest.SplitCommands(command)...)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	i.parseInterface(output, i.matrix)

	i.client.Metadata.NumCalls.Store(1)
	i.client.Metadata.BytesRx.Store(uint64(len(output.Raw)))
	i.client.Metadata.PluginInstances.Store(uint64(len(i.matrix.GetInstances())))

	return []*matrix.Matrix{i.matrix}, i.client.Metadata, nil
}

func (i *Interface) parseInterface(output gjson.Result, ifMat *matrix.Matrix) {

	rowQuery := "0.interfaces"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		i.SLogger.Warn("Unable to parse interfaces because interfaces are missing", slog.String("query", rowQuery))
		return
	}

	crcErrorsMetric := ifMat.MustGetMetric(crcErrors)
	receiveBytesMetric := ifMat.MustGetMetric(receiveBytes)
	receiveErrorsMetric := ifMat.MustGetMetric(receiveErrors)
	transmitBytesMetric := ifMat.MustGetMetric(transmitBytes)
	transmitErrorsMetric := ifMat.MustGetMetric(transmitErrors)
	receiveMulticastMetric := ifMat.MustGetMetric(receiveMulticast)
	receiveBroadcastMetric := ifMat.MustGetMetric(receiveBroadcast)
	receiveDropsMetric := ifMat.MustGetMetric(receiveDrops)
	transmitDropsMetric := ifMat.MustGetMetric(transmitDrops)
	adminUpMetric := ifMat.MustGetMetric(adminUp)
	upMetric := ifMat.MustGetMetric(up)
	errorStatusMetric := ifMat.MustGetMetric(errorStatus)

	rows.ForEach(func(key, value gjson.Result) bool {
		interfaceName := value.Get("name").ClonedString()
		if interfaceName == "" {
			interfaceName = key.ClonedString()
		}

		// Only collect Ethernet interfaces (skip VLAN, Loopback, Port-Channel, Management, etc.)
		if value.Get("hardware").ClonedString() != "ethernet" {
			return true
		}

		macAddr := value.Get("physicalAddress").ClonedString()
		desc := strings.TrimSpace(rest.TrimQuotes(value.Get("description").ClonedString()))
		interfaceStatus := value.Get("interfaceStatus").ClonedString()
		lineProtocol := value.Get("lineProtocolStatus").ClonedString()

		counters := value.Get("interfaceCounters")
		inBytes := counters.Get("inOctets").Float()
		outBytes := counters.Get("outOctets").Float()
		inErrors := counters.Get("totalInErrors").Float()
		outErrors := counters.Get("totalOutErrors").Float()
		inMcast := counters.Get("inMulticastPkts").Float()
		inBcast := counters.Get("inBroadcastPkts").Float()
		inDiscards := counters.Get("inDiscards").Float()
		outDiscards := counters.Get("outDiscards").Float()
		crc := counters.Get("inputErrorsDetail.fcsErrors").Float()

		instanceKey := interfaceName + "_" + macAddr

		instance, err := ifMat.NewInstance(instanceKey)
		if err != nil {
			i.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
			return true
		}

		instance.SetLabel("interface", interfaceName)
		instance.SetLabel("mac", macAddr)
		instance.SetLabel("description", desc)

		crcErrorsMetric.SetValueFloat64(instance, crc)
		receiveBytesMetric.SetValueFloat64(instance, inBytes)
		receiveErrorsMetric.SetValueFloat64(instance, inErrors)
		transmitBytesMetric.SetValueFloat64(instance, outBytes)
		transmitErrorsMetric.SetValueFloat64(instance, outErrors)
		receiveMulticastMetric.SetValueFloat64(instance, inMcast)
		receiveBroadcastMetric.SetValueFloat64(instance, inBcast)
		receiveDropsMetric.SetValueFloat64(instance, inDiscards)
		transmitDropsMetric.SetValueFloat64(instance, outDiscards)

		// interfaceStatus is one of: connected, notconnect, disabled, errdisabled
		adminEnabled := interfaceStatus != "disabled"
		if adminEnabled {
			adminUpMetric.SetValueFloat64(instance, 1)
		} else {
			adminUpMetric.SetValueFloat64(instance, 0)
		}

		linkUp := lineProtocol == "up"
		if linkUp {
			upMetric.SetValueFloat64(instance, 1)
		} else {
			upMetric.SetValueFloat64(instance, 0)
		}

		// Flag an error when the interface is administratively enabled but the line protocol is down.
		if adminEnabled != linkUp {
			errorStatusMetric.SetValueFloat64(instance, 1)
		} else {
			errorStatusMetric.SetValueFloat64(instance, 0)
		}

		return true
	})
}
