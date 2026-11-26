package rest

import (
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/pkg/slice"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
	_, _ = benchRest.pollData(fullPollData, set.New(), nil)

	os.Exit(m.Run())
}

func BenchmarkRestPerf_PollData(b *testing.B) {
	now := time.Now().Truncate(time.Second)

	for b.Loop() {
		now = now.Add(time.Minute * 15)
		_, _ = benchRest.pollData(fullPollData, set.New(), nil)
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
			numMetrics:    8748,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := newRest("Volume", "volume.yaml", "testdata/conf")
			pollData := collectors.JSONToGson(tt.pollDataPath1, true)

			mcount, parseD := r.pollData(pollData, set.New(), nil)
			mecount, apiD, endpointParseD := r.ProcessEndPoints(r.Matrix[r.Object], volumeEndpoints, set.New())

			metricCount := mcount + mecount
			r.postPollData(apiD, parseD+endpointParseD, metricCount, set.New())
			m := r.Matrix["Volume"]

			assert.Equal(t, tt.numInstances, len(m.GetInstances()))

			metadata := r.Metadata
			numMetrics, _ := metadata.GetMetric("metrics").GetValueInt64(metadata.GetInstance("data"))
			assert.Equal(t, tt.numMetrics, int(numMetrics))
		})
	}
}

func volumeEndpoints(e *EndPoint, processBatch func([]gjson.Result, int64) error) error {
	return pagedVolumeEndpoints(0)(e, processBatch)
}

// pagedVolumeEndpoints serves the volume endpoint testdata in pages of pageSize records, or as one page when pageSize is 0.
func pagedVolumeEndpoints(pageSize int) func(*EndPoint, func([]gjson.Result, int64) error) error {
	return func(e *EndPoint, processBatch func([]gjson.Result, int64) error) error {
		path := "testdata/" + strings.ReplaceAll(e.Prop.Query, "/", "-") + ".json.gz"
		records := collectors.JSONToGson(path, true)
		if len(records) == 0 {
			return errs.New(errs.ErrNoInstance, "no instances found")
		}
		size := pageSize
		if size == 0 {
			size = len(records)
		}
		for page := range slices.Chunk(records, size) {
			if err := processBatch(page, time.Now().Unix()); err != nil {
				return err
			}
		}
		return nil
	}
}

func TestProcessEndPointsPagedMatchesSinglePage(t *testing.T) {
	pollData := collectors.JSONToGson("testdata/volume-1.json.gz", true)

	single := newRest("Volume", "volume.yaml", "testdata/conf")
	_, _ = single.pollData(pollData, set.New(), nil)
	singleCount, _, _ := single.ProcessEndPoints(single.Matrix[single.Object], pagedVolumeEndpoints(0), set.New())

	paged := newRest("Volume", "volume.yaml", "testdata/conf")
	_, _ = paged.pollData(pollData, set.New(), nil)
	pagedCount, _, _ := paged.ProcessEndPoints(paged.Matrix[paged.Object], pagedVolumeEndpoints(7), set.New())

	assert.Equal(t, pagedCount, singleCount)

	want := single.Matrix[single.Object]
	got := paged.Matrix[paged.Object]
	assert.Equal(t, len(got.GetInstances()), len(want.GetInstances()))
	assert.Equal(t, len(got.GetMetrics()), len(want.GetMetrics()))
	for key, wantInstance := range want.GetInstances() {
		gotInstance := got.GetInstance(key)
		assert.NotNil(t, gotInstance)
		assert.Equal(t, cmp.Diff(gotInstance.GetLabels(), wantInstance.GetLabels()), "")
		for name, wantMetric := range want.GetMetrics() {
			wantValue, wantOk := wantMetric.GetValueFloat64(wantInstance)
			gotValue, gotOk := got.GetMetric(name).GetValueFloat64(gotInstance)
			assert.Equal(t, gotOk, wantOk)
			assert.Equal(t, gotValue, wantValue)
		}
	}
}

func TestProcessEndPointsCountsEveryEndpoint(t *testing.T) {
	pollData := collectors.JSONToGson("testdata/volume-1.json.gz", true)
	r := newRest("Volume", "volume.yaml", "testdata/conf")
	_, _ = r.pollData(pollData, set.New(), nil)
	assert.True(t, len(r.Endpoints) > 1)

	var perEndpoint uint64
	endpoints := r.Endpoints
	for _, e := range endpoints {
		r.Endpoints = []*EndPoint{e}
		c, _, _ := r.ProcessEndPoints(r.Matrix[r.Object], volumeEndpoints, set.New())
		perEndpoint += c
	}
	r.Endpoints = endpoints

	total, _, parseD := r.ProcessEndPoints(r.Matrix[r.Object], volumeEndpoints, set.New())
	assert.Equal(t, total, perEndpoint)
	assert.True(t, parseD > 0)
}

