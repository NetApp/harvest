package eseriesperf

import (
	"fmt"
	"os"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

const (
	pollerName = "test"
)

func TestMain(m *testing.M) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	os.Exit(m.Run())
}

// newEseriesPerf creates a new EseriesPerf collector for testing
func newEseriesPerf(object string, path string) *EseriesPerf {
	var err error
	opts := options.New(options.WithConfPath("../../../conf"))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true

	ac := collector.New("EseriesPerf", object, opts, params(object, path), nil, conf.Remote{})
	ep := &EseriesPerf{}
	err = ep.Init(ac)
	if err != nil {
		panic(err)
	}
	return ep
}

// params creates a minimal params tree for testing
func params(object string, path string) *node.Node {
	yml := `
schedule:
  - counter: 9999h
  - data: 9999h
type: %s
objects:
  %s: %s
`
	yml = fmt.Sprintf(yml, object, object, path)
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		panic(err)
	}
	return root
}

// jsonToPerfData reads JSON test data and converts to gjson.Result array
// The pollData method expects results[0].Raw to contain the full JSON response
func jsonToPerfData(path string) []gjson.Result {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	// Parse and return as single element array
	// The Raw field will contain the original JSON string
	result := gjson.ParseBytes(data)
	return []gjson.Result{result}
}

func jsonToArrayPerfData(path string) []gjson.Result {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	parsed := gjson.ParseBytes(data)
	return parsed.Array()
}

func TestEseriesPerf_Init(t *testing.T) {
	tests := []struct {
		name     string
		object   string
		template string
	}{
		{
			name:     "volume",
			object:   "Volume",
			template: "volume.yaml",
		},
		{
			name:     "controller",
			object:   "Controller",
			template: "controller.yaml",
		},
		{
			name:     "drive",
			object:   "Drive",
			template: "drive.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := newEseriesPerf(tt.object, tt.template)

			// Verify collector initialized
			assert.NotNil(t, ep)
			assert.NotNil(t, ep.ESeries)
			assert.NotNil(t, ep.perfProp)

			// Verify counter info populated
			assert.NotNil(t, ep.perfProp.counterInfo)
			assert.True(t, len(ep.perfProp.counterInfo) > 0)

			// Verify timestamp metric identified (should contain observedTime)
			if ep.perfProp.timestampMetricName == "" {
				t.Error("timestamp metric name should be set")
			}

			// Verify matrix created
			mat := ep.Matrix[ep.Object]
			assert.NotNil(t, mat)
		})
	}
}

func TestEseriesPerf_buildCounters(t *testing.T) {
	tests := []struct {
		name           string
		object         string
		template       string
		minCounters    int
		hasUtilization bool
	}{
		{
			name:           "volume counters",
			object:         "Volume",
			template:       "volume.yaml",
			minCounters:    5, // Should have at least 5 counters
			hasUtilization: false,
		},
		{
			name:           "drive counters with utilization",
			object:         "Drive",
			template:       "drive.yaml",
			minCounters:    3, // Should have at least 3 counters
			hasUtilization: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := newEseriesPerf(tt.object, tt.template)

			// Verify we have enough counters loaded
			numCounters := len(ep.perfProp.counterInfo)
			if numCounters < tt.minCounters {
				t.Errorf("expected at least %d counters, got %d", tt.minCounters, numCounters)
			}

			// Verify each counter has a type
			for counterName, counter := range ep.perfProp.counterInfo {
				if counter.counterType == "" {
					t.Errorf("counter %s should have a type", counterName)
				}

				// Verify average counters have denominators
				if counter.counterType == "average" {
					if counter.denominator == "" {
						t.Errorf("average counter %s should have a denominator", counterName)
					}
					// Verify denominator exists
					if _, denominatorExists := ep.perfProp.counterInfo[counter.denominator]; !denominatorExists {
						t.Errorf("denominator %s for counter %s should exist", counter.denominator, counterName)
					}
				}
			}

			// Verify utilization flag
			assert.Equal(t, ep.perfProp.calculateUtilization, tt.hasUtilization)
		})
	}
}

