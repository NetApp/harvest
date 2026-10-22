package statperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/poller/plugin/aggregator"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"math"
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
	curMat := prevMat.CloneForCollection()
	curMat.Reset()
	metricCount, _, _ := s.processPerfRecords(pollData, curMat, prevMat)
	got, err := s.cookCounters(curMat, prevMat)
	return got, metricCount, err
}

func pollFlexlog(t *testing.T, s *StatPerf) *matrix.Matrix {
	t.Helper()
	counters, err := getDataJSON("testdata/flexlog/counters.txt")
	assert.Nil(t, err)
	assert.Nil(t, s.pollCounter(counters, 0))

	instances, err := getDataJSON("testdata/flexlog/instances.txt")
	assert.Nil(t, err)
	_, err = s.pollInstance(s.Matrix[s.Object], instances, 0)
	assert.Nil(t, err)

	var got map[string]*matrix.Matrix
	for _, path := range []string{"testdata/flexlog/data_1.txt", "testdata/flexlog/data_2.txt"} {
		data, err := getDataJSON(path)
		assert.Nil(t, err)
		got, _, err = processAndCookCounters(s, data, s.Matrix[s.Object])
		assert.Nil(t, err)
	}
	return got[s.Object]
}

func flexlogValue(t *testing.T, m *matrix.Matrix, key string, nodeName string) float64 {
	t.Helper()
	metric := m.GetMetric(key)
	if metric == nil {
		t.Fatalf("metric %s missing", key)
	}
	for _, instance := range m.GetInstances() {
		if instance.GetLabel("node") == nodeName {
			v, ok := metric.GetValueFloat64(instance)
			assert.True(t, ok)
			return v
		}
	}
	t.Fatalf("instance for node %s missing", nodeName)
	return 0
}

func TestCookCounters_ArrayBaseCounter(t *testing.T) {
	s := newStatPerf("wafl_flexlog", "wafl_flexlog.yaml")
	m := pollFlexlog(t, s)

	tests := []struct {
		key  string
		node string
		want float64
	}{
		{"ldma_latency#wafl_spinnp_write", "panaffa400x05", 4646045.0 / 113031},
		{"rdma_latency#wafl_spinnp_write", "panaffa400x05", 5044872.0 / 113031},
		{"ldma_latency#wafl_write", "panaffa400x05", 14639.0 / 1435},
		{"ldma_latency#wafl_spinnp_write", "panaffa400x06", 1697122.0 / 37701},
		{"ldma_latency#wafl_write", "panaffa400x06", 5045.0 / 497},
		{"ldma_latency#null_message", "panaffa400x05", 0},
		{"logging_count#wafl_spinnp_write", "panaffa400x05", 113031},
	}
	for _, tt := range tests {
		got := flexlogValue(t, m, tt.key, tt.node)
		assert.True(t, math.Abs(got-tt.want) < 1e-9)
	}
}

func TestCookCounters_ArrayBaseLatencyThreshold(t *testing.T) {
	s := newStatPerf("wafl_flexlog", "wafl_flexlog.yaml")
	// 61 s between samples, so an element needs more than 610 logging_count ops
	s.perfProp.latencyIoReqd = 10
	m := pollFlexlog(t, s)

	got := flexlogValue(t, m, "ldma_latency#wafl_write", "panaffa400x05")
	assert.True(t, math.Abs(got-14639.0/1435) < 1e-9)

	got = flexlogValue(t, m, "ldma_latency#wafl_write", "panaffa400x06")
	assert.Equal(t, got, 0.0)
}

func TestCookCounters_ArrayBaseAggregator(t *testing.T) {
	s := newStatPerf("wafl_flexlog", "wafl_flexlog.yaml")
	m := pollFlexlog(t, s)
	for _, instance := range m.GetInstances() {
		instance.SetLabel("cluster", "cluster_1")
	}

	params := node.NewS("Aggregator")
	params.NewChildS("", "cluster")
	agg := &aggregator.Aggregator{AbstractPlugin: plugin.New("Test", nil, params, nil, m.Object, nil)}
	assert.Nil(t, agg.Init(conf.Remote{}))

	results, _, err := agg.Run(map[string]*matrix.Matrix{m.Object: m})
	assert.Nil(t, err)
	assert.Equal(t, len(results), 1)

	cluster := results[0].GetInstance("cluster_1")
	assert.NotNil(t, cluster)
	got, ok := results[0].GetMetric("ldma_latency#wafl_spinnp_write").GetValueFloat64(cluster)
	assert.True(t, ok)
	// weighted by each node's logging_count delta
	assert.True(t, math.Abs(got-(4646045.0+1697122.0)/(113031+37701)) < 1e-9)
}

func TestDenominatorKey(t *testing.T) {
	m := matrix.New("uuid", "object", "object")
	for _, key := range []string{"lat#a", "lat#b", "count#a", "ops", "scalar_lat"} {
		metric, err := m.NewMetricFloat64(key)
		assert.Nil(t, err)
		metric.SetArray(strings.Contains(key, arrayKeyToken))
	}

	tests := []struct {
		key         string
		denominator string
		want        string
	}{
		{"lat#a", "count", "count#a"},
		{"lat#b", "count", "count"},
		{"lat#a", "ops", "ops"},
		{"scalar_lat", "ops", "ops"},
		{"lat#a", "", ""},
	}
	for _, tt := range tests {
		got := denominatorKey(m, m.GetMetric(tt.key), tt.key, &counter{denominator: tt.denominator})
		assert.Equal(t, got, tt.want)
	}
}

func TestParseRows_MultiLineInstanceUUID(t *testing.T) {
	s := newStatPerf("nvm_mirror", "nvm_mirror.yaml")

	content, err := os.ReadFile("testdata/nvm_mirror_data.txt")
	assert.Nil(t, err)

	groups, err := s.parseRows(string(content))
	assert.Nil(t, err)
	assert.Equal(t, len(groups), 2)

	assert.Equal(t, groups[0]["instance_name"], "NVM Mirror")
	assert.Equal(t, groups[0]["instance_uuid"], "sa-tme-flexpod-a800-rdma-01:kernel:NVM Mirror")
	assert.Equal(t, groups[0]["node_name"], "sa-tme-flexpod-a800-rdma-01")
	assert.Equal(t, groups[0]["write_throughput"], "13847639326958")

	assert.Equal(t, groups[1]["instance_name"], "NVM Mirror")
	assert.Equal(t, groups[1]["instance_uuid"], "sa-tme-flexpod-a800-rdma-02:kernel:NVM Mirror")
	assert.Equal(t, groups[1]["node_name"], "sa-tme-flexpod-a800-rdma-02")
	assert.Equal(t, groups[1]["write_throughput"], "629627001678")
}