func TestProcessEndPointsErrors(t *testing.T) {
	pollData := collectors.JSONToGson("testdata/volume-1.json.gz", true)
	r := newRest("Volume", "volume.yaml", "testdata/conf")
	_, _ = r.pollData(pollData, set.New(), nil)
	assert.True(t, len(r.Endpoints) > 1)

	var logs strings.Builder
	r.Logger = slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))

	failing := r.Endpoints[0]
	empty := r.Endpoints[1]
	endpointFunc := func(e *EndPoint, processBatch func([]gjson.Result, int64) error) error {
		switch e {
		case failing:
			return errors.New("connection refused")
		case empty:
			return errs.New(errs.ErrNoInstance, "no instances found")
		}
		return volumeEndpoints(e, processBatch)
	}

	count, _, _ := r.ProcessEndPoints(r.Matrix[r.Object], endpointFunc, set.New())

	var others uint64
	for _, e := range r.Endpoints[2:] {
		others += newEndpointCount(t, pollData, e)
	}
	assert.Equal(t, count, others)
	assert.True(t, strings.Contains(logs.String(), `level=ERROR msg="failed to fetch endpoint"`))
	assert.True(t, strings.Contains(logs.String(), "connection refused"))
	assert.True(t, strings.Contains(logs.String(), `level=DEBUG msg="no instances on cluster"`))
}

// newEndpointCount returns the metric count of one endpoint on a fresh Volume collector.
func newEndpointCount(t *testing.T, pollData []gjson.Result, e *EndPoint) uint64 {
	t.Helper()
	r := newRest("Volume", "volume.yaml", "testdata/conf")
	_, _ = r.pollData(pollData, set.New(), nil)
	for _, re := range r.Endpoints {
		if re.Prop.Query == e.Prop.Query {
			r.Endpoints = []*EndPoint{re}
			count, _, _ := r.ProcessEndPoints(r.Matrix[r.Object], volumeEndpoints, set.New())
			return count
		}
	}
	t.Fatalf("endpoint %s not found", e.Prop.Query)
	return 0
}

func TestProcessEndPointsLogsDuplicatesAcrossPages(t *testing.T) {
	pollData := collectors.JSONToGson("testdata/volume-1.json.gz", true)
	r := newRest("Volume", "volume.yaml", "testdata/conf")
	_, _ = r.pollData(pollData, set.New(), nil)
	r.Endpoints = r.Endpoints[:1]

	var logs strings.Builder
	r.Logger = slog.New(slog.NewTextHandler(&logs, nil))

	endpointFunc := func(e *EndPoint, processBatch func([]gjson.Result, int64) error) error {
		path := "testdata/" + strings.ReplaceAll(e.Prop.Query, "/", "-") + ".json.gz"
		rec := collectors.JSONToGson(path, true)[0]
		for range 2 { // same record on two pages
			if err := processBatch([]gjson.Result{rec}, time.Now().Unix()); err != nil {
				return err
			}
		}
		return nil
	}

	_, _, _ = r.ProcessEndPoints(r.Matrix[r.Object], endpointFunc, set.New())
	assert.True(t, strings.Contains(logs.String(), "instKey is not unique"))
}

func TestPollDataLogsDuplicatesAcrossPages(t *testing.T) {
	pollData := collectors.JSONToGson("testdata/volume-1.json.gz", true)
	page := pollData[:1]

	t.Run("one seen set across pages finds the duplicate", func(t *testing.T) {
		r := newRest("Volume", "volume.yaml", "testdata/conf")
		var logs strings.Builder
		r.Logger = slog.New(slog.NewTextHandler(&logs, nil))

		seen := set.New()
		_, _ = r.pollData(page, set.New(), seen)
		_, _ = r.pollData(page, set.New(), seen)
		assert.True(t, strings.Contains(logs.String(), "instKey is not unique"))
	})

	t.Run("nil seen checks one page only", func(t *testing.T) {
		r := newRest("Volume", "volume.yaml", "testdata/conf")
		var logs strings.Builder
		r.Logger = slog.New(slog.NewTextHandler(&logs, nil))

		_, _ = r.pollData(page, set.New(), nil)
		_, _ = r.pollData(page, set.New(), nil)
		assert.False(t, strings.Contains(logs.String(), "instKey is not unique"))
	})
}

