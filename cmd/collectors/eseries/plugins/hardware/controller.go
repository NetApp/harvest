package hardware

import (
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// speedRegex matches patterns like "10", "2pt5", "22pt5" followed by "gig" or "meg"
var speedRegex = regexp.MustCompile(`(?i)^(\d+)(pt(\d+))?(gig|meg)$`)

// speedToMB converts E-Series speed strings to MB values or cleaned strings
// e.g., "speed10gig" -> "10000", "speed100meg" -> "100", "speed2pt5Gig" -> "2500", "speedAuto" -> "Auto"
func speedToMB(speed string) string {
	if speed == "" {
		return ""
	}

	// Handle special cases
	if speed == "__UNDEFINED" {
		return "__UNDEFINED"
	}

	s := strings.TrimPrefix(speed, "speed")

	matches := speedRegex.FindStringSubmatch(s)
	if len(matches) < 5 {
		return s
	}

	// matches[1] = whole number part (e.g., "10", "2", "22")
	// matches[3] = decimal part after "pt" (e.g., "5" from "pt5"), empty if no "pt"
	// matches[4] = unit (e.g., "gig", "meg")

	wholeNum, _ := strconv.ParseFloat(matches[1], 64)
	unit := strings.ToLower(matches[4])

	// Handle decimal part (e.g., "2pt5" = 2.5)
	if matches[3] != "" {
		decimalPart, _ := strconv.ParseFloat("0."+matches[3], 64)
		wholeNum += decimalPart
	}

	var mbValue float64
	switch unit {
	case "gig":
		mbValue = wholeNum * 1000
	case "meg":
		mbValue = wholeNum
	default:
		return s
	}

	return strconv.FormatFloat(mbValue, 'f', 0, 64)
}

// cleanLinkState removes the "link" prefix from port state values
// e.g., "linkUp" -> "Up", "linkDown" -> "Down"
func cleanLinkState(state string) string {
	return strings.TrimPrefix(state, "link")
}

// interfaceTypeToDataKey maps the interfaceType field value to the actual JSON data key
// in ioInterfaceTypeData. The eSeries swagger IOInterfaceTypeData definition shows two
// cases where the interfaceType enum value does not match the JSON property key:
//   - "fc"                 → "fibre"               (FibreInterface)
//   - "nvmeCouplingDriver" → "couplingDriverNvme"   (CouplingDriverNVMeInterface)
//
// All other types (ib, iscsi, sas, sata, scsi, ethernet, pcie) use identical names.
func interfaceTypeToDataKey(interfaceType string) string {
	switch interfaceType {
	case "fc":
		return "fibre"
	case "nvmeCouplingDriver":
		return "couplingDriverNvme"
	default:
		return interfaceType
	}
}

func cleanConfigType(configType string) string {
	switch strings.ToLower(configType) {
	case "stat":
		return "Static"
	default:
		return configType
	}
}

// getDNSServerAddress extracts the IP address from a DNS server entry
func getDNSServerAddress(server gjson.Result) string {
	addressType := server.Get("addressType").ClonedString()
	switch addressType {
	case "ipv4":
		return server.Get("ipv4Address").ClonedString()
	case "ipv6":
		return server.Get("ipv6Address").ClonedString()
	default:
		return ""
	}
}

func (h *Hardware) initControllerMatrix() {
	mat := matrix.New(h.Parent+"."+controllerMatrix, controllerMatrix, controllerMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"controller_id", "controller"},
		[]string{"app_version", "boot_version", "manufacturer", "model", "part_number", "controller", "serial_number", "status"},
	))

	if _, err := mat.NewMetricFloat64("used_cache_memory"); err != nil {
		h.SLogger.Error("Failed to create used_cache_memory metric", slogx.Err(err))
	}
	if _, err := mat.NewMetricFloat64("total_cache_memory"); err != nil {
		h.SLogger.Error("Failed to create total_cache_memory metric", slogx.Err(err))
	}
	if _, err := mat.NewMetricFloat64("processor_memory"); err != nil {
		h.SLogger.Error("Failed to create processor_memory metric", slogx.Err(err))
	}

	h.data[controllerMatrix] = mat
}

