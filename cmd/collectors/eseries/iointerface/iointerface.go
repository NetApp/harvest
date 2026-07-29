package iointerface

import (
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// InterfaceTypeToDataKey maps the interfaceType field value to the actual JSON data key
// in ioInterfaceTypeData. The eSeries swagger IOInterfaceTypeData definition shows two
// cases where the interfaceType enum value does not match the JSON property key:
//   - "fc"                 → "fibre"               (FibreInterface)
//   - "nvmeCouplingDriver" → "couplingDriverNvme"   (CouplingDriverNVMeInterface)
//
// All other types (ib, iscsi, sas, sata, scsi, ethernet, pcie) use identical names.
func InterfaceTypeToDataKey(interfaceType string) string {
	switch interfaceType {
	case "fc":
		return "fibre"
	case "nvmeCouplingDriver":
		return "couplingDriverNvme"
	default:
		return interfaceType
	}
}

func ResolveInterfaceData(iface gjson.Result) (string, gjson.Result, bool) {
	interfaceType := iface.Get("ioInterfaceTypeData.interfaceType").ClonedString()
	dataKey := InterfaceTypeToDataKey(interfaceType)
	data := iface.Get("ioInterfaceTypeData." + dataKey)
	if !data.Exists() || data.Type == gjson.Null {
		return interfaceType, gjson.Result{}, false
	}
	return interfaceType, data, true
}

func BuildPortLabelMap(response gjson.Result) map[string]string {
	portLabelMap := make(map[string]string)

	channelPorts := response.Get("channelPorts")
	if !channelPorts.Exists() || !channelPorts.IsArray() {
		return portLabelMap
	}

	for _, cp := range channelPorts.Array() {
		if cp.Get("channelType").ClonedString() != "hostside" {
			continue
		}
		channel := cp.Get("channel").ClonedString()
		label := cp.Get("physicalLocation.label").ClonedString()
		if channel != "" && label != "" {
			portLabelMap[channel] = label
		}
	}

	return portLabelMap
}

func ResolvePortLabel(channel string, portLabelMap map[string]string, physicalLocationLabel string) string {
	if label, ok := portLabelMap[channel]; ok && label != "" {
		return label
	}
	return physicalLocationLabel
}
