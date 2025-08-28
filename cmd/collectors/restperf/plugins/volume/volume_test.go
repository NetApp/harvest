package volume_test

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	volume2 "github.com/netapp/harvest/v2/cmd/collectors/restperf/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/collectors/zapiperf/plugins/volume"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"log/slog"
	"strconv"
	"testing"

	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

const OpsKeyPrefix = "temp_"
const StyleType = "style"

// Common test logic for RestPerf/ZapiPerf Volume plugin
func runVolumeTest(t *testing.T, createVolume func(params *node.Node) plugin.Plugin, includeConstituents string, expectedCount int, setMetricNaN bool) {
	params := node.NewS("Volume")
	params.NewChildS("include_constituents", includeConstituents)
	v := createVolume(params)
	volumesMap := make(map[string]string)

	// Initialize the plugin
	if err := v.Init(conf.Remote{}); err != nil {
		t.Fatalf("failed to initialize plugin: %v", err)
	}

	// Create test data
	data := matrix.New("volume", "volume", "volume")
	instance1, _ := data.NewInstance("RahulTest__0001")
	instance1.SetLabel("volume", "RahulTest__0001")
	instance1.SetLabel("svm", "svm1")
	instance1.SetLabel("aggr", "aggr1")
	volumesMap["svm1"+"RahulTest__0001"] = "flexgroup_constituent"

	instance2, _ := data.NewInstance("RahulTest__0002")
	instance2.SetLabel("volume", "RahulTest__0002")
	instance2.SetLabel("svm", "svm1")
	instance2.SetLabel("aggr", "aggr2")
	volumesMap["svm1"+"RahulTest__0002"] = "flexgroup_constituent"

	instance3, _ := data.NewInstance("RahulTest__0003")
	instance3.SetLabel("volume", "RahulTest__0003")
	instance3.SetLabel("svm", "svm1")
	instance3.SetLabel("aggr", "aggr3")
	volumesMap["svm1"+"RahulTest__0003"] = "flexgroup_constituent"

	// Create a simple volume instance
	simpleInstance, _ := data.NewInstance("SimpleVolume")
	simpleInstance.SetLabel("volume", "SimpleVolume")
	simpleInstance.SetLabel("svm", "svm1")
	simpleInstance.SetLabel("aggr", "aggr4")
	volumesMap["svm1"+"SimpleVolume"] = "flexvol"

	// Create latency and ops metrics
	latencyMetric, _ := data.NewMetricFloat64("read_latency")
	latencyMetric.SetComment("read_ops")
	latencyMetric.SetProperty("average")

	opsMetric, _ := data.NewMetricFloat64("read_ops")
	opsMetric.SetProperty("rate")

	// Set metric values for the instances
	latencyMetric.SetValueFloat64(instance1, 20)
	opsMetric.SetValueFloat64(instance1, 4)

	latencyMetric.SetValueFloat64(instance2, 30)
	opsMetric.SetValueFloat64(instance2, 6)

	latencyMetric.SetValueFloat64(instance3, 40)
	opsMetric.SetValueFloat64(instance3, 10)

	// Optionally set one metric value to NaN
	if setMetricNaN {
		latencyMetric.SetValueNAN(instance2)
		opsMetric.SetValueNAN(instance2)
	}

	// Set metric values for the simple volume instance
	latencyMetric.SetValueFloat64(simpleInstance, 50)
	opsMetric.SetValueFloat64(simpleInstance, 5)

	// Run the plugin
	boolValue, _ := strconv.ParseBool(includeConstituents)
	output, _, err := collectors.ProcessFlexGroupData(slog.Default(), data, StyleType, boolValue, OpsKeyPrefix, volumesMap, true)
	assert.Nil(t, err)

	// Verify the output
	assert.Equal(t, len(output), 2)

	cache := output[0]
	volumeAggrmetric := output[1]

	// Check for flexgroup instance
	flexgroupInstance := cache.GetInstance("svm1.RahulTest")
	assert.NotNil(t, flexgroupInstance)

	// Check for flexgroup constituents
	if includeConstituents == "true" {
		for _, suffix := range []string{"0001", "0002", "0003"} {
			instance := data.GetInstance("RahulTest__" + suffix)
			assert.NotNil(t, instance)
			assert.Equal(t, instance.GetLabel("volume"), "RahulTest__"+suffix)
		}
	}

	// Check for aggregated metrics
	flexgroupMetricInstance := volumeAggrmetric.GetInstance("svm1.RahulTest")
	assert.NotNil(t, flexgroupMetricInstance)
	assert.Equal(t, flexgroupMetricInstance.GetLabel("volume"), "RahulTest")

	// Verify aggregated ops metric
	if setMetricNaN {
		assert.True(t, flexgroupMetricInstance.IsExportable())
		_, ok := cache.GetMetric("read_ops").GetValueFloat64(flexgroupInstance)
		assert.False(t, ok)
	} else {
		value, ok := cache.GetMetric("read_ops").GetValueFloat64(flexgroupInstance)
		assert.True(t, ok)
		assert.Equal(t, value, 20.0)
	}

	// Verify aggregated latency metric (weighted average)
	if setMetricNaN {
		_, ok := cache.GetMetric("read_latency").GetValueFloat64(flexgroupInstance)
		assert.False(t, ok)
	} else {
		expectedLatency := (20*4 + 30*6 + 40*10) / 20.0
		value, ok := cache.GetMetric("read_latency").GetValueFloat64(flexgroupInstance)
		assert.True(t, ok)
		assert.Equal(t, value, expectedLatency)
	}

	// Check for simple volume instance
	simpleVolumeInstance := cache.GetInstance("svm1.SimpleVolume")
	assert.Nil(t, simpleVolumeInstance)

	// count instances in both data and cache
	currentCount := 0
	for _, i := range data.GetInstances() {
		if i.IsExportable() {
			currentCount++
		}
	}

	for _, i := range cache.GetInstances() {
		if i.IsExportable() {
			currentCount++
		}
	}

	// Verify the number of instances in the cache
	assert.Equal(t, currentCount, expectedCount)
}