func TestPollDataLogsDuplicateAcrossServerPages(t *testing.T) {
	r := newRestWithHandler(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		record := `{"name":"vol1","svm":{"name":"svm1"},"state":"online"}`
		if req.URL.Query().Get("page") == "2" {
			_, _ = fmt.Fprintf(w, `{"records":[%s],"num_records":1}`, record)
			return
		}
		_, _ = fmt.Fprintf(w, `{"records":[%s],"num_records":1,"_links":{"next":{"href":"/api/storage/volumes?page=2"}}}`, record)
	})
	var logs strings.Builder
	r.Logger = slog.New(slog.NewTextHandler(&logs, nil))
	r.Prop.Href = "api/storage/volumes"
	r.Endpoints = nil

	_, err := r.PollData()

	assert.Nil(t, err)
	assert.Equal(t, len(r.Matrix[r.Object].GetInstances()), 1)
	assert.Equal(t, strings.Count(logs.String(), "instKey is not unique"), 1)
}

func TestProcessEndPointsEntryNotExistIsNoInstances(t *testing.T) {
	r := newRestWithHandler(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, `{"error":{"message":%q,"code":"%d"}}`, errs.EntryNotExist.Name, errs.EntryNotExist.Code)
	})
	var logs strings.Builder
	r.Logger = slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	r.Endpoints = []*EndPoint{{Prop: &prop{Query: "api/security/authentication/cluster/ad-proxy", Href: "api/security/authentication/cluster/ad-proxy"}}}

	count, _, _ := r.ProcessEndPoints(r.Matrix[r.Object], r.ProcessEndPoint, set.New())

	assert.Equal(t, count, uint64(0))
	assert.True(t, strings.Contains(logs.String(), `level=DEBUG msg="no instances on cluster"`))
	assert.False(t, strings.Contains(logs.String(), "failed to fetch endpoint"))
}

// newRestWithHandler returns a Volume collector whose client talks to a test server using handler.
func newRestWithHandler(t *testing.T, handler http.HandlerFunc) *Rest {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	r := newRest("Volume", "volume.yaml", "testdata/conf")
	insecure := true
	poller := &conf.Poller{
		Addr:           strings.TrimPrefix(server.URL, "https://"),
		Username:       "test-user",
		Password:       "test-password",
		UseInsecureTLS: &insecure,
	}
	client, err := rest.New(poller, 5*time.Second, auth.NewCredentials(poller, slog.New(slog.DiscardHandler)))
	assert.Nil(t, err)
	r.Client = client
	return r
}

func TestProcessEndPointStreamsPages(t *testing.T) {
	pages := map[string]string{
		"/api/storage/volumes": `{"records":[{"uuid":"a"},{"uuid":"b"}],"num_records":2,"_links":{"next":{"href":"/api/storage/volumes?page=2"}}}`,
		"page=2":               `{"records":[{"uuid":"c"}],"num_records":1}`,
	}
	r := newRestWithHandler(t, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if req.URL.RawQuery == "page=2" {
			_, _ = w.Write([]byte(pages["page=2"]))
			return
		}
		_, _ = w.Write([]byte(pages["/api/storage/volumes"]))
	})

	var got [][]string
	e := &EndPoint{Prop: &prop{Href: "api/storage/volumes"}}
	err := r.ProcessEndPoint(e, func(records []gjson.Result, _ int64) error {
		uuids := make([]string, 0, len(records))
		for _, rec := range records {
			uuids = append(uuids, rec.Get("uuid").String())
		}
		got = append(got, uuids)
		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, cmp.Diff(got, [][]string{{"a", "b"}, {"c"}}), "")
}

