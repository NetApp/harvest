package volume

import (
	"strings"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

type Volume struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (v *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]

	for _, instance := range data.GetInstances() {
		if wwid := instance.GetLabel("wwid"); wwid != "" {
			formattedWWID := formatWWID(wwid)
			instance.SetLabel("wwid", formattedWWID)
		}
	}

	return nil, nil, nil
}

// formatWWID converts WWID from continuous hex string to colon-separated format
// Example: 6D039EA000DCD9780000268E695FC9BE -> 6D:03:9E:A0:00:DC:D9:78:00:00:26:8E:69:5F:C9:BE
func formatWWID(wwid string) string {
	if wwid == "" {
		return wwid
	}

	var result strings.Builder
	result.Grow(len(wwid) + (len(wwid) / 2))

	for i := 0; i < len(wwid); i += 2 {
		if i > 0 {
			result.WriteString(":")
		}

		if i+2 <= len(wwid) {
			result.WriteString(wwid[i : i+2])
		} else {
			result.WriteString(wwid[i:])
		}
	}

	return result.String()
}
