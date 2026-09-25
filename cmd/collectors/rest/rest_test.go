package rest

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slice"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"slices"
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
			numMetrics, _ := metadata.GetMetric("metrics").GetValueInt64(metadata.GetInstance("data"))
			assert.Equal(t, tt.numMetrics, int(numMetrics))
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
	r := &Rest{AbstractCollector: ac}
	err = r.Init(ac)
	if err != nil {
		panic(err)
	}
	return r
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
			result := tt.r.isValidFormat(tt.p.Fields)
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
			name: "Test with array index fields",
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
			expectedResult: []string{
				"uuid",
				"cloud_storage.stores.cloud_store.name",
				"block_storage.primary.raid_type",
			},
		},
		{
			name: "Test with array fields",
			r: &Rest{
				isIgnoreUnknownFieldsEnabled: true,
			},
			p: &prop{
				Fields: []string{
					"uuid",
					"cloud_storage.stores.#.cloud_store.name",
					"block_storage.primary.raid_type",
				},
				IsPublic: true,
			},
			expectedResult: []string{
				"uuid",
				"cloud_storage.stores.cloud_store.name",
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
					"friends.#(last==\"Murphy\")#.first",
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

func TestRequestFields(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want   []string
	}{
		{name: "no index", fields: []string{"uuid", "version.full"}, want: []string{"uuid", "version.full"}},
		{name: "one index", fields: []string{"ha.partners.0.name"}, want: []string{"ha.partners.name"}},
		{name: "nested indexes", fields: []string{"a.0.b.12.c"}, want: []string{"a.b.c"}},
		{name: "duplicates removed", fields: []string{"users.0.name", "users.1.name"}, want: []string{"users.name"}},
		{name: "digits inside a name are kept", fields: []string{"counter_v2.value"}, want: []string{"counter_v2.value"}},
		{name: "only an index is left alone", fields: []string{"0"}, want: []string{"0"}},
		{name: "array", fields: []string{"volumes.#.name"}, want: []string{"volumes.name"}},
		{name: "array count", fields: []string{"block_storage.plexes.#"}, want: []string{"block_storage.plexes"}},
		{name: "array in array", fields: []string{"origins.#.svm.name"}, want: []string{"origins.svm.name"}},
		{
			name:   "multipath",
			fields: []string{"{interfaces.#.name,interfaces.#.ip.address}"},
			want:   []string{"interfaces.name", "interfaces.ip.address"},
		},
		{name: "only an array is left alone", fields: []string{"#"}, want: []string{"#"}},
		{name: "query is left alone", fields: []string{"friends.#(last==\"Murphy\")#.first"}, want: []string{"friends.#(last==\"Murphy\")#.first"}},
		{name: "modifier is left alone", fields: []string{"children|@case:upper"}, want: []string{"children|@case:upper"}},
		{name: "unclosed multipath stays invalid", fields: []string{"{interfaces.#.name"}, want: []string{"{interfaces.name"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, cmp.Diff(requestFields(tt.fields), tt.want), "")
		})
	}
}

// Templates with array counters, e.g. ha.partners.0.name or volumes.#.name, used to request fields=*.
// On large AFX clusters fields=* on api/cluster/nodes includes controller.bezel, which is slow enough to time out.
func TestTemplatesWithArraysDoNotRequestAllFields(t *testing.T) {
	tests := []struct {
		object string
		path   string
		want   []string
	}{
		{object: "Node", path: "node.yaml", want: []string{"ha.partners.name"}},
		{object: "Quota", path: "quota.yaml", want: []string{"users.id", "users.name"}},
		{object: "CIFSSession", path: "cifs_session.yaml", want: []string{"volumes.name"}},
		{object: "EmsDestination", path: "ems_destination.yaml", want: []string{"filters.name"}},
		{
			object: "FlexCache",
			path:   "flexcache.yaml",
			want:   []string{"aggregates.name", "origins.cluster.name", "origins.svm.name", "origins.volume.name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.object, func(t *testing.T) {
			r := newRest(tt.object, tt.path, "../../../conf")
			r.isIgnoreUnknownFieldsEnabled = true
			fields := r.Fields(r.Prop)
			assert.False(t, slices.Contains(fields, "*"))
			for _, w := range tt.want {
				assert.True(t, slices.Contains(fields, w))
			}
			for _, f := range fields {
				assert.False(t, strings.Contains(f, ".0."))
				assert.False(t, strings.Contains(f, "#"))
			}
		})
	}
}

