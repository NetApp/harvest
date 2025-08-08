package zapiperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"os"
	"testing"
)

func TestZapiPerfPollCounter(t *testing.T) {
	z := NewZapiPerf("Volume", "volume.yaml")

	expectedCounter := 27

	z.testFilePath = "testdata/pollCounter.xml"
	if _, err := z.PollCounter(); err != nil {
		t.Fatalf("Failed to fetch poll counter %v", err)
	}

	if len(z.scalarCounters) != expectedCounter {
		t.Errorf("counter count got=%d, expected=%d", len(z.scalarCounters), expectedCounter)
	}

	z.testFilePath = "testdata/pollCounter.xml"
	if _, err := z.PollCounter(); err != nil {
		t.Fatalf("Failed to fetch poll counter %v", err)
	}

	if len(z.scalarCounters) != expectedCounter {
		t.Errorf("counter count got=%d, expected=%d", len(z.scalarCounters), expectedCounter)
	}
}

func TestZapiPerfSequence(t *testing.T) {
	// Initialize the ZapiPerf collector for Volume object
	z := NewZapiPerf("Volume", "volume.yaml")

	// PollCounter to update the counter detail in cache
	z.testFilePath = "testdata/pollCounter.xml"
	if _, err := z.PollCounter(); err != nil {
		t.Fatalf("Failed to fetch poll counter %v", err)
	}

	// First poll
	t.Log("Running First poll")
	z.testPollInstanceAndData(t, "testdata/pollInstance1.xml", "testdata/pollData1.xml", 0)
	if t.Failed() {
		t.Fatal("First poll failed")
	}

	// Case1: pollInstance has 5 records and pollData has 5 records, expected exported instances are 5
	t.Log("Running Case 1")
	z.testPollInstanceAndData(t, "testdata/pollInstance1.xml", "testdata/pollData1.xml", 5)
	if t.Failed() {
		t.Fatal("Case 1 failed")
	}

	// Case2: pollInstance has 6 records and pollData has 7 records, expected exported instances are 6
	t.Log("Running Case 2")
	z.testPollInstanceAndData(t, "testdata/pollInstance2.xml", "testdata/pollData2.xml", 6)
	if t.Failed() {
		t.Fatal("Case 2 failed")
	}

	// Case3: pollInstance has 5 records and pollData has 3 records, expected exported instances are 3
	t.Log("Running Case 3")
	z.testPollInstanceAndData(t, "testdata/pollInstance3.xml", "testdata/pollData3.xml", 3)
	if t.Failed() {
		t.Fatal("Case 3 failed")
	}
}

func TestSkipsSequence(t *testing.T) {
	z := NewZapiPerf("Qtree", "qtree.yaml")

	z.testFilePath = "testdata/skips/pollCounter.xml"
	if _, err := z.PollCounter(); err != nil {
		t.Fatalf("Failed to fetch poll counter %v", err)
	}
	z.testFilePath = "testdata/skips/pollInstance1.xml"
	if _, err := z.PollInstance(); err != nil {
		t.Fatalf("Failed to fetch poll instance %v", err)
	}

	// First Poll
	t.Log("Running First Poll")
	z.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData1.xml", 0, 0)
	if t.Failed() {
		t.Fatal("First Poll failed")
	}

	// Complete Poll
	t.Log("Running Complete Poll")
	z.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData1.xml", 1, 4)
	if t.Failed() {
		t.Fatal("Complete Poll failed")
	}

	// Skips Poll
	t.Log("Running Skips Poll")
	z.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData2.xml", 1, 3)
	if t.Failed() {
		t.Fatal("Skips Poll failed")
	}

	// Skips Poll 2
	t.Log("Running Skips Poll 2")
	z.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData3.xml", 1, 3)
	if t.Failed() {
		t.Fatal("Skips Poll 2 failed")
	}

	t.Log("Running Poll 3")
	z.testPollInstanceAndDataWithMetrics(t, "testdata/skips/pollData4.xml", 1, 4)
	if t.Failed() {
		t.Fatal("Poll 3 failed")
	}
}

type testCase struct {
	name                    string
	dataFile                string
	expectedExportedInst    int
	expectedExportedMetrics int
}

func testPartialAggregationHelper(t *testing.T, z *ZapiPerf, testDataDir string, cases []testCase) {
	for _, tc := range cases {
		t.Log("Running " + tc.name)
		z.testPollInstanceAndDataWithMetrics(t, testDataDir+tc.dataFile, tc.expectedExportedInst, tc.expectedExportedMetrics)
		if t.Failed() {
			t.Fatal(tc.name + " failed")
		}
	}
}

