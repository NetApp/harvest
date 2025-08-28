/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package aggregator

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
)

func newAggregator() *Aggregator {
	params := node.NewS("Aggregator")
	params.NewChildS("", "node")

	abc := plugin.New("Test", nil, params, nil, "", nil)
	p := &Aggregator{AbstractPlugin: abc}

	if err := p.Init(conf.Remote{}); err != nil {
		panic(err)
	}
	return p
}

func TestRuleSimpleAggregation(t *testing.T) {
	var (
		n                *matrix.Matrix
		instanceA        *matrix.Instance
		metricA, metricB *matrix.Metric
	)
	m := newArtificialData()
	p := newAggregator()

	// run the plugin
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	results, _, err := p.Run(dataMap)
	assert.Nil(t, err)

	assert.Equal(t, len(results), 1)
	n = results[0]

	// check aggregated values

	assert.Equal(t, len(n.GetInstances()), 1)

	instanceA = n.GetInstance("nodeA")
	assert.NotNil(t, instanceA)

	metricA = n.GetMetric("metricA")
	assert.NotNil(t, metricA)

	metricB = n.GetMetric("metricB")
	assert.NotNil(t, metricB)

	value, ok := metricA.GetValueUint8(instanceA)
	assert.True(t, ok)
	assert.Equal(t, value, uint8(20))

	value, ok = metricB.GetValueUint8(instanceA)
	assert.True(t, ok)
	assert.Equal(t, value, uint8(10))
}

func TestRuleIncludeAllLabels(t *testing.T) {
	m := newArtificialData()
	p := newAggregator()

	var n *matrix.Matrix

	params := node.NewS("Aggregator")
	params.NewChildS("", "svm ...")

	p.Params = params

	err := p.Init(conf.Remote{})
	assert.Nil(t, err)

	for _, instance := range m.GetInstances() {
		instance.SetLabel("svm", "svmA")
		instance.SetLabel("datacenter", "DatacenterA")
	}

	// run the plugin
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	results, _, err := p.Run(dataMap)
	assert.Nil(t, err)

	assert.Equal(t, len(results), 1)
	n = results[0]

	for _, instance := range n.GetInstances() {
		assert.Equal(t, instance.GetLabel("node"), "nodeA")
		assert.Equal(t, instance.GetLabel("svm"), "svmA")
		assert.Equal(t, instance.GetLabel("datacenter"), "DatacenterA")
		break
	}
}

func TestComplexRuleRegex(t *testing.T) {

	var n *matrix.Matrix
	var A, B, C, D, instance *matrix.Instance
	var metricA *matrix.Metric
	var err error

	params := node.NewS("Aggregator")
	params.NewChildS("", "volume<`_\\d{4}$`>flexgroup aggr,svm")
	p := newAggregator()

	p.Params = params
	m := newArtificialData()

	err = p.Init(conf.Remote{})
	assert.Nil(t, err)

	m.PurgeInstances()

	// should match rule
	A, err = m.NewInstance("A")
	assert.Nil(t, err)

	A.SetLabel("volume", "A_1234")
	A.SetLabel("aggr", "aggrA")
	A.SetLabel("svm", "svmA")

	// should match
	B, err = m.NewInstance("B")
	assert.Nil(t, err)

	B.SetLabel("volume", "A_1234")
	B.SetLabel("aggr", "aggrA")
	B.SetLabel("svm", "svmA")
	B.SetLabel("node", "nodeA")

	// should NOT match rule
	C, err = m.NewInstance("C")
	assert.Nil(t, err)

	C.SetLabel("volume", "C_12345") // not 4 digits
	C.SetLabel("aggr", "aggrA")
	C.SetLabel("svm", "svmA")
	B.SetLabel("node", "nodeA")

	// should match
	D, err = m.NewInstance("D")
	assert.Nil(t, err)

	D.SetLabel("volume", "D_1111")
	D.SetLabel("aggr", "aggrB")
	D.SetLabel("svm", "svmB")
	B.SetLabel("node", "nodeA")

	// flush data from previous tests
	m.Reset()

	metricA = m.GetMetric("metricA")
	assert.NotNil(t, metricA)

	metricA.SetValueUint8(A, 2)
	metricA.SetValueUint8(B, 2)
	metricA.SetValueUint8(C, 2)
	metricA.SetValueUint8(D, 2)

	// run the plugin
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	results, _, err := p.Run(dataMap)
	assert.Nil(t, err)

	assert.Equal(t, len(results), 1)
	n = results[0]

	// expecting new matrix with two instances
	// where A+B is new instance
	// C is discarded
	// and D is as it was

	assert.Equal(t, len(n.GetInstances()), 2)
	assert.Equal(t, n.Object, "flexgroup")

	metricA = n.GetMetric("metricA")
	assert.NotNil(t, metricA)

	key := "A_1234.aggrA.svmA"
	expected := uint8(4)
	instance = n.GetInstance(key)
	assert.NotNil(t, instance)
	assert.Equal(t, instance.GetLabel("aggr"), "aggrA")
	assert.Equal(t, instance.GetLabel("svm"), "svmA")
	assert.Equal(t, instance.GetLabel("node"), "")

	v, ok := metricA.GetValueUint8(instance)
	assert.True(t, ok)
	assert.Equal(t, v, expected)

	key = "D_1111.aggrB.svmB"
	expected = uint8(2)
	instance = n.GetInstance(key)

	assert.NotNil(t, instance)
	assert.Equal(t, instance.GetLabel("aggr"), "aggrB")
	assert.Equal(t, instance.GetLabel("svm"), "svmB")
	assert.Equal(t, instance.GetLabel("node"), "")

	v, ok = metricA.GetValueUint8(instance)
	assert.True(t, ok)
	assert.Equal(t, v, expected)

	key = "C_12345.aggrA.svmA"
	instance = n.GetInstance(key)
	assert.Nil(t, instance)
}

