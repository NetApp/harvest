package volume_test

import (
	"log/slog"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	keyperfVolume "github.com/netapp/harvest/v2/cmd/collectors/keyperf/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

const opsKeyPrefix = "temp_"
const styleType = "style"

// TestKeyPerfLatencyRawFlexGroup tests that FlexGroup latency aggregation works correctly
// when metric keys use KeyPerf-style API paths (e.g., "statistics.latency_raw.read")
// instead of RestPerf-style display names (e.g., "read_latency").
func TestKeyPerfLatencyRawFlexGroup(t *testing.T) {
	data := matrix.New("volume", "volume", "volume")
	volumesMap := make(map[string]string)

	// --- Create 3 FlexGroup constituents simulating KeyPerf data ---
	// Constituent 1: latency=20µs, ops=4
	instance1, _ := data.NewInstance("flexgroupharvest__0001")
	instance1.SetLabel("volume", "flexgroupharvest__0001")
	instance1.SetLabel("svm", "svm1")
	instance1.SetLabel("aggr", "aggr1")
	volumesMap["svm1flexgroupharvest__0001"] = "flexgroup_constituent"

	// Constituent 2: latency=30µs, ops=6
	instance2, _ := data.NewInstance("flexgroupharvest__0002")
	instance2.SetLabel("volume", "flexgroupharvest__0002")
	instance2.SetLabel("svm", "svm1")
	instance2.SetLabel("aggr", "aggr2")
	volumesMap["svm1flexgroupharvest__0002"] = "flexgroup_constituent"

	// Constituent 3: latency=40µs, ops=10
	instance3, _ := data.NewInstance("flexgroupharvest__0003")
	instance3.SetLabel("volume", "flexgroupharvest__0003")
	instance3.SetLabel("svm", "svm1")
	instance3.SetLabel("aggr", "aggr3")
	volumesMap["svm1flexgroupharvest__0003"] = "flexgroup_constituent"

	readLatency, _ := data.NewMetricFloat64("statistics.latency_raw.read", "read_latency")
	readLatency.SetComment("statistics.iops_raw.read") // denominator (ops key)
	readLatency.SetProperty("average")

	readOps, _ := data.NewMetricFloat64("statistics.iops_raw.read", "read_ops")
	readOps.SetProperty("rate")

	writeLatency, _ := data.NewMetricFloat64("statistics.latency_raw.write", "write_latency")
	writeLatency.SetComment("statistics.iops_raw.write")
	writeLatency.SetProperty("average")

	writeOps, _ := data.NewMetricFloat64("statistics.iops_raw.write", "write_ops")
	writeOps.SetProperty("rate")

	totalLatency, _ := data.NewMetricFloat64("statistics.latency_raw.total", "avg_latency")
	totalLatency.SetComment("statistics.iops_raw.total")
	totalLatency.SetProperty("average")

	totalOps, _ := data.NewMetricFloat64("statistics.iops_raw.total", "total_ops")
	totalOps.SetProperty("rate")

	readData, _ := data.NewMetricFloat64("statistics.throughput_raw.read", "read_data")
	readData.SetProperty("rate")

	// Constituent 1
	readLatency.SetValueFloat64(instance1, 20)
	readOps.SetValueFloat64(instance1, 4)
	writeLatency.SetValueFloat64(instance1, 100)
	writeOps.SetValueFloat64(instance1, 10)
	totalLatency.SetValueFloat64(instance1, 50)
	totalOps.SetValueFloat64(instance1, 14)
	readData.SetValueFloat64(instance1, 1000)

	// Constituent 2
	readLatency.SetValueFloat64(instance2, 30)
	readOps.SetValueFloat64(instance2, 6)
	writeLatency.SetValueFloat64(instance2, 200)
	writeOps.SetValueFloat64(instance2, 20)
	totalLatency.SetValueFloat64(instance2, 80)
	totalOps.SetValueFloat64(instance2, 26)
	readData.SetValueFloat64(instance2, 2000)

	// Constituent 3
	readLatency.SetValueFloat64(instance3, 40)
	readOps.SetValueFloat64(instance3, 10)
	writeLatency.SetValueFloat64(instance3, 300)
	writeOps.SetValueFloat64(instance3, 30)
	totalLatency.SetValueFloat64(instance3, 120)
	totalOps.SetValueFloat64(instance3, 40)
	readData.SetValueFloat64(instance3, 3000)

	_, _, err := collectors.ProcessFlexGroupData(
		slog.Default(), data, styleType, false, opsKeyPrefix, volumesMap, false)
	assert.Nil(t, err)

	// --- Verify FlexGroup instance was created ---
	fg := data.GetInstance("svm1.flexgroupharvest")
	if fg == nil {
		t.Fatal("expected flexgroup instance 'svm1.flexgroupharvest' to be created")
	}
	assert.Equal(t, fg.GetLabel("volume"), "flexgroupharvest")
	assert.Equal(t, fg.GetLabel(styleType), "flexgroup")

	// --- Verify weighted average latency (the bug fix) ---
	// read_latency = sum(latency_i * ops_i) / sum(ops_i)
	//             = (20*4 + 30*6 + 40*10) / (4+6+10)
	//             = (80 + 180 + 400) / 20
	//             = 660 / 20 = 33.0
	expectedReadLatency := (20.0*4 + 30.0*6 + 40.0*10) / (4.0 + 6.0 + 10.0)
	actualReadLatency, ok := data.GetMetric("statistics.latency_raw.read").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, actualReadLatency, expectedReadLatency)

	// write_latency = (100*10 + 200*20 + 300*30) / (10+20+30)
	//              = (1000 + 4000 + 9000) / 60
	//              = 14000 / 60 ≈ 233.33
	expectedWriteLatency := (100.0*10 + 200.0*20 + 300.0*30) / (10.0 + 20.0 + 30.0)
	actualWriteLatency, ok := data.GetMetric("statistics.latency_raw.write").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, actualWriteLatency, expectedWriteLatency)

	// total latency = (50*14 + 80*26 + 120*40) / (14+26+40)
	//              = (700 + 2080 + 4800) / 80
	//              = 7580 / 80 = 94.75
	expectedTotalLatency := (50.0*14 + 80.0*26 + 120.0*40) / (14.0 + 26.0 + 40.0)
	actualTotalLatency, ok := data.GetMetric("statistics.latency_raw.total").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, actualTotalLatency, expectedTotalLatency)

	// --- Verify ops are summed (not averaged) ---
	actualReadOps, ok := data.GetMetric("statistics.iops_raw.read").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, actualReadOps, 20.0) // 4+6+10

	actualWriteOps, ok := data.GetMetric("statistics.iops_raw.write").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, actualWriteOps, 60.0) // 10+20+30

	actualTotalOps, ok := data.GetMetric("statistics.iops_raw.total").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, actualTotalOps, 80.0) // 14+26+40

	// --- Verify throughput is summed ---
	actualReadData, ok := data.GetMetric("statistics.throughput_raw.read").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, actualReadData, 6000.0) // 1000+2000+3000
}