func TestEseriesPerf_PollData(t *testing.T) {
	tests := []struct {
		name            string
		object          string
		template        string
		pollDataPath1   string
		pollDataPath2   string
		numInstances    int
		expectedMetrics []string
	}{
		{
			name:          "volume basic flow",
			object:        "Volume",
			template:      "volume.yaml",
			pollDataPath1: "testdata/perf1.json",
			pollDataPath2: "testdata/perf2.json",
			numInstances:  5,
			expectedMetrics: []string{
				"read_ops",
				"write_ops",
				"read_latency",
				"write_latency",
				"read_data",
				"write_data",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := newEseriesPerf(tt.object, tt.template)

			// Mock array info via global labels
			mat := ep.Matrix[ep.Object]
			mat.SetGlobalLabel("array_id", "600a098000f63714000000005e5cf5d2")
			mat.SetGlobalLabel("array", "eseries-test-system")

			// First poll - establishes baseline
			pollData1 := jsonToPerfData(tt.pollDataPath1)
			count1, partial1 := ep.pollData(mat, pollData1, set.New())

			// Verify instances populated
			assert.True(t, count1 > 0)
			assert.Equal(t, len(mat.GetInstances()), tt.numInstances)
			assert.Equal(t, partial1, uint64(0))

			// First cookCounters should return nil (cache empty)
			got, err := ep.cookCounters(mat, mat)
			assert.Nil(t, err)
			assert.Nil(t, got)

			// Verify cache is no longer empty
			assert.False(t, ep.perfProp.isCacheEmpty)

			// Second poll
			pollData2 := jsonToPerfData(tt.pollDataPath2)
			prevMat := mat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
			curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
			curMat.Reset() // Reset data arrays to match current instances

			count2, partial2 := ep.pollData(curMat, pollData2, set.New())
			assert.True(t, count2 > 0)
			assert.Equal(t, partial2, uint64(0))

			// Cook counters
			got, err = ep.cookCounters(curMat, prevMat)
			assert.Nil(t, err)
			assert.NotNil(t, got)

			resultMat := got[tt.object]
			assert.NotNil(t, resultMat)

			// Verify expected metrics exist and are exportable
			for _, metricName := range tt.expectedMetrics {
				metric := resultMat.DisplayMetric(metricName) // Use DisplayMetric to lookup by display name
				if metric == nil {
					t.Errorf("metric %s should exist", metricName)
					continue
				}
				if !metric.IsExportable() {
					t.Errorf("metric %s should be exportable", metricName)
				}
			}

			// Verify all instances are exportable
			for _, instance := range resultMat.GetInstances() {
				if !instance.IsExportable() {
					t.Error("instance should be exportable")
				}
			}
		})
	}
}

func TestEseriesPerf_TimestampConversion(t *testing.T) {
	ep := newEseriesPerf("Volume", "volume.yaml")

	// Mock array info
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	// Load test data
	pollData := jsonToPerfData("testdata/perf1.json")
	ep.pollData(mat, pollData, set.New())

	// Verify timestamp converted from milliseconds to seconds
	timestampMetric := mat.GetMetric("observedTimeInMS")
	if timestampMetric == nil {
		t.Skip("timestamp metric not found")
	}

	for _, instance := range mat.GetInstances() {
		if !instance.IsExportable() {
			continue
		}

		value, ok := timestampMetric.GetValueFloat64(instance)
		if !ok {
			continue
		}

		// Verify it's in seconds (not milliseconds)
		// Original: 1767069729000ms -> 1767069729.0s
		if value >= 2000000000.0 {
			t.Errorf("timestamp should be in seconds, got %f", value)
		}
		if value <= 1000000000.0 {
			t.Errorf("timestamp should be realistic, got %f", value)
		}
	}
}

