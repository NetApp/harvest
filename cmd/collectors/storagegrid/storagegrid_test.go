package storagegrid

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const (
	pollerName = "test"
)

// newStorageGrid initializes a new StorageGrid instance for testing
func newStorageGrid() (*StorageGrid, error) {
	const (
		object = "Tenant"
		path   = "tenant.yaml"
	)
	opts := options.New(options.WithConfPath("testdata/conf"))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true
	r := StorageGrid{}
	rest.NewClientFunc = func(_ string, _ string, _ *auth.Credentials) (*rest.Client, error) {
		return rest.NewDummyClient(), nil
	}
	ac := collector.New("StorageGrid", object, opts, collectors.Params(object, path), nil, conf.Remote{})
	err := r.Init(ac)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func Test_SGAddRemoveRestInstances(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	sg, err := newStorageGrid()
	assert.Nil(t, err)

	testFile(t, sg, "testdata/tenant.json", 9)
	testFile(t, sg, "testdata/tenant_delete.json", 8)
	testFile(t, sg, "testdata/tenant_add.json", 10)
}

func testFile(t *testing.T, sg *StorageGrid, filename string, expectedLen int) {
	output, err := os.ReadFile(filename)
	assert.Nil(t, err)

	data := gjson.Get(string(output), "data")
	_ = sg.handleResults(data.Array())

	assert.Equal(t, len(sg.Matrix[sg.Object].GetInstances()), expectedLen)
}

// emptyMetricQueryServer serves an empty JSON object for the Prometheus proxy
// endpoint, which is what a grid returns for a metric that has no data on it --
// a brand new cluster, a grid with no tenants, or a metric that does not exist
// in that StorageGrid version.
func emptyMetricQueryServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestGetMetricWithNoDataReturnsEmptyMatrix is the regression test for
// GetMetric returning (nil, nil).
//
// Both callers check only err != nil and then dereference the matrix:
// pollPrometheusMetrics calls len(mat.GetInstances()) and the Tenant plugin sets
// mat.Object. matrix.GetInstances() has no nil-receiver guard, so a metric with
// no data on the grid panicked the collector. makePromMetrics already returns an
// empty matrix for its own no-results case, so GetMetric must do the same.
func TestGetMetricWithNoDataReturnsEmptyMatrix(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	sg, err := newStorageGrid()
	assert.Nil(t, err)

	srv := emptyMetricQueryServer(t)
	sg.client = rest.NewDummyClientWithBaseURL(srv.URL)

	mat, err := sg.GetMetric("storagegrid_tenant_usage_data_bytes", "usage_data_bytes", nil)

	assert.Nil(t, err)
	// Pre-fix this was nil, and the next line is exactly what the callers do.
	assert.NotNil(t, mat)
	assert.Equal(t, len(mat.GetInstances()), 0)

	// The empty matrix must be shaped like the one makePromMetrics builds, so
	// downstream exporters and plugins treat it identically.
	assert.Equal(t, mat.Object, sg.Props.Object)
	assert.True(t, strings.HasSuffix(mat.UUID, ".usage_data_bytes"))
}

// TestGetMetricWithNoDataCallersDoNotPanic reproduces the two call sites'
// dereferences directly, since those are where the panic surfaced.
func TestGetMetricWithNoDataCallersDoNotPanic(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	sg, err := newStorageGrid()
	assert.Nil(t, err)

	srv := emptyMetricQueryServer(t)
	sg.client = rest.NewDummyClientWithBaseURL(srv.URL)

	mat, err := sg.GetMetric("storagegrid_tenant_usage_quota_bytes", "usage_quota_bytes", nil)
	assert.Nil(t, err)

	// pollPrometheusMetrics does this (storagegrid.go).
	numInstances := len(mat.GetInstances())
	assert.Equal(t, numInstances, 0)

	// The Tenant plugin does this (tenant.go).
	mat.Object = "storagegrid"
	assert.Equal(t, mat.Object, "storagegrid")
}

// TestGetMetricUsesDisplayNameForUUID pins that the display name, not the raw
// metric name, is what suffixes the UUID -- the same rule makePromMetrics
// applies -- so the empty and non-empty paths stay consistent.
func TestGetMetricUsesDisplayNameForUUID(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/config.yml")

	sg, err := newStorageGrid()
	assert.Nil(t, err)

	srv := emptyMetricQueryServer(t)
	sg.client = rest.NewDummyClientWithBaseURL(srv.URL)

	// Empty display falls back to the metric name.
	mat, err := sg.GetMetric("storagegrid_private_load_balancer_storage_request_count", "", nil)
	assert.Nil(t, err)
	assert.NotNil(t, mat)
	assert.True(t, strings.HasSuffix(mat.UUID, ".storagegrid_private_load_balancer_storage_request_count"))
}
