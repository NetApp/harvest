package zapiperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/rs/zerolog/log"
	"testing"
)

func Test_ZapiPerf(t *testing.T) {
	// Initialize the ZapiPerf collector for Volume object
	z := NewZapiPerf("Volume", "volume.yaml")

	// PollCounter to update the counter detail in cache
	z.testFilePath = "testdata/pollCounter.xml"
	_, _ = z.PollCounter()

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

func NewZapiPerf(object, path string) *ZapiPerf {
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

	z.Object = object
	z.instanceKey = "name"
	z.isCacheEmpty = false
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

func (z *ZapiPerf) testPollInstanceAndData(t *testing.T, pollInstanceFile, pollDataFile string, exportableExportedInstance int) {
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
	for _, instance := range data[z.Object].GetInstances() {
		if instance.IsExportable() {
			exportableInstance++
		}
	}

	if exportableInstance != exportableExportedInstance {
		t.Errorf("Exported instances got= %d, expected: %d", exportableInstance, exportableExportedInstance)
	}
}