// initHostInterfaceMatrix creates the matrix for host interface data
func (h *Hardware) initHostInterfaceMatrix() {
	mat := matrix.New(h.Parent+"."+hostInterfaceMatrix, hostInterfaceMatrix, hostInterfaceMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"controller_id", "channel"},
		[]string{
			"controller_id", "controller", "interface_type", "port", "channel", "link_state", "speed",
			"physical_port_state", "nvme_supported", "max_transmission_unit", "ip_address",
		},
	))

	h.data[hostInterfaceMatrix] = mat
}

// initDriveInterfaceMatrix creates the matrix for drive interface data
func (h *Hardware) initDriveInterfaceMatrix() {
	mat := matrix.New(h.Parent+"."+driveInterfaceMatrix, driveInterfaceMatrix, driveInterfaceMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"controller_id", "channel"},
		[]string{"controller_id", "controller", "interface_type", "channel", "current_speed", "maximum_speed", "protection_info_capable"},
	))

	h.data[driveInterfaceMatrix] = mat
}

// initCodeVersionMatrix creates the matrix for code version data
func (h *Hardware) initCodeVersionMatrix() {
	mat := matrix.New(h.Parent+"."+codeVersionMatrix, codeVersionMatrix, codeVersionMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"controller_id", "code_module"},
		[]string{"controller_id", "controller", "code_module", "version"},
	))

	h.data[codeVersionMatrix] = mat
}

// initDNSMatrix creates the matrix for DNS property data
func (h *Hardware) initDNSMatrix() {
	mat := matrix.New(h.Parent+"."+dnsPropertyMatrix, dnsPropertyMatrix, dnsPropertyMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"controller_id", "dns_server"},
		[]string{"controller_id", "controller", "dns_server", "address_type", "acquisition_type"},
	))

	h.data[dnsPropertyMatrix] = mat
}

// initNetInterfaceMatrix creates the matrix for network interface (management port) data
func (h *Hardware) initNetInterfaceMatrix() {
	mat := matrix.New(h.Parent+"."+netInterfaceMatrix, netInterfaceMatrix, netInterfaceMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"controller_id", "interface_name"},
		[]string{
			"controller_id", "controller", "interface_name", "alias", "port", "mac_address", "link_status",
			"ipv4_enabled", "ipv4_address", "ipv4_subnet_mask", "ipv4_gateway", "ipv4_config_method",
			"ipv6_enabled", "ipv6_local_address", "ipv6_routable_address", "ipv6_config_method",
			"full_duplex", "configured_speed", "current_speed", "remote_access_enabled",
			"dns_config_method", "primary_dns_server", "backup_dns_server", "ntp_service",
		},
	))

	h.data[netInterfaceMatrix] = mat
}

// processControllers processes the controllers array and its nested data
func (h *Hardware) processControllers(response gjson.Result, _ map[string]string) {
	controllers := response.Get("controllers")

	if !controllers.Exists() || !controllers.IsArray() {
		h.SLogger.Debug("No controllers found in response")
		return
	}

	for _, controller := range controllers.Array() {
		controllerID := controller.Get("controllerRef").ClonedString()
		if controllerID == "" {
			controllerID = controller.Get("id").ClonedString()
		}
		if controllerID == "" {
			continue
		}

		controllerLocation := controller.Get("physicalLocation.label").ClonedString()

		h.processController(controller, controllerID)

		h.processCodeVersions(controller, controllerID, controllerLocation)
		h.processDNSProperties(controller, controllerID, controllerLocation)
		h.processNetInterfaces(controller, controllerID, controllerLocation)
	}
}