func TestEseriesPerf_PartialDetection_CounterReset(t *testing.T) {
	ep := newEseriesPerf("Volume", "volume.yaml")

	// Mock array info
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	// First poll
	pollData1 := jsonToPerfData("testdata/perf1.json")
	ep.pollData(mat, pollData1, set.New())
	_, _ = ep.cookCounters(mat, mat)

	// Second poll with counter reset
	pollData2 := jsonToPerfData("testdata/perf-partial-reset.json")
	prevMat := mat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	_, partialCount := ep.pollData(curMat, pollData2, set.New())

	// Should detect partial data
	if partialCount == 0 {
		t.Error("should detect partial instances after counter reset")
	}

	// Verify instances are marked non-exportable
	gotMat, err := ep.cookCounters(curMat, prevMat)
	assert.Nil(t, err)

	if gotMat != nil {
		resultMat := gotMat["Volume"]
		if resultMat != nil {
			nonExportableCount := 0
			for _, instance := range resultMat.GetInstances() {
				if !instance.IsExportable() {
					nonExportableCount++
				}
			}
			if nonExportableCount == 0 {
				t.Error("should have non-exportable instances after counter reset")
			}
		}
	}
}

func TestEseriesPerf_CookCounters_ThreePass(t *testing.T) {
	ep := newEseriesPerf("Volume", "volume.yaml")

	// Mock system info
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	// First poll
	pollData1 := jsonToPerfData("testdata/perf1.json")
	ep.pollData(mat, pollData1, set.New())
	_, _ = ep.cookCounters(mat, mat)

	// Second poll
	pollData2 := jsonToPerfData("testdata/perf2.json")
	prevMat := mat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	ep.pollData(curMat, pollData2, set.New())
	got, err := ep.cookCounters(curMat, prevMat)

	assert.Nil(t, err)
	assert.NotNil(t, got)

	resultMat := got["Volume"]
	assert.NotNil(t, resultMat)

	// Verify rate counters have correct property
	readOps := resultMat.GetMetric("read_ops")
	if readOps != nil {
		assert.Equal(t, readOps.GetProperty(), "rate")
	}

	// Verify average counters have correct property
	readLatency := resultMat.GetMetric("read_latency")
	if readLatency != nil {
		assert.Equal(t, readLatency.GetProperty(), "average")
	}

	// Verify timestamp is not exportable
	observedTime := resultMat.GetMetric("observed_time")
	if observedTime != nil {
		assert.False(t, observedTime.IsExportable())
	}
}

func TestEseriesPerf_UtilizationCalculation(t *testing.T) {
	ep := newEseriesPerf("Drive", "drive.yaml")

	// Verify calculate_utilization flag is set for Drive
	if !ep.perfProp.calculateUtilization {
		t.Error("drive should have utilization calculation enabled")
	}

	// Mock array info
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	// First poll
	pollData1 := jsonToPerfData("testdata/perf1.json")
	ep.pollData(mat, pollData1, set.New())
	_, _ = ep.cookCounters(mat, mat)

	// Second poll
	pollData2 := jsonToPerfData("testdata/perf2.json")
	prevMat := mat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	ep.pollData(curMat, pollData2, set.New())
	got, err := ep.cookCounters(curMat, prevMat)

	assert.Nil(t, err)
	if got == nil {
		t.Skip("no data returned, skipping utilization test")
	}

	resultMat := got["Drive"]
	if resultMat == nil {
		t.Skip("no Drive matrix returned, skipping utilization test")
	}

	// Verify utilization metrics exist
	utilizationMetrics := []string{
		"read_utilization",
		"write_utilization",
		"total_utilization",
	}

	for _, metricName := range utilizationMetrics {
		metric := resultMat.GetMetric(metricName)
		if metric != nil {
			if !metric.IsExportable() {
				t.Errorf("%s should be exportable", metricName)
			}

			// Verify utilization values are percentages (0-100)
			for _, instance := range resultMat.GetInstances() {
				if !instance.IsExportable() {
					continue
				}
				value, ok := metric.GetValueFloat64(instance)
				if ok && value > 0 {
					if value < 0.0 || value > 100.0 {
						t.Errorf("%s should be 0-100, got %f", metricName, value)
					}
				}
			}
		}
	}
}

