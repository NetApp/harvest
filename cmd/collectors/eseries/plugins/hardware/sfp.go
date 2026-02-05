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

// initSFPMatrix creates the matrix for SFP data
func (h *Hardware) initSFPMatrix() {
	mat := matrix.New(h.Parent+"."+sfpMatrix, sfpMatrix, sfpMatrix)
	exportOptions := node.NewS("export_options")

	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "id")

	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceLabels.NewChildS("", "controller")
	instanceLabels.NewChildS("", "port")
	instanceLabels.NewChildS("", "status")
	instanceLabels.NewChildS("", "part_number")
	instanceLabels.NewChildS("", "serial_number")
	instanceLabels.NewChildS("", "vendor")

	mat.SetExportOptions(exportOptions)

	h.data[sfpMatrix] = mat
}

// processSFPs processes the sfps array from hardware-inventory
func (h *Hardware) processSFPs(response gjson.Result, controllerLabelMap map[string]string, portLabelMap map[string]string) {
	mat := h.data[sfpMatrix]
	sfps := response.Get("sfps")

	if !sfps.Exists() || !sfps.IsArray() {
		h.SLogger.Debug("No SFPs found in response")
		return
	}

	for _, sfp := range sfps.Array() {
		id := sfp.Get("id").ClonedString()
		if id == "" {
			id = sfp.Get("sfpRef").ClonedString()
		}
		if id == "" {
			continue
		}

		inst, err := mat.NewInstance(id)
		if err != nil {
			h.SLogger.Error("Failed to create SFP instance", slogx.Err(err), slog.String("id", id))
			continue
		}

		inst.SetLabelTrimmed("id", id)
		inst.SetLabelTrimmed("status", sfp.Get("status").ClonedString())
		inst.SetLabelTrimmed("part_number", sfp.Get("sfpType.vendorPN").ClonedString())
		inst.SetLabelTrimmed("serial_number", sfp.Get("sfpType.vendorSN").ClonedString())
		inst.SetLabelTrimmed("vendor", sfp.Get("sfpType.vendorName").ClonedString())

		// Get controller label from parentController
		controllerRef := sfp.Get("parentData.controllerSFP.parentController").ClonedString()
		var controllerLabel string
		if label, ok := controllerLabelMap[controllerRef]; ok {
			controllerLabel = label
		}
		inst.SetLabelTrimmed("controller", controllerLabel)

		// Get port label from channel using portLabelMap
		channel := sfp.Get("parentData.controllerSFP.channel").ClonedString()
		var portLabel string
		if label, ok := portLabelMap[channel]; ok {
			portLabel = label
		}
		inst.SetLabelTrimmed("port", portLabel)

	}

	h.SLogger.Debug("Processed SFPs", slog.Int("count", len(mat.GetInstances())))
}
