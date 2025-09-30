package igroup

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"slices"
	"strings"
)

type Igroup struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Igroup{AbstractPlugin: p}
}

func (i *Igroup) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[i.Object]
	var initiators []string
	for _, instance := range data.GetInstances() {
		initiatorsData := gjson.Result{Type: gjson.JSON, Raw: "[" + instance.GetLabel("initiators") + "]"}
		array := initiatorsData.Array()
		initiators = make([]string, 0, len(array))
		for _, mapData := range array {
			initiator := mapData.Get("name").ClonedString()
			initiators = append(initiators, initiator)
		}
		slices.Sort(initiators)
		instance.SetLabel("initiator", strings.Join(initiators, ","))
	}
	return nil, nil, nil
}