// processController processes a single controller
func (h *Hardware) processController(controller gjson.Result, controllerID string) {
	mat := h.data[controllerMatrix]

	inst, err := mat.NewInstance(controllerID)
	if err != nil {
		h.SLogger.Error("Failed to create controller instance", slogx.Err(err), slog.String("id", controllerID))
		return
	}

	inst.SetLabelTrimmed("controller_id", controllerID)
	inst.SetLabelTrimmed("app_version", controller.Get("appVersion").ClonedString())
	inst.SetLabelTrimmed("boot_version", controller.Get("bootVersion").ClonedString())
	inst.SetLabelTrimmed("manufacturer", controller.Get("manufacturer").ClonedString())
	inst.SetLabelTrimmed("model", controller.Get("modelName").ClonedString())
	inst.SetLabelTrimmed("part_number", controller.Get("partNumber").ClonedString())
	inst.SetLabelTrimmed("controller", controller.Get("physicalLocation.label").ClonedString())
	inst.SetLabelTrimmed("serial_number", controller.Get("serialNumber").ClonedString())
	inst.SetLabelTrimmed("status", controller.Get("status").ClonedString())

	usedCacheMemoryMetric := mat.MustGetMetric("used_cache_memory")
	totalCacheMemoryMetric := mat.MustGetMetric("total_cache_memory")
	processorMemoryMetric := mat.MustGetMetric("processor_memory")

	// Set metrics
	if val := controller.Get("cacheMemorySize"); val.Exists() {
		usedCacheMemoryMetric.SetValueFloat64(inst, float64(val.Uint()*1024*1024))
	}
	if val := controller.Get("physicalCacheMemorySize"); val.Exists() {
		totalCacheMemoryMetric.SetValueFloat64(inst, float64(val.Uint()*1024*1024))
	}
	if val := controller.Get("processorMemorySize"); val.Exists() {
		processorMemoryMetric.SetValueFloat64(inst, float64(val.Uint()*1024*1024))
	}

	h.SLogger.Debug("Processed controller", slog.String("id", controllerID))
}

func extractPortIPAddress(iface gjson.Result, interfaceType string, interfaceData gjson.Result) string {
	// iSCSI: IPv4 address is directly inside the iscsi interface type data.
	if interfaceType == "iscsi" {
		return interfaceData.Get("ipv4Data.ipv4AddressData.ipv4Address").ClonedString()
	}

	// All other protocols: IP (if any) is carried inside commandProtocolPropertiesList
	// at the top level of the IoInterface object.
	cmdProtoList := iface.Get("commandProtocolPropertiesList.commandProtocolProperties")
	if !cmdProtoList.Exists() || !cmdProtoList.IsArray() {
		return ""
	}

	for _, cp := range cmdProtoList.Array() {
		switch cp.Get("commandProtocol").ClonedString() {
		case "nvme":
			nvmeofProps := cp.Get("nvmeProperties.nvmeofProperties")
			if !nvmeofProps.Exists() || nvmeofProps.Type == gjson.Null {
				continue
			}
			provider := nvmeofProps.Get("provider").ClonedString()
			switch provider {
			case "providerRocev2", "providerRoce", "providerIwarp":
				if ip := nvmeofProps.Get("roceV2Properties.ipv4Data.ipv4AddressData.ipv4Address").ClonedString(); ip != "" {
					return ip
				}
			case "providerInfiniband":
				if ip := nvmeofProps.Get("ibProperties.ipAddressData.ipv4Data.ipv4Address").ClonedString(); ip != "" {
					return ip
				}
			}
		case "scsi":
			if cp.Get("scsiProperties.scsiProtocolType").ClonedString() == "iser" {
				if ip := cp.Get("scsiProperties.iserProperties.ipv4Data.ipv4AddressData.ipv4Address").ClonedString(); ip != "" {
					return ip
				}
			}
		}
	}
	return ""
}

