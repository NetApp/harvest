/*
 * Copyright NetApp Inc, 2024 All rights reserved
 */

package snapshotpolicy

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
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
		copies := strings.Split(instance.GetLabel("copies"), ",")
		schedules := strings.Split(instance.GetLabel("schedules"), ",")

		var scheduleVal string
		var copiesValue int
		if len(copies) > 1 {
			for index, copiesData := range copies {
				countVal, _ := strconv.Atoi(copiesData)
				schedule := schedules[index]
				scheduleVal = scheduleVal + schedule + ":" + copiesData + ","
				copiesValue += countVal
			}
			instance.SetLabel("schedules", strings.TrimSuffix(scheduleVal, ","))
			instance.SetLabel("copies", strconv.Itoa(copiesValue))
		}
	}

	return nil, nil, nil
}
