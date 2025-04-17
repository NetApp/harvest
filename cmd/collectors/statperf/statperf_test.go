package statperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

const (
	pollerName = "test"
)

func TestRestPerf_pollData(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	tests := []struct {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newStatPerf("flexcache_per_volume", "flexcache.yaml")

			counters, err := getDataJSON(tt.pollCounters)
			if err != nil {
				t.Fatalf("error: %v", err)
			}

			_, err = s.pollCounter(counters, 0)
			if err != nil {
				t.Fatal(err)
			}

			instances, err := getDataJSON(tt.pollInstance)
			if err != nil {
				t.Fatalf("error: %v", err)
			}

			_, err = s.pollInstance(s.Matrix[s.Object], instances, 0)
			if err != nil {
				t.Fatal(err)
			}

			curTime := time.Now().UnixNano() / util.BILLION

			data, err := getDataJSON(tt.pollDataPath1)
			if err != nil {
				t.Fatalf("error: %v", err)
			}

			prevMat := s.Matrix[s.Object]
			_, _, err = processAndCookCounters(s, data, prevMat, float64(curTime))
			if err != nil {
				t.Fatal(err)
			}

			future := curTime + 60

			data2, err := getDataJSON(tt.pollDataPath2)
			if err != nil {
				t.Fatalf("error: %v", err)
			}

			prevMat = s.Matrix[s.Object]
			got, metricCount, err := processAndCookCounters(s, data2, prevMat, float64(future))
			if (err != nil) != tt.wantErr {
				t.Errorf("pollData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			m := got["flexcache_per_volume"]
			if len(m.GetInstances()) != tt.numInstances {
				t.Errorf("pollData() numInstances got=%v, want=%v", len(m.GetInstances()), tt.numInstances)
			}

			if metricCount != tt.numMetrics {
				t.Errorf("pollData() numMetrics got=%v, want=%v", metricCount, tt.numMetrics)
			}

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

func TestPollCounter(t *testing.T) {
	var (
		err error
	)
	s := newStatPerf("flexcache_per_volume", "flexcache.yaml")

	counters, err := getDataJSON("testdata/counters.txt")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	_, err = s.pollCounter(counters, 0)
	if err != nil {
		t.Fatalf("Failed to fetch poll counter %v", err)
	}

	if len(s.Prop.Metrics) != len(s.perfProp.counterInfo) {
		t.Errorf("Prop metrics and counterInfo size should be same")
	}
}

func getDataJSON(filePath string) ([]gjson.Result, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	jsonContent := `"` + strings.TrimSpace(string(content)) + `"`
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

func processAndCookCounters(s *StatPerf, pollData []gjson.Result, prevMat *matrix.Matrix, ts float64) (map[string]*matrix.Matrix, uint64, error) {
	curMat := prevMat.Clone(matrix.With{Data: false, Metrics: true, Instances: true, ExportInstances: true})
	curMat.Reset()
	metricCount, _ := s.processPerfRecords(pollData, curMat, prevMat, ts)
	got, err := s.cookCounters(curMat, prevMat)
	return got, metricCount, err
}
