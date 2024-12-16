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
		if len(copies) > 1 {
			var copiesValue int
			for _, c := range copies {
				val, _ := strconv.Atoi(c)
				copiesValue += val
			}
			instance.SetLabel("copies", strconv.Itoa(copiesValue))
		}
	}

	return nil, nil, nil
}
