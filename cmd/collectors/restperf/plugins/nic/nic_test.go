package nic

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"testing"
)

// Common test logic for RestPerf/ZapiPerf Volume plugin
func runNicTest(t *testing.T, createRestNic func(params *node.Node) plugin.Plugin, expectedCount int) {
	params := node.NewS("NicCommon")
	n := createRestNic(params)

	// Initialize the plugin
	if err := n.Init(conf.Remote{}); err != nil {
		t.Fatalf("failed to initialize plugin: %v", err)
	}

	// Create test data
	data := matrix.New("nic", "nic", "nic")
	instanceA1, _ := data.NewInstance("rtp-a700s-02:e5a")
	instanceA1.SetLabel("id", "rtp-a700s-02:e5a")
	instanceA1.SetLabel("speed", "0")
	instanceA1.SetLabel("node", "rtp-a700s-02")
	instanceA1.SetLabel("type", "nic_ixl")

	instanceA2, _ := data.NewInstance("rtp-a700s-02:e5b")
	instanceA2.SetLabel("id", "rtp-a700s-02:e5b")
	instanceA2.SetLabel("speed", "0")
	instanceA2.SetLabel("node", "rtp-a700s-02")
	instanceA2.SetLabel("type", "nic_ixl")

	instanceA3, _ := data.NewInstance("rtp-a700s-02:e5c")
	instanceA3.SetLabel("id", "rtp-a700s-02:e5c")
	instanceA3.SetLabel("speed", "0")
	instanceA3.SetLabel("node", "rtp-a700s-02")
	instanceA3.SetLabel("type", "nic_ixl")

	instanceA4, _ := data.NewInstance("rtp-a700s-02:e5d")
	instanceA4.SetLabel("id", "rtp-a700s-02:e5d")
	instanceA4.SetLabel("speed", "0")
	instanceA4.SetLabel("node", "rtp-a700s-02")
	instanceA4.SetLabel("type", "nic_ixl")

	instanceA5, _ := data.NewInstance("rtp-a700s-02:e5e")
	instanceA5.SetLabel("id", "rtp-a700s-02:e5e")
	instanceA5.SetLabel("speed", "0")
	instanceA5.SetLabel("node", "rtp-a700s-02")
	instanceA5.SetLabel("type", "nic_ixl")

	instanceB1, _ := data.NewInstance("rtp-a700s-01:f5v")
	instanceB1.SetLabel("id", "rtp-a700s-01:f5v")
	instanceB1.SetLabel("speed", "10000M")
	instanceB1.SetLabel("node", "rtp-a700s-01")
	instanceB1.SetLabel("type", "nic_ixl")

	instanceB2, _ := data.NewInstance("rtp-a700s-01:f5w")
	instanceB2.SetLabel("id", "rtp-a700s-01:f5w")
	instanceB2.SetLabel("speed", "10000M")
	instanceB2.SetLabel("node", "rtp-a700s-01")
	instanceB2.SetLabel("type", "nic_ixl")

	instanceB3, _ := data.NewInstance("rtp-a700s-01:f5x")
	instanceB3.SetLabel("id", "rtp-a700s-01:f5x")
	instanceB3.SetLabel("speed", "10000M")
	instanceB3.SetLabel("node", "rtp-a700s-01")
	instanceB3.SetLabel("type", "nic_ixl")

	instanceB4, _ := data.NewInstance("rtp-a700s-01:f5y")
	instanceB4.SetLabel("id", "rtp-a700s-01:f5y")
	instanceB4.SetLabel("speed", "10000M")
	instanceB4.SetLabel("node", "rtp-a700s-01")
	instanceB4.SetLabel("type", "nic_ixl")

	instanceB5, _ := data.NewInstance("rtp-a700s-01:f5z")
	instanceB5.SetLabel("id", "rtp-a700s-01:f5z")
	instanceB5.SetLabel("speed", "10000M")
	instanceB5.SetLabel("node", "rtp-a700s-01")
	instanceB5.SetLabel("type", "nic_ixl")

	// Create latency and ops metrics
	receiveBytes, _ := data.NewMetricFloat64("receive_bytes")
	transmitBytes, _ := data.NewMetricFloat64("transmit_bytes")

	// Set metric values for the instances
	receiveBytes.SetValueFloat64(instanceA1, 2861802356977)
	transmitBytes.SetValueFloat64(instanceA1, 5789662182305)

	receiveBytes.SetValueFloat64(instanceA2, 2861802356977)
	transmitBytes.SetValueFloat64(instanceA2, 5789662182305)

	receiveBytes.SetValueFloat64(instanceA3, 2861802356977)
	transmitBytes.SetValueFloat64(instanceA3, 5789662182305)

	receiveBytes.SetValueFloat64(instanceA4, 2861802356977)
	transmitBytes.SetValueFloat64(instanceA4, 5789662182305)

	receiveBytes.SetValueFloat64(instanceA5, 2861802356977)
	transmitBytes.SetValueFloat64(instanceA5, 5789662182305)

	receiveBytes.SetValueFloat64(instanceB1, 2861802356977)
	transmitBytes.SetValueFloat64(instanceB1, 5789662182305)

	receiveBytes.SetValueFloat64(instanceB2, 2861802356977)
	transmitBytes.SetValueFloat64(instanceB2, 5789662182305)

	receiveBytes.SetValueFloat64(instanceB3, 2861802356977)
	transmitBytes.SetValueFloat64(instanceB3, 5789662182305)

	receiveBytes.SetValueFloat64(instanceB4, 2861802356977)
	transmitBytes.SetValueFloat64(instanceB4, 5789662182305)

	receiveBytes.SetValueFloat64(instanceB5, 2861802356977)
	transmitBytes.SetValueFloat64(instanceB5, 5789662182305)

	dataMap := map[string]*matrix.Matrix{
		"nic": data,
	}

	// Run the plugin
	output, _, err := n.Run(dataMap)
	assert.Nil(t, err)

	// Verify the output
	assert.Equal(t, len(output), 1)

	ifgroupData := output[0]

	// Check for ifgroup instance
	ifgroupInstance1 := ifgroupData.GetInstance("rtp-a700s-01a0a")
	assert.NotNil(t, ifgroupInstance1)

	// Check for ifgroup instance
	ifgroupInstance2 := ifgroupData.GetInstance("rtp-a700s-02a0b")
	assert.Nil(t, ifgroupInstance2)

	assert.Equal(t, ifgroupInstance1.GetLabel("ports"), "f5w,f5x,f5y,f5z")

	// count ifgroup instances
	ifgroupCount := 0
	for _, i := range ifgroupData.GetInstances() {
		if i.IsExportable() {
			ifgroupCount++
		}
	}

	// Verify the number of instances in the ifgroup
	assert.Equal(t, ifgroupCount, expectedCount)

	value, ok := ifgroupData.GetMetric("rx_bytes").GetValueFloat64(ifgroupInstance1)
	assert.True(t, ok)
	assert.Equal(t, value, 11447209427908.0)
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
	n.client = &rest.Client{Metadata: &collector.Metadata{}}
	n.testFilePath = "../../testdata/port-test.json"
	return n
}