func TestRuleSimpleLatencyAggregation(t *testing.T) {

	params := node.NewS("Aggregator")
	params.NewChildS("", "node")
	p := newAggregator()

	p.Params = params

	err := p.Init(conf.Remote{})
	assert.Nil(t, err)

	m := newArtificialData()
	var n *matrix.Matrix

	metricA, err := m.NewMetricUint8("read_latency")
	assert.Nil(t, err)

	metricA.SetComment("total_read_ops")
	metricA.SetProperty("average")

	metricB, err := m.NewMetricUint8("total_read_ops")
	assert.Nil(t, err)
	metricB.SetProperty("rate")

	m.RemoveInstance("InstanceA")
	instanceA, err := m.NewInstance("InstanceA")
	assert.Nil(t, err)
	instanceA.SetLabel("node", "nodeA")

	m.RemoveInstance("InstanceB")
	instanceB, err := m.NewInstance("InstanceB")
	assert.Nil(t, err)
	instanceB.SetLabel("node", "nodeA")

	metricA.SetValueUint8(instanceA, 20)
	metricB.SetValueUint8(instanceA, 4)
	metricA.SetValueUint8(instanceB, 30)
	metricB.SetValueUint8(instanceB, 6)

	// run the plugin
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	results, _, err := p.Run(dataMap)
	assert.Nil(t, err)

	assert.Equal(t, len(results), 1)
	n = results[0]

	// check aggregated values

	assert.Equal(t, len(n.GetInstances()), 1)

	instanceA = n.GetInstance("nodeA")
	assert.NotNil(t, instanceA)

	metricA = n.GetMetric("read_latency")
	assert.NotNil(t, metricA)

	metricB = n.GetMetric("total_read_ops")
	assert.NotNil(t, metricB)

	value, ok := metricA.GetValueUint8(instanceA)
	assert.True(t, ok)
	assert.Equal(t, value, 26)

	value, ok = metricB.GetValueUint8(instanceA)
	assert.True(t, ok)
	assert.Equal(t, value, 10)
}

func TestRuleSimpleLatencyZeroAggregation(t *testing.T) {

	params := node.NewS("Aggregator")
	params.NewChildS("", "node")
	p := newAggregator()

	p.Params = params

	err := p.Init(conf.Remote{})
	assert.Nil(t, err)

	m := newArtificialData()
	var n *matrix.Matrix

	metricA, err := m.NewMetricUint8("read_latency")
	assert.Nil(t, err)

	metricA.SetComment("total_read_ops")
	metricA.SetProperty("average")

	metricB, err := m.NewMetricUint8("total_read_ops")
	assert.Nil(t, err)
	metricB.SetProperty("rate")

	m.RemoveInstance("InstanceA")
	instanceA, err := m.NewInstance("InstanceA")
	assert.Nil(t, err)
	instanceA.SetLabel("node", "nodeA")

	m.RemoveInstance("InstanceB")
	instanceB, err := m.NewInstance("InstanceB")
	assert.Nil(t, err)
	instanceB.SetLabel("node", "nodeA")

	metricA.SetValueUint8(instanceA, 20)
	metricB.SetValueUint8(instanceA, 0)
	metricA.SetValueUint8(instanceB, 21)
	metricB.SetValueUint8(instanceB, 0)

	// run the plugin
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	results, _, err := p.Run(dataMap)
	assert.Nil(t, err)

	assert.Equal(t, len(results), 1)
	n = results[0]

	// check aggregated values

	assert.Equal(t, len(n.GetInstances()), 1)

	instanceA = n.GetInstance("nodeA")
	assert.NotNil(t, instanceA)

	metricA = n.GetMetric("read_latency")
	assert.NotNil(t, metricA)

	metricB = n.GetMetric("total_read_ops")
	assert.NotNil(t, metricB)

	value, ok := metricA.GetValueUint8(instanceA)
	assert.True(t, ok)
	assert.Equal(t, value, 0)

	value, ok = metricB.GetValueUint8(instanceA)
	assert.True(t, ok)
	assert.Equal(t, value, 0)
}

func newArtificialData() *matrix.Matrix {
	m := matrix.New("", "", "")

	metricA, err := m.NewMetricUint8("metricA")
	if err != nil {
		panic(err)
	}
	metricB, err := m.NewMetricUint8("metricB")
	if err != nil {
		panic(err)
	}
	metricB.SetProperty("average")

	instanceA, err := m.NewInstance("InstanceA")
	if err != nil {
		panic(err)
	}
	instanceA.SetLabel("node", "nodeA")

	instanceB, err := m.NewInstance("InstanceB")
	if err != nil {
		panic(err)
	}
	instanceB.SetLabel("node", "nodeA")

	metricA.SetValueUint8(instanceA, 10)
	metricA.SetValueUint8(instanceB, 10)
	metricB.SetValueUint8(instanceA, 10)

	return m
}