// TestKeyPerfLatencyRawWithNaN tests that FlexGroup latency aggregation handles
// missing (NaN) constituent values correctly with latency_raw metric keys.
func TestKeyPerfLatencyRawWithNaN(t *testing.T) {
	data := matrix.New("volume", "volume", "volume")
	volumesMap := make(map[string]string)

	// Create 3 constituents, but set one to NaN
	instance1, _ := data.NewInstance("flexgroupharvest__0001")
	instance1.SetLabel("volume", "flexgroupharvest__0001")
	instance1.SetLabel("svm", "svm1")
	instance1.SetLabel("aggr", "aggr1")
	volumesMap["svm1flexgroupharvest__0001"] = "flexgroup_constituent"

	instance2, _ := data.NewInstance("flexgroupharvest__0002")
	instance2.SetLabel("volume", "flexgroupharvest__0002")
	instance2.SetLabel("svm", "svm1")
	instance2.SetLabel("aggr", "aggr2")
	volumesMap["svm1flexgroupharvest__0002"] = "flexgroup_constituent"

	instance3, _ := data.NewInstance("flexgroupharvest__0003")
	instance3.SetLabel("volume", "flexgroupharvest__0003")
	instance3.SetLabel("svm", "svm1")
	instance3.SetLabel("aggr", "aggr3")
	volumesMap["svm1flexgroupharvest__0003"] = "flexgroup_constituent"

	// Create KeyPerf-style metrics
	readLatency, _ := data.NewMetricFloat64("statistics.latency_raw.read", "read_latency")
	readLatency.SetComment("statistics.iops_raw.read")
	readLatency.SetProperty("average")

	readOps, _ := data.NewMetricFloat64("statistics.iops_raw.read", "read_ops")
	readOps.SetProperty("rate")

	// Set values — instance2 has NaN
	readLatency.SetValueFloat64(instance1, 20)
	readOps.SetValueFloat64(instance1, 4)

	readLatency.SetValueNAN(instance2)
	readOps.SetValueNAN(instance2)

	readLatency.SetValueFloat64(instance3, 40)
	readOps.SetValueFloat64(instance3, 10)

	_, _, err := collectors.ProcessFlexGroupData(
		slog.Default(), data, styleType, false, opsKeyPrefix, volumesMap, false,
	)
	assert.Nil(t, err)

	fg := data.GetInstance("svm1.flexgroupharvest")
	if fg == nil {
		t.Fatal("expected flexgroup instance 'svm1.flexgroupharvest' to be created")
	}

	// When a constituent has NaN, the FlexGroup latency should also be NaN
	_, ok := data.GetMetric("statistics.latency_raw.read").GetValueFloat64(fg)
	assert.False(t, ok)
}

