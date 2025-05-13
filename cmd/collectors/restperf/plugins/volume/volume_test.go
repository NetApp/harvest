package volume_test

import (
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
	if err != nil {
		t.Fatalf("Run method failed: %v", err)
	}

	// Verify the output
	if len(output) != 2 {
		t.Fatalf("expected 2 output matrices, got %d", len(output))
	}

	cache := output[0]
	volumeAggrmetric := output[1]

	// Check for flexgroup instance
	flexgroupInstance := cache.GetInstance("svm1.RahulTest")
	if flexgroupInstance == nil {
		t.Fatalf("expected flexgroup instance 'svm1.RahulTest' not found")
	}

	// Check for flexgroup constituents
	if includeConstituents == "true" {
		for _, suffix := range []string{"0001", "0002", "0003"} {
			instance := data.GetInstance("RahulTest__" + suffix)
			if instance == nil {
				t.Fatalf("expected flexgroup constituent 'svm1.RahulTest__%s' not found", suffix)
			}
			if label := instance.GetLabel("volume"); label != "RahulTest__"+suffix {
				t.Fatalf("expected instance label 'volume' to be 'RahulTest__%s', got '%s'", suffix, label)
			}
		}
	}

	// Check for aggregated metrics
	flexgroupMetricInstance := volumeAggrmetric.GetInstance("svm1.RahulTest")
	if flexgroupMetricInstance == nil {
		t.Fatalf("expected flexgroup metric instance 'svm1.RahulTest' not found")
	}
	if label := flexgroupMetricInstance.GetLabel("volume"); label != "RahulTest" {
		t.Fatalf("expected flexgroup metric instance label 'volume' to be 'RahulTest', got '%s'", label)
	}

	// Verify aggregated ops metric
	if setMetricNaN {
		if _, ok := cache.GetMetric("read_ops").GetValueFloat64(flexgroupInstance); ok {
			t.Errorf("expected metric 'read_ops' for flexgroup instance 'svm1.RahulTest' to be NaN")
		}
	} else if value, ok := cache.GetMetric("read_ops").GetValueFloat64(flexgroupInstance); !ok {
		t.Error("Value [read_ops] missing")
	} else if value != 20 {
		t.Errorf("Value [read_ops] = (%f) incorrect", value)
	}

	// Verify aggregated latency metric (weighted average)
	if setMetricNaN {
		if _, ok := cache.GetMetric("read_latency").GetValueFloat64(flexgroupInstance); ok {
			t.Errorf("expected metric 'read_latency' for flexgroup instance 'svm1.RahulTest' to be NaN")
		}
	} else {
		expectedLatency := (20*4 + 30*6 + 40*10) / 20.0
		if value, ok := cache.GetMetric("read_latency").GetValueFloat64(flexgroupInstance); !ok {
			t.Error("Value [read_latency] missing")
		} else if value != expectedLatency {
			t.Errorf("Value [read_latency] = (%f) incorrect, expected (%f)", value, expectedLatency)
		}
	}

	// Check for simple volume instance
	simpleVolumeInstance := cache.GetInstance("svm1.SimpleVolume")
	if simpleVolumeInstance != nil {
		t.Fatalf("expected simple volume instance 'svm1.SimpleVolume' found")
	}

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
	if currentCount != expectedCount {
		t.Errorf("expected %d instances in the matrix, got %d", expectedCount, currentCount)
	}
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

	aggr := flexgroupInstance.GetLabel("aggr")
	if aggr != "aggr1,aggr2,aggr3" {
		t.Fatalf("expected flexgroup instance 'aggr1,aggr2,aggr3' to be created got '%s'", aggr)
	}

	if value, ok := cache.GetMetric("volume_blocks_footprint_bin0").GetValueFloat64(flexgroupInstance); !ok {
		t.Error("Value [volume_blocks_footprint_bin0] missing")
	} else if value != 70 {
		t.Errorf("Value [volume_blocks_footprint_bin0] = (%f) incorrect, expected 70", value)
	}
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
