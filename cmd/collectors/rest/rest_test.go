package rest

import (
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slice"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	pollerName = "test"
)

var (
	benchRest    *Rest
	fullPollData []gjson.Result
)

func TestMain(m *testing.M) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	benchRest = newRest("Volume", "volume.yaml", "testdata/conf")
	fullPollData = collectors.JSONToGson("testdata/volume-1.json.gz", true)
	_, _ = benchRest.pollData(fullPollData, set.New())

	os.Exit(m.Run())
}

func BenchmarkRestPerf_PollData(b *testing.B) {
	now := time.Now().Truncate(time.Second)

	for b.Loop() {
		now = now.Add(time.Minute * 15)
		_, _ = benchRest.pollData(fullPollData, set.New())
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
			pollData := collectors.JSONToGson(tt.pollDataPath1, true)

			mcount, parseD := r.pollData(pollData, set.New())
			mecount, apiD := r.ProcessEndPoints(r.Matrix[r.Object], volumeEndpoints, set.New())

			metricCount := mcount + mecount
			r.postPollData(apiD, parseD, metricCount, set.New())
			m := r.Matrix["Volume"]

			assert.Equal(t, tt.numInstances, len(m.GetInstances()))

			metadata := r.Metadata
			numMetrics, _ := metadata.GetMetric("metrics").GetValueInt(metadata.GetInstance("data"))
			assert.Equal(t, tt.numMetrics, numMetrics)
		})
	}
}

func volumeEndpoints(e *EndPoint) ([]gjson.Result, time.Duration, error) {
	path := "testdata/" + strings.ReplaceAll(e.Prop.Query, "/", "-") + ".json.gz"
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
			assert.Equal(t, tt.expectedResult, result)
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
			assert.Equal(t, diff, "")
		})
	}
}

func TestQuotas(t *testing.T) {
	r := newRest("Quota", "quota.yaml", "../../../conf")
	var instanceKeys []string
	result, err := collectors.InvokeRestCallWithTestFile(r.Client, "", "testdata/quota.json")
	assert.Nil(t, err)

	for _, quotaInstanceData := range result {
		var instanceKey string
		if len(r.Prop.InstanceKeys) != 0 {
			// extract instance key(s)
			for _, k := range r.Prop.InstanceKeys {
				value := quotaInstanceData.Get(k)
				if value.Exists() {
					instanceKey += value.ClonedString()
				}
			}

			if instanceKey == "" {
				continue
			}
			instanceKeys = append(instanceKeys, instanceKey)
		}
	}

	assert.False(t, slice.HasDuplicates(instanceKeys))
}