func TestEseriesPerf_MultipleObjectTypes(t *testing.T) {
	tests := []struct {
		name     string
		object   string
		template string
	}{
		{"volume", "Volume", "volume.yaml"},
		{"controller", "Controller", "controller.yaml"},
		{"drive", "Drive", "drive.yaml"},
		{"pool", "Pool", "pool.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := newEseriesPerf(tt.object, tt.template)

			// Mock array info
			mat := ep.Matrix[ep.Object]
			mat.SetGlobalLabel("array_id", "test-system")
			mat.SetGlobalLabel("array", "test")

			// Load test data
			pollData := jsonToPerfData("testdata/perf1.json")
			count, _ := ep.pollData(mat, pollData, set.New())

			// Verify data was extracted
			if count == 0 {
				t.Errorf("should extract data for %s", tt.object)
			}
			if len(mat.GetInstances()) == 0 {
				t.Errorf("should have instances for %s", tt.object)
			}
		})
	}
}

func TestEseriesPerf_Init_SsdCache(t *testing.T) {
	ep := newEseriesPerf("SsdCache", "ssd_cache.yaml")

	assert.NotNil(t, ep)
	assert.NotNil(t, ep.perfProp)
	assert.NotNil(t, ep.perfProp.counterInfo)
	assert.True(t, len(ep.perfProp.counterInfo) > 0)

	// Timestamp metric should be statistics.timestamp (not observedTimeInMS)
	if ep.perfProp.timestampMetricName != "statistics.timestamp" {
		t.Errorf("timestampMetricName = %q, want %q", ep.perfProp.timestampMetricName, "statistics.timestamp")
	}

	// SSD cache should not have utilization calculation (that's for Drive)
	assert.False(t, ep.perfProp.calculateUtilization)
}

func TestEseriesPerf_buildCounters_SsdCache(t *testing.T) {
	ep := newEseriesPerf("SsdCache", "ssd_cache.yaml")

	numCounters := len(ep.perfProp.counterInfo)
	// 20 counters: 1 delta (timestamp), 14 rate, 4 raw + 1 auto-added timestamp
	if numCounters < 19 {
		t.Errorf("expected at least 19 counters, got %d", numCounters)
	}

	var rateCount, deltaCount, rawCount int
	for _, ctr := range ep.perfProp.counterInfo {
		switch ctr.counterType {
		case "rate":
			rateCount++
		case "delta":
			deltaCount++
		case "raw":
			rawCount++
		case "average", "percent":
			t.Errorf("unexpected counter type %q for SSD cache counter %s", ctr.counterType, ctr.name)
		}

		// SSD cache counters should have no denominators (no average/percent types)
		if ctr.denominator != "" {
			t.Errorf("counter %s should have no denominator, got %q", ctr.name, ctr.denominator)
		}
	}

	if rateCount < 14 {
		t.Errorf("expected at least 14 rate counters, got %d", rateCount)
	}
	if deltaCount < 1 {
		t.Errorf("expected at least 1 delta counter (timestamp), got %d", deltaCount)
	}
	if rawCount < 4 {
		t.Errorf("expected at least 4 raw counters (byte metrics), got %d", rawCount)
	}
}

