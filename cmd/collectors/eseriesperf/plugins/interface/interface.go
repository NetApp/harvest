package interfacename

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/iointerface"
	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

type InterfaceName struct {
	*plugin.AbstractPlugin
	client          *rest.Client
	schedule        int
	interfaceLabels map[string]string // Maps interfaceRef -> port label
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &InterfaceName{AbstractPlugin: p}
}

func (n *InterfaceName) Init(remote conf.Remote) error {
	if err := n.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	poller, err := conf.PollerNamed(n.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, n.SLogger)
	if n.client, err = rest.New(poller, timeout, credentials, ""); err != nil {
		return err
	}

	if !n.Options.IsTest {
		if err := n.client.Init(1, remote); err != nil {
			return err
		}
	}

	n.interfaceLabels = make(map[string]string)
	n.schedule = n.SetPluginInterval()
	return nil
}

func (n *InterfaceName) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[n.Object]

	arrayID := n.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		n.SLogger.Warn("arrayID not found in ParentParams, skipping interface names")
		return nil, nil, nil
	}

	if n.schedule >= n.PluginInvocationRate {
		n.schedule = 0
		n.refreshInterfaceLabels(arrayID)
	}
	n.schedule++

	n.applyInterfaceLabels(data)

	return nil, nil, nil
}

func (n *InterfaceName) refreshInterfaceLabels(arrayID string) {
	n.interfaceLabels = make(map[string]string)

	interfaceLabels, err := n.buildInterfaceLabelMap(arrayID)
	if err != nil {
		n.SLogger.Warn("Failed to build interface label map", slogx.Err(err))
		return
	}

	n.interfaceLabels = interfaceLabels
	n.SLogger.Debug("Refreshed interface labels", slog.Int("count", len(n.interfaceLabels)))
}

// buildPortLabelMap builds a map of channel number -> port label (e.g. "1" -> "1a")
// from the channelPorts array in hardware-inventory, filtered to hostside channels.
func (n *InterfaceName) buildPortLabelMap(arrayID string) (map[string]string, error) {
	apiPath := n.client.APIPath + "/storage-systems/" + arrayID + "/hardware-inventory"
	results, err := n.client.Fetch(apiPath, nil)
	if err != nil {
		return map[string]string{}, fmt.Errorf("failed to fetch hardware-inventory: %w", err)
	}
	if len(results) == 0 {
		return map[string]string{}, nil
	}

	portLabelMap := iointerface.BuildPortLabelMap(results[0])
	n.SLogger.Debug("Built port label map", slog.Int("entries", len(portLabelMap)))
	return portLabelMap, nil
}

func (n *InterfaceName) fetchInterfaces(arrayID string, channelType string) ([]gjson.Result, error) {
	apiPath := n.client.APIPath + "/storage-systems/" + arrayID + "/interfaces?interfaceType=&channelType=" + channelType
	results, err := n.client.Fetch(apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch interfaces for channelType=%s: %w", channelType, err)
	}
	return results, nil
}

func (n *InterfaceName) buildInterfaceLabelMap(arrayID string) (map[string]string, error) {
	interfaceLabels := make(map[string]string)

	portLabelMap, err := n.buildPortLabelMap(arrayID)
	if err != nil {
		return interfaceLabels, err
	}

	for _, channelType := range []string{"hostside", "driveside"} {
		results, err := n.fetchInterfaces(arrayID, channelType)
		if err != nil {
			return interfaceLabels, err
		}

		for _, iface := range results {
			interfaceRef := iface.Get("interfaceRef").ClonedString()
			if interfaceRef == "" {
				continue
			}

			_, interfaceData, ok := iointerface.ResolveInterfaceData(iface)
			if !ok {
				continue
			}

			channel := interfaceData.Get("channel").ClonedString()

			label := iointerface.ResolvePortLabel(channel, portLabelMap, interfaceData.Get("physicalLocation.label").ClonedString())
			if label != "" {
				interfaceLabels[interfaceRef] = label
			}
		}
	}

	n.SLogger.Debug("Built interface label map", slog.Int("count", len(interfaceLabels)))
	return interfaceLabels, nil
}

func (n *InterfaceName) applyInterfaceLabels(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		id := instance.GetLabel("id")
		if id == "" {
			continue
		}

		if label, ok := n.interfaceLabels[id]; ok {
			instance.SetLabel("interface", label)
		} else {
			n.SLogger.Debug("Interface label not found in cache", slog.String("id", id))
		}
	}
}
