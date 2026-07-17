// Copyright NetApp Inc, 2021 All rights reserved

package simple

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

func TestNodeMonPollDataSupportsConfiguredCounterSubset(t *testing.T) {
	mat := matrix.New("NodeMon", "nodemon", "nodemon")
	statusMetric, err := mat.NewMetricUint8("status")
	assert.Nil(t, err)
	allocMetric, err := mat.NewMetricInt64("alloc")
	assert.Nil(t, err)
	instance, err := mat.NewInstance("simple")
	assert.Nil(t, err)

	n := &NodeMon{AbstractCollector: collector.New("NodeMon", "nodemon", options.New(), nil, nil, conf.Remote{})}
	n.SetMatrix(map[string]*matrix.Matrix{"nodemon": mat})

	_, err = n.PollData()
	assert.Nil(t, err)

	status, ok := statusMetric.GetValueUint8(instance)
	assert.True(t, ok)
	assert.Equal(t, uint8(0), status)
	_, ok = allocMetric.GetValueUint64(instance)
	assert.True(t, ok)
}

func TestNodeMonPollDataRecordsDefaultCounters(t *testing.T) {
	mat := matrix.New("NodeMon", "nodemon", "nodemon")
	_, err := mat.NewMetricUint8("status")
	assert.Nil(t, err)
	allocMetric, err := mat.NewMetricInt64("alloc")
	assert.Nil(t, err)
	numGcMetric, err := mat.NewMetricInt64("num_gc")
	assert.Nil(t, err)
	numCPUMetric, err := mat.NewMetricInt64("num_cpu")
	assert.Nil(t, err)
	instance, err := mat.NewInstance("simple")
	assert.Nil(t, err)

	n := &NodeMon{AbstractCollector: collector.New("NodeMon", "nodemon", options.New(), nil, nil, conf.Remote{})}
	n.SetMatrix(map[string]*matrix.Matrix{"nodemon": mat})

	_, err = n.PollData()
	assert.Nil(t, err)

	_, ok := allocMetric.GetValueUint64(instance)
	assert.True(t, ok)
	_, ok = numGcMetric.GetValueUint64(instance)
	assert.True(t, ok)
	numCPU, ok := numCPUMetric.GetValueInt64(instance)
	assert.True(t, ok)
	assert.Equal(t, int64(0) < numCPU, true)
}
