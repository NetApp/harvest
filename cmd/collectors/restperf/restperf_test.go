package restperf

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	rest2 "github.com/netapp/harvest/v2/cmd/collectors/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"
)

const (
	pollerName = "test"
)

func Test_parseMetricResponse(t *testing.T) {
	data, err := os.ReadFile("testdata/submetrics.json")
	if err != nil {
		t.Fatal(err)
	}
	instanceData := gjson.GetBytes(data, "records.0")
	type args struct {
		instanceData gjson.Result
		metric       string
	}
	tests := []struct {
		name string
		args args
		want *metricResponse
	}{
		{
			name: "rss_matrix",
			args: args{
				instanceData: instanceData, metric: "rss_matrix",
			}, want: &metricResponse{
				label:   "tx_frames#queue_0,tx_frames#queue_1,tx_frames#queue_2,tx_bytes#queue_0,tx_bytes#queue_1,tx_bytes#queue_2",
				value:   "6177010,1605252882,0,3,1,4",
				isArray: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := map[string]*rest2.Metric{
				tt.args.metric: {
					Name: tt.args.metric,
				},
			}
			metricResponses := parseMetricResponses(tt.args.instanceData, metrics)
			got := metricResponses[tt.args.metric]
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\ngot  %v\nwant %v", got, tt.want)
			}
		})
	}
}

var (
	ms             []*matrix.Matrix
	benchPerf      *RestPerf
	fullPollData   []rest.PerfRecord
	propertiesData []rest.PerfRecord
)

func TestMain(m *testing.M) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	benchPerf = newRestPerf("Volume", "volume.yaml")
	counters := jsonToPerfRecords("testdata/volume-counters.json")
	_, _ = benchPerf.pollCounter(counters[0].Records.Array())
	now := time.Now().Truncate(time.Second)
	propertiesData = jsonToPerfRecords("testdata/volume-poll-properties.json.gz")
	fullPollData = jsonToPerfRecords("testdata/volume-poll-full.json.gz")
	fullPollData[0].Timestamp = now.UnixNano()
	_, _ = benchPerf.pollInstance(propertiesData[0].Records.Array())
	_, _ = benchPerf.pollData(now, fullPollData)

	os.Exit(m.Run())
}

func BenchmarkRestPerf_PollData(b *testing.B) {
	ms = make([]*matrix.Matrix, 0)
	now := time.Now().Truncate(time.Second)
	fullPollData[0].Timestamp = now.UnixNano()

	for i := 0; i < b.N; i++ {
		now = now.Add(time.Minute * 15)
		fullPollData[0].Timestamp = now.UnixNano()
		mi, _ := benchPerf.pollInstance(propertiesData[0].Records.Array())
		for _, mm := range mi {
			ms = append(ms, mm)
		}
		m, err := benchPerf.pollData(now, fullPollData)
		if err != nil {
			b.Errorf("error: %v", err)
		}
		for _, mm := range m {
			ms = append(ms, mm)
		}
	}
}

