package zapiperf

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"math"
	"os"
	"testing"
)

func TestZapiPerfPollCounter(t *testing.T) {
	z := NewZapiPerf("Lun", "lun.yaml")

	expectedCounter := 20

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
	z := NewZapiPerf("Lun", "lun.yaml")

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

	// Case1: pollInstance has 12 records and pollData has 12 records, expected exported instances are 12
	t.Log("Running Case 1")
	z.testPollInstanceAndData(t, "testdata/pollInstance1.xml", "testdata/pollData1.xml", 12)
	if t.Failed() {
		t.Fatal("Case 1 failed")
	}

	// Case2: pollInstance has 13 records and pollData has 12 records, expected exported instances are 13
	t.Log("Running Case 2")
	z.testPollInstanceAndData(t, "testdata/pollInstance2.xml", "testdata/pollData2.xml", 13)
	if t.Failed() {
		t.Fatal("Case 2 failed")
	}

	// Case3: pollInstance has 12 records and pollData has 11 records, expected exported instances are 11
	t.Log("Running Case 3")
	z.testPollInstanceAndData(t, "testdata/pollInstance3.xml", "testdata/pollData3.xml", 11)
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
			for _, instance := range mat.GetInstances() {
				if _, recorded := met.GetValueFloat64(instance); recorded {
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

// TestPollInstanceKeepsCacheOnError checks that a failed instance request is propagated instead of
// being reported as ErrNoInstance, and that it does not wipe the instance cache. Reporting a
// connection error as ErrNoInstance parks the object for the max of 5m and the instance schedule
// (10m by default) instead of retrying in 4s.
func TestPollInstanceKeepsCacheOnError(t *testing.T) {
	z := NewZapiPerf("Lun", "lun.yaml")

	z.testFilePath = "testdata/pollInstance1.xml"
	_, err := z.PollInstance()
	assert.Nil(t, err)

	want := len(z.Matrix[z.Object].GetInstances())
	if want == 0 {
		t.Fatal("instance cache is empty, nothing to protect")
	}

	z.testFilePath = "testdata/doesNotExist.xml"
	_, err = z.PollInstance()

	// the request error must reach the framework, which classifies it, e.g. ErrConnection
	assert.NotNil(t, err)
	assert.False(t, errors.Is(err, errs.ErrNoInstance))
	// a failed poll must not purge the cache
	assert.Equal(t, len(z.Matrix[z.Object].GetInstances()), want)
}

// TestPollInstancePurgesOnSuccess checks that a successful poll still removes instances that ONTAP
// no longer reports.
func TestPollInstancePurgesOnSuccess(t *testing.T) {
	z := NewZapiPerf("Lun", "lun.yaml")

	z.testFilePath = "testdata/pollInstance2.xml"
	_, err := z.PollInstance()
	assert.Nil(t, err)
	assert.Equal(t, len(z.Matrix[z.Object].GetInstances()), 13)

	z.testFilePath = "testdata/pollInstance3.xml"
	_, err = z.PollInstance()
	assert.Nil(t, err)
	assert.Equal(t, len(z.Matrix[z.Object].GetInstances()), 12)
}

// TestPollInstanceNoInstances checks that an object which legitimately has no instances still
// reports ErrNoInstance, and that the stale cache is purged.
func TestPollInstanceNoInstances(t *testing.T) {
	z := NewZapiPerf("Lun", "lun.yaml")

	z.testFilePath = "testdata/pollInstance1.xml"
	_, err := z.PollInstance()
	assert.Nil(t, err)

	z.testFilePath = "testdata/pollInstanceEmpty.xml"
	_, err = z.PollInstance()

	assert.ErrorIs(t, err, errs.ErrNoInstance)
	assert.Equal(t, len(z.Matrix[z.Object].GetInstances()), 0)
}

// newWorkloadZapiPerf returns a ZapiPerf for a workload template. In test mode the Zapi initializer
// loads the template without setting Query, which selects the workload handling.
func newWorkloadZapiPerf(object, path string) *ZapiPerf {
	z := NewZapiPerf(object, path)
	z.Query = z.Params.GetChildContentS("query")
	z.instanceKeys = []string{z.Params.GetChildContentS("instance_key")}
	return z
}

// workload_queue_combined instances and labels must match CmPerf and StatPerf, which read
// api/storage/qos/workloads. The fixtures come from an ONTAP 9.21.1 cluster, plus a qtree workload.
func TestWorkloadQueueInstancesMatchRest(t *testing.T) {
	z := newWorkloadZapiPerf("WorkloadQueue", "workload_queue_combined.yaml")
	assert.Equal(t, z.instanceKeys, []string{"uuid"})

	z.testFilePath = "testdata/workloadQueue/pollCounter.xml"
	_, err := z.PollCounter()
	assert.Nil(t, err)

	z.testFilePath = "testdata/workloadQueue/pollInstance.xml"
	_, err = z.PollInstance()
	assert.Nil(t, err)

	mat := z.Matrix[z.Object]
	// q1-wid48042 is a qtree workload, which is not volume-level
	assert.Equal(t, len(mat.GetInstances()), 3)
	assert.Nil(t, mat.GetInstance("281b4f44-dfbc-11f1-929c-005056a70d2d"))

	tests := []struct {
		uuid string
		want map[string]string
	}{
		// system_defined and preset workloads report the admin SVM in ZAPI but no SVM in REST
		{uuid: "2b14acd7-dfbb-11f1-929c-005056a70d2c", want: map[string]string{
			"class": "system_defined", "policy_group": "_System-Best-Effort", "wid": "1", "workload": "System-Default",
		}},
		{uuid: "2b17f860-dfbb-11f1-929c-005056a70d2c", want: map[string]string{
			"class": "preset", "policy_group": "User-Best-Effort", "wid": "2", "workload": "User-Default",
		}},
		{uuid: "281b4f44-dfbc-11f1-929c-005056a70d2c", want: map[string]string{
			"class": "autovolume", "policy_group": "User-Best-Effort", "svm": "vs0", "volume": "root_vs0", "wid": "48041", "workload": "root_vs0-wid48041",
		}},
	}
	for _, tt := range tests {
		instance := mat.GetInstance(tt.uuid)
		assert.NotNil(t, instance)
		assert.Equal(t, instance.GetLabels(), tt.want)
	}
}

func TestUpdateWorkloadQuery(t *testing.T) {
	tests := []struct {
		object, template string
		wantClass        string
	}{
		{object: "WorkloadQueue", template: "workload_queue_combined.yaml", wantClass: ""},
		{object: "Workload", template: "workload.yaml", wantClass: objWorkloadClass},
	}
	for _, tt := range tests {
		t.Run(tt.object, func(t *testing.T) {
			z := newWorkloadZapiPerf(tt.object, tt.template)
			query := node.NewS("qos-workload-info")
			z.updateWorkloadQuery(query)
			assert.Equal(t, query.GetChildContentS("workload-class"), tt.wantClass)
			assert.Equal(t, query.GetChildS("workload-class") != nil, tt.wantClass != "")
		})
	}
}

// workload_queue_combined averages over iops_total are latencies, so latency_io_reqd zeroes them
// for workloads with too few IOPS, as it does in CmPerf and StatPerf. ZapiPerf timestamps each poll
// with the current time, so the cached timestamps are moved back 60s to give a known interval.
func TestWorkloadQueueLatencyThreshold(t *testing.T) {
	const (
		busyUUID = "281b4f44-dfbc-11f1-929c-005056a70d2c" // 600 ops, 10 IOPS
		idleUUID = "2b14acd7-dfbb-11f1-929c-005056a70d2c" // 2 ops, under 1 IOPS
	)
	z := newWorkloadZapiPerf("WorkloadQueue", "workload_queue_combined.yaml")
	z.testFilePath = "testdata/workloadQueue/pollCounter.xml"
	if _, err := z.PollCounter(); err != nil {
		t.Fatalf("PollCounter: %v", err)
	}
	z.testFilePath = "testdata/workloadQueue/pollInstance.xml"
	if _, err := z.PollInstance(); err != nil {
		t.Fatalf("PollInstance: %v", err)
	}
	z.testFilePath = "testdata/workloadQueue/pollData1.xml"
	if _, err := z.PollData(); err != nil {
		t.Fatalf("first PollData: %v", err)
	}

	cache := z.Matrix[z.Object]
	ts := cache.GetMetric(collectors.TimestampMetricName)
	if ts == nil {
		t.Fatal("timestamp metric missing from cache")
	}
	for _, instance := range cache.GetInstances() {
		if v, ok := ts.GetValueFloat64(instance); ok {
			ts.SetValueFloat64(instance, v-60)
		}
	}

	z.testFilePath = "testdata/workloadQueue/pollData2.xml"
	data, err := z.PollData()
	if err != nil {
		t.Fatalf("second PollData: %v", err)
	}
	m := data[z.Object]
	busy, idle := m.GetInstance(busyUUID), m.GetInstance(idleUUID)
	if busy == nil || idle == nil {
		t.Fatalf("instances missing: busy=%v idle=%v", busy != nil, idle != nil)
	}

	for _, key := range []string{"cpu_dblade_residence_time", "cpu_dblade_background_service_time.direct"} {
		metric := m.GetMetric(key)
		if metric == nil {
			t.Fatalf("metric %s missing", key)
		}
		got, ok := metric.GetValueFloat64(busy)
		assert.True(t, ok)
		assert.Equal(t, got, float64(90-30)/600)

		// without the threshold, this would be (90-30)/2 = 30
		got, ok = metric.GetValueFloat64(idle)
		assert.True(t, ok)
		assert.Equal(t, got, float64(0))
	}

	// iops_total is overridden to rate, but the averages above still divide by its delta
	iops := m.GetMetric("iops_total")
	assert.True(t, iops.IsExportable())
	got, ok := iops.GetValueFloat64(busy)
	assert.True(t, ok)
	if math.Abs(got-10) > 0.1 {
		t.Errorf("iops_total got=%v want=about 10", got)
	}
}
