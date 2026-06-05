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
	transmitDiscard  = "transmit_discards"
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
	transmitDiscard,
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

	i.matrix = matrix.New(i.Parent+".Interface", i.templateObject, i.templateObject)

	return nil
}

func (i *Interface) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[i.Object]
	i.client.Metadata.Reset()

	interfaceMat, err := i.initMatrix(i.templateObject)
	if err != nil {
		return nil, nil, fmt.Errorf("error while initializing matrix: %w", err)
	}

	// Set all global labels if they don't already exist
	interfaceMat.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := i.ParentParams.GetChildContentS("query")
	output, err := i.client.RunCmds(rest.SplitCommands(command)...)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	i.parseInterface(output, interfaceMat)

	i.client.Metadata.NumCalls.Store(1)
	i.client.Metadata.BytesRx.Store(uint64(len(output.Raw)))
	i.client.Metadata.PluginInstances.Store(uint64(len(interfaceMat.GetInstances())))

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

func (i *Interface) parseInterface(output gjson.Result, ifMat *matrix.Matrix) {

	rowQuery := "0.interfaces"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		i.SLogger.Warn("Unable to parse interfaces because interfaces are missing", slog.String("query", rowQuery))
		return
	}

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

		ifMat.GetMetric(crcErrors).SetValueFloat64(instance, crc)
		ifMat.GetMetric(transmitDiscard).SetValueFloat64(instance, outDiscards)
		ifMat.GetMetric(receiveBytes).SetValueFloat64(instance, inBytes)
		ifMat.GetMetric(receiveErrors).SetValueFloat64(instance, inErrors)
		ifMat.GetMetric(transmitBytes).SetValueFloat64(instance, outBytes)
		ifMat.GetMetric(transmitErrors).SetValueFloat64(instance, outErrors)
		ifMat.GetMetric(receiveMulticast).SetValueFloat64(instance, inMcast)
		ifMat.GetMetric(receiveBroadcast).SetValueFloat64(instance, inBcast)
		ifMat.GetMetric(receiveDrops).SetValueFloat64(instance, inDiscards)
		ifMat.GetMetric(transmitDrops).SetValueFloat64(instance, outDiscards)

		// interfaceStatus is one of: connected, notconnect, disabled, errdisabled
		adminEnabled := interfaceStatus != "disabled"
		if adminEnabled {
			ifMat.GetMetric(adminUp).SetValueFloat64(instance, 1)
		} else {
			ifMat.GetMetric(adminUp).SetValueFloat64(instance, 0)
		}

		linkUp := lineProtocol == "up"
		if linkUp {
			ifMat.GetMetric(up).SetValueFloat64(instance, 1)
		} else {
			ifMat.GetMetric(up).SetValueFloat64(instance, 0)
		}

		// Flag an error when the interface is administratively enabled but the line protocol is down.
		if adminEnabled != linkUp {
			ifMat.GetMetric(errorStatus).SetValueFloat64(instance, 1)
		} else {
			ifMat.GetMetric(errorStatus).SetValueFloat64(instance, 0)
		}

		return true
	})
}