func TestRestPerf_pollData(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	tests := []struct {
		name          string
		wantErr       bool
		pollInstance  string
		pollDataPath1 string
		pollDataPath2 string
		numInstances  int
		numMetrics    int
		sum           int64
		pollCounters  string
		counter       string
	}{
		{
			name:          "bytes_read",
			counter:       "bytes_read",
			pollCounters:  "testdata/volume-counters.json",
			pollInstance:  "testdata/volume-poll-instance.json",
			pollDataPath1: "testdata/volume-poll-1.json",
			pollDataPath2: "testdata/volume-poll-2.json",
			numInstances:  2,
			numMetrics:    40,
			sum:           26,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRestPerf("Volume", "volume.yaml")

			counters := jsonToPerfRecords(tt.pollCounters)
			_, err := r.pollCounter(counters[0].Records.Array())
			if err != nil {
				t.Fatal(err)
			}
			pollInstance := jsonToPerfRecords(tt.pollInstance)
			pollData := jsonToPerfRecords(tt.pollDataPath1)
			_, err = r.pollInstance(pollInstance[0].Records.Array())
			if err != nil {
				t.Fatal(err)
			}

			now := time.Now().Truncate(time.Second)
			pollData[0].Timestamp = now.UnixNano()
			_, err = r.pollData(now, pollData)
			if err != nil {
				t.Fatal(err)
			}

			future := now.Add(time.Minute * 15)
			pollData = jsonToPerfRecords(tt.pollDataPath2)
			pollData[0].Timestamp = future.UnixNano()

			got, err := r.pollData(future, pollData)
			if (err != nil) != tt.wantErr {
				t.Errorf("pollData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			m := got["Volume"]
			if len(m.GetInstances()) != tt.numInstances {
				t.Errorf("pollData() numInstances got=%v, want=%v", len(m.GetInstances()), tt.numInstances)
			}

			metadata := r.Metadata
			numMetrics, _ := metadata.GetMetric("metrics").GetValueInt(metadata.GetInstance("data"))
			if numMetrics != tt.numMetrics {
				t.Errorf("pollData() numMetrics got=%v, want=%v", numMetrics, tt.numMetrics)
			}

			var sum int64
			var names []string
			for n := range m.GetInstances() {
				names = append(names, n)
			}
			sort.Strings(names)
			bytesRead := m.GetMetric(tt.counter)
			for _, name := range names {
				i := m.GetInstance(name)
				val, recorded := bytesRead.GetValueInt64(i)
				if !recorded {
					t.Errorf("pollData() recorded = false, want true")
				}
				sum += val
			}
			if sum != tt.sum {
				t.Errorf("pollData() sum got=%v, want=%v", sum, tt.sum)
			}
		})
	}
}

func newRestPerf(object string, path string) *RestPerf {
	var err error
	opts := options.Options{
		Poller:   pollerName,
		HomePath: "testdata",
		IsTest:   true,
	}
	ac := collector.New("RestPerf", object, &opts, params(object, path), nil)
	r := RestPerf{}
	err = r.Init(ac)
	if err != nil {
		panic(err)
	}
	return &r
}

func jsonToPerfRecords(path string) []rest.PerfRecord {
	var (
		perfRecords []rest.PerfRecord
		p           rest.PerfRecord
	)
	gson := collectors.JSONToGson(path, false)
	p = rest.PerfRecord{Records: gson[0], Timestamp: time.Now().UnixNano()}
	perfRecords = append(perfRecords, p)

	return perfRecords
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

func TestQosVolume(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	tests := []struct {
		name          string
		wantErr       bool
		pollDataPath1 string
		pollDataPath2 string
		numInstances  int
		numMetrics    int
		sum           int64
		pollCounters  string
		pollInstance  string
	}{
		{
			name:          "qos_volume_read_latency",
			pollCounters:  "testdata/qos-volume-counters.json",
			pollInstance:  "testdata/qos-volume-getInstances.json",
			pollDataPath1: "testdata/qos-volume-poll-1.json",
			pollDataPath2: "testdata/qos-volume-poll-2.json",
			numInstances:  9,
			numMetrics:    234,
			sum:           81,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRestPerf("WorkloadVolume", "workload_volume.yaml")

			counters := jsonToPerfRecords(tt.pollCounters)
			_, err := r.pollCounter(counters[0].Records.Array())
			if err != nil {
				t.Fatal(err)
			}

			pollInst := jsonToPerfRecords(tt.pollInstance)
			_, err = r.pollInstance(pollInst[0].Records.Array())
			if err != nil {
				t.Fatal(err)
			}

			pollData := jsonToPerfRecords(tt.pollDataPath1)
			now := time.Now().Truncate(time.Second)
			pollData[0].Timestamp = now.UnixNano()
			_, err = r.pollData(now, pollData)
			if err != nil {
				t.Fatal(err)
			}

			future := now.Add(time.Minute * 15)
			pollData = jsonToPerfRecords(tt.pollDataPath2)
			pollData[0].Timestamp = future.UnixNano()

			got, err := r.pollData(future, pollData)
			if (err != nil) != tt.wantErr {
				t.Errorf("pollData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			m := got["WorkloadVolume"]
			if len(m.GetInstances()) != tt.numInstances {
				t.Errorf("pollData() numInstances got=%v, want=%v", len(m.GetInstances()), tt.numInstances)
			}

			metadata := r.Metadata
			numMetrics, _ := metadata.GetMetric("metrics").GetValueInt(metadata.GetInstance("data"))
			if numMetrics != tt.numMetrics {
				t.Errorf("pollData() numMetrics got=%v, want=%v", numMetrics, tt.numMetrics)
			}

			var sum int64
			var names []string
			for n := range m.GetInstances() {
				names = append(names, n)
			}
			sort.Strings(names)
			readLatency := m.GetMetric("read_latency")
			for _, name := range names {
				i := m.GetInstance(name)
				val, recorded := readLatency.GetValueInt64(i)
				if !recorded {
					t.Errorf("pollData() recorded = false, want true")
				}
				sum += val
			}
			if sum != tt.sum {
				t.Errorf("pollData() sum got=%v, want=%v", sum, tt.sum)
			}
		})
	}
}
