/*
 * Copyright NetApp Inc, 2025 All rights reserved
 */

package hardware

import (
	"fmt"
	"log/slog"

	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

// supportCRUInfo holds information about a support CRU (fan/power supply canister)
type supportCRUInfo struct {
	controllerRef string
	cruType       string
	label         string
}

// initFanMatrix creates the matrix for fan data
func (h *Hardware) initFanMatrix() {
	mat := matrix.New(h.Parent+"."+fanMatrix, fanMatrix, fanMatrix)
	exportOptions := node.NewS("export_options")

	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "id")

	instanceLabels := exportOptions.NewChildS("instance_labels", "")
	instanceLabels.NewChildS("", "location")
	instanceLabels.NewChildS("", "status")
	instanceLabels.NewChildS("", "slot")
	instanceLabels.NewChildS("", "controller")
	instanceLabels.NewChildS("", "parent_type")
	instanceLabels.NewChildS("", "fan_number")

	mat.SetExportOptions(exportOptions)

	h.data[fanMatrix] = mat
}

// buildSupportCRUMap creates a map of supportCRU reference to its info
// This is used to look up parent controller/canister information for fans and power supplies
func (h *Hardware) buildSupportCRUMap(response gjson.Result) map[string]supportCRUInfo {
	cruMap := make(map[string]supportCRUInfo)

	supportCRUs := response.Get("supportCRUs")
	if !supportCRUs.Exists() || !supportCRUs.IsArray() {
		return cruMap
	}

	for _, cru := range supportCRUs.Array() {
		ref := cru.Get("supportCRURef").ClonedString()
		if ref == "" {
			ref = cru.Get("id").ClonedString()
		}
		if ref == "" {
			continue
		}

		info := supportCRUInfo{
			controllerRef: cru.Get("physicalLocation.locationParent.controllerRef").ClonedString(),
			cruType:       cru.Get("type").ClonedString(),
			label:         cru.Get("physicalLocation.label").ClonedString(),
		}

		cruMap[ref] = info
	}

	h.SLogger.Debug("Built supportCRU map", slog.Int("entries", len(cruMap)))
	return cruMap
}

// processFans processes the fans array from hardware-inventory
func (h *Hardware) processFans(response gjson.Result, controllerLabelMap map[string]string) {
	mat := h.data[fanMatrix]
	fans := response.Get("fans")

	if !fans.Exists() || !fans.IsArray() {
		h.SLogger.Debug("No fans found in response")
		return
	}

	// Build supportCRU lookup map
	supportCRUMap := h.buildSupportCRUMap(response)

	for _, fan := range fans.Array() {
		id := fan.Get("id").ClonedString()
		if id == "" {
			id = fan.Get("fanRef").ClonedString()
		}
		if id == "" {
			continue
		}

		inst, err := mat.NewInstance(id)
		if err != nil {
			h.SLogger.Error("Failed to create fan instance", slogx.Err(err), slog.String("id", id))
			continue
		}

		// Set basic labels
		inst.SetLabelTrimmed("id", id)
		inst.SetLabelTrimmed("status", fan.Get("status").ClonedString())
		slot := fan.Get("physicalLocation.slot").ClonedString()
		inst.SetLabelTrimmed("slot", slot)

		parentSupportCRU := fan.Get("rtrAttributes.parentCru.parentSupportCru").ClonedString()

		var controllerLabel string
		var parentType string
		fanNumber := fan.Get("physicalLocation.locationPosition").ClonedString()
		if fanNumber == "" {
			fanNumber = fan.Get("physicalLocation.label").ClonedString()
		}

		if parentSupportCRU != "" {
			if cruInfo, ok := supportCRUMap[parentSupportCRU]; ok {
				if cruInfo.controllerRef != "" {
					if label, ok := controllerLabelMap[cruInfo.controllerRef]; ok {
						controllerLabel = label
					}
				}

				switch cruInfo.cruType {
				case "fan":
					parentType = "controller"
				case "powerFan":
					parentType = "power_fan_canister"
				default:
					parentType = cruInfo.cruType
				}
			}
		}

		// Build the location label similar to UI: "Controller A, Fan 1" or "Power/fan canister 1, Fan 1"
		var location string
		switch {
		case parentType == "controller" && controllerLabel != "":
			location = fmt.Sprintf("Controller %s, Fan %s", controllerLabel, fanNumber)
		case parentType == "power_fan_canister":
			location = fmt.Sprintf("Power/fan canister %s, Fan %s", slot, fanNumber)
		default:
			location = fan.Get("physicalLocation.label").ClonedString()
		}

		inst.SetLabelTrimmed("location", location)
		inst.SetLabelTrimmed("controller", controllerLabel)
		inst.SetLabelTrimmed("parent_type", parentType)
		inst.SetLabelTrimmed("fan_number", fanNumber)
	}

	h.SLogger.Debug("Processed fans", slog.Int("count", len(mat.GetInstances())))
}