// processHostInterfacesFromAPI processes host interfaces from the /interfaces?channelType=hostside API endpoint
func (h *Hardware) processHostInterfacesFromAPI(results []gjson.Result, controllerLabelMap, portLabelMap map[string]string) {
	mat := h.data[hostInterfaceMatrix]

	for _, iface := range results {
		interfaceRef := iface.Get("interfaceRef").ClonedString()
		if interfaceRef == "" {
			continue
		}

		controllerRef := iface.Get("controllerRef").ClonedString()
		controllerName := controllerLabelMap[controllerRef]

		interfaceType := iface.Get("ioInterfaceTypeData.interfaceType").ClonedString()
		dataKey := interfaceTypeToDataKey(interfaceType)
		interfaceData := iface.Get("ioInterfaceTypeData." + dataKey)

		if !interfaceData.Exists() || interfaceData.Type == gjson.Null {
			continue
		}

		channel := interfaceData.Get("channel").ClonedString()

		key := controllerRef + "_" + interfaceRef
		inst, err := mat.NewInstance(key)
		if err != nil {
			h.SLogger.Error("Failed to create host interface instance", slogx.Err(err), slog.String("key", key))
			continue
		}

		inst.SetLabelTrimmed("controller_id", controllerRef)
		inst.SetLabelTrimmed("controller", controllerName)
		inst.SetLabelTrimmed("interface_type", interfaceType)
		inst.SetLabelTrimmed("channel", channel)

		portLabel := portLabelMap[channel]
		if portLabel == "" {
			portLabel = interfaceData.Get("physicalLocation.label").ClonedString()
		}
		inst.SetLabelTrimmed("port", portLabel)

		// FC interfaces use "linkStatus" and "currentInterfaceSpeed"; iSCSI/ethernet use "linkState" and "currentSpeed"
		linkState := interfaceData.Get("linkState").ClonedString()
		if linkState == "" {
			linkState = interfaceData.Get("linkStatus").ClonedString()
		}
		inst.SetLabelTrimmed("link_state", linkState)

		speedVal := interfaceData.Get("currentSpeed").ClonedString()
		if speedVal == "" {
			speedVal = interfaceData.Get("currentInterfaceSpeed").ClonedString()
		}
		inst.SetLabelTrimmed("speed", speedToMB(speedVal))
		inst.SetLabelTrimmed("physical_port_state", cleanLinkState(interfaceData.Get("physPortState").ClonedString()))
		inst.SetLabelTrimmed("nvme_supported", interfaceData.Get("isNVMeSupported").ClonedString())
		inst.SetLabelTrimmed("max_transmission_unit", interfaceData.Get("maximumTransmissionUnit").ClonedString())
		inst.SetLabelTrimmed("ip_address", extractPortIPAddress(iface, interfaceType, interfaceData))
	}

	h.SLogger.Debug("Processed host interfaces from API", slog.Int("count", len(results)))
}

// processDriveInterfacesFromAPI processes drive interfaces from the /interfaces?channelType=driveside API endpoint
func (h *Hardware) processDriveInterfacesFromAPI(results []gjson.Result, controllerLabelMap map[string]string) {
	mat := h.data[driveInterfaceMatrix]

	for _, iface := range results {
		interfaceRef := iface.Get("interfaceRef").ClonedString()
		if interfaceRef == "" {
			continue
		}

		controllerRef := iface.Get("controllerRef").ClonedString()
		controllerName := controllerLabelMap[controllerRef]

		interfaceType := iface.Get("ioInterfaceTypeData.interfaceType").ClonedString()
		interfaceData := iface.Get("ioInterfaceTypeData." + interfaceType)

		if !interfaceData.Exists() || interfaceData.Type == gjson.Null {
			continue
		}

		channel := interfaceData.Get("channel").ClonedString()

		key := controllerRef + "_drive_" + interfaceRef
		inst, err := mat.NewInstance(key)
		if err != nil {
			h.SLogger.Error("Failed to create drive interface instance", slogx.Err(err), slog.String("key", key))
			continue
		}

		inst.SetLabelTrimmed("controller_id", controllerRef)
		inst.SetLabelTrimmed("controller", controllerName)
		inst.SetLabelTrimmed("interface_type", interfaceType)
		inst.SetLabelTrimmed("channel", channel)
		inst.SetLabelTrimmed("current_speed", speedToMB(interfaceData.Get("currentInterfaceSpeed").ClonedString()))
		inst.SetLabelTrimmed("maximum_speed", speedToMB(interfaceData.Get("maximumInterfaceSpeed").ClonedString()))
		inst.SetLabelTrimmed("protection_info_capable", interfaceData.Get("protectionInformationCapable").ClonedString())
	}

	h.SLogger.Debug("Processed drive interfaces from API", slog.Int("count", len(results)))
}