func TestKeyPerfVolumePlugin(t *testing.T) {
	params := node.NewS("Volume")
	params.NewChildS("include_constituents", "false")

	opts := options.New()
	opts.IsTest = true
	kpv := keyperfVolume.New(plugin.New("volume", opts, params, nil, "volume", nil))
	if err := kpv.Init(conf.Remote{}); err != nil {
		t.Fatalf("failed to initialize KeyPerf volume plugin: %v", err)
	}

	data := matrix.New("volume", "volume", "volume")

	c1, _ := data.NewInstance("testvol__0001")
	c1.SetLabel("volume", "testvol__0001")
	c1.SetLabel("svm", "svm1")
	c1.SetLabel("aggr", "aggr1")
	c1.SetLabel("style", "flexgroup_constituent")

	c2, _ := data.NewInstance("testvol__0002")
	c2.SetLabel("volume", "testvol__0002")
	c2.SetLabel("svm", "svm1")
	c2.SetLabel("aggr", "aggr2")
	c2.SetLabel("style", "flexgroup_constituent")

	latency, _ := data.NewMetricFloat64("statistics.latency_raw.read", "read_latency")
	latency.SetComment("statistics.iops_raw.read")
	latency.SetProperty("average")

	ops, _ := data.NewMetricFloat64("statistics.iops_raw.read", "read_ops")
	ops.SetProperty("rate")

	latency.SetValueFloat64(c1, 100)
	ops.SetValueFloat64(c1, 50)
	latency.SetValueFloat64(c2, 200)
	ops.SetValueFloat64(c2, 150)

	dataMap := map[string]*matrix.Matrix{"volume": data}
	_, _, err := kpv.Run(dataMap)
	assert.Nil(t, err)

	fg := data.GetInstance("svm1.testvol")
	if fg == nil {
		t.Fatal("expected flexgroup instance 'svm1.testvol' to be created")
	}

	// Weighted average: (100*50 + 200*150) / (50+150) = (5000+30000)/200 = 175.0
	expectedLatency := (100.0*50 + 200.0*150) / (50.0 + 150.0)
	actualLatency, ok := data.GetMetric("statistics.latency_raw.read").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, actualLatency, expectedLatency)

	actualOps, ok := data.GetMetric("statistics.iops_raw.read").GetValueFloat64(fg)
	assert.True(t, ok)
	assert.Equal(t, actualOps, 200.0) // 50+150
}
