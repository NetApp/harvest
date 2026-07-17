package hardware

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
)

func (h *Hardware) initHostBoardMatrix() {
	mat := matrix.New(h.Parent+"."+hostBoardMatrix, hostBoardMatrix, hostBoardMatrix)
	mat.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"id"},
		[]string{"controller", "status", "type", "slot", "part_number", "serial_number", "vendor", "fru_type", "number_of_ports"},
	))

	h.data[hostBoardMatrix] = mat
}

func (h *Hardware) processHostBoards(response gjson.Result, controllerLabelMap map[string]string) {
	mat := h.data[hostBoardMatrix]
	hostBoards := response.Get("hostBoards")

	if !hostBoards.Exists() || !hostBoards.IsArray() {
		h.SLogger.Debug("No host boards found in response")
		return
	}

	for _, hb := range hostBoards.Array() {
		id := hb.Get("id").ClonedString()
		if id == "" {
			id = hb.Get("hostBoardRef").ClonedString()
		}
		if id == "" {
			continue
		}

		inst, err := mat.NewInstance(id)
		if err != nil {
			h.SLogger.Error("Failed to create host board instance", slogx.Err(err), slog.String("id", id))
			continue
		}

		inst.SetLabelTrimmed("id", id)
		inst.SetLabelTrimmed("status", hb.Get("status").ClonedString())
		inst.SetLabelTrimmed("type", hb.Get("type").ClonedString())
		inst.SetLabelTrimmed("slot", hb.Get("physicalLocation.slot").ClonedString())
		inst.SetLabelTrimmed("part_number", hb.Get("partNumber").ClonedString())
		inst.SetLabelTrimmed("serial_number", hb.Get("serialNumber").ClonedString())
		inst.SetLabelTrimmed("vendor", hb.Get("vendorName").ClonedString())
		inst.SetLabelTrimmed("fru_type", hb.Get("fruType").ClonedString())
		inst.SetLabelTrimmed("number_of_ports", hb.Get("numberOfPorts").ClonedString())

		controllerRef := hb.Get("physicalLocation.locationParent.controllerRef").ClonedString()
		if controllerLabel, ok := controllerLabelMap[controllerRef]; ok {
			inst.SetLabelTrimmed("controller", controllerLabel)
		}
	}

	h.SLogger.Debug("Processed host boards", slog.Int("count", len(mat.GetInstances())))
}
