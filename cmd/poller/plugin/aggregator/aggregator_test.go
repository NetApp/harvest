/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package aggregator

import (
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
)

func newAggregator() *Aggregator {
	params := node.NewS("Aggregator")
	params.NewChildS("", "node")

	abc := plugin.New("Test", nil, params, nil, "", nil)
	p := &Aggregator{AbstractPlugin: abc}

	if err := p.Init(); err != nil {
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
	results, err := p.Run(dataMap)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 1 {
		n = results[0]
	} else {
		t.Fatalf("Plugin output has %d matrices, 1 was expected\n", len(results))
	}

	// check aggregated values

	if len(n.GetInstances()) != 1 {
		t.Fatalf("Number of instances is %d, 1 was expected\n", len(n.GetInstances()))
	}

	if instanceA = n.GetInstance("nodeA"); instanceA == nil {
		t.Fatal("Instance [nodeA] missing")
	}

	if metricA = n.GetMetric("metricA"); metricA == nil {
		t.Fatal("Metric [metricA] missing")
	}

	if metricB = n.GetMetric("metricB"); metricB == nil {
		t.Fatal("Metric [metricB] missing")
	}

	if value, ok := metricA.GetValueUint8(instanceA); !ok {
		t.Error("Value [metricA] missing")
	} else if value != 20 {
		t.Errorf("Value [metricA] = (%d) incorrect", value)
	} else {
		t.Logf("Value [metricA] = (%d) correct!", value)
	}

	if value, ok := metricB.GetValueUint8(instanceA); !ok {
		t.Error("Value [metricB] missing")
	} else if value != 10 {
		t.Errorf("Value [metricB] = (%d) incorrect", value)
	} else {
		t.Logf("Value [metricB] = (%d) correct!", value)
	}
}

func TestRuleIncludeAllLabels(t *testing.T) {
	m := newArtificialData()
	p := newAggregator()

	var n *matrix.Matrix

	params := node.NewS("Aggregator")
	params.NewChildS("", "svm ...")

	p.Params = params

	if err := p.Init(); err != nil {
		t.Fatal(err)
	}

	for _, instance := range m.GetInstances() {
		instance.SetLabel("svm", "svmA")
		instance.SetLabel("datacenter", "DatacenterA")
	}

	// run the plugin
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	results, err := p.Run(dataMap)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 1 {
		n = results[0]
	} else {
		t.Fatalf("plugin output has %d matrices, 1 was expected\n", len(results))
	}

	for _, instance := range n.GetInstances() {

		if instance.GetLabel("node") != "nodeA" {
			t.Errorf("label [node] has not expected value: %s", instance.GetLabel("node"))
		} else {
			t.Logf("label [svm] set: %s", instance.GetLabel("svm"))
		}

		if instance.GetLabel("svm") != "svmA" {
			t.Errorf("label [svm] has not expected value: %s", instance.GetLabel("svm"))
		} else {
			t.Logf("label [svm] set: %s", instance.GetLabel("svm"))
		}

		if instance.GetLabel("datacenter") != "DatacenterA" {
			t.Errorf("label [datacenter] has not expected value: %s", instance.GetLabel("datacenter"))
		} else {
			t.Logf("label [datacenter] set: %s", instance.GetLabel("datacenter"))
		}

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

	if err := p.Init(); err != nil {
		t.Fatal(err)
	}

	m.PurgeInstances()

	// should match rule
	if A, err = m.NewInstance("A"); err != nil {
		t.Fatal(err)
	}
	A.SetLabel("volume", "A_1234")
	A.SetLabel("aggr", "aggrA")
	A.SetLabel("svm", "svmA")

	// should match
	if B, err = m.NewInstance("B"); err != nil {
		t.Fatal(err)
	}
	B.SetLabel("volume", "A_1234")
	B.SetLabel("aggr", "aggrA")
	B.SetLabel("svm", "svmA")
	B.SetLabel("node", "nodeA")

	// should NOT match rule
	if C, err = m.NewInstance("C"); err != nil {
		t.Fatal(err)
	}
	C.SetLabel("volume", "C_12345") // not 4 digits
	C.SetLabel("aggr", "aggrA")
	C.SetLabel("svm", "svmA")
	B.SetLabel("node", "nodeA")

	// should match
	if D, err = m.NewInstance("D"); err != nil {
		t.Fatal(err)
	}
	D.SetLabel("volume", "D_1111")
	D.SetLabel("aggr", "aggrB")
	D.SetLabel("svm", "svmB")
	B.SetLabel("node", "nodeA")

	// flush data from previous tests
	m.Reset()

	if metricA = m.GetMetric("metricA"); metricA == nil {
		t.Fatal("missing [metricA]")
	}

	if err = metricA.SetValueUint8(A, 2); err != nil {
		t.Fatal(err)
	}

	if err = metricA.SetValueUint8(B, 2); err != nil {
		t.Fatal(err)
	}

	if err = metricA.SetValueUint8(C, 2); err != nil {
		t.Fatal(err)
	}

	if err = metricA.SetValueUint8(D, 2); err != nil {
		t.Fatal(err)
	}

	// run the plugin
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	results, err := p.Run(dataMap)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 1 {
		n = results[0]
	} else {
		t.Fatalf("plugin output has %d matrices, 1 was expected", len(results))
	}

	// expecting new matrix with two instances
	// where A+B is new instance
	// C is discarded
	// and D is as it was

	if len(n.GetInstances()) == 2 {
		t.Logf("OK - matrix has %d instances as expected", len(n.GetInstances()))
	} else {
		t.Fatalf("matrix has %d instances, 2 was expected", len(n.GetInstances()))
	}

	if n.Object == "flexgroup" {
		t.Logf("OK - matrix object is (%s)", n.Object)
	} else {
		t.Errorf("matrix object is (%s), expected (flexgroup)", n.Object)
	}

	if metricA = n.GetMetric("metricA"); metricA == nil {
		t.Fatal("missing [metricA]")
	}

	key := "A_1234.aggrA.svmA"
	expected := uint8(4)
	if instance = n.GetInstance(key); instance == nil {
		t.Errorf("instance [%s] missing", key)
	} else {

		if instance.GetLabel("svm") == "svmA" && instance.GetLabel("aggr") == "aggrA" && instance.GetLabel("node") == "" {
			t.Logf("OK - instance has expected labels: %v", instance.GetLabels())
		} else {
			t.Errorf("instance has not expected labels: %v", instance.GetLabels())
		}
		if v, ok := metricA.GetValueUint8(instance); !ok {
			t.Errorf("value [metricA] not set")
		} else if v != expected {
			t.Errorf("value [metricA] = %d, expected %d", v, expected)
		} else {
			t.Logf("OK - value [metricA] = %d", v)
		}
	}

	key = "D_1111.aggrB.svmB"
	expected = uint8(2)
	if instance = n.GetInstance(key); instance == nil {
		t.Errorf("instance [%s] missing", key)
	} else {

		if instance.GetLabel("svm") == "svmB" && instance.GetLabel("aggr") == "aggrB" && instance.GetLabel("node") == "" {
			t.Logf("OK - instance has expected labels: %v", instance.GetLabels())
		} else {
			t.Errorf("instance has not expected labels: %v", instance.GetLabels())
		}

		if v, ok := metricA.GetValueUint8(instance); !ok {
			t.Errorf("value [metricA] not set")
		} else if v != expected {
			t.Errorf("value [metricA] = %d, expected %d", v, expected)
		} else {
			t.Logf("OK - value [metricA] = %d", v)
		}
	}

	key = "C_12345.aggrA.svmA"
	if instance = n.GetInstance(key); instance == nil {
		t.Logf("OK - no instance [%s] added (did not match regex)", key)
	} else {
		t.Errorf("instance [%s] was added, however should not match regex", key)
	}
}

func TestRuleSimpleLatencyAggregation(t *testing.T) {

	params := node.NewS("Aggregator")
	params.NewChildS("", "node")
	p := newAggregator()

	p.Params = params

	if err := p.Init(); err != nil {
		t.Fatal(err)
	}

	m := newArtificialData()
	var n *matrix.Matrix

	metricA, err := m.NewMetricUint8("read_latency")
	if err != nil {
		t.Fatal(err)
	}
	metricA.SetComment("total_read_ops")
	metricA.SetProperty("average")

	metricB, err := m.NewMetricUint8("total_read_ops")
	if err != nil {
		t.Fatal(err)
	}
	metricB.SetProperty("rate")

	m.RemoveInstance("InstanceA")
	instanceA, err := m.NewInstance("InstanceA")
	if err != nil {
		t.Fatal(err)
	}
	instanceA.SetLabel("node", "nodeA")

	m.RemoveInstance("InstanceB")
	instanceB, err := m.NewInstance("InstanceB")
	if err != nil {
		t.Fatal(err)
	}
	instanceB.SetLabel("node", "nodeA")

	if err = metricA.SetValueUint8(instanceA, 20); err != nil {
		t.Fatal(err)
	}

	if err = metricB.SetValueUint8(instanceA, 4); err != nil {
		t.Fatal(err)
	}

	if err = metricA.SetValueUint8(instanceB, 30); err != nil {
		t.Fatal(err)
	}

	if err = metricB.SetValueUint8(instanceB, 6); err != nil {
		t.Fatal(err)
	}

	// run the plugin
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	results, err := p.Run(dataMap)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 1 {
		n = results[0]
	} else {
		t.Fatalf("Plugin output has %d matrices, 1 was expected\n", len(results))
	}

	// check aggregated values

	if len(n.GetInstances()) != 1 {
		t.Fatalf("Number of instances is %d, 1 was expected\n", len(n.GetInstances()))
	}

	if instanceA = n.GetInstance("nodeA"); instanceA == nil {
		t.Fatal("Instance [nodeA] missing")
	}

	if metricA = n.GetMetric("read_latency"); metricA == nil {
		t.Fatal("Metric [read_latency] missing")
	}

	if metricB = n.GetMetric("total_read_ops"); metricB == nil {
		t.Fatal("Metric [total_read_ops] missing")
	}

	if value, ok := metricA.GetValueUint8(instanceA); !ok {
		t.Error("Value [read_latency] missing")
	} else if value != 26 {
		t.Errorf("Value [read_latency] = (%d) incorrect", value)
	} else {
		t.Logf("Value [read_latency] = (%d) correct!", value)
	}

	if value, ok := metricB.GetValueUint8(instanceA); !ok {
		t.Error("Value [total_read_ops] missing")
	} else if value != 10 {
		t.Errorf("Value [total_read_ops] = (%d) incorrect", value)
	} else {
		t.Logf("Value [total_read_ops] = (%d) correct!", value)
	}
}

func TestRuleSimpleLatencyZeroAggregation(t *testing.T) {

	params := node.NewS("Aggregator")
	params.NewChildS("", "node")
	p := newAggregator()

	p.Params = params

	if err := p.Init(); err != nil {
		t.Fatal(err)
	}

	m := newArtificialData()
	var n *matrix.Matrix

	metricA, err := m.NewMetricUint8("read_latency")
	if err != nil {
		t.Fatal(err)
	}
	metricA.SetComment("total_read_ops")
	metricA.SetProperty("average")

	metricB, err := m.NewMetricUint8("total_read_ops")
	if err != nil {
		t.Fatal(err)
	}
	metricB.SetProperty("rate")

	m.RemoveInstance("InstanceA")
	instanceA, err := m.NewInstance("InstanceA")
	if err != nil {
		t.Fatal(err)
	}
	instanceA.SetLabel("node", "nodeA")

	m.RemoveInstance("InstanceB")
	instanceB, err := m.NewInstance("InstanceB")
	if err != nil {
		t.Fatal(err)
	}
	instanceB.SetLabel("node", "nodeA")

	if err = metricA.SetValueUint8(instanceA, 20); err != nil {
		t.Fatal(err)
	}

	if err = metricB.SetValueUint8(instanceA, 0); err != nil {
		t.Fatal(err)
	}

	if err = metricA.SetValueUint8(instanceB, 21); err != nil {
		t.Fatal(err)
	}

	if err = metricB.SetValueUint8(instanceB, 0); err != nil {
		t.Fatal(err)
	}

	// run the plugin
	dataMap := map[string]*matrix.Matrix{
		m.Object: m,
	}
	results, err := p.Run(dataMap)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 1 {
		n = results[0]
	} else {
		t.Fatalf("Plugin output has %d matrices, 1 was expected\n", len(results))
	}

	// check aggregated values

	if len(n.GetInstances()) != 1 {
		t.Fatalf("Number of instances is %d, 1 was expected\n", len(n.GetInstances()))
	}

	if instanceA = n.GetInstance("nodeA"); instanceA == nil {
		t.Fatal("Instance [nodeA] missing")
	}

	if metricA = n.GetMetric("read_latency"); metricA == nil {
		t.Fatal("Metric [read_latency] missing")
	}

	if metricB = n.GetMetric("total_read_ops"); metricB == nil {
		t.Fatal("Metric [total_read_ops] missing")
	}

	if value, ok := metricA.GetValueUint8(instanceA); !ok {
		t.Error("Value [read_latency] missing")
	} else if value != 0 {
		t.Errorf("Value [read_latency] = (%d) incorrect", value)
	} else {
		t.Logf("Value [read_latency] = (%d) correct!", value)
	}

	if value, ok := metricB.GetValueUint8(instanceA); !ok {
		t.Error("Value [total_read_ops] missing")
	} else if value != 0 {
		t.Errorf("Value [total_read_ops] = (%d) incorrect", value)
	} else {
		t.Logf("Value [total_read_ops] = (%d) correct!", value)
	}
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

	if err = metricA.SetValueUint8(instanceA, 10); err != nil {
		panic(err)
	}

	if err = metricA.SetValueUint8(instanceB, 10); err != nil {
		panic(err)
	}

	if err = metricB.SetValueUint8(instanceA, 10); err != nil {
		panic(err)
	}

	return m
}