func TestRunForAllImplementations(t *testing.T) {
	testCases := []struct {
		name                string
		createVolume        func(params *node.Node) plugin.Plugin
		includeConstituents string
		expectedCount       int
		setMetricNaN        bool
	}{
		{
			name:                "REST include_constituents=true",
			createVolume:        createRestVolume,
			includeConstituents: "true",
			expectedCount:       5, // 3 constituents + 1 flexgroup + 1 flexvol
			setMetricNaN:        false,
		},
		{
			name:                "REST include_constituents=false",
			createVolume:        createRestVolume,
			includeConstituents: "false",
			expectedCount:       2, // Only 1 flexgroup + 1 flexvol
			setMetricNaN:        false,
		},
		{
			name:                "ZAPI include_constituents=true",
			createVolume:        createZapiVolume,
			includeConstituents: "true",
			expectedCount:       5, // 3 constituents + 1 flexgroup + 1 flexvol
			setMetricNaN:        false,
		},
		{
			name:                "ZAPI include_constituents=false",
			createVolume:        createZapiVolume,
			includeConstituents: "false",
			expectedCount:       2, // Only 1 flexgroup + 1 flexvol
			setMetricNaN:        false,
		},
		{
			name:                "REST include_constituents=true with NaN metric",
			createVolume:        createRestVolume,
			includeConstituents: "true",
			expectedCount:       5, // 3 constituents + 1 flexgroup + 1 flexvol
			setMetricNaN:        true,
		},
		{
			name:                "ZAPI include_constituents=true with NaN metric",
			createVolume:        createZapiVolume,
			includeConstituents: "true",
			expectedCount:       5, // 3 constituents + 1 flexgroup + 1 flexvol
			setMetricNaN:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runVolumeTest(t, tc.createVolume, tc.includeConstituents, tc.expectedCount, tc.setMetricNaN)
		})
	}
}

func TestProcessFlexGroupFootPrint(t *testing.T) {
	logger := slog.Default()
	data := matrix.New("volume", "volume", "volume")

	// Create test data
	instance1, _ := data.NewInstance("RahulTest__0001")
	instance1.SetLabel("volume", "RahulTest__0001")
	instance1.SetLabel("svm", "svm1")
	instance1.SetLabel("style", "flexgroup_constituent")
	instance1.SetLabel("aggr", "aggr1")

	instance2, _ := data.NewInstance("RahulTest__0002")
	instance2.SetLabel("volume", "RahulTest__0002")
	instance2.SetLabel("svm", "svm1")
	instance2.SetLabel("style", "flexgroup_constituent")
	instance2.SetLabel("aggr", "aggr2")

	instance3, _ := data.NewInstance("RahulTest__0003")
	instance3.SetLabel("volume", "RahulTest__0003")
	instance3.SetLabel("svm", "svm1")
	instance3.SetLabel("style", "flexgroup_constituent")
	instance3.SetLabel("aggr", "aggr3")

	footprintMetric, _ := data.NewMetricFloat64("volume_blocks_footprint_bin0")
	footprintMetric.SetValueFloat64(instance1, 20)
	footprintMetric.SetValueFloat64(instance2, 50)
	// Intentionally leave instance2 without a footprint value to test missing data handling

	cache := collectors.ProcessFlexGroupFootPrint(data, logger)

	flexgroupInstance := cache.GetInstance("svm1.RahulTest")
	if flexgroupInstance == nil {
		t.Fatalf("expected flexgroup instance 'svm1.RahulTest' to be created")
	}

	assert.Equal(t, flexgroupInstance.GetLabel("aggr"), "aggr1,aggr2,aggr3")

	value, ok := cache.GetMetric("volume_blocks_footprint_bin0").GetValueFloat64(flexgroupInstance)
	assert.True(t, ok)
	assert.Equal(t, value, float64(70))
}

func createRestVolume(params *node.Node) plugin.Plugin {
	opts := options.New(options.WithConfPath("testdata/conf"))
	opts.IsTest = true
	v := &volume2.Volume{AbstractPlugin: plugin.New("volume", opts, params, nil, "volume", nil)}
	v.SLogger = slog.Default()
	return v
}

func createZapiVolume(params *node.Node) plugin.Plugin {
	opts := options.New(options.WithConfPath("testdata/conf"))
	opts.IsTest = true
	v := &volume.Volume{AbstractPlugin: plugin.New("volume", opts, params, nil, "volume", nil)}
	v.SLogger = slog.Default()
	return v
}