func TestPartialAggregationSequence(t *testing.T) {
	z := NewZapiPerf("Workload", "workload.yaml")
	z.testFilePath = "testdata/partialAggregation/pollCounter.xml"
	if _, err := z.PollCounter(); err != nil {
		t.Fatalf("Failed to fetch poll counter %v", err)
	}
	z.testFilePath = "testdata/partialAggregation/pollInstance1.xml"
	if _, err := z.PollInstance(); err != nil {
		t.Fatalf("Failed to fetch poll instance %v", err)
	}
	testPartialAggregationHelper(t, z, "testdata/partialAggregation/", []testCase{
		{"First Poll", "pollData1.xml", 0, 0},
		{"Complete Poll", "pollData1.xml", 2, 48},
		{"Partial Poll", "pollData2.xml", 0, 0},
		{"Partial Poll 2", "pollData2.xml", 0, 0},
		{"First Complete Poll After Partial", "pollData1.xml", 2, 0},
		{"Second Complete Poll After Partial", "pollData1.xml", 2, 48},
		{"Partial Poll 3", "pollData2.xml", 0, 0},
	})
}

func TestAllowPartialAggregationSequence(t *testing.T) {
	z := NewZapiPerf("SystemNode", "system_node.yaml")
	z.testFilePath = "testdata/allowPartialAggregation/pollCounter.xml"
	if _, err := z.PollCounter(); err != nil {
		t.Fatalf("Failed to fetch poll counter %v", err)
	}
	z.testFilePath = "testdata/allowPartialAggregation/pollInstance1.xml"
	if _, err := z.PollInstance(); err != nil {
		t.Fatalf("Failed to fetch poll instance %v", err)
	}
	testPartialAggregationHelper(t, z, "testdata/allowPartialAggregation/", []testCase{
		{"First Poll", "pollData1.xml", 0, 0},
		{"Complete Poll", "pollData1.xml", 4, 8},
		{"Partial Poll", "pollData2.xml", 3, 6},
		{"Partial Poll 2", "pollData2.xml", 3, 6},
		{"First Complete Poll After Partial", "pollData1.xml", 4, 6},
		{"Second Complete Poll After Partial", "pollData1.xml", 4, 8},
		{"Partial Poll 3", "pollData2.xml", 3, 6},
	})
}

// New method specifically for TestPartialAggregation
func (z *ZapiPerf) testPollInstanceAndDataWithMetrics(t *testing.T, pollDataFile string, expectedExportedInst, expectedExportedMetrics int) {
	// Additional logic to count metrics
	z.testFilePath = pollDataFile
	data, err := z.PollData()
	if err != nil {
		t.Fatalf("Failed to fetch poll data %v", err)
	}

	totalMetrics := 0
	exportableInstance := 0
	mat := data[z.Object]
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
}

func NewZapiPerf(object, path string) *ZapiPerf {
	// homepath is harvest directory level
	homePath := "../../../"
	conf.TestLoadHarvestConfig("testdata/config.yml")
	opts := options.New(options.WithConfPath(homePath + "/conf"))
	opts.Poller = "testZapiperf"
	opts.HomePath = homePath
	opts.IsTest = true

	ac := collector.New("Zapiperf", object, opts, params(object, path), nil, conf.Remote{})
	z := &ZapiPerf{}
	if err := z.Init(ac); err != nil {
		slog.Error("", slogx.Err(err))
		os.Exit(1)
	}

	z.Object = object
	z.instanceKeys = []string{"name"}
	z.isCacheEmpty = true
	mx := matrix.New(z.Object, z.Object, z.Object)
	z.Matrix = make(map[string]*matrix.Matrix)
	z.Matrix[z.Object] = mx
	return z
}

func params(object string, path string) *node.Node {
	yml := `
schedule:
  - counter: 9999h
  - instance: 9999h
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

func (z *ZapiPerf) testPollInstanceAndData(t *testing.T, pollInstanceFile, pollDataFile string, expectedExportedInst int) {
	// PollInstance test
	z.testFilePath = pollInstanceFile
	_, _ = z.PollInstance()

	// PollData test
	z.testFilePath = pollDataFile
	data, err := z.PollData()
	if err != nil {
		t.Fatalf("Failed to fetch poll data %v", err)
	}

	exportableInstance := 0
	mat := data[z.Object]
	if mat != nil {
		for _, instance := range mat.GetInstances() {
			if instance.IsExportable() {
				exportableInstance++
			}
		}
	}

	if exportableInstance != expectedExportedInst {
		t.Errorf("Exported instances got= %d, expected: %d", exportableInstance, expectedExportedInst)
	}
}
