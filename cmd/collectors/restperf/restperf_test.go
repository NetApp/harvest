package restperf

import (
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

const (
	pollerName = "test"
)

func Test_parseMetricResponse(t *testing.T) {
	bytes, err := os.ReadFile("testdata/submetrics.json")
	if err != nil {
		t.Fatal(err)
	}
	instanceData := gjson.GetBytes(bytes, "records.0")
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
			got := parseMetricResponse(tt.args.instanceData, tt.args.metric)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\ngot  %v\nwant %v", got, tt.want)
			}
		})
	}
}

func TestRestPerf_pollData(t *testing.T) {
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
	}{
		{
			name:          "bytes_read",
			pollCounters:  "testdata/volume-counters.json",
			pollDataPath1: "testdata/volume-poll-1.json",
			pollDataPath2: "testdata/volume-poll-2.json",
			numInstances:  215,
			numMetrics:    3225,
			sum:           237306987,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRestPerf(t)

			counters := jsonToPerfRecords(t, tt.pollCounters)
			_, err := r.pollCounter(counters[0].Records.Array())
			if err != nil {
				t.Fatal(err)
			}

			pollData := jsonToPerfRecords(t, tt.pollDataPath1)
			_, err = r.pollInstance(pollData[0].Records.Array())
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
			pollData = jsonToPerfRecords(t, tt.pollDataPath2)
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
			bytesRead := m.GetMetric("bytes_read")
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

func newRestPerf(t *testing.T) *RestPerf {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	splits := strings.Split(pwd, "/")
	homePath := strings.Join(splits[:len(splits)-3], "/")
	opts := options.Options{
		Poller:   pollerName,
		HomePath: homePath,
	}
	ac := collector.New("RestPerf", "Volume", &opts, params(), nil)
	ac.IsTest = true
	r := RestPerf{}
	err = r.Init(ac)
	if err != nil {
		t.Fatal(err)
	}
	return &r
}

func jsonToPerfRecords(t *testing.T, path string) []rest.PerfRecord {
	var (
		perfRecords []rest.PerfRecord
		p           rest.PerfRecord
	)
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	output := gjson.GetManyBytes(bytes, "records", "num_records", "_links.next.href")

	data := output[0]
	numRecords := output[1]

	if !data.Exists() {
		contentJSON := `{"records":[]}`
		response, err := sjson.SetRawBytes([]byte(contentJSON), "records.-1", bytes)
		if err != nil {
			t.Fatal(err)
		}
		p = rest.PerfRecord{Records: gjson.GetBytes(response, "records")}
	}
	if numRecords.Int() > 0 {
		p = rest.PerfRecord{Records: data, Timestamp: time.Now().UnixNano()}
	}
	perfRecords = append(perfRecords, p)

	return perfRecords
}

func params() *node.Node {
	yml := `
schedule:
  - counter: 9999h
  - instance: 9999h
  - data: 9999h
objects:
  Volume: volume.yaml
`
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		panic(err)
	}
	return root
}
