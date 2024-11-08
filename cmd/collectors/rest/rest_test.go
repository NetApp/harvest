package rest

import (
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"os"
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
)

func TestMain(m *testing.M) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	benchRest = newRest("Volume", "volume.yaml", "testdata/conf")
	fullPollData = collectors.JSONToGson("testdata/volume-1.json.gz", true)
	now := time.Now().Truncate(time.Second)
	_, _ = benchRest.pollData(now, fullPollData, volumeEndpoints)

	os.Exit(m.Run())
}

func BenchmarkRestPerf_PollData(b *testing.B) {
	var err error
	ms = make([]*matrix.Matrix, 0)
	now := time.Now().Truncate(time.Second)

	for range b.N {
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

			r := newRest("Volume", "volume.yaml", "testdata/conf")
			now := time.Now().Truncate(time.Second)
			pollData := collectors.JSONToGson(tt.pollDataPath1, true)

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

func volumeEndpoints(e *EndPoint) ([]gjson.Result, time.Duration, error) {
	path := "testdata/" + strings.ReplaceAll(e.prop.Query, "/", "-") + ".json.gz"
	gson := collectors.JSONToGson(path, true)
	return gson, 0, nil
}

func newRest(object string, path string, confPath string) *Rest {
	var err error
	opts := options.New(options.WithConfPath(confPath))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true
	ac := collector.New("Rest", object, opts, collectors.Params(object, path), nil, conf.Remote{})
	r := Rest{}
	err = r.Init(ac)
	if err != nil {
		panic(err)
	}
	return &r
}

func TestIsValidFormat(t *testing.T) {

	tests := []struct {
		name           string
		r              *Rest
		p              *prop
		expectedResult bool
	}{
		{
			name: "Test with valid fields 1",
			r:    &Rest{},
			p: &prop{
				Fields: []string{
					"uuid",
					"block_storage.primary.disk_type",
					"block_storage.primary.raid_type",
				},
				IsPublic: true,
			},
			expectedResult: true,
		},

		{
			name: "Test with invalid fields 2",
			r:    &Rest{},
			p: &prop{
				Fields: []string{
					"uuid",
					"cloud_storage.stores.#.cloud_store.name",
					"block_storage.primary.raid_type",
				},
				IsPublic: true,
			},
			expectedResult: false,
		},
		{
			name: "Test with invalid fields 3",
			r:    &Rest{},
			p: &prop{
				Fields: []string{
					"uuid",
					"cloud_storage.stores.0.cloud_store.name",
					"block_storage.primary.raid_type",
				},
				IsPublic: true,
			},
			expectedResult: false,
		},
		{
			name: "Test with invalid fields 4",
			r:    &Rest{},
			p: &prop{
				Fields: []string{
					"uuid",
					"{interfaces.#.name,interfaces.#.ip.address}",
				},
				IsPublic: true,
			},
			expectedResult: false,
		},
		{
			name: "Test with invalid fields 5",
			r:    &Rest{},
			p: &prop{
				Fields: []string{
					"uuid",
					"friends.#(last==\"Murphy\")#.first",
				},
				IsPublic: true,
			},
			expectedResult: false,
		},
		{
			name: "Test with invalid fields 6",
			r:    &Rest{},
			p: &prop{
				Fields: []string{
					"uuid",
					"children|@case:upper",
				},
				IsPublic: true,
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r.isValidFormat(tt.p)
			if result != tt.expectedResult {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		name           string
		r              *Rest
		p              *prop
		expectedResult []string
	}{
		{
			name: "Test with valid fields",
			r: &Rest{
				isIgnoreUnknownFieldsEnabled: true,
			},
			p: &prop{
				Fields: []string{
					"uuid",
					"block_storage.primary.disk_type",
					"block_storage.primary.raid_type",
				},
				IsPublic: true,
			},
			expectedResult: []string{
				"uuid",
				"block_storage.primary.disk_type",
				"block_storage.primary.raid_type",
			},
		},
		{
			name: "Test with invalid fields",
			r: &Rest{
				isIgnoreUnknownFieldsEnabled: true,
			},
			p: &prop{
				Fields: []string{
					"uuid",
					"cloud_storage.stores.0.cloud_store.name",
					"block_storage.primary.raid_type",
				},
				IsPublic: true,
			},
			expectedResult: []string{"*"},
		},
		{
			name: "Test with valid fields and prior versions to 9.11.1",
			r: &Rest{
				isIgnoreUnknownFieldsEnabled: false,
			},
			p: &prop{
				Fields: []string{
					"uuid",
					"block_storage.primary.disk_type",
					"block_storage.primary.raid_type",
				},
				IsPublic: true,
			},
			expectedResult: []string{
				"*",
			},
		},
		{
			name: "Test with valid fields for private API",
			r: &Rest{
				isIgnoreUnknownFieldsEnabled: false,
			},
			p: &prop{
				Fields: []string{
					"uuid",
					"block_storage.primary.disk_type",
					"block_storage.primary.raid_type",
				},
				IsPublic: false,
			},
			expectedResult: []string{
				"uuid",
				"block_storage.primary.disk_type",
				"block_storage.primary.raid_type",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.r.Fields(tt.p)
			diff := cmp.Diff(result, tt.expectedResult)
			if diff != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestQuotas(t *testing.T) {
	r := newRest("Quota", "quota.yaml", "../../../conf")
	var instanceKeys []string
	result, err := collectors.InvokeRestCallWithTestFile(r.Client, "", "testdata/quota.json")
	if err != nil {
		t.Errorf("Error while invoking quota rest api call")
	}

	for _, quotaInstanceData := range result {
		var instanceKey string
		if len(r.Prop.InstanceKeys) != 0 {
			// extract instance key(s)
			for _, k := range r.Prop.InstanceKeys {
				value := quotaInstanceData.Get(k)
				if value.Exists() {
					instanceKey += value.String()
				}
			}

			if instanceKey == "" {
				continue
			}
			instanceKeys = append(instanceKeys, instanceKey)
		}
	}

	if util.HasDuplicates(instanceKeys) {
		t.Errorf("Duplicate instanceKeys found for quota rest api")
	}
}
