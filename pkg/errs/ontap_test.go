package errs

import (
	"errors"
	"testing"
)

func TestErrAuthFailed(t *testing.T) {
	err := NewRest().Error(ErrAuthFailed).Build()
	if !errors.Is(err, ErrAuthFailed) {
		t.Errorf("err should be ErrAuthFailed but isn't")
	}
}
