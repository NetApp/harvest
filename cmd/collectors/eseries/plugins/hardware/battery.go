/*
 * Copyright NetApp Inc, 2025 All rights reserved
 */

package hardware

import (
	"log/slog"

	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

func (h *Hardware) initBatteryMatrix() {
	mat := matrix.New(h.Parent+"."+batteryMatrix, batteryMatrix, batteryMatrix)
	exportOptions := node.NewS("export_options")

	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "id")

	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceLabels.NewChildS("", "fru_type")
	instanceLabels.NewChildS("", "location")
	instanceLabels.NewChildS("", "status")
	instanceLabels.NewChildS("", "vendor")
	instanceLabels.NewChildS("", "vendor_part_number")
	instanceLabels.NewChildS("", "vendor_serial_number")
	instanceLabels.NewChildS("", "controller")

	mat.SetExportOptions(exportOptions)

	h.data[batteryMatrix] = mat
}

func (h *Hardware) processBatteries(response gjson.Result, controllerLabelMap map[string]string) {
	mat := h.data[batteryMatrix]
	batteries := response.Get("batteries")

	if !batteries.Exists() || !batteries.IsArray() {
		h.SLogger.Debug("No batteries found in response")
		return
	}

	for _, battery := range batteries.Array() {
		id := battery.Get("id").ClonedString()
		if id == "" {
			id = battery.Get("batteryRef").ClonedString()
		}
		if id == "" {
			continue
		}

		inst, err := mat.NewInstance(id)
		if err != nil {
			h.SLogger.Error("Failed to create battery instance", slogx.Err(err), slog.String("id", id))
			continue
		}

		inst.SetLabelTrimmed("id", id)
		inst.SetLabelTrimmed("fru_type", battery.Get("fruType").ClonedString())
		inst.SetLabelTrimmed("location", battery.Get("physicalLocation.label").ClonedString())
		inst.SetLabelTrimmed("status", battery.Get("status").ClonedString())
		inst.SetLabelTrimmed("vendor", battery.Get("vendorName").ClonedString())
		inst.SetLabelTrimmed("vendor_part_number", battery.Get("vendorPN").ClonedString())
		inst.SetLabelTrimmed("vendor_serial_number", battery.Get("vendorSN").ClonedString())

		controllerRef := battery.Get("batteryTypeData.parentController").ClonedString()
		if controllerLabel, ok := controllerLabelMap[controllerRef]; ok {
			inst.SetLabelTrimmed("controller", controllerLabel)
		}
	}

	h.SLogger.Debug("Processed batteries", slog.Int("count", len(mat.GetInstances())))
}