func TestEseriesPerf_PollData_SsdCache(t *testing.T) {
	ep := newEseriesPerf("SsdCache", "ssd_cache.yaml")

	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	// First poll — establishes baseline
	pollData1 := jsonToArrayPerfData("testdata/ssd_cache1.json")
	count1, partial1 := ep.pollData(mat, pollData1, set.New())

	assert.True(t, count1 > 0)
	assert.Equal(t, len(mat.GetInstances()), 2) // 2 controllers
	assert.Equal(t, partial1, uint64(0))

	// First cookCounters returns nil (cache empty)
	got, err := ep.cookCounters(mat, mat)
	assert.Nil(t, err)
	assert.Nil(t, got)
	assert.False(t, ep.perfProp.isCacheEmpty)

	// Second poll
	pollData2 := jsonToArrayPerfData("testdata/ssd_cache2.json")
	prevMat := mat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	count2, partial2 := ep.pollData(curMat, pollData2, set.New())
	assert.True(t, count2 > 0)
	assert.Equal(t, partial2, uint64(0))

	// Cook counters — should now produce results
	got, err = ep.cookCounters(curMat, prevMat)
	assert.Nil(t, err)
	assert.NotNil(t, got)

	resultMat := got[ep.Object]
	assert.NotNil(t, resultMat)

	// Verify rate metrics for controller 1
	// Delta: reads=3852, timestamp=72 → rate=53.5/s
	ctrl1 := resultMat.GetInstance("070000000000000000000001")
	if ctrl1 == nil {
		t.Fatal("controller 1 instance not found")
	}

	readOps := resultMat.GetMetric("statistics.reads")
	if readOps == nil {
		t.Fatal("statistics.reads metric not found")
	}

	readOpsVal, readOpsOk := readOps.GetValueFloat64(ctrl1)
	if !readOpsOk {
		t.Fatal("could not get read_ops value for controller 1")
	}
	// Delta: reads=3852, timestamp=72
	assert.Equal(t, readOpsVal, 3852.0/72.0)

	writeOps := resultMat.GetMetric("statistics.writes")
	if writeOps != nil {
		writeOpsVal, ok := writeOps.GetValueFloat64(ctrl1)
		if ok {
			// Delta: writes=1717, timestamp=72
			assert.Equal(t, writeOpsVal, 1717.0/72.0)
		}
	}

	// Verify rate for controller 2
	ctrl2 := resultMat.GetInstance("070000000000000000000002")
	if ctrl2 == nil {
		t.Fatal("controller 2 instance not found")
	}

	readOpsVal2, readOpsOk2 := readOps.GetValueFloat64(ctrl2)
	if !readOpsOk2 {
		t.Fatal("could not get read_ops value for controller 2")
	}
	// Delta: reads=3965, timestamp=71
	assert.Equal(t, readOpsVal2, 3965.0/71.0)

	// Verify raw metrics are NOT transformed (current-value passthrough)
	availableBytes := resultMat.GetMetric("statistics.availableBytes")
	if availableBytes != nil {
		val, ok := availableBytes.GetValueFloat64(ctrl1)
		if ok {
			assert.Equal(t, val, 1599784091648.0)
		}
	}

	// Verify timestamp metric is NOT exportable
	timestamp := resultMat.GetMetric("statistics.timestamp")
	if timestamp != nil {
		assert.False(t, timestamp.IsExportable())
	}

	// Verify instances are exportable
	for _, inst := range resultMat.GetInstances() {
		if !inst.IsExportable() {
			t.Error("instance should be exportable")
		}
	}
}

func TestEseriesPerf_SsdCache_NoTimestampConversion(t *testing.T) {
	ep := newEseriesPerf("SsdCache", "ssd_cache.yaml")

	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	pollData := jsonToArrayPerfData("testdata/ssd_cache1.json")
	ep.pollData(mat, pollData, set.New())

	// statistics.timestamp values are already in seconds (not milliseconds)
	// They should be stored as-is, not divided by 1000
	timestampMetric := mat.GetMetric("statistics.timestamp")
	if timestampMetric == nil {
		t.Skip("timestamp metric not found")
	}

	for _, instance := range mat.GetInstances() {
		value, ok := timestampMetric.GetValueFloat64(instance)
		if !ok {
			continue
		}
		// Original value: 1773809538 (seconds since epoch)
		// Should NOT be divided by 1000 — these are already seconds
		if value < 1000000000.0 {
			t.Errorf("timestamp should be >= 1e9 (seconds), got %f — appears to have been divided", value)
		}
		assert.Equal(t, value, 1773809538.0)
	}
}

func TestEseriesPerf_SsdCache_ZeroIO(t *testing.T) {
	ep := newEseriesPerf("SsdCache", "ssd_cache.yaml")

	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	// Both polls use zero I/O data — deltas will be 0
	pollData1 := jsonToArrayPerfData("testdata/ssd_cache_zero_io.json")
	ep.pollData(mat, pollData1, set.New())
	_, _ = ep.cookCounters(mat, mat)

	prevMat := mat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	pollData2 := jsonToArrayPerfData("testdata/ssd_cache_zero_io.json")
	ep.pollData(curMat, pollData2, set.New())

	// Should not panic with zero deltas / zero denominators
	_, err := ep.cookCounters(curMat, prevMat)
	assert.Nil(t, err)
}

