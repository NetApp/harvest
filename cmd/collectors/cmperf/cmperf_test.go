package cmperf

import (
	"errors"
	"strings"
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/errs"
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
