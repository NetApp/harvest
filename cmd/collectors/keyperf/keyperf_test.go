package keyperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"sort"
	"testing"
	"time"
)

const (
	pollerName = "test"
)

func TestPartialAggregationSequence(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	kp := newKeyPerf("Volume", "volume.yaml")

	// First Poll
	t.Log("Running First Poll")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/volume-poll-1.json", 0, 0)

	// Complete Poll
	t.Log("Running Complete Poll")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/volume-poll-2.json", 4, 48)

	// Partial Poll
	t.Log("Running Partial Poll")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/volume-poll-partial.json", 4, 36)

	// Partial Poll 2
	t.Log("Running Partial Poll 2")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/volume-poll-partial.json", 4, 36)
	if t.Failed() {
		t.Fatal("Partial Poll 2 failed")
	}

	// First Complete Poll After Partial
	t.Log("Running First Complete Poll After Partial")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/volume-poll-3.json", 4, 36)
	if t.Failed() {
		t.Fatal("First Complete Poll After Partial failed")
	}

	// Second Complete Poll After Partial
	t.Log("Running First Complete Poll After Partial")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/volume-poll-3.json", 4, 48)
	if t.Failed() {
		t.Fatal("First Complete Poll After Partial failed")
	}

	// Partial Poll 3
	t.Log("Running Partial Poll 3")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/volume-poll-partial-2.json", 4, 36)
	if t.Failed() {
		t.Fatal("Partial Poll 3 failed")
	}
}

func (kp *KeyPerf) testPollInstanceAndDataWithMetrics(t *testing.T, pollDataFile string, expectedExportedInst, expectedExportedMetrics int) *matrix.Matrix {
	// Additional logic to count metrics
	pollData := collectors.JSONToGson(pollDataFile, true)
	now := time.Now().Truncate(time.Second)
	data, err := kp.pollData(now, pollData, nil)
	if err != nil {
		t.Fatal(err)
	}

	totalMetrics := 0
	exportableInstance := 0
	mat := data[kp.Object]
	if mat != nil {
		for _, instance := range mat.GetInstances() {
			if instance.IsExportable() {
				exportableInstance++
			}
		}
		for _, met := range mat.GetMetrics() {
			if !met.IsExportable() {
				continue
			}
			records := met.GetRecords()
			for _, v := range records {
				if v {
					totalMetrics++
				}
			}
		}
	}

	if exportableInstance != expectedExportedInst {
		t.Errorf("Exported instances got=%d, expected=%d", exportableInstance, expectedExportedInst)
	}

	// Check if the total number of metrics matches the expected value
	if totalMetrics != expectedExportedMetrics {
		t.Errorf("Total metrics got=%d, expected=%d", totalMetrics, expectedExportedMetrics)
	}
	return mat
}

func TestKeyPerf_pollData(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	tests := []struct {
		name          string
		wantErr       bool
		pollDataPath1 string
		pollDataPath2 string
		counter       string
		sum           int64
		numInstances  int
		numMetrics    int
		record        bool
	}{
		{
			name:          "statistics.iops_raw.read",
			counter:       "statistics.iops_raw.read",
			pollDataPath1: "testdata/volume-poll-1.json",
			pollDataPath2: "testdata/volume-poll-2.json",
			numInstances:  4,
			numMetrics:    48,
			sum:           4608,
			record:        true,
		},
		{
			name:          "statistics.latency_raw.read",
			counter:       "statistics.latency_raw.read",
			pollDataPath1: "testdata/volume-poll-1.json",
			pollDataPath2: "testdata/volume-poll-2.json",
			numInstances:  4,
			numMetrics:    48,
			sum:           1114,
			record:        true,
		},
		{
			name:          "statistics.latency_raw.read",
			counter:       "statistics.latency_raw.read",
			pollDataPath1: "testdata/missingStats/volume-poll-1.json",
			pollDataPath2: "testdata/missingStats/volume-poll-2.json",
			numInstances:  1,
			numMetrics:    0,
			sum:           0,
			record:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			kp := newKeyPerf("Volume", "volume.yaml")
			// First poll data
			kp.testPollInstanceAndDataWithMetrics(t, tt.pollDataPath1, 0, 0)
			// Complete Poll
			m := kp.testPollInstanceAndDataWithMetrics(t, tt.pollDataPath2, tt.numInstances, tt.numMetrics)

			var sum int64
			var names []string
			for n := range m.GetInstances() {
				names = append(names, n)
			}
			sort.Strings(names)
			metric := m.GetMetric(tt.counter)
			for _, name := range names {
				i := m.GetInstance(name)
				val, recorded := metric.GetValueInt64(i)
				if recorded != tt.record {
					t.Errorf("pollData() recorded got=%v, want=%v", recorded, tt.record)
				}
				sum += val
			}
			if sum != tt.sum {
				t.Errorf("pollData() sum got=%v, want=%v", sum, tt.sum)
			}
		})
	}
}

func newKeyPerf(object string, path string) *KeyPerf {
	var err error
	opts := options.New(options.WithConfPath("testdata/conf"))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true

	ac := collector.New("KeyPerf", object, opts, params(object, path), nil, conf.Remote{})
	kp := KeyPerf{}
	err = kp.Init(ac)
	if err != nil {
		panic(err)
	}
	return &kp
}

func params(object string, path string) *node.Node {
	yml := `
schedule:
  - counter: 9999h
  - data: 9999h
objects:
  %s: %s
`
	yml = fmt.Sprintf(yml, object, path)
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		panic(err)
	}
	return root
}
