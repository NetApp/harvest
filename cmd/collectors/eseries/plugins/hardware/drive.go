/*
 * Copyright NetApp Inc, 2025 All rights reserved
 */

package hardware

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// boolToString converts a boolean to "true" or "false" string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// formatWWID formats a WWID from "38483830598005500025384700000001" to "38:48:38:30:59:80:05:50:00:25:38:47:00:00:00:01"
func formatWWID(wwid string) string {
	if len(wwid) != 32 {
		return wwid
	}
	parts := make([]string, 16)
	for i := range 16 {
		parts[i] = wwid[i*2 : i*2+2]
	}
	return strings.Join(parts, ":")
}

// initDriveMatrix creates the matrix for drive data
func (h *Hardware) initDriveMatrix() {
	mat := matrix.New(h.Parent+"."+driveMatrix, driveMatrix, driveMatrix)
	exportOptions := node.NewS("export_options")

	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "id")
	instanceKeys.NewChildS("", "location")

	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceLabels.NewChildS("", "status")
	instanceLabels.NewChildS("", "mode")
	instanceLabels.NewChildS("", "assigned_to")
	instanceLabels.NewChildS("", "media_type")
	instanceLabels.NewChildS("", "interface_type")
	instanceLabels.NewChildS("", "pcie_generation")
	instanceLabels.NewChildS("", "current_speed")
	instanceLabels.NewChildS("", "max_speed")
	instanceLabels.NewChildS("", "firmware_version")
	instanceLabels.NewChildS("", "product_id")
	instanceLabels.NewChildS("", "serial_number")
	instanceLabels.NewChildS("", "manufacturer")
	instanceLabels.NewChildS("", "wwid")
	instanceLabels.NewChildS("", "tray_id")
	instanceLabels.NewChildS("", "slot")
	instanceLabels.NewChildS("", "secure_capable")
	instanceLabels.NewChildS("", "secure_enabled")
	instanceLabels.NewChildS("", "da_capable")
	instanceLabels.NewChildS("", "dulbe_capable")
	instanceLabels.NewChildS("", "fips_capable")
	instanceLabels.NewChildS("", "rw_accessible")
	instanceLabels.NewChildS("", "redundant_path")

	mat.SetExportOptions(exportOptions)

	// Define metrics
	_, _ = mat.NewMetricFloat64("percent_endurance_used")
	_, _ = mat.NewMetricFloat64("capacity")
	_, _ = mat.NewMetricFloat64("block_size")
	_, _ = mat.NewMetricFloat64("block_size_physical")

	h.data[driveMatrix] = mat
}

// processDrives processes the drives array from hardware-inventory
func (h *Hardware) processDrives(response gjson.Result, trayLabelMap, poolNames map[string]string) {
	mat := h.data[driveMatrix]
	drives := response.Get("drives")

	if !drives.Exists() || !drives.IsArray() {
		h.SLogger.Debug("No drives found in response")
		return
	}

	for _, drive := range drives.Array() {
		id := drive.Get("id").ClonedString()
		if id == "" {
			id = drive.Get("driveRef").ClonedString()
		}
		if id == "" {
			continue
		}

		inst, err := mat.NewInstance(id)
		if err != nil {
			h.SLogger.Error("Failed to create drive instance", slogx.Err(err), slog.String("id", id))
			continue
		}

		inst.SetLabelTrimmed("id", id)
		inst.SetLabelTrimmed("status", drive.Get("status").ClonedString())
		inst.SetLabelTrimmed("manufacturer", drive.Get("manufacturer").ClonedString())
		inst.SetLabelTrimmed("product_id", drive.Get("productID").ClonedString())
		inst.SetLabelTrimmed("serial_number", drive.Get("serialNumber").ClonedString())
		inst.SetLabelTrimmed("firmware_version", drive.Get("firmwareVersion").ClonedString())
		inst.SetLabelTrimmed("wwid", formatWWID(drive.Get("worldWideName").ClonedString()))

		trayRef := drive.Get("physicalLocation.trayRef").ClonedString()
		slot := drive.Get("physicalLocation.slot").ClonedString()
		label := drive.Get("physicalLocation.label").ClonedString()

		trayID := ""
		if trayRef != "" {
			if tid, ok := trayLabelMap[trayRef]; ok {
				trayID = tid
			}
		}
		inst.SetLabelTrimmed("tray_id", trayID)
		inst.SetLabelTrimmed("slot", slot)

		var location string
		switch {
		case trayID != "" && label != "":
			location = "Shelf " + trayID + ", Bay " + label
		case trayID != "" && slot != "":
			location = "Shelf " + trayID + ", Slot " + slot
		case slot != "":
			location = "Slot " + slot
		}
		inst.SetLabelTrimmed("location", location)

		mode := h.determineDriveMode(drive)
		inst.SetLabelTrimmed("mode", mode)

		volumeGroupRef := drive.Get("currentVolumeGroupRef").ClonedString()
		assignedTo := ""
		if volumeGroupRef != "" && volumeGroupRef != "0000000000000000000000000000000000000000" {
			if poolName, ok := poolNames[volumeGroupRef]; ok {
				assignedTo = poolName
			}
		}
		inst.SetLabelTrimmed("assigned_to", assignedTo)

		inst.SetLabelTrimmed("media_type", drive.Get("driveMediaType").ClonedString())

		inst.SetLabelTrimmed("interface_type", drive.Get("interfaceType.driveType").ClonedString())

		inst.SetLabelTrimmed("pcie_generation", drive.Get("pcieDriveGen").ClonedString())

		inst.SetLabelTrimmed("current_speed", speedToMB(drive.Get("currentSpeed").ClonedString()))
		inst.SetLabelTrimmed("max_speed", speedToMB(drive.Get("maxSpeed").ClonedString()))

		if endurance := drive.Get("ssdWearLife.percentEnduranceUsed"); endurance.Exists() {
			h.setMetricValue(mat, inst, "percent_endurance_used", endurance.Float())
		}

		if capacityStr := drive.Get("usableCapacity").ClonedString(); capacityStr != "" {
			if capacity, err := strconv.ParseFloat(capacityStr, 64); err == nil {
				h.setMetricValue(mat, inst, "capacity", capacity)
			}
		}

		if blkSize := drive.Get("blkSize"); blkSize.Exists() {
			h.setMetricValue(mat, inst, "block_size", blkSize.Float())
		}
		if blkSizePhysical := drive.Get("blkSizePhysical"); blkSizePhysical.Exists() {
			h.setMetricValue(mat, inst, "block_size_physical", blkSizePhysical.Float())
		}

		inst.SetLabelTrimmed("secure_capable", boolToString(drive.Get("fdeCapable").Bool()))
		inst.SetLabelTrimmed("secure_enabled", boolToString(drive.Get("fdeEnabled").Bool()))
		inst.SetLabelTrimmed("da_capable", boolToString(drive.Get("protectionInformationCapable").Bool()))
		inst.SetLabelTrimmed("dulbe_capable", boolToString(drive.Get("dulbeCapable").Bool()))
		inst.SetLabelTrimmed("fips_capable", boolToString(drive.Get("fipsCapable").Bool()))

		inst.SetLabelTrimmed("rw_accessible", boolToString(!drive.Get("offline").Bool()))
		inst.SetLabelTrimmed("redundant_path", boolToString(!drive.Get("nonRedundantAccess").Bool()))
	}
}

