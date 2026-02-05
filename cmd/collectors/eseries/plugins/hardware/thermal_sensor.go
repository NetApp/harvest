/*
 * Copyright NetApp Inc, 2025 All rights reserved
 */

package hardware

import (
	"log/slog"

	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// initThermalSensorMatrix initializes the thermal sensor matrix
func (h *Hardware) initThermalSensorMatrix() {
	mat := matrix.New(h.Parent+"."+thermalSensorMatrix, thermalSensorMatrix, thermalSensorMatrix)
	exportOptions := node.NewS("export_options")

	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "id")

	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceLabels.NewChildS("", "status")
	instanceLabels.NewChildS("", "slot")
	instanceLabels.NewChildS("", "location")

	mat.SetExportOptions(exportOptions)

	h.data[thermalSensorMatrix] = mat
}

// processThermalSensors processes the thermalSensors array from hardware-inventory
func (h *Hardware) processThermalSensors(response gjson.Result) {
	mat := h.data[thermalSensorMatrix]
	thermalSensors := response.Get("thermalSensors")

	if !thermalSensors.Exists() || !thermalSensors.IsArray() {
		h.SLogger.Debug("No thermalSensors found in response")
		return
	}

	for _, sensor := range thermalSensors.Array() {
		sensorID := sensor.Get("id").ClonedString()
		if sensorID == "" {
			sensorID = sensor.Get("thermalSensorRef").ClonedString()
		}
		if sensorID == "" {
			continue
		}

		inst, err := mat.NewInstance(sensorID)
		if err != nil {
			h.SLogger.Warn("Failed to create thermal sensor instance", slog.String("id", sensorID))
			continue
		}

		inst.SetLabelTrimmed("id", sensorID)
		inst.SetLabelTrimmed("status", sensor.Get("status").ClonedString())
		inst.SetLabelTrimmed("slot", sensor.Get("physicalLocation.slot").ClonedString())
		inst.SetLabelTrimmed("location", sensor.Get("physicalLocation.label").ClonedString())
	}

	h.SLogger.Debug("Processed thermal sensors", slog.Int("count", len(mat.GetInstances())))
}
