package statperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"sort"
	"strings"
	"testing"
)

const (
	pollerName = "test"
)

func TestStatPerf_pollData(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	tests := []struct {
		object        string
		path          string
		name          string
		wantErr       bool
		pollInstance  string
		pollDataPath1 string
		pollDataPath2 string
		numInstances  int
		numMetrics    uint64
		sum           int64
		pollCounters  string
		counter       string
		record        bool
	}{
		{
			object:        "flexcache_per_volume",
			path:          "flexcache.yaml",
			name:          "blocks_requested_from_client",
			counter:       "blocks_requested_from_client",
			pollCounters:  "testdata/counters.txt",
			pollInstance:  "testdata/instances.txt",
			pollDataPath1: "testdata/data.txt",
			pollDataPath2: "testdata/data-2.txt",
			numInstances:  6,
			numMetrics:    108,
			sum:           30000000,
			record:        true,
		},
		{
			object:        "flexcache_per_volume",
			path:          "flexcache.yaml",
			name:          "blocks_requested_from_client",
			counter:       "blocks_requested_from_client",
			pollCounters:  "testdata/counters_1.txt",
			pollInstance:  "testdata/instances_1.txt",
			pollDataPath1: "testdata/data_1.txt",
			pollDataPath2: "testdata/data-2.txt",
			numInstances:  6,
			numMetrics:    108,
			sum:           30000000,
			record:        true,
		},
		{
			object:        "flexcache_per_volume",
			path:          "flexcache.yaml",
			name:          "blocks_requested_from_client",
			counter:       "blocks_requested_from_client",
			pollCounters:  "testdata/counters_2.txt",
			pollInstance:  "testdata/instances_2.txt",
			pollDataPath1: "testdata/data_2.txt",
			pollDataPath2: "testdata/data-2.txt",
			numInstances:  6,
			numMetrics:    108,
			sum:           30000000,
			record:        true,
		},
		{
			object:        "flexcache_per_volume",
			path:          "flexcache.yaml",
			name:          "blocks_requested_from_client",
			counter:       "blocks_requested_from_client",
			pollCounters:  "testdata/counters_2.txt",
			pollInstance:  "testdata/instances_2.txt",
			pollDataPath1: "testdata/data.txt",
			pollDataPath2: "testdata/data-2-partial.txt",
			numInstances:  0,
			numMetrics:    108,
			sum:           30000000,
			record:        false,
		},
		{
			object:        "lun",
			path:          "lun.yaml",
			name:          "read_align_histo",
			counter:       "read_align_histo#0",
			pollCounters:  "testdata/array/lun_counters.txt",
			pollInstance:  "testdata/array/lun_instances.txt",
			pollDataPath1: "testdata/array/lun_data.txt",
			pollDataPath2: "testdata/array/lun_data_1.txt",
			numInstances:  5,
			numMetrics:    195,
			sum:           90,
			record:        true,
		},
		{
			object:        "lun",
			path:          "lun.yaml",
			name:          "read_ops",
			counter:       "read_ops",
			pollCounters:  "testdata/array/lun_counters.txt",
			pollInstance:  "testdata/array/lun_instances.txt",
			pollDataPath1: "testdata/partialAggregation/data_1.txt",
			pollDataPath2: "testdata/partialAggregation/data_2.txt",
			numInstances:  0,
			numMetrics:    195,
			sum:           10000,
			record:        false,
		},
		{
			object:        "node",
			path:          "system_node.yaml",
			name:          "cpu_busy",
			counter:       "cpu_busy",
			pollCounters:  "testdata/allowPartialAggregation/counters.txt",
			pollInstance:  "testdata/allowPartialAggregation/instances.txt",
			pollDataPath1: "testdata/allowPartialAggregation/data_1.txt",
			pollDataPath2: "testdata/allowPartialAggregation/data_2.txt",
			numInstances:  3,
			numMetrics:    150,
			sum:           26,
			record:        true,
		},
		{
			object:        "nvm_mirror",
			path:          "nvm_mirror.yaml",
			name:          "write_throughput",
			counter:       "write_throughput",
			pollCounters:  "testdata/space/counters.txt",
			pollInstance:  "testdata/space/instances.txt",
			pollDataPath1: "testdata/space/data_1.txt",
			pollDataPath2: "testdata/space/data_2.txt",
			numInstances:  2,
			numMetrics:    8,
			sum:           2711128,
			record:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStatPerf(tt.object, tt.path)

			counters, err := getDataJSON(tt.pollCounters)
			assert.Nil(t, err)

			err = s.pollCounter(counters, 0)
			assert.Nil(t, err)

			instances, err := getDataJSON(tt.pollInstance)
			assert.Nil(t, err)

			_, err = s.pollInstance(s.Matrix[s.Object], instances, 0)
			assert.Nil(t, err)

			data, err := getDataJSON(tt.pollDataPath1)
			assert.Nil(t, err)

			prevMat := s.Matrix[s.Object]
			_, _, err = processAndCookCounters(s, data, prevMat)
			assert.Nil(t, err)

			data2, err := getDataJSON(tt.pollDataPath2)
			assert.Nil(t, err)

			prevMat = s.Matrix[s.Object]
			got, metricCount, err := processAndCookCounters(s, data2, prevMat)
			if err != nil {
				assert.True(t, tt.wantErr)
				return
			}

			m := got[tt.object]
			var exportInstances int
			// collect exported instances
			for _, instance := range m.GetInstances() {
				if instance.IsExportable() {
					exportInstances++
				}
			}
			assert.Equal(t, exportInstances, tt.numInstances)
			assert.Equal(t, metricCount, tt.numMetrics)

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
				assert.Equal(t, recorded, tt.record)
				sum += val
			}
			assert.Equal(t, sum, tt.sum)
		})
	}
}

func TestPollCounter(t *testing.T) {
	var (
		err error
	)
	s := newStatPerf("flexcache_per_volume", "flexcache.yaml")

	counters, err := getDataJSON("testdata/counters.txt")
	assert.Nil(t, err)
	err = s.pollCounter(counters, 0)
	assert.Nil(t, err)

	assert.Equal(t, len(s.Prop.Metrics), len(s.perfProp.counterInfo))
}

func getDataJSON(filePath string) ([]gjson.Result, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	jsonContent := `"` + strings.ReplaceAll(strings.TrimSpace(string(content)), `"`, ``) + `"`
	var counters = []gjson.Result{gjson.Parse(jsonContent)}
	return counters, nil
}

func newStatPerf(object string, path string) *StatPerf {
	var err error
	homePath := "../../../"
	conf.TestLoadHarvestConfig("testdata/config.yml")
	opts := options.New(options.WithConfPath(homePath + "/conf"))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true

	ac := collector.New("StatPerf", object, opts, params(object, path), nil, conf.Remote{})
	s := StatPerf{}
	err = s.Init(ac)
	if err != nil {
		panic(err)
	}
	return &s
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

func processAndCookCounters(s *StatPerf, pollData []gjson.Result, prevMat *matrix.Matrix) (map[string]*matrix.Matrix, uint64, error) {
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()
	metricCount, _, _ := s.processPerfRecords(pollData, curMat, prevMat)
	got, err := s.cookCounters(curMat, prevMat)
	return got, metricCount, err
}