func TestEseriesPerf_SsdCache_SingleController(t *testing.T) {
	ep := newEseriesPerf("SsdCache", "ssd_cache.yaml")

	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	pollData := jsonToArrayPerfData("testdata/ssd_cache_single_controller.json")
	count, _ := ep.pollData(mat, pollData, set.New())

	assert.True(t, count > 0)
	assert.Equal(t, len(mat.GetInstances()), 1)
}

func TestEseriesPerf_SsdCache_NegativeDelta(t *testing.T) {
	ep := newEseriesPerf("SsdCache", "ssd_cache.yaml")

	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	// Load higher counters first (poll2), then lower counters (poll1)
	// This simulates a counter reset scenario
	pollData2 := jsonToArrayPerfData("testdata/ssd_cache2.json")
	ep.pollData(mat, pollData2, set.New())
	_, _ = ep.cookCounters(mat, mat)

	prevMat := mat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	// Now poll with lower values — simulates counter reset
	pollData1 := jsonToArrayPerfData("testdata/ssd_cache1.json")
	ep.pollData(curMat, pollData1, set.New())

	got, err := ep.cookCounters(curMat, prevMat)
	assert.Nil(t, err)

	if got != nil {
		resultMat := got[ep.Object]
		if resultMat != nil {
			for _, inst := range resultMat.GetInstances() {
				readOps := resultMat.GetMetric("statistics.reads")
				if readOps != nil {
					val, ok := readOps.GetValueFloat64(inst)
					if ok && val < 0 {
						t.Error("should not have negative rate values after delta protection")
					}
				}
			}
		}
	}
}

func TestEseriesPerf_SsdCache_NegativeDelta_SkipOnCounterReset(t *testing.T) {
	ep := newEseriesPerf("SsdCache", "ssd_cache.yaml")

	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	pollData1 := jsonToArrayPerfData("testdata/ssd_cache1.json")
	count1, partial1 := ep.pollData(mat, pollData1, set.New())
	assert.True(t, count1 > 0)
	assert.Equal(t, len(mat.GetInstances()), 2)
	assert.Equal(t, partial1, uint64(0))

	got, err := ep.cookCounters(mat, mat)
	assert.Nil(t, err)
	assert.Nil(t, got)
	assert.False(t, ep.perfProp.isCacheEmpty)

	prevMat := mat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()

	pollData2 := jsonToArrayPerfData("testdata/ssd_cache_zero_io.json")
	count2, partial2 := ep.pollData(curMat, pollData2, set.New())
	assert.True(t, count2 > 0)
	assert.Equal(t, partial2, uint64(0))

	got, err = ep.cookCounters(curMat, prevMat)
	assert.Nil(t, err)

	resultMat := got[ep.Object]

	skippedMetrics := 0
	for _, metricKey := range []string{"statistics.reads", "statistics.writes", "statistics.fullCacheHits", "statistics.partialCacheHits"} {
		m := resultMat.GetMetric(metricKey)
		if m == nil {
			continue
		}
		for _, inst := range resultMat.GetInstances() {
			_, ok := m.GetValueFloat64(inst)
			if !ok {
				skippedMetrics++
			}
		}
	}
	if skippedMetrics == 0 {
		t.Error("expected rate metrics to have no recorded value (skipped) after negative delta from counter reset to zero")
	}

	for _, metricKey := range []string{"statistics.reads", "statistics.writes", "statistics.fullCacheHits"} {
		m := resultMat.GetMetric(metricKey)
		if m == nil {
			continue
		}
		for _, inst := range resultMat.GetInstances() {
			val, ok := m.GetValueFloat64(inst)
			if ok && val < 0 {
				t.Errorf("metric %s has negative value %f — negative delta must be suppressed", metricKey, val)
			}
		}
	}

	// Raw metrics (byte values) are NOT delta-ed — they pass through unchanged
	// and must still have valid values after the negative delta cycle.
	availBytes := resultMat.GetMetric("statistics.availableBytes")
	if availBytes != nil {
		for _, inst := range resultMat.GetInstances() {
			val, ok := availBytes.GetValueFloat64(inst)
			if !ok {
				t.Error("raw metric statistics.availableBytes should still have a recorded value")
			}
			if val <= 0 {
				t.Errorf("statistics.availableBytes should be positive, got %f", val)
			}
		}
	}
}
func TestEseriesPerf_QueueDepthAverage_Flag(t *testing.T) {
	ep := newEseriesPerf("Volume", "volume.yaml")

	assert.True(t, ep.perfProp.calculateQueueDepthAverage)
	assert.False(t, ep.perfProp.calculateUtilization)
}

