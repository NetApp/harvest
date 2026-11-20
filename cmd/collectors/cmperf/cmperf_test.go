package cmperf

import (
	"errors"
	"strings"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func TestCanonicalSamplePeriod(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "1m", raw: "1m", want: "1m"},
		{name: "60s normalizes to 1m", raw: "60s", want: "1m"},
		{name: "1m0s normalizes to 1m", raw: "1m0s", want: "1m"},
		{name: "5m", raw: "5m", want: "5m"},
		{name: "300s normalizes to 5m", raw: "300s", want: "5m"},
		{name: "10m", raw: "10m", want: "10m"},
		{name: "30m", raw: "30m", want: "30m"},
		{name: "1h", raw: "1h", want: "1h"},
		{name: "60m normalizes to 1h", raw: "60m", want: "1h"},
		{name: "empty", raw: "", wantErr: errs.ErrMissingParam},
		{name: "unsupported 3m", raw: "3m", wantErr: errs.ErrInvalidParam},
		{name: "unsupported 90s", raw: "90s", wantErr: errs.ErrInvalidParam},
		{name: "unsupported 2h", raw: "2h", wantErr: errs.ErrInvalidParam},
		{name: "unparseable", raw: "abc", wantErr: errs.ErrInvalidParam},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CanonicalSamplePeriod(tt.raw)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				if errors.Is(tt.wantErr, errs.ErrInvalidParam) {
					assert.True(t, strings.Contains(err.Error(), allowedSamplePeriodList))
				}
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestIsWorkloadObject(t *testing.T) {
	tests := []struct {
		query string
		want  bool
	}{
		{query: "workload", want: true},
		{query: "workload_volume", want: true},
		{query: "workload_queue_combined", want: true},
		{query: "volume", want: false},
		{query: "wafl", want: false},
		{query: "workload_queue_combined:constituent", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			assert.Equal(t, isWorkloadObject(tt.query), tt.want)
		})
	}
}

func TestWorkloadFilterFor(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		classes []string // workload_class template param; nil means unset
		want    []string
	}{
		{name: "workload default", query: "workload",
			want: []string{"workload_class=user_defined|system_defined|preset"}},
		{name: "workload_volume default", query: "workload_volume",
			want: []string{"workload_class=autovolume"}},
		{name: "workload_queue_combined default", query: "workload_queue_combined",
			want: []string{"qtree=null", "lun=null", "file=null"}},
		{name: "unknown default", query: "unknown",
			want: []string{"workload_class=user_defined|system_defined|preset"}},
		{name: "workload param", query: "workload", classes: []string{"user_defined", "preset"},
			want: []string{"workload_class=user_defined|preset"}},
		{name: "workload_volume param", query: "workload_volume", classes: []string{"user_defined"},
			want: []string{"workload_class=user_defined"}},
		{name: "workload_queue_combined param", query: "workload_queue_combined", classes: []string{"autovolume"},
			want: []string{"qtree=null", "lun=null", "file=null", "workload_class=autovolume"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := node.NewS("params")
			if tt.classes != nil {
				wc := params.NewChildS("workload_class", "")
				for _, c := range tt.classes {
					wc.NewChildS("", c)
				}
			}
			assert.Equal(t, workloadFilterFor(tt.query, params), tt.want)
		})
	}
}