// processCodeVersions processes the codeVersions array within a controller
func (h *Hardware) processCodeVersions(controller gjson.Result, controllerID, controllerLocation string) {
	mat := h.data[codeVersionMatrix]
	versions := controller.Get("codeVersions")

	if !versions.Exists() || !versions.IsArray() {
		return
	}

	for _, ver := range versions.Array() {
		codeModule := ver.Get("codeModule").ClonedString()
		versionString := ver.Get("versionString").ClonedString()

		if codeModule == "" {
			continue
		}

		key := controllerID + "_" + codeModule
		inst, err := mat.NewInstance(key)
		if err != nil {
			h.SLogger.Error("Failed to create code version instance", slogx.Err(err), slog.String("key", key))
			continue
		}

		inst.SetLabelTrimmed("controller_id", controllerID)
		inst.SetLabelTrimmed("controller", controllerLocation)
		inst.SetLabelTrimmed("code_module", codeModule)
		inst.SetLabelTrimmed("version", versionString)
	}
}

// processDNSProperties processes the dnsProperties within a controller
func (h *Hardware) processDNSProperties(controller gjson.Result, controllerID, controllerLocation string) {
	mat := h.data[dnsPropertyMatrix]
	dnsProps := controller.Get("dnsProperties")

	if !dnsProps.Exists() {
		return
	}

	// DNS properties structure: {"acquisitionProperties": {"dnsAcquisitionType": "...", "dnsServers": [...]}}
	acquisitionType := dnsProps.Get("acquisitionProperties.dnsAcquisitionType").ClonedString()
	dnsServers := dnsProps.Get("acquisitionProperties.dnsServers")

	if !dnsServers.Exists() || !dnsServers.IsArray() {
		return
	}

	for _, server := range dnsServers.Array() {
		addressType := server.Get("addressType").ClonedString()
		var dnsServer string

		switch addressType {
		case "ipv4":
			dnsServer = server.Get("ipv4Address").ClonedString()
		case "ipv6":
			dnsServer = server.Get("ipv6Address").ClonedString()
		}

		if dnsServer == "" {
			continue
		}

		key := controllerID + "_" + dnsServer
		inst, err := mat.NewInstance(key)
		if err != nil {
			h.SLogger.Error("Failed to create DNS instance", slogx.Err(err), slog.String("key", key))
			continue
		}

		inst.SetLabelTrimmed("controller_id", controllerID)
		inst.SetLabelTrimmed("controller", controllerLocation)
		inst.SetLabelTrimmed("dns_server", dnsServer)
		inst.SetLabelTrimmed("address_type", addressType)
		inst.SetLabelTrimmed("acquisition_type", acquisitionType)
	}
}

// formatMacAddress formats a MAC address from "D039EADCD97C" to "D0:39:EA:DC:D9:7C"
func formatMacAddress(mac string) string {
	if len(mac) != 12 {
		return mac
	}
	parts := make([]string, 6)
	for i := range 6 {
		parts[i] = mac[i*2 : i*2+2]
	}
	return strings.Join(parts, ":")
}

