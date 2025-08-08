package keyperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	collector2 "github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"sort"
	"testing"
	"time"
)

const (
	pollerName = "test"
)

func TestSkipsSequence(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	kp := newKeyPerf("Volume", "volume.yaml")

	// First Poll
	t.Log("Running First Poll")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData1.json", 0, 0)
	if t.Failed() {
		t.Fatal("First Poll failed")
	}

	// Complete Poll
	t.Log("Running Complete Poll")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData1.json", 1, 12)
	if t.Failed() {
		t.Fatal("Complete Poll failed")
	}

	// Skips Poll
	t.Log("Running Skips Poll")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData2.json", 1, 10)
	if t.Failed() {
		t.Fatal("Skips Poll failed")
	}

	// Skips Poll 2
	t.Log("Running Skips Poll 2")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData3.json", 1, 10)
	if t.Failed() {
		t.Fatal("Skips Poll 2 failed")
	}

	t.Log("Running Poll 3")
	kp.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData4.json", 1, 12)
	if t.Failed() {
		t.Fatal("Poll 3 failed")
	}
}

func TestPartialAndAllowPartialAggregationSequence(t *testing.T) {
	cases := []struct {
		name      string
		object    string
		yaml      string
		pollFiles []struct {
			label                   string
			file                    string
			expectedExportedInst    int
			expectedExportedMetrics int
		}
	}{
		{
			name:   "PartialAggregationSequence",
			object: "Volume",
			yaml:   "volume.yaml",
			pollFiles: []struct {
				label                   string
				file                    string
				expectedExportedInst    int
				expectedExportedMetrics int
			}{
				{"First Poll", "testdata/partialAggregation/poll-1.json", 0, 0},
				{"Complete Poll", "testdata/partialAggregation/poll-2.json", 4, 48},
				{"Partial Poll", "testdata/partialAggregation/poll-partial.json", 3, 36},
				{"Partial Poll 2", "testdata/partialAggregation/poll-partial.json", 3, 36},
				{"First Complete Poll After Partial", "testdata/partialAggregation/poll-3.json", 4, 36},
				{"Second Complete Poll After Partial", "testdata/partialAggregation/poll-3.json", 4, 48},
				{"Partial Poll 3", "testdata/partialAggregation/poll-partial-2.json", 3, 36},
			},
		},
		{
			name:   "AllowPartialAggregationSequence",
			object: "SystemNode",
			yaml:   "system_node.yaml",
			pollFiles: []struct {
				label                   string
				file                    string
				expectedExportedInst    int
				expectedExportedMetrics int
			}{
				{"First Poll", "testdata/allowPartialAggregation/poll-1.json", 0, 0},
				{"Complete Poll", "testdata/allowPartialAggregation/poll-2.json", 4, 4},
				{"Partial Poll", "testdata/allowPartialAggregation/poll-partial.json", 4, 3},
				{"Partial Poll 2", "testdata/allowPartialAggregation/poll-partial.json", 4, 3},
				{"First Complete Poll After Partial", "testdata/allowPartialAggregation/poll-3.json", 4, 3},
				{"Second Complete Poll After Partial", "testdata/allowPartialAggregation/poll-3.json", 4, 4},
				{"Partial Poll 3", "testdata/allowPartialAggregation/poll-partial-2.json", 4, 3},
			},
		},
	}

	conf.TestLoadHarvestConfig("testdata/config.yml")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kp := newKeyPerf(tc.object, tc.yaml)
			for _, pf := range tc.pollFiles {
				t.Log("Running " + pf.label)
				kp.testPollInstanceAndDataWithMetrics(t, pf.file, pf.expectedExportedInst, pf.expectedExportedMetrics)
				if t.Failed() {
					t.Fatalf("%s failed", pf.label)
				}
			}
		})
	}
}

func (kp *KeyPerf) testPollInstanceAndDataWithMetrics(t *testing.T, pollDataFile string, expectedExportedInst, expectedExportedMetrics int) *matrix.Matrix {
	// Additional logic to count metrics
	prevMat := kp.Matrix[kp.Object]
	pollData := collectors.JSONToGson(pollDataFile, true)
	got, _, err := processAndCookCounters(kp, pollData, prevMat)
	if err != nil {
		t.Fatal(err)
	}

	totalMetrics := 0
	exportableInstance := 0
	mat := got[kp.Object]
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

func processAndCookCounters(kp *KeyPerf, pollData []gjson.Result, prevMat *matrix.Matrix) (map[string]*matrix.Matrix, uint64, error) {
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()
	metricCount, _, _ := kp.processPerfRecords(pollData, curMat, set.New(), time.Now().UnixNano()/collector2.BILLION)
	got, err := kp.cookCounters(curMat, prevMat)
	return got, metricCount, err
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
		checksum      bool
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
			checksum:      true,
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
			checksum:      true,
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
			checksum:      true,
		},
		{
			name:          "statistics.latency_raw.read",
			counter:       "statistics.latency_raw.read",
			pollDataPath1: "testdata/missingTimestamp/volume-poll-1.json",
			pollDataPath2: "testdata/missingTimestamp/volume-poll-2.json",
			numInstances:  4,
			numMetrics:    48,
			sum:           4608,
			record:        true,
			checksum:      false, // given timestamp is synthesized so calculated metrics value may differ with very run
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
			if sum != tt.sum && tt.checksum {
				t.Errorf("pollData() sum got=%v, want=%v", sum, tt.sum)
			}
		})
	}
}

func newKeyPerf(object string, path string) *KeyPerf {
	var err error
	homePath := "../../../"
	opts := options.New(options.WithConfPath(homePath + "/conf"))
	// An additional histogram metric is being tested, so load the data from the test dataset.
	if path == "volume.yaml" {
		opts = options.New(options.WithConfPath("testdata/conf"))
	}
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