func TestProcessEndPointErrors(t *testing.T) {
	noBatch := func([]gjson.Result, int64) error {
		t.Fatal("processBatch must not be called")
		return nil
	}

	t.Run("empty href is a config error", func(t *testing.T) {
		r := newRest("Volume", "volume.yaml", "testdata/conf")
		err := r.ProcessEndPoint(&EndPoint{Prop: &prop{}}, noBatch)
		assert.ErrorIs(t, err, errs.ErrConfig)
	})

	t.Run("no records is ErrNoInstance", func(t *testing.T) {
		r := newRestWithHandler(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"records":[],"num_records":0}`))
		})
		err := r.ProcessEndPoint(&EndPoint{Prop: &prop{Href: "api/storage/volumes"}}, noBatch)
		assert.ErrorIs(t, err, errs.ErrNoInstance)
	})

	t.Run("entry doesn't exist means no records", func(t *testing.T) {
		r := newRestWithHandler(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintf(w, `{"error":{"message":%q,"code":"%d"}}`, errs.EntryNotExist.Name, errs.EntryNotExist.Code)
		})
		err := r.ProcessEndPoint(&EndPoint{Prop: &prop{Href: "api/storage/volumes"}}, noBatch)
		assert.ErrorIs(t, err, errs.ErrNoInstance)
	})

	t.Run("other REST errors are fetch failures", func(t *testing.T) {
		r := newRestWithHandler(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintf(w, `{"error":{"message":%q,"code":"%d"}}`, errs.APINotFound.Name, errs.APINotFound.Code)
		})
		err := r.ProcessEndPoint(&EndPoint{Prop: &prop{Href: "api/storage/volumes"}}, noBatch)
		assert.NotNil(t, err)
		assert.True(t, errs.IsRestErr(err, errs.APINotFound))
		assert.True(t, strings.Contains(err.Error(), "failed to fetch data"))
	})
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

// TestHandleErrorMissingAPIOrEntry covers GitHub issue #4559.
// Unlike TestHandleError, the input is a plain ONTAP REST error that is not already wrapped in ErrAPIRequestRejected.
func TestHandleErrorMissingAPIOrEntry(t *testing.T) {
	r := &Rest{}

	t.Run("entry doesn't exist means no records, not an error", func(t *testing.T) {
		// e.g., security/authentication/cluster/ad-proxy when no AD proxy is configured
		in := errs.NewRest().StatusCode(404).Message("entry doesn't exist").Code(errs.EntryNotExist.Code).Build()

		got, err := r.handleError(in)

		assert.Nil(t, got)
		assert.Nil(t, err)
	})

	t.Run("API not found stays an error the collector can classify", func(t *testing.T) {
		// e.g., cluster/web on ONTAP 9.9.1. The collector puts the task in standby when it finds APINotFound,
		// so handleError must return an error that still wraps the ONTAP code.
		in := errs.NewRest().StatusCode(404).Message("API not found").Code(errs.APINotFound.Code).Build()

		got, err := r.handleError(in)

		assert.Nil(t, got)
		assert.NotNil(t, err)
		assert.True(t, errs.IsRestErr(err, errs.APINotFound))
	})
}

// newRestWithServer returns a DomainTunnel collector whose client talks to a test server that answers every
// request with the given ONTAP REST error. It seeds one instance, as if an earlier poll had found an AD proxy.
func newRestWithServer(t *testing.T, ontapErr errs.OntapRestCode) *Rest {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, `{"error":{"message":%q,"code":"%d"}}`, ontapErr.Name, ontapErr.Code)
	}))
	t.Cleanup(server.Close)

	r := newRest("DomainTunnel", "domain_tunnel.yaml", "../../../conf")
	insecure := true
	poller := &conf.Poller{
		Addr:           strings.TrimPrefix(server.URL, "https://"),
		Username:       "test-user",
		Password:       "test-password",
		UseInsecureTLS: &insecure,
	}
	client, err := rest.New(poller, 5*time.Second, auth.NewCredentials(poller, slog.New(slog.DiscardHandler)))
	assert.Nil(t, err)
	r.Client = client

	_, err = r.Matrix[r.Object].NewInstance("svm1")
	assert.Nil(t, err)
	return r
}

// TestPollDataEntryNotExist covers GitHub issue #4559: when the AD proxy is removed, ONTAP answers
// "entry doesn't exist". The poll succeeds and drops the old instance so its labels are no longer exported.
func TestPollDataEntryNotExist(t *testing.T) {
	r := newRestWithServer(t, errs.EntryNotExist)

	data, err := r.PollData()

	assert.Nil(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, len(r.Matrix[r.Object].GetInstances()), 0)
	instances, _ := r.Metadata.GetMetric("instances").GetValueUint64(r.Metadata.GetInstance("data"))
	assert.Equal(t, instances, uint64(0))
}

// Contrast with TestPollDataEntryNotExist: any other REST error fails the poll and keeps the old instances.
func TestPollDataAPINotFoundKeepsInstances(t *testing.T) {
	r := newRestWithServer(t, errs.APINotFound)

	data, err := r.PollData()

	assert.Nil(t, data)
	assert.True(t, errs.IsRestErr(err, errs.APINotFound))
	assert.Equal(t, len(r.Matrix[r.Object].GetInstances()), 1)
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
