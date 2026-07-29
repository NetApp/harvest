/*
 * Copyright NetApp Inc, 2025 All rights reserved
 */

package hardware

import (
	"log/slog"

	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// initPowerSupplyMatrix creates the matrix for power supply data
func (h *Hardware) initPowerSupplyMatrix() {
	mat := matrix.New(h.Parent+"."+powerSupplyMatrix, powerSupplyMatrix, powerSupplyMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"id"},
		[]string{"firmware_version", "fru_type", "part_number", "serial_number", "status", "vendor", "slot"},
	))

	h.data[powerSupplyMatrix] = mat
}

// processPowerSupplies processes the powerSupplies array from hardware-inventory
func (h *Hardware) processPowerSupplies(response gjson.Result) {
	mat := h.data[powerSupplyMatrix]
	powerSupplies := response.Get("powerSupplies")

	if !powerSupplies.Exists() || !powerSupplies.IsArray() {
		h.SLogger.Debug("No power supplies found in response")
		return
	}

	for _, ps := range powerSupplies.Array() {
		id := ps.Get("id").ClonedString()
		if id == "" {
			id = ps.Get("powerSupplyRef").ClonedString()
		}
		if id == "" {
			continue
		}

		inst, err := mat.NewInstance(id)
		if err != nil {
			h.SLogger.Error("Failed to create power supply instance", slogx.Err(err), slog.String("id", id))
			continue
		}

		inst.SetLabelTrimmed("id", id)
		inst.SetLabelTrimmed("firmware_version", ps.Get("firmwareRevision").ClonedString())
		inst.SetLabelTrimmed("fru_type", ps.Get("fruType").ClonedString())
		inst.SetLabelTrimmed("part_number", ps.Get("partNumber").ClonedString())
		inst.SetLabelTrimmed("serial_number", ps.Get("serialNumber").ClonedString())
		inst.SetLabelTrimmed("status", ps.Get("status").ClonedString())
		inst.SetLabelTrimmed("vendor", ps.Get("vendorName").ClonedString())
		inst.SetLabelTrimmed("slot", ps.Get("physicalLocation.slot").ClonedString())
	}

	h.SLogger.Debug("Processed power supplies", slog.Int("count", len(mat.GetInstances())))
}