// formatIPv6Address converts a hex IPv6 address to standard notation
// e.g., "FE80000000000000D239EAFFFEDCD97C" -> "fe80::d239:eaff:fedc:d97c"
// Follows RFC 5952 for IPv6 address compression
func formatIPv6Address(hexAddr string) string {
	if len(hexAddr) != 32 {
		return hexAddr
	}

	hexAddr = strings.ToLower(hexAddr)
	groups := make([]string, 8)
	for i := range 8 {
		groups[i] = strings.TrimLeft(hexAddr[i*4:i*4+4], "0")
		if groups[i] == "" {
			groups[i] = "0"
		}
	}

	var longestStart, longestLen, currentStart, currentLen int
	for i, group := range groups {
		if group == "0" {
			if currentLen == 0 {
				currentStart = i
			}
			currentLen++
		} else {
			if currentLen > longestLen {
				longestStart = currentStart
				longestLen = currentLen
			}
			currentLen = 0
		}
	}
	if currentLen > longestLen {
		longestStart = currentStart
		longestLen = currentLen
	}

	if longestLen >= 2 {
		before := groups[:longestStart]
		after := groups[longestStart+longestLen:]

		switch {
		case longestStart == 0 && longestStart+longestLen == 8:
			return "::"
		case longestStart == 0:
			return "::" + strings.Join(after, ":")
		case longestStart+longestLen == 8:
			return strings.Join(before, ":") + "::"
		default:
			return strings.Join(before, ":") + "::" + strings.Join(after, ":")
		}
	}

	return strings.Join(groups, ":")
}

// processNetInterfaces processes the netInterfaces array from controller data (management ports)
func (h *Hardware) processNetInterfaces(controller gjson.Result, controllerID, controllerLocation string) {
	mat := h.data[netInterfaceMatrix]
	netInterfaces := controller.Get("netInterfaces")

	if !netInterfaces.Exists() || !netInterfaces.IsArray() {
		return
	}

	remoteAccessEnabled := controller.Get("networkSettings.remoteAccessEnabled").ClonedString()

	for _, netIface := range netInterfaces.Array() {
		ethernet := netIface.Get("ethernet")
		if !ethernet.Exists() {
			continue
		}

		interfaceName := ethernet.Get("interfaceName").ClonedString()
		if interfaceName == "" {
			continue
		}

		key := controllerID + "_" + interfaceName
		inst, err := mat.NewInstance(key)
		if err != nil {
			h.SLogger.Error("Failed to create net interface instance", slogx.Err(err), slog.String("key", key))
			continue
		}

		inst.SetLabelTrimmed("controller_id", controllerID)
		inst.SetLabelTrimmed("controller", controllerLocation)
		inst.SetLabelTrimmed("interface_name", interfaceName)
		inst.SetLabelTrimmed("alias", ethernet.Get("alias").ClonedString())
		inst.SetLabelTrimmed("port", ethernet.Get("physicalLocation.label").ClonedString())
		inst.SetLabelTrimmed("mac_address", formatMacAddress(ethernet.Get("macAddr").ClonedString()))
		inst.SetLabelTrimmed("link_status", ethernet.Get("linkStatus").ClonedString())

		inst.SetLabelTrimmed("ipv4_enabled", ethernet.Get("ipv4Enabled").ClonedString())
		inst.SetLabelTrimmed("ipv4_address", ethernet.Get("ipv4Address").ClonedString())
		inst.SetLabelTrimmed("ipv4_subnet_mask", ethernet.Get("ipv4SubnetMask").ClonedString())
		inst.SetLabelTrimmed("ipv4_gateway", ethernet.Get("ipv4GatewayAddress").ClonedString())

		ipv4ConfigMethod := strings.TrimPrefix(ethernet.Get("ipv4AddressConfigMethod").ClonedString(), "config")
		inst.SetLabelTrimmed("ipv4_config_method", ipv4ConfigMethod)

		inst.SetLabelTrimmed("ipv6_enabled", ethernet.Get("ipv6Enabled").ClonedString())

		ipv6LocalHex := ethernet.Get("ipv6LocalAddress.address").ClonedString()
		if ipv6LocalHex != "" && ipv6LocalHex != "00000000000000000000000000000000" {
			inst.SetLabelTrimmed("ipv6_local_address", formatIPv6Address(ipv6LocalHex))
		}

		ipv6RoutableHex := ethernet.Get("ipv6PortStaticRoutableAddress.address").ClonedString()
		if ipv6RoutableHex != "" && ipv6RoutableHex != "00000000000000000000000000000000" {
			inst.SetLabelTrimmed("ipv6_routable_address", formatIPv6Address(ipv6RoutableHex))
		}

		// Clean up IPv6 config method (e.g., "configStateless" -> "Stateless")
		ipv6ConfigMethod := strings.TrimPrefix(ethernet.Get("ipv6AddressConfigMethod").ClonedString(), "config")
		inst.SetLabelTrimmed("ipv6_config_method", ipv6ConfigMethod)

		// Duplex mode
		fullDuplex := ethernet.Get("fullDuplex").Bool()
		if fullDuplex {
			inst.SetLabelTrimmed("full_duplex", "true")
		} else {
			inst.SetLabelTrimmed("full_duplex", "false")
		}

		// Speed settings (e.g., "speedAutoNegotiated" -> "AutoNegotiated")
		configuredSpeed := strings.TrimPrefix(ethernet.Get("configuredSpeedSetting").ClonedString(), "speed")
		inst.SetLabelTrimmed("configured_speed", configuredSpeed)

		inst.SetLabelTrimmed("current_speed", speedToMB(ethernet.Get("currentSpeed").ClonedString()))

		inst.SetLabelTrimmed("remote_access_enabled", remoteAccessEnabled)

		dnsAcqType := ethernet.Get("dnsProperties.acquisitionProperties.dnsAcquisitionType").ClonedString()
		inst.SetLabelTrimmed("dns_config_method", cleanConfigType(dnsAcqType))

		dnsServers := ethernet.Get("dnsProperties.acquisitionProperties.dnsServers")
		if dnsServers.Exists() && dnsServers.IsArray() {
			serverArray := dnsServers.Array()
			if len(serverArray) > 0 {
				primaryServer := getDNSServerAddress(serverArray[0])
				inst.SetLabelTrimmed("primary_dns_server", primaryServer)
			}
			if len(serverArray) > 1 {
				backupServer := getDNSServerAddress(serverArray[1])
				inst.SetLabelTrimmed("backup_dns_server", backupServer)
			}
		}

		ntpAcqType := ethernet.Get("ntpProperties.acquisitionProperties.ntpAcquisitionType").ClonedString()
		inst.SetLabelTrimmed("ntp_service", cleanConfigType(ntpAcqType))
	}
}

