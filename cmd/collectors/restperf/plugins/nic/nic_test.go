package nic

import (
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"log/slog"
	"testing"
)

// Common test logic for RestPerf/ZapiPerf Volume plugin
func runNicTest(t *testing.T, createRestNic func(params *node.Node) plugin.Plugin, expectedCount int) {
	params := node.NewS("NicCommon")
	n := createRestNic(params)

	// Initialize the plugin
	if err := n.Init(); err != nil {
		t.Fatalf("failed to initialize plugin: %v", err)
	}

	// Create test data
	data := matrix.New("nic", "nic", "nic")
	instance1, _ := data.NewInstance("rtp-a700s-02:e5a")
	instance1.SetLabel("id", "rtp-a700s-02:e5a")
	instance1.SetLabel("speed", "0")
	instance1.SetLabel("node", "rtp-a700s-02")
	instance1.SetLabel("type", "nic_ixl")

	instance2, _ := data.NewInstance("rtp-a700s-02:e5b")
	instance2.SetLabel("id", "rtp-a700s-02:e5b")
	instance2.SetLabel("speed", "0")
	instance2.SetLabel("node", "rtp-a700s-02")
	instance2.SetLabel("type", "nic_ixl")

	instance3, _ := data.NewInstance("rtp-a700s-01:e5c")
	instance3.SetLabel("id", "rtp-a700s-01:e5c")
	instance3.SetLabel("speed", "10000M")
	instance3.SetLabel("node", "rtp-a700s-01")
	instance3.SetLabel("type", "nic_ixl")

	instance4, _ := data.NewInstance("rtp-a700s-01:e5d")
	instance4.SetLabel("id", "rtp-a700s-01:e5d")
	instance4.SetLabel("speed", "10000M")
	instance4.SetLabel("node", "rtp-a700s-01")
	instance4.SetLabel("type", "nic_ixl")

	// Create latency and ops metrics
	receiveBytes, _ := data.NewMetricFloat64("receive_bytes")
	transmitBytes, _ := data.NewMetricFloat64("transmit_bytes")

	// Set metric values for the instances
	_ = receiveBytes.SetValueFloat64(instance1, 2861802356977)
	_ = transmitBytes.SetValueFloat64(instance1, 5789662182305)

	_ = receiveBytes.SetValueFloat64(instance2, 2861802356977)
	_ = transmitBytes.SetValueFloat64(instance2, 5789662182305)

	_ = receiveBytes.SetValueFloat64(instance3, 2861802356977)
	_ = transmitBytes.SetValueFloat64(instance3, 5789662182305)

	_ = receiveBytes.SetValueFloat64(instance4, 2861802356977)
	_ = transmitBytes.SetValueFloat64(instance4, 5789662182305)

	dataMap := map[string]*matrix.Matrix{
		"nic": data,
	}

	// Run the plugin
	output, _, err := n.Run(dataMap)
	if err != nil {
		t.Fatalf("Run method failed: %v", err)
	}

	// Verify the output
	if len(output) != 1 {
		t.Fatalf("expected 2 output matrices, got %d", len(output))
	}

	ifgroupData := output[0]

	// Check for ifgroup instance
	ifgroupInstance1 := ifgroupData.GetInstance("rtp-a700s-01a0a")
	if ifgroupInstance1 == nil {
		t.Fatalf("expected ifgroup instance 'rtp-a700s-01a0a' not found")
	}

	// Check for ifgroup instance
	ifgroupInstance2 := ifgroupData.GetInstance("rtp-a700s-02a0b")
	if ifgroupInstance2 != nil {
		t.Fatalf("expected ifgroup instance 'rtp-a700s-02a0b' found")
	}

	if label := ifgroupInstance1.GetLabel("ports"); label != "e5c,e5d" {
		t.Fatalf("expected ifgroup metric instance label 'ports' to be 'e5c,e5d', got '%s'", label)
	}

	// count ifgroup instances
	ifgroupCount := 0
	for _, i := range ifgroupData.GetInstances() {
		if i.IsExportable() {
			ifgroupCount++
		}
	}

	// Verify the number of instances in the ifgroup
	if ifgroupCount != expectedCount {
		t.Errorf("expected %d instances in the matrix, got %d", expectedCount, ifgroupCount)
	}

	if value, ok := ifgroupData.GetMetric("rx_bytes").GetValueFloat64(ifgroupInstance1); !ok {
		t.Error("Value [rx_bytes] missing")
	} else if value != 5723604713954.0 {
		t.Errorf("Value [rx_bytes] = (%f) incorrect", value)
	}
}

func TestRunForAllImplementations(t *testing.T) {
	t.Run("RestPerf nic_common with ifgrp", func(t *testing.T) {
		runNicTest(t, createRestNic, 1) // Only 1 ifgroup instance would be exported as 2nd ifgroup has speed=0
	})
}

func createRestNic(params *node.Node) plugin.Plugin {
	o := options.Options{IsTest: true}
	n := &Nic{AbstractPlugin: plugin.New("nic", &o, params, nil, "nic", nil)}
	n.SLogger = slog.Default()
	n.client = &rest.Client{Metadata: &util.Metadata{}}
	n.testFilePath = "../../testdata/port-test.json"
	return n
}
