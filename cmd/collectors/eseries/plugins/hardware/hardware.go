/*
 * Copyright NetApp Inc, 2025 All rights reserved
 */

package hardware

import (
	"log/slog"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// Matrix names for hardware components
const (
	batteryMatrix           = "eseries_battery"
	controllerMatrix        = "eseries_controller"
	hostInterfaceMatrix     = "eseries_controller_host_interface"
	driveInterfaceMatrix    = "eseries_controller_drive_interface"
	codeVersionMatrix       = "eseries_controller_code_version"
	dnsPropertyMatrix       = "eseries_controller_dns"
	netInterfaceMatrix      = "eseries_controller_net_interface"
	cacheMemoryDimmMatrix   = "eseries_cache_memory_dimm"
	cacheBackupDeviceMatrix = "eseries_cache_backup_device"
	driveMatrix             = "eseries_drive"
	fanMatrix               = "eseries_fan"
	powerSupplyMatrix       = "eseries_power_supply"
	sfpMatrix               = "eseries_sfp"
	thermalSensorMatrix     = "eseries_thermal_sensor"
)

// Hardware is the main plugin that processes all hardware-inventory data
type Hardware struct {
	*plugin.AbstractPlugin
	client *rest.Client
	data   map[string]*matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Hardware{AbstractPlugin: p}
}

// Init initializes the plugin, creates REST client and all matrices
func (h *Hardware) Init(_ conf.Remote) error {
	if err := h.InitAbc(); err != nil {
		return err
	}

	if err := h.initClient(); err != nil {
		return err
	}

	h.data = make(map[string]*matrix.Matrix)
	h.initBatteryMatrix()
	h.initControllerMatrix()
	h.initHostInterfaceMatrix()
	h.initDriveInterfaceMatrix()
	h.initCodeVersionMatrix()
	h.initDNSMatrix()
	h.initNetInterfaceMatrix()
	h.initCacheMemoryDimmMatrix()
	h.initCacheBackupDeviceMatrix()
	h.initDriveMatrix()
	h.initFanMatrix()
	h.initPowerSupplyMatrix()
	h.initSFPMatrix()
	h.initThermalSensorMatrix()

	return nil
}

// initClient initializes the REST client for making API calls
func (h *Hardware) initClient() error {
	var err error

	clientTimeout := h.ParentParams.GetChildContentS("client_timeout")
	if clientTimeout == "" {
		clientTimeout = rest.DefaultTimeout
	}

	duration, err := time.ParseDuration(clientTimeout)
	if err != nil {
		h.SLogger.Info("Using default timeout", slog.String("timeout", rest.DefaultTimeout))
		duration, _ = time.ParseDuration(rest.DefaultTimeout)
	}

	poller, err := conf.PollerNamed(h.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, h.SLogger)

	if h.client, err = rest.New(poller, duration, credentials, ""); err != nil {
		return err
	}

	if err := h.client.Init(1, conf.Remote{}); err != nil {
		return err
	}

	return nil
}

// Run fetches hardware-inventory data and processes all components
func (h *Hardware) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[h.Object]
	globalLabels := data.GetGlobalLabels()

	arrayID := h.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		h.SLogger.Error("array_id not found in parent params")
		return nil, nil, nil
	}

	for _, mat := range h.data {
		mat.PurgeInstances()
		mat.Reset()
		mat.SetGlobalLabels(globalLabels)
	}

	query := "storage-systems/" + arrayID + "/hardware-inventory"
	results, err := h.client.Fetch(h.client.APIPath+"/"+query, nil)
	if err != nil {
		h.SLogger.Error("Failed to fetch hardware-inventory", slogx.Err(err))
		return nil, nil, err
	}

	if len(results) == 0 {
		h.SLogger.Warn("No hardware-inventory data returned")
		return nil, nil, nil
	}

	response := results[0]

	portLabelMap := h.buildPortLabelMap(response)

	controllerLabelMap := h.buildControllerLabelMap(response)

	trayLabelMap := h.buildTrayLabelMap(response)

	poolNames := h.buildPoolLabelMap(arrayID)

	// Process all hardware components from hardware-inventory
	h.processBatteries(response, controllerLabelMap)
	h.processControllers(response, portLabelMap)
	h.processCacheMemoryDimms(response, controllerLabelMap)
	h.processCacheBackupDevices(response, controllerLabelMap)
	h.processDrives(response, trayLabelMap, poolNames)
	h.processFans(response, controllerLabelMap)
	h.processPowerSupplies(response)
	h.processSFPs(response, controllerLabelMap, portLabelMap)
	h.processThermalSensors(response)

	hostInterfaceQuery := "storage-systems/" + arrayID + "/interfaces?interfaceType=&channelType=hostside"
	hostInterfaceResults, err := h.client.Fetch(h.client.APIPath+"/"+hostInterfaceQuery, nil)
	if err != nil {
		h.SLogger.Warn("Failed to fetch host interfaces", slogx.Err(err))
	} else if len(hostInterfaceResults) > 0 {
		h.processHostInterfacesFromAPI(hostInterfaceResults, controllerLabelMap, portLabelMap)
	}

	driveInterfaceQuery := "storage-systems/" + arrayID + "/interfaces?interfaceType=&channelType=driveside"
	driveInterfaceResults, err := h.client.Fetch(h.client.APIPath+"/"+driveInterfaceQuery, nil)
	if err != nil {
		h.SLogger.Warn("Failed to fetch drive interfaces", slogx.Err(err))
	} else if len(driveInterfaceResults) > 0 {
		h.processDriveInterfacesFromAPI(driveInterfaceResults, controllerLabelMap)
	}

	matrices := make([]*matrix.Matrix, 0, len(h.data))
	totalInstances := 0
	for _, mat := range h.data {
		matrices = append(matrices, mat)
		totalInstances += len(mat.GetInstances())
	}

	h.SLogger.Info(
		"Hardware plugin",
		slog.Int("totalInstances", totalInstances),
		slog.Int("batteries", len(h.data[batteryMatrix].GetInstances())),
		slog.Int("controllers", len(h.data[controllerMatrix].GetInstances())),
		slog.Int("hostInterfaces", len(h.data[hostInterfaceMatrix].GetInstances())),
		slog.Int("driveInterfaces", len(h.data[driveInterfaceMatrix].GetInstances())),
		slog.Int("codeVersions", len(h.data[codeVersionMatrix].GetInstances())),
		slog.Int("netInterfaces", len(h.data[netInterfaceMatrix].GetInstances())),
		slog.Int("cacheMemoryDimms", len(h.data[cacheMemoryDimmMatrix].GetInstances())),
		slog.Int("cacheBackupDevices", len(h.data[cacheBackupDeviceMatrix].GetInstances())),
		slog.Int("drives", len(h.data[driveMatrix].GetInstances())),
		slog.Int("fans", len(h.data[fanMatrix].GetInstances())),
		slog.Int("powerSupplies", len(h.data[powerSupplyMatrix].GetInstances())),
		slog.Int("sfps", len(h.data[sfpMatrix].GetInstances())),
		slog.Int("thermalSensors", len(h.data[thermalSensorMatrix].GetInstances())),
	)

	metadata := &collector.Metadata{}
	//nolint:gosec
	metadata.PluginInstances = uint64(totalInstances)

	return matrices, metadata, nil
}

// buildPortLabelMap creates a map of channel number to port label (e.g., "1" -> "1a")
// from the channelPorts array in hardware-inventory
func (h *Hardware) buildPortLabelMap(response gjson.Result) map[string]string {
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

	h.SLogger.Debug("Built port label map", slog.Int("entries", len(portLabelMap)))
	return portLabelMap
}

// buildControllerLabelMap creates a map of controllerRef to controller label (e.g., "070000..." -> "A")
// from the controllers array in hardware-inventory
func (h *Hardware) buildControllerLabelMap(response gjson.Result) map[string]string {
	controllerLabelMap := make(map[string]string)

	controllers := response.Get("controllers")
	if !controllers.Exists() || !controllers.IsArray() {
		return controllerLabelMap
	}

	for _, controller := range controllers.Array() {
		controllerRef := controller.Get("controllerRef").ClonedString()
		if controllerRef == "" {
			controllerRef = controller.Get("id").ClonedString()
		}

		label := controller.Get("physicalLocation.label").ClonedString()

		if controllerRef != "" && label != "" {
			controllerLabelMap[controllerRef] = label
		}
	}

	h.SLogger.Debug("Built controller label map", slog.Int("entries", len(controllerLabelMap)))
	return controllerLabelMap
}