// initCacheMemoryDimmMatrix creates the matrix for cache memory DIMM data
func (h *Hardware) initCacheMemoryDimmMatrix() {
	mat := matrix.New(h.Parent+"."+cacheMemoryDimmMatrix, cacheMemoryDimmMatrix, cacheMemoryDimmMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"dimm_id", "controller"},
		[]string{"dimm_id", "controller_id", "controller", "slot", "status", "serial_number", "part_number", "manufacturer_part_number", "manufacturer"},
	))

	_, _ = mat.NewMetricFloat64("capacity")

	h.data[cacheMemoryDimmMatrix] = mat
}

// initCacheBackupDeviceMatrix creates the matrix for cache backup device data
func (h *Hardware) initCacheBackupDeviceMatrix() {
	mat := matrix.New(h.Parent+"."+cacheBackupDeviceMatrix, cacheBackupDeviceMatrix, cacheBackupDeviceMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"device_id", "controller"},
		[]string{"device_id", "controller_id", "slot", "status", "device_type", "serial_number", "part_number", "product_id", "manufacturer"},
	))

	_, _ = mat.NewMetricFloat64("capacity")

	h.data[cacheBackupDeviceMatrix] = mat
}

// processCacheMemoryDimms processes the cacheMemoryDimms array from hardware-inventory
func (h *Hardware) processCacheMemoryDimms(response gjson.Result, controllerLabelMap map[string]string) {
	mat := h.data[cacheMemoryDimmMatrix]
	dimms := response.Get("cacheMemoryDimms")

	if !dimms.Exists() || !dimms.IsArray() {
		h.SLogger.Debug("No cacheMemoryDimms found in response")
		return
	}

	capacityMetric := mat.MustGetMetric("capacity")

	for _, dimm := range dimms.Array() {
		dimmRef := dimm.Get("cacheMemoryDimmRef").ClonedString()
		if dimmRef == "" {
			continue
		}

		inst, err := mat.NewInstance(dimmRef)
		if err != nil {
			h.SLogger.Error("Failed to create cache memory DIMM instance", slogx.Err(err), slog.String("id", dimmRef))
			continue
		}

		controllerRef := dimm.Get("physicalLocation.locationParent.controllerRef").ClonedString()
		controllerLabel := controllerLabelMap[controllerRef]

		inst.SetLabelTrimmed("dimm_id", dimmRef)
		inst.SetLabelTrimmed("controller_id", controllerRef)
		inst.SetLabelTrimmed("controller", controllerLabel)
		inst.SetLabelTrimmed("slot", dimm.Get("physicalLocation.label").ClonedString())
		inst.SetLabelTrimmed("status", dimm.Get("status").ClonedString())
		if val := dimm.Get("capacityInMegabytes"); val.Exists() {
			capacityMetric.SetValueFloat64(inst, float64(val.Uint()*1024*1024))
		}
		inst.SetLabelTrimmed("serial_number", dimm.Get("serialNumber").ClonedString())
		inst.SetLabelTrimmed("part_number", dimm.Get("partNumber").ClonedString())
		inst.SetLabelTrimmed("manufacturer_part_number", dimm.Get("manufacturerPartNumber").ClonedString())
		inst.SetLabelTrimmed("manufacturer", dimm.Get("manufacturer").ClonedString())
	}

	h.SLogger.Debug("Processed cache memory DIMMs", slog.Int("count", len(dimms.Array())))
}

