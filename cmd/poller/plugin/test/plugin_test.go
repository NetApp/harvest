package test

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/aggregator"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/labelagent"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
)

func TestMultipleRule(t *testing.T) {
	var Plugins []plugin.Plugin

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
			abc := plugin.New("Test", nil, params, nil, "")
			lb = &labelagent.LabelAgent{AbstractPlugin: abc}
			if err := lb.Init(); err != nil {
				t.Fatal(err)
			}
			Plugins = append(Plugins, lb)
		}

		if name == "Aggregator" {
			params := node.NewS("Aggregator")
			for _, x1 := range x.GetChildren() {
				params.NewChildS("", x1.GetContentS())
			}
			abc := plugin.New("Test", nil, params, nil, "")
			ag := &aggregator.Aggregator{AbstractPlugin: abc}

			if err := ag.Init(); err != nil {
				t.Fatal(err)
			}
			Plugins = append(Plugins, ag)
		}
	}

	m := matrix.New("TestLabelAgent", "test", "test")

	metricA, err := m.NewMetricUint8("metricA")
	if err != nil {
		t.Fatal(err)
	}

	metricB, err := m.NewMetricUint8("metricB")
	if err != nil {
		t.Fatal(err)
	}
	metricB.SetProperty("average")

	// should match
	instanceA, _ := m.NewInstance("0")
	instanceA.SetLabel("A", "aaa bbb ccc")
	instanceA.SetLabel("B", "abc")
	instanceA.SetLabel("state", "online") // "status" should be 1
	instanceA.SetLabel("node", "nodeA")

	if err = metricA.SetValueUint8(instanceA, 10); err != nil {
		t.Fatal(err)
	}

	if err = metricB.SetValueUint8(instanceA, 10); err != nil {
		t.Fatal(err)
	}

	// should not match
	instanceNo, _ := m.NewInstance("1")
	instanceNo.SetLabel("A", "aaa bbb")
	instanceNo.SetLabel("B", "aaa bbb ccc")
	instanceNo.SetLabel("node", "nodeB")

	results := make([]*matrix.Matrix, 0)
	results = append(results, m)
	for _, plg := range Plugins {
		if pluginData, err := plg.Run(m); err != nil {
			panic(err)
		} else if pluginData != nil {
			results = append(results, pluginData...)
		}
	}

	if instanceA.IsExportable() {
		t.Error("InstanceYes should have been excluded")
	}

	if !instanceNo.IsExportable() {
		t.Error("instanceNo should not have been excluded")
	}

	if instanceA.GetLabel("B") != "xyz" {
		t.Errorf("metric [status]: value for InstanceA is %s, expected %s", instanceA.GetLabel("B"), "xyz")
	}

	var status matrix.Metric
	var expected uint8
	if status = m.GetMetric("new_status"); status == nil {
		t.Error("metric [status] missing")
	}

	expected = 1
	if v, ok, pass := status.GetValueUint8(instanceA); !ok || !pass {
		t.Error("metric [status]: value for InstanceA not set")
	} else if v != expected {
		t.Errorf("metric [status]: value for InstanceA is %d, expected %d", v, expected)
	} else {
		t.Logf("OK - metric [status]: value for instanceA set to %d", v)
	}

	n := results[1]
	// one instance present as instanceA exported was false
	if len(n.GetInstances()) != 1 {
		t.Fatalf("Number of instances is %d, 1 was expected\n", len(n.GetInstances()))
	}

}
