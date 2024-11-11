package restperf

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
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
				label:   "queue_0#tx_frames,queue_1#tx_frames,queue_2#tx_frames,queue_0#tx_bytes,queue_1#tx_bytes,queue_2#tx_bytes",
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
			diff := cmp.Diff(got, tt.want, cmp.AllowUnexported(metricResponse{}))
			if diff != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff)
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
	counters := jsonToPerfRecords("testdata/volume-counters-1.json")
	_, _ = benchPerf.pollCounter(counters[0].Records.Array(), 0)
	now := time.Now().Truncate(time.Second)
	propertiesData = jsonToPerfRecords("testdata/volume-poll-properties.json.gz")
	fullPollData = jsonToPerfRecords("testdata/volume-poll-full.json.gz")
	fullPollData[0].Timestamp = now.UnixNano()
	mat := matrix.New("Volume", "Volume", "Volume")
	_, _ = benchPerf.pollInstance(mat, perfToJSON(propertiesData), 0)
	_, _ = benchPerf.pollData(now, fullPollData)

	os.Exit(m.Run())
}

func BenchmarkRestPerf_PollData(b *testing.B) {
	ms = make([]*matrix.Matrix, 0)
	now := time.Now().Truncate(time.Second)
	fullPollData[0].Timestamp = now.UnixNano()

	for range b.N {
		now = now.Add(time.Minute * 15)
		fullPollData[0].Timestamp = now.UnixNano()
		mat := matrix.New("Volume", "Volume", "Volume")
		mi, _ := benchPerf.pollInstance(mat, perfToJSON(propertiesData), 0)
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
		pollCounters2 string
		counter       string
		record        bool
	}{
		{
			name:          "bytes_read",
			counter:       "bytes_read",
			pollCounters:  "testdata/volume-counters-1.json",
			pollCounters2: "testdata/volume-counters-2.json",
			pollInstance:  "testdata/volume-poll-instance.json",
			pollDataPath1: "testdata/volume-poll-1.json",
			pollDataPath2: "testdata/volume-poll-2.json",
			numInstances:  2,
			numMetrics:    40,
			sum:           26,
			record:        true,
		},
		{
			name:          "abc",
			counter:       "abc",
			pollCounters:  "testdata/volume-counters-1.json",
			pollCounters2: "testdata/volume-counters-2.json",
			pollInstance:  "testdata/volume-poll-instance.json",
			pollDataPath1: "testdata/volume-poll-1.json",
			pollDataPath2: "testdata/volume-poll-3.json",
			numInstances:  2,
			numMetrics:    42,
			sum:           526336,
			record:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRestPerf("Volume", "volume.yaml")

			counters := jsonToPerfRecords(tt.pollCounters)
			_, err := r.pollCounter(counters[0].Records.Array(), 0)
			if err != nil {
				t.Fatal(err)
			}
			pollInstance := jsonToPerfRecords(tt.pollInstance)
			pollData := jsonToPerfRecords(tt.pollDataPath1)
			_, err = r.pollInstance(r.Matrix[r.Object], perfToJSON(pollInstance), 0)
			if err != nil {
				t.Fatal(err)
			}

			now := time.Now().Truncate(time.Second)
			pollData[0].Timestamp = now.UnixNano()
			_, err = r.pollData(now, pollData)
			if err != nil {
				t.Fatal(err)
			}
			counters = jsonToPerfRecords(tt.pollCounters2)
			_, err = r.pollCounter(counters[0].Records.Array(), 0)
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

func (r *RestPerf) testPollInstanceAndDataWithMetrics(t *testing.T, pollDataFile string, expectedExportedInst, expectedExportedMetrics int) {
	// Additional logic to count metrics
	pollData := jsonToPerfRecords(pollDataFile)
	data, err := r.pollData(time.Now().Truncate(time.Second), pollData)
	if err != nil {
		t.Fatal(err)
	}

	totalMetrics := 0
	exportableInstance := 0
	mat := data[r.Object]
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

func TestPollCounter(t *testing.T) {
	var (
		err error
	)
	r := newRestPerf("Workload", "workload.yaml")

	counters := jsonToPerfRecords("testdata/partialAggregation/qos-counters.json")
	_, err = r.pollCounter(counters[0].Records.Array(), 0)
	if err != nil {
		t.Fatalf("Failed to fetch poll counter %v", err)
	}

	if len(r.Prop.Metrics) != len(r.perfProp.counterInfo) {
		t.Errorf("Prop metrics and counterInfo size should be same")
	}
}

func TestPartialAggregationSequence(t *testing.T) {
	var (
		err error
	)
	r := newRestPerf("Workload", "workload.yaml")

	counters := jsonToPerfRecords("testdata/partialAggregation/qos-counters.json")
	_, err = r.pollCounter(counters[0].Records.Array(), 0)
	if err != nil {
		t.Fatalf("Failed to fetch poll counter %v", err)
	}
	pollInstance := jsonToPerfRecords("testdata/partialAggregation/qos-poll-instance.json")
	_, err = r.pollInstance(r.Matrix[r.Object], perfToJSON(pollInstance), 0)
	if err != nil {
		t.Fatal(err)
	}

	// First Poll
	pollData := jsonToPerfRecords("testdata/partialAggregation/qos-poll-data-1.json")
	now := time.Now().Truncate(time.Second)
	pollData[0].Timestamp = now.UnixNano()
	t.Log("Running First Poll")
	r.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/qos-poll-data-1.json", 0, 0)

	// Complete Poll
	t.Log("Running Complete Poll")
	r.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/qos-poll-data-1.json", 2, 48)

	// Partial Poll
	t.Log("Running Partial Poll")
	r.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/qos-poll-data-2.json", 2, 0)

	// Partial Poll 2
	t.Log("Running Partial Poll 2")
	r.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/qos-poll-data-2.json", 2, 0)
	if t.Failed() {
		t.Fatal("Partial Poll 2 failed")
	}

	// First Complete Poll After Partial
	t.Log("Running First Complete Poll After Partial")
	r.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/qos-poll-data-1.json", 2, 0)
	if t.Failed() {
		t.Fatal("First Complete Poll After Partial failed")
	}

	// Second Complete Poll After Partial
	t.Log("Running Second Complete Poll After Partial")
	r.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/qos-poll-data-1.json", 2, 48)
	if t.Failed() {
		t.Fatal("Second Complete Poll After Partial failed")
	}

	// Partial Poll 3
	t.Log("Running Partial Poll 3")
	r.testPollInstanceAndDataWithMetrics(t, "testdata/partialAggregation/qos-poll-data-2.json", 2, 0)
	if t.Failed() {
		t.Fatal("Partial Poll 3 failed")
	}
}

func newRestPerf(object string, path string) *RestPerf {
	var err error
	opts := options.New(options.WithConfPath("testdata/conf"))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true

	ac := collector.New("RestPerf", object, opts, params(object, path), nil, conf.Remote{})
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
			sum:           18,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRestPerf("WorkloadVolume", "workload_volume.yaml")

			counters := jsonToPerfRecords(tt.pollCounters)
			_, err := r.pollCounter(counters[0].Records.Array(), 0)
			if err != nil {
				t.Fatal(err)
			}

			pollInst := jsonToPerfRecords(tt.pollInstance)
			_, err = r.pollInstance(r.Matrix[r.Object], perfToJSON(pollInst), 0)
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

			future := now.Add(time.Minute * 1)
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
