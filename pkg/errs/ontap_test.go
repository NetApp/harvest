package errs

import (
	"github.com/netapp/harvest/v2/assert"
	"testing"
)

func TestErrAuthFailed(t *testing.T) {
	err := NewRest().Error(ErrAuthFailed).Build()
	assert.ErrorIs(t, err, ErrAuthFailed)
}