func TestQuotas(t *testing.T) {
	r := newRest("Quota", "quota.yaml", "../../../conf")
	var instanceKeys []string
	result, err := collectors.InvokeRestCallWithTestFile(r.Client, "", "testdata/quota.json")
	assert.Nil(t, err)

	for _, quotaInstanceData := range result {
		var instanceKey strings.Builder
		if len(r.Prop.InstanceKeys) != 0 {
			// extract instance key(s)
			for _, k := range r.Prop.InstanceKeys {
				value := quotaInstanceData.Get(k)
				if value.Exists() {
					instanceKey.WriteString(value.ClonedString())
				}
			}

			instKey := instanceKey.String()
			if instKey == "" {
				continue
			}
			instanceKeys = append(instanceKeys, instKey)
		}
	}

	assert.False(t, slice.HasDuplicates(instanceKeys))
}

// TestHandleError covers the classification step behind GitHub issue #4197
// "MetroclusterCheck collector frequent failed status", which shipped with the
// status/untested label.
//
// handleError decides whether a REST failure is a transient, expected condition
// that the poller should stand by on, or a genuine fetch failure that drives the
// collector to failed status. Before the fix, code 2428841 had no branch here
// and fell through to the generic "failed to fetch data".
func TestHandleError(t *testing.T) {
	tests := []struct {
		name           string
		code           int64
		wantRejected   bool
		wantSentinel   error
		wantGenericMsg bool
	}{
		{
			name:         "MetroCluster check in progress is rejected, not a fetch failure",
			code:         2428841,
			wantRejected: true,
			wantSentinel: errs.ErrMetroClusterCheckInProgress,
		},
		{
			name:         "MetroCluster not configured is rejected, not a fetch failure",
			code:         2426405,
			wantRejected: true,
			wantSentinel: errs.ErrMetroClusterNotConfigured,
		},
		{
			name:           "any other code is a generic fetch failure",
			code:           8585320,
			wantRejected:   false,
			wantGenericMsg: true,
		},
	}

	r := &Rest{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := errs.NewRest().Error(errs.ErrAPIRequestRejected).Code(tc.code).Build()

			got, err := r.handleError(in)

			assert.Nil(t, got)
			assert.NotNil(t, err)

			if tc.wantRejected {
				// ErrAPIRequestRejected is what the collector keys standby off.
				assert.ErrorIs(t, err, errs.ErrAPIRequestRejected)
				assert.ErrorIs(t, err, tc.wantSentinel)
				assert.False(t, strings.Contains(err.Error(), "failed to fetch data"))
			}
			if tc.wantGenericMsg {
				assert.True(t, strings.Contains(err.Error(), "failed to fetch data"))
				assert.False(t, errors.Is(err, errs.ErrMetroClusterCheckInProgress))
				assert.False(t, errors.Is(err, errs.ErrMetroClusterNotConfigured))
			}
		})
	}
}

// TestHandleErrorPassesThroughPlainError pins that a transport-level error,
// which carries no ONTAP code at all, is still reported as a fetch failure
// rather than being silently classified as an expected condition.
func TestHandleErrorPassesThroughPlainError(t *testing.T) {
	r := &Rest{}

	got, err := r.handleError(errors.New("connection refused"))

	assert.Nil(t, got)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to fetch data"))
	assert.True(t, strings.Contains(err.Error(), "connection refused"))
}
