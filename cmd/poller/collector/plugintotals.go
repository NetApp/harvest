/*
Copyright NetApp Inc, 2026 All rights reserved

This file has the type that adds the metadata from the plugins. The type writes
the totals to the metadata matrix of the collector one time for each data poll.
*/

package collector

import (
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

// pluginTotals adds the metadata from all the plugins for one data poll.
//
// The type keeps the totals in memory. It writes them to the metadata matrix one time.
// The code must not write the pluginInstances total one time for each plugin.
// matrix.Metric.AddValue* adds the new value to the value in the metric.
// It does not read the record flag.
// Matrix.ResetInstance clears the record flag, but it keeps the value.
// If the code adds a value for each plugin, the total becomes too large after each poll.
// Metric.AddValueString does this operation correctly.
type pluginTotals struct {
	bytesRx   uint64
	numCalls  uint64
	instances uint64
}

// add adds the metadata from one plugin to the totals.
// If md is nil, add does nothing. You can give the result from Plugin.Run to add.
func (t *pluginTotals) add(md *collector.Metadata) {
	if md == nil {
		return
	}
	t.bytesRx += md.BytesRx.Load()
	t.numCalls += md.NumCalls.Load()
	t.instances += md.PluginInstances.Load()
}

// write puts the totals in instance i of the metadata matrix of the collector.
//
// write adds the bytesRx and numCalls totals to the values in the matrix.
// PollData writes the numbers for the API calls of the collector to these values first.
// write sets the pluginInstances value, because the plugins are the only source for
// this total. t.instances is already the total for all the plugins.
func (t *pluginTotals) write(md *matrix.Matrix, i *matrix.Instance) {
	md.MustAddValueUint64("bytesRx", i, t.bytesRx)
	md.MustAddValueUint64("numCalls", i, t.numCalls)
	md.MustSetValueUint64("pluginInstances", i, t.instances)
}
