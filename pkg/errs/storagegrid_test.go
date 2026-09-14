package errs

import (
	"errors"
	"strings"
	"testing"

	"github.com/netapp/harvest/v2/assert"
)

// TestNewStorageGridErrNonJSONBody covers GitHub issue #4035
// "StorageGrid: Invalid Character Error on gather remote info".
//
// When a StorageGrid collector is aimed at something that is not StorageGrid,
// the error body is an HTML page rather than JSON. json.Unmarshal then fails
// and the raw text has to be surfaced, truncated, so the operator can see what
// actually came back instead of only "invalid character '<'".
func TestNewStorageGridErrNonJSONBody(t *testing.T) {
	body := []byte(`<!DOCTYPE html><html><head><title>404 Not Found</title></head></html>`)

	err := NewStorageGridErr(404, body)

	assert.NotNil(t, err)
	// It must not masquerade as a StorageGridError, because callers key their
	// 401 retry off that type.
	_, ok := errors.AsType[StorageGridError](err)
	assert.False(t, ok)
	assert.True(t, strings.Contains(err.Error(), "failed to unmarshal storage grid err"))
	assert.True(t, strings.Contains(err.Error(), "<!DOCTYPE html>"))
}

// TestNewStorageGridErrTruncatesLongBody pins the 100-byte cap on the echoed
// body, so a large HTML page cannot flood the log on every poll.
func TestNewStorageGridErrTruncatesLongBody(t *testing.T) {
	body := []byte("<html>" + strings.Repeat("x", 500) + "</html>")

	err := NewStorageGridErr(500, body)

	assert.NotNil(t, err)
	msg := err.Error()
	// 100 bytes of body, and nothing beyond it.
	assert.True(t, strings.Contains(msg, "<html>"+strings.Repeat("x", 94)))
	assert.False(t, strings.Contains(msg, strings.Repeat("x", 101)))
}

// TestNewStorageGridErrShortBodyNotTruncated guards the min() boundary: a body
// shorter than the cap must be echoed whole and must not slice out of range.
func TestNewStorageGridErrShortBodyNotTruncated(t *testing.T) {
	body := []byte("<b>nope</b>")

	err := NewStorageGridErr(500, body)

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "<b>nope</b>"))
}

// TestNewStorageGridErrEmptyBody is the degenerate case -- an empty body still
// fails to unmarshal, and must not panic on the slice.
func TestNewStorageGridErrEmptyBody(t *testing.T) {
	err := NewStorageGridErr(500, []byte(""))

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to unmarshal storage grid err"))
}

// TestNewStorageGridErrAuthFailure covers the classification behind GitHub
// issue #4009: a 401 has to come back as a StorageGridError whose IsAuthErr
// reports true, because that is the only thing that triggers the client's
// token-refresh retry.
func TestNewStorageGridErrAuthFailure(t *testing.T) {
	body := []byte(`{"message":{"text":"unauthorized","key":"auth"},"status":"error"}`)

	err := NewStorageGridErr(401, body)

	sgErr, ok := errors.AsType[StorageGridError](err)
	assert.True(t, ok)
	assert.True(t, sgErr.IsAuthErr())
	assert.Equal(t, sgErr.Code, 401)
	// The message is replaced with the canonical auth-failure text, and the
	// original body is preserved in Status for diagnosis.
	assert.Equal(t, sgErr.Message.Text, ErrAuthFailed.Error())
	assert.Equal(t, sgErr.Status, string(body))
}

// TestNewStorageGridErrNon401JSON is the control: a well-formed non-401 error
// must be preserved as-is and must not be treated as an auth failure, otherwise
// the client would retry-and-reauthorize on every ordinary error.
func TestNewStorageGridErrNon401JSON(t *testing.T) {
	body := []byte(`{"message":{"text":"not found"},"code":404,"status":"error"}`)

	err := NewStorageGridErr(404, body)

	sgErr, ok := errors.AsType[StorageGridError](err)
	assert.True(t, ok)
	assert.False(t, sgErr.IsAuthErr())
	assert.Equal(t, sgErr.Message.Text, "not found")
	assert.Equal(t, sgErr.Error(), "not found")
}

// TestStorageGridErrIsAuthErr pins IsAuthErr on the code alone, since the
// client's retry decision depends only on it.
func TestStorageGridErrIsAuthErr(t *testing.T) {
	tests := []struct {
		name string
		code int
		want bool
	}{
		{name: "401 is an auth error", code: 401, want: true},
		{name: "403 is not", code: 403, want: false},
		{name: "500 is not", code: 500, want: false},
		{name: "zero is not", code: 0, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := StorageGridError{Code: tc.code}
			assert.Equal(t, e.IsAuthErr(), tc.want)
		})
	}
}