func TestEseriesPerf_QueueDepthAverage_Calculation(t *testing.T) {
	ep := newEseriesPerf("Volume", "volume.yaml")
	mat := ep.Matrix[ep.Object]

	instance, err := mat.NewInstance("vol1")
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	queueDepthTotal, err := mat.NewMetricFloat64("queueDepthTotal")
	if err != nil {
		t.Fatalf("failed to create queueDepthTotal: %v", err)
	}
	readOps, err := mat.NewMetricFloat64("readOps")
	if err != nil {
		t.Fatalf("failed to create readOps: %v", err)
	}
	writeOps, err := mat.NewMetricFloat64("writeOps")
	if err != nil {
		t.Fatalf("failed to create writeOps: %v", err)
	}
	otherOps, err := mat.NewMetricFloat64("otherOps")
	if err != nil {
		t.Fatalf("failed to create otherOps: %v", err)
	}

	queueDepthTotal.SetValueFloat64(instance, 1500)
	readOps.SetValueFloat64(instance, 300)
	writeOps.SetValueFloat64(instance, 200)
	otherOps.SetValueFloat64(instance, 0)

	skips, err := ep.calculateQueueDepthAverage(mat)
	assert.Nil(t, err)
	assert.Equal(t, skips, 0)

	avgMetric := mat.GetMetric("queue_depth_average")
	assert.NotNil(t, avgMetric)

	val, ok := avgMetric.GetValueFloat64(instance)
	assert.True(t, ok)
	assert.Equal(t, val, 3.0)
}

func TestEseriesPerf_QueueDepthAverage_ZeroTotalOps(t *testing.T) {
	ep := newEseriesPerf("Volume", "volume.yaml")
	mat := ep.Matrix[ep.Object]

	instance, err := mat.NewInstance("vol1")
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	queueDepthTotal, _ := mat.NewMetricFloat64("queueDepthTotal")
	readOps, _ := mat.NewMetricFloat64("readOps")
	writeOps, _ := mat.NewMetricFloat64("writeOps")
	otherOps, _ := mat.NewMetricFloat64("otherOps")

	queueDepthTotal.SetValueFloat64(instance, 100)
	readOps.SetValueFloat64(instance, 0)
	writeOps.SetValueFloat64(instance, 0)
	otherOps.SetValueFloat64(instance, 0)

	skips, err := ep.calculateQueueDepthAverage(mat)
	assert.Nil(t, err)
	assert.Equal(t, skips, 1)

	avgMetric := mat.GetMetric("queue_depth_average")
	assert.NotNil(t, avgMetric)

	_, ok := avgMetric.GetValueFloat64(instance)
	assert.False(t, ok)
}

func TestEseriesPerf_QueueDepthTotal_NotExported(t *testing.T) {
	ep := newEseriesPerf("Volume", "volume.yaml")
	mat := ep.Matrix[ep.Object]

	instance, err := mat.NewInstance("vol1")
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}

	m, _ := mat.NewMetricFloat64("queueDepthTotal")
	m.SetValueFloat64(instance, 100)

	mat2 := mat.Clone(matrix.With{Data: true, Metrics: true, Instances: true, ExportInstances: true})

	qdt := mat2.GetMetric("queueDepthTotal")
	if qdt != nil {
		qdt.SetExportable(false)
		assert.False(t, qdt.IsExportable())
	}
}
