package loginbanner

import (
	"net/http"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

func TestBannerStateFor(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        []byte
		wantPresent bool
		wantOK      bool
	}{
		{
			name:        "200 with banner text",
			status:      http.StatusOK,
			body:        []byte("Authorized users only."),
			wantPresent: true,
			wantOK:      true,
		},
		{
			name:   "200 with whitespace-only body",
			status: http.StatusOK,
			body:   []byte("   \n\t"),
			wantOK: true,
		},
		{
			name:   "200 with empty body",
			status: http.StatusOK,
			body:   []byte(""),
			wantOK: true,
		},
		{
			name:   "204 no content",
			status: http.StatusNoContent,
			body:   nil,
			wantOK: true,
		},
		{
			name:   "422 unprocessable (documented error response)",
			status: http.StatusUnprocessableEntity,
			body:   []byte(`{"errorMessage":"unable to return the login banner"}`),
			wantOK: false,
		},
		{
			name:   "404 not found (outside documented contract)",
			status: http.StatusNotFound,
			body:   nil,
			wantOK: false,
		},
		{
			name:   "500 internal server error (outside documented contract)",
			status: http.StatusInternalServerError,
			body:   nil,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			present, ok := bannerStateFor(tt.status, tt.body)
			assert.Equal(t, ok, tt.wantOK)
			if tt.wantOK {
				assert.Equal(t, present, tt.wantPresent)
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		want string
	}{
		{
			name: "CallResponse with errorMessage",
			body: []byte(`{"errorMessage":"Unable to return the login banner.","retcode":"unknown","codeType":"devicemgrerror","developerMessage":"","invalidFieldsIfKnown":[]}`),
			want: "Unable to return the login banner.",
		},
		{
			name: "empty body",
			body: nil,
			want: "",
		},
		{
			name: "non-JSON body",
			body: []byte("plain text error"),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, errorMessage(tt.body), tt.want)
		})
	}
}

func newTestPlugin(t *testing.T) *LoginBanner {
	t.Helper()
	mat := matrix.New("test.LoginBanner", loginBannerMatrix, loginBannerMatrix)
	mat.SetExportOptions(matrix.NewExportOptions("wwn"))
	if _, err := mat.NewMetricFloat64(metricName); err != nil {
		t.Fatal(err)
	}
	return &LoginBanner{data: mat}
}

// TestExportOptionsHaveNoInstanceLabels guards against reintroducing an
// instance_labels export option that duplicates the wwn instance key: the
// Prometheus exporter emits a separate eseries_array_login_banner_labels
// pseudo-metric for any instance_labels entry, which would be redundant here
// since wwn is already attached to eseries_array_login_banner_enabled.
func TestExportOptionsHaveNoInstanceLabels(t *testing.T) {
	l := newTestPlugin(t)
	options := l.data.GetExportOptions()
	assert.Nil(t, options.GetChildS("instance_labels"))

	keys := options.GetChildS("instance_keys")
	if keys == nil {
		t.Fatal("expected instance_keys")
	}
	assert.Equal(t, keys.GetAllChildContentS(), []string{"wwn"})
}

func TestApplyState(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		l := newTestPlugin(t)
		l.data.PurgeInstances()
		l.data.Reset()
		assert.Nil(t, l.applyState("wwn1", true))
		inst := l.data.GetInstance("wwn1")
		if inst == nil {
			t.Fatal("expected instance")
		}
		value, ok := l.data.GetMetric(metricName).GetValueFloat64(inst)
		assert.True(t, ok)
		assert.Equal(t, value, 1.0)
		assert.Equal(t, inst.GetLabel("wwn"), "wwn1")
	})

	t.Run("absent", func(t *testing.T) {
		l := newTestPlugin(t)
		l.data.PurgeInstances()
		l.data.Reset()
		assert.Nil(t, l.applyState("wwn1", false))
		inst := l.data.GetInstance("wwn1")
		if inst == nil {
			t.Fatal("expected instance")
		}
		value, ok := l.data.GetMetric(metricName).GetValueFloat64(inst)
		assert.True(t, ok)
		assert.Equal(t, value, 0.0)
		assert.Equal(t, inst.GetLabel("wwn"), "wwn1")
	})

	t.Run("absent after purge does not carry over a stale instance", func(t *testing.T) {
		l := newTestPlugin(t)
		l.data.PurgeInstances()
		l.data.Reset()
		assert.Nil(t, l.applyState("wwn1", true))
		assert.Equal(t, len(l.data.GetInstances()), 1)

		l.data.PurgeInstances()
		l.data.Reset()
		assert.Nil(t, l.applyState("wwn1", false))
		assert.Equal(t, len(l.data.GetInstances()), 1)
		inst := l.data.GetInstance("wwn1")
		if inst == nil {
			t.Fatal("expected instance")
		}
		value, ok := l.data.GetMetric(metricName).GetValueFloat64(inst)
		assert.True(t, ok)
		assert.Equal(t, value, 0.0)
	})
}