// determineDriveMode determines the drive mode based on status flags
func (h *Hardware) determineDriveMode(drive gjson.Result) string {
	if drive.Get("hotSpare").Bool() {
		return "hotSpare"
	}
	if drive.Get("available").Bool() {
		return "unassigned"
	}
	if drive.Get("offline").Bool() {
		return "offline"
	}

	// Check if assigned to a volume group
	volumeGroupRef := drive.Get("currentVolumeGroupRef").ClonedString()
	if volumeGroupRef != "" && volumeGroupRef != "0000000000000000000000000000000000000000" {
		return "assigned"
	}

	return "unknown"
}

// buildTrayLabelMap creates a map of trayRef to tray ID (shelf number)
func (h *Hardware) buildTrayLabelMap(response gjson.Result) map[string]string {
	trayMap := make(map[string]string)

	trays := response.Get("trays")
	if !trays.Exists() || !trays.IsArray() {
		return trayMap
	}

	for _, tray := range trays.Array() {
		trayRef := tray.Get("trayRef").ClonedString()
		if trayRef == "" {
			trayRef = tray.Get("id").ClonedString()
		}
		if trayRef == "" {
			continue
		}

		trayID := tray.Get("trayId").ClonedString()
		if trayID == "" {
			// Fall back to label if trayId not available
			trayID = tray.Get("physicalLocation.label").ClonedString()
		}

		if trayID != "" {
			trayMap[trayRef] = trayID
		}
	}

	h.SLogger.Debug("Built tray label map", slog.Int("entries", len(trayMap)))
	return trayMap
}

// buildPoolLabelMap fetches storage pools and creates a map of poolRef to pool name
func (h *Hardware) buildPoolLabelMap(arrayID string) map[string]string {
	poolNames := make(map[string]string)

	apiPath := h.client.APIPath + "/storage-systems/" + arrayID + "/storage-pools"
	pools, err := h.client.Fetch(apiPath, nil)
	if err != nil {
		h.SLogger.Warn("Failed to fetch storage pools for drive assignment", slogx.Err(err))
		return poolNames
	}

	for _, pool := range pools {
		poolRef := pool.Get("id").ClonedString()
		if poolRef == "" {
			poolRef = pool.Get("volumeGroupRef").ClonedString()
		}
		poolName := pool.Get("name").ClonedString()
		if poolName == "" {
			poolName = pool.Get("label").ClonedString()
		}
		if poolRef != "" && poolName != "" {
			poolNames[poolRef] = poolName
		}
	}

	h.SLogger.Debug("Built pool label map", slog.Int("entries", len(poolNames)))
	return poolNames
}

// setMetricValue sets a float metric value
func (h *Hardware) setMetricValue(mat *matrix.Matrix, inst *matrix.Instance, metricName string, value float64) {
	if m := mat.GetMetric(metricName); m != nil {
		m.SetValueFloat64(inst, value)
	}
}
