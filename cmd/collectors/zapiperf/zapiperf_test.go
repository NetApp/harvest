package zapiperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/rs/zerolog/log"
	"os"
	"testing"
	"time"
)

func Test_ZapiPerf(t *testing.T) {
	// Initialize the ZapiPerf collector for Volume object
	z := NewZapiperf("Volume", "volume.yaml")

	// pollCounter to update the counter detail in cache
	z.testPollCounter(t, "testdata/pollCounter.xml")

	// Case1: pollInstance has 5 records and pollData has 5 records, expected exported instances are 5
	expectedInstances := 5
	z.testPollInstanceAndData(t, "testdata/pollInstance1.xml", "testdata/pollData1.xml", expectedInstances)

	// Case2: pollInstance has 6 records and pollData has 7 records, expected exported instances are 6
	expectedInstances = 6
	z.testPollInstanceAndData(t, "testdata/pollInstance2.xml", "testdata/pollData2.xml", expectedInstances)

	// Case3: pollInstance has 5 records and pollData has 3 records, expected exported instances are 3
	expectedInstances = 3
	z.testPollInstanceAndData(t, "testdata/pollInstance3.xml", "testdata/pollData3.xml", expectedInstances)
}

func NewZapiperf(object, path string) *ZapiPerf {
	// homepath is harvest directory level
	homePath := "../../../"
	zapiperfPoller := "testZapiperf"

	conf.TestLoadHarvestConfig("testdata/config.yml")
	opts := options.Options{
		Poller:   zapiperfPoller,
		HomePath: homePath,
		IsTest:   true,
	}

	ac := collector.New("Zapiperf", object, &opts, params(object, path), nil)
	z := &ZapiPerf{}
	if err := z.Init(ac); err != nil {
		log.Fatal().Err(err)
	}

	// Test for volume object
	z.Object = object
	z.instanceKey = "name"
	z.isCacheEmpty = false
	mx := matrix.New(z.Object, z.Object, z.Object)
	z.Matrix = make(map[string]*matrix.Matrix)
	z.Matrix[z.Object] = mx
	return z
}

func readFile(path string) *node.Node {
	var root *node.Node
	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if root, err = tree.LoadXML(bytes); err != nil {
		panic(err)
	}
	return root
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

func (z *ZapiPerf) testPollInstanceAndData(t *testing.T, pollInstanceFile, pollDataFile string, exportableExportedInstance int) {
	// pollInstance test
	instancesAttr := "attributes-list"
	nameAttr := "name"
	uuidAttr := "uuid"
	keyAttr := z.instanceKey
	oldInstances := set.New()
	mat := z.Matrix[z.Object]
	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}
	pollInstanceResult := readFile(pollInstanceFile)
	_ = z.pollInstance(pollInstanceResult, instancesAttr, keyAttr, nameAttr, uuidAttr, oldInstances, mat)

	// pollData test
	apiT := 1 * time.Second
	parseT := 1 * time.Second
	start := time.Now()
	keyName := "instances"
	pollDataResult := readFile(pollDataFile)
	prevMat := z.Matrix[z.Object]
	// clone matrix without numeric data and non-exportable all instances
	curMat := prevMat.Clone(false, true, true, false)
	curMat.Reset()
	timestamp := curMat.GetMetric("timestamp")
	if timestamp == nil {
		t.Fatal("missing timestamp metric")
	}
	// for updating metadata
	count := uint64(0)
	apiT += time.Since(start.Add(2 * time.Second))
	parseT += time.Since(start)
	if err := z.pollData(pollDataResult, curMat, timestamp, &count); err != nil {
		t.Fatalf("Failed to fetch poll data %v", err)
	}

	// processData test
	data, e := z.processData(&apiT, &parseT, curMat, prevMat, keyName, count, uint64(len(curMat.GetInstanceKeys())))
	if e != nil {
		t.Fatalf("Failed to process data %v", e)
	}

	exportableInstance := 0
	for _, instance := range data[z.Object].GetInstances() {
		if instance.IsExportable() {
			exportableInstance++
		}
	}

	if exportableInstance != exportableExportedInstance {
		t.Errorf("Exported instances got= %d, expected: %d", exportableInstance, exportableExportedInstance)
	}
}

func (z *ZapiPerf) testPollCounter(t *testing.T, pollCounterFile string) {
	wanted, oldLabels, oldMetrics, mat, oldLabelsSize, oldMetricsSize, err := z.parseCounters()
	if err != nil {
		t.Fatalf("Error missing parameter counters %v", err)
	}
	pollCounterResult := readFile(pollCounterFile)
	// pollCounter test
	_, _ = z.pollCounter(pollCounterResult, wanted, oldLabels, oldMetrics, mat, oldLabelsSize, oldMetricsSize)
}
