package aggregator

import (
    "testing"
    "goharvest2/share/tree/node"
    "goharvest2/share/matrix"
    "goharvest2/poller/plugin"
    "goharvest2/share/logger"
)

var p *Aggregator

func TestInitPlugin(t *testing.T) {

    logger.SetLevel(0)

    params := node.NewS("Aggregator")
    params.NewChildS("", "node")

    abc := plugin.New("Test", nil, params, nil)
    p = &Aggregator{AbstractPlugin: abc}

    if err := p.Init(); err != nil {
        t.Fatal(err)
    }
}

func TestRun(t *testing.T) {

    // create artifical data
    m := matrix.New("", "", "")

    metricA, err := m.AddMetricUint8("metricA")
    if err != nil {
        t.Fatal(err)
    }
    metricB, err := m.AddMetricUint8("metricB")
    if err != nil {
        t.Fatal(err)
    }
    metricB.SetProperty("average")

    instanceA, err := m.AddInstance("InstanceA")
    if err != nil {
        t.Fatal(err)
    }
    instanceA.SetLabel("node", "nodeA")

    instanceB, err := m.AddInstance("InstanceB")
    if err != nil {
        t.Fatal(err)
    }
    instanceB.SetLabel("node", "nodeA")

    if err := m.Reset(); err != nil {
        t.Fatal(err)
    }

    if err = metricA.SetValueUint8(instanceA, 10); err != nil {
        t.Fatal(err)
    }

    if err = metricA.SetValueUint8(instanceB, 10); err != nil {
        t.Fatal(err)
    }

    if err = metricB.SetValueUint8(instanceA, 10); err != nil {
        t.Fatal(err)
    }

    // run the plugin
    results, err := p.Run(m)
    if err != nil {
        t.Fatal(err)
    }

    if len(results) == 1 {
        m = results[0]
    } else {
        t.Fatalf("Plugin output has %d matrices, however 1 was expected\n", len(results))
    }

    // check aggregated values

    if len(m.GetInstances()) != 1 {
        t.Fatalf("Number of instances is %d, however 1 was expected\n", len(m.GetInstances()))
    }

    if instanceA = m.GetInstance("nodeA"); instanceA == nil {
        t.Fatal("Instance [nodeA] missing")
    }

    if metricA = m.GetMetric("metricA"); metricA == nil {
        t.Fatal("Metric [metricA] missing")
    }

    if metricB = m.GetMetric("metricB"); metricB == nil {
        t.Fatal("Metric [metricB] missing")
    }

    if value, ok := metricA.GetValueUint8(instanceA); ! ok {
        t.Error("Value [metricA] missing")
    } else if value != 20 {
        t.Errorf("Value [metricA] = (%d) incorrect", value)
    } else {
        t.Logf("Value [metricA] = (%d) correct!", value)
    }

    if value, ok := metricB.GetValueUint8(instanceA); ! ok {
        t.Error("Value [metricB] missing")
    } else if value != 10 {
        t.Errorf("Value [metricB] = (%d) incorrect", value)
    } else {
        t.Logf("Value [metricB] = (%d) correct!", value)
    }
}
