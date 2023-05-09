package rest

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	pollerName = "test"
)

func Test_getFieldName(t *testing.T) {

	type test struct {
		name   string
		source string
		parent string
		want   int
	}

	var tests = []test{
		{
			name:   "Test1",
			source: readFile("testdata/cluster.json"),
			parent: "",
			want:   52,
		},
		{
			name:   "Test2",
			source: readFile("testdata/sample.json"),
			parent: "",
			want:   3,
		},
		{
			name:   "Test3",
			source: readFile("testdata/test.json"),
			parent: "",
			want:   9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFieldName(tt.source, tt.parent); len(got) != tt.want {
				t.Errorf("length of output slice = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func readFile(path string) string {
	b, _ := os.ReadFile(path)
	return string(b)
}

var (
	ms           []*matrix.Matrix
	benchRest    *Rest
	fullPollData []gjson.Result
	gsonCache    = make(map[string][]gjson.Result)
)

func TestMain(m *testing.M) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	benchRest = newRest("Volume", "volume.yaml")
	fullPollData = jsonToGson("testdata/volume-1.json.gz")
	now := time.Now().Truncate(time.Second)
	_, _ = benchRest.pollData(now, fullPollData, volumeEndpoints)

	os.Exit(m.Run())
}

func BenchmarkRestPerf_PollData(b *testing.B) {
	var err error
	ms = make([]*matrix.Matrix, 0)
	now := time.Now().Truncate(time.Second)

	for i := 0; i < b.N; i++ {
		now = now.Add(time.Minute * 15)
		mi, _ := benchRest.pollData(now, fullPollData, volumeEndpoints)

		for _, mm := range mi {
			ms = append(ms, mm)
		}
		mi, err = benchRest.pollData(now, fullPollData, volumeEndpoints)
		if err != nil {
			b.Errorf("error: %v", err)
		}
		for _, mm := range mi {
			ms = append(ms, mm)
		}
	}
}

func Test_pollDataVolume(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")
	tests := []struct {
		name          string
		wantErr       bool
		pollDataPath1 string
		numInstances  int
		numMetrics    int
	}{
		{
			name:          "sar",
			pollDataPath1: "testdata/volume-1.json.gz",
			numInstances:  185,
			numMetrics:    6916,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := newRest("Volume", "volume.yaml")
			now := time.Now().Truncate(time.Second)
			pollData := jsonToGson(tt.pollDataPath1)

			mm, err := r.pollData(now, pollData, volumeEndpoints)
			if err != nil {
				t.Fatal(err)
			}
			m := mm["Volume"]

			if len(m.GetInstances()) != tt.numInstances {
				t.Errorf("pollData() numInstances got=%v, want=%v", len(m.GetInstances()), tt.numInstances)
			}

			metadata := r.Metadata
			numMetrics, _ := metadata.GetMetric("metrics").GetValueInt(metadata.GetInstance("data"))
			if numMetrics != tt.numMetrics {
				t.Errorf("pollData() numMetrics got=%v, want=%v", numMetrics, tt.numMetrics)
			}
		})
	}
}

func volumeEndpoints(e *endPoint) ([]gjson.Result, error) {
	path := "testdata/" + strings.ReplaceAll(e.prop.Query, "/", "-") + ".json.gz"
	toGjson := jsonToGson(path)
	return toGjson, nil
}

func jsonToGson(path string) []gjson.Result {

	var (
		result []gjson.Result
		err    error
	)
	results, ok := gsonCache[path]
	if ok {
		return results
	}

	var reader io.Reader
	if filepath.Ext(path) == ".gz" {
		open, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer open.Close()
		reader, err = gzip.NewReader(open)
		if err != nil {
			panic(err)
		}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		reader = bytes.NewReader(data)
	}
	var b bytes.Buffer
	_, err = io.Copy(&b, reader) //nolint:gosec
	if err != nil {
		return nil
	}
	bb := b.Bytes()
	output := gjson.GetManyBytes(bb, "records", "num_records", "_links.next.href")

	data := output[0]
	numRecords := output[1]
	isNonIterRestCall := !data.Exists()

	if isNonIterRestCall {
		contentJSON := `{"records":[]}`
		response, err := sjson.SetRawBytes([]byte(contentJSON), "records.-1", bb)
		if err != nil {
			panic(err)
		}
		value := gjson.GetBytes(response, "records")
		result = append(result, value)
	} else if numRecords.Int() > 0 {
		result = append(result, data.Array()...)
	}

	gsonCache[path] = result
	return result
}

func newRest(object string, path string) *Rest {
	var err error
	opts := options.Options{
		Poller:   pollerName,
		HomePath: "testdata",
		IsTest:   true,
	}
	ac := collector.New("Rest", object, &opts, params(object, path), nil)
	ac.IsTest = true
	r := Rest{}
	err = r.Init(ac)
	if err != nil {
		panic(err)
	}
	return &r
}

func params(object string, path string) *node.Node {
	yml := `
schedule:
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