// processCacheBackupDevices processes the cacheBackupDevices array from hardware-inventory
func (h *Hardware) processCacheBackupDevices(response gjson.Result, controllerLabelMap map[string]string) {
	mat := h.data[cacheBackupDeviceMatrix]
	devices := response.Get("cacheBackupDevices")

	if !devices.Exists() || !devices.IsArray() {
		h.SLogger.Debug("No cacheBackupDevices found in response")
		return
	}

	capacityMetric := mat.MustGetMetric("capacity")

	for _, device := range devices.Array() {
		deviceRef := device.Get("backupDeviceRef").ClonedString()
		if deviceRef == "" {
			deviceRef = device.Get("id").ClonedString()
		}
		if deviceRef == "" {
			continue
		}

		inst, err := mat.NewInstance(deviceRef)
		if err != nil {
			h.SLogger.Error("Failed to create cache backup device instance", slogx.Err(err), slog.String("id", deviceRef))
			continue
		}

		controllerRef := device.Get("parentController").ClonedString()
		controllerLabel := controllerLabelMap[controllerRef]

		inst.SetLabelTrimmed("device_id", deviceRef)
		inst.SetLabelTrimmed("controller_id", controllerRef)
		inst.SetLabelTrimmed("controller", controllerLabel)
		inst.SetLabelTrimmed("slot", device.Get("physicalLocation.label").ClonedString())
		inst.SetLabelTrimmed("status", device.Get("backupDeviceStatus").ClonedString())
		inst.SetLabelTrimmed("device_type", device.Get("backupDeviceType").ClonedString())
		if val := device.Get("backupDeviceCapacity"); val.Exists() {
			capacityMetric.SetValueFloat64(inst, float64(val.Uint()*1024*1024))
		}
		inst.SetLabelTrimmed("serial_number", device.Get("backupDeviceVpd.serialNumber").ClonedString())
		inst.SetLabelTrimmed("part_number", device.Get("backupDeviceVpd.partNumber").ClonedString())
		inst.SetLabelTrimmed("product_id", device.Get("backupDeviceVpd.productId").ClonedString())
		inst.SetLabelTrimmed("manufacturer", device.Get("backupDeviceVpd.manufacturer").ClonedString())
	}

	h.SLogger.Debug("Processed cache backup devices", slog.Int("count", len(devices.Array())))
}
