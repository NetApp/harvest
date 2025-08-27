package test

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/aggregator"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/labelagent"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
)

func TestMultipleRule(t *testing.T) {
	var Plugins []plugin.Plugin
	remote := conf.Remote{}

	defaultTemplate, _ := tree.ImportYaml("testdata/sample.yaml")
	p := defaultTemplate.GetChildS("plugins")
	for _, x := range p.GetChildren() {

		name := x.GetNameS()
		if name == "" {
			name = x.GetContentS() // some plugins are defined as list elements others as dicts
			x.SetNameS(name)
		}

		if name == "LabelAgent" {
			var lb *labelagent.LabelAgent

			params := node.NewS("LabelAgent")
			for _, x1 := range x.GetChildren() {
				params1 := params.NewChildS(x1.GetNameS(), "")
				for _, x2 := range x1.GetChildren() {
					params1.NewChildS("", x2.GetContentS())
				}
			}
			abc := plugin.New("Test", nil, params, nil, "", nil)
			lb = &labelagent.LabelAgent{AbstractPlugin: abc}
			err := lb.Init(remote)
			assert.Nil(t, err)
			Plugins = append(Plugins, lb)
		}

		if name == "Aggregator" {
			params := node.NewS("Aggregator")
			for _, x1 := range x.GetChildren() {
				params.NewChildS("", x1.GetContentS())
			}
			abc := plugin.New("Test", nil, params, nil, "", nil)
			ag := &aggregator.Aggregator{AbstractPlugin: abc}

			err := ag.Init(remote)
			assert.Nil(t, err)
			Plugins = append(Plugins, ag)
		}
	}

	m := matrix.New("TestLabelAgent", "", "test")

	metricA, err := m.NewMetricUint8("metricA")
	assert.Nil(t, err)

	metricB, err := m.NewMetricUint8("metricB")
	assert.Nil(t, err)
	metricB.SetProperty("average")

	// should match
	instanceA, _ := m.NewInstance("0")
	instanceA.SetLabel("A", "aaa bbb ccc")
	instanceA.SetLabel("B", "abc")
	instanceA.SetLabel("state", "online") // "status" should be 1
	instanceA.SetLabel("node", "nodeA")

	metricA.SetValueUint8(instanceA, 10)

	metricB.SetValueUint8(instanceA, 10)

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa bbb")
	instanceNo.SetLabel("B", "aaa bbb ccc")
	instanceNo.SetLabel("node", "nodeB")

	var results []*matrix.Matrix
	results = append(results, m)
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	for _, plg := range Plugins {
		if pluginData, _, err := plg.Run(dataMap); err != nil {
			panic(err)
		} else if pluginData != nil {
			results = append(results, pluginData...)
		}
	}

	assert.False(t, instanceA.IsExportable())
	assert.True(t, instanceNo.IsExportable())

	assert.Equal(t, instanceA.GetLabel("B"), "xyz")

	var status *matrix.Metric
	var expected uint8
	status = m.GetMetric("new_status")
	assert.NotNil(t, status)

	expected = 1
	v, ok := status.GetValueUint8(instanceA)
	assert.True(t, ok)
	assert.Equal(t, expected, v)

	n := results[1]
	// one instance present as instanceA exported was false
	assert.Equal(t, len(n.GetInstances()), 1)
}
