/*
 * Copyright NetApp Inc, 2024 All rights reserved
 */

package snapshotpolicy

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"strconv"
	"strings"
)

type SnapshotPolicy struct {
	*plugin.AbstractPlugin
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &SnapshotPolicy{AbstractPlugin: p}
}

func (m *SnapshotPolicy) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	// Purge and reset data
	data := dataMap[m.Object]

	for _, instance := range data.GetInstances() {
		copies := instance.GetLabel("copies")
		copiesJSON := gjson.Result{Type: gjson.JSON, Raw: "[" + copies + "]"}
		var scheduleVal string
		var copiesValue int
		for _, copiesData := range copiesJSON.Array() {
			count := copiesData.Get("count").String()
			countVal, _ := strconv.Atoi(count)
			schedule := copiesData.Get("schedule.name").ClonedString()
			scheduleVal = scheduleVal + schedule + ":" + count + ","
			copiesValue += countVal
		}
		instance.SetLabel("schedules", strings.TrimSuffix(scheduleVal, ","))
		instance.SetLabel("copies", strconv.Itoa(copiesValue))
	}

	return nil, nil, nil
}
