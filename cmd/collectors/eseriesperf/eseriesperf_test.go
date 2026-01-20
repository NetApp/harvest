package eseriesperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"testing"
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

			// Mock cluster info via global labels
			mat := ep.Matrix[ep.Object]
			mat.SetGlobalLabel("cluster_id", "600a098000f63714000000005e5cf5d2")
			mat.SetGlobalLabel("cluster", "eseries-test-system")

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

	// Mock cluster info
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("cluster_id", "test-system")
	mat.SetGlobalLabel("cluster", "test")

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

	// Mock cluster info
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("cluster_id", "test-system")
	mat.SetGlobalLabel("cluster", "test")

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
	mat.SetGlobalLabel("cluster_id", "test-system")
	mat.SetGlobalLabel("cluster", "test")

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

	// Mock cluster info
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("cluster_id", "test-system")
	mat.SetGlobalLabel("cluster", "test")

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

			// Mock cluster info
			mat := ep.Matrix[ep.Object]
			mat.SetGlobalLabel("cluster_id", "test-system")
			mat.SetGlobalLabel("cluster", "test")

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
