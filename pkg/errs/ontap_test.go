package errs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/netapp/harvest/v2/assert"
)

func TestErrAuthFailed(t *testing.T) {
	err := NewRest().Error(ErrAuthFailed).Build()
	assert.ErrorIs(t, err, ErrAuthFailed)
}

// TestIsRestErr covers the REST error-code classification that decides whether
// a collector reports a transient, expected condition or goes to failed status.
//
// Code 2428841 is the case from GitHub issue #4197 "MetroclusterCheck collector
// frequent failed status", which shipped with the status/untested label: the
// check being in progress is expected and must be distinguishable from a real
// failure.
func TestIsRestErr(t *testing.T) {
	tests := []struct {
		name     string
		code     int64
		sentinel OntapRestCode
		want     bool
	}{
		{name: "MetroCluster check in progress", code: 2428841, sentinel: MetroClusterCheckInProgress, want: true},
		{name: "MetroCluster not configured", code: 2426405, sentinel: MetroClusterNotConfigured, want: true},
		{name: "API not found", code: 3, sentinel: APINotFound, want: true},
		{name: "entry does not exist", code: 4, sentinel: EntryNotExist, want: true},
		{name: "table not found", code: 8585320, sentinel: TableNotFound, want: true},
		{name: "CM reject", code: 8585368, sentinel: CMReject, want: true},
		{
			name:     "check-in-progress code is not read as not-configured",
			code:     2428841,
			sentinel: MetroClusterNotConfigured,
			want:     false,
		},
		{
			name:     "not-configured code is not read as check-in-progress",
			code:     2426405,
			sentinel: MetroClusterCheckInProgress,
			want:     false,
		},
		{name: "unrelated code matches nothing", code: 999999, sentinel: MetroClusterCheckInProgress, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := NewRest().Error(ErrAPIRequestRejected).Code(tc.code).Build()
			assert.Equal(t, IsRestErr(err, tc.sentinel), tc.want)
		})
	}
}

// TestIsRestErrOnNonRestError pins that a plain error, or nil, is never
// classified as an ONTAP REST code.
func TestIsRestErrOnNonRestError(t *testing.T) {
	assert.False(t, IsRestErr(errors.New("connection refused"), MetroClusterCheckInProgress))
	assert.False(t, IsRestErr(nil, MetroClusterCheckInProgress))
}

// TestIsRestErrThroughWrappedError pins that classification survives wrapping,
// since handleError and the collectors wrap these errors before they are
// inspected again further up.
func TestIsRestErrThroughWrappedError(t *testing.T) {
	inner := NewRest().Error(ErrAPIRequestRejected).Code(2428841).Build()
	wrapped := fmt.Errorf("failed to fetch data: %w", inner)

	assert.True(t, IsRestErr(wrapped, MetroClusterCheckInProgress))
}
