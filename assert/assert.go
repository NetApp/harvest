package assert

// Inspired by https://www.alexedwards.net/blog/the-9-go-test-assertions-i-use

import (
	"errors"
	"reflect"
	"testing"
)

func Equal[T any](t *testing.T, got T, want T) {
	t.Helper()
	if !isEqual(got, want) {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func NotEqual[T any](t *testing.T, got, want T) {
	t.Helper()
	if isEqual(got, want) {
		t.Errorf("got: %v; expected values to be different", got)
	}
}

func True(t *testing.T, got bool) {
	t.Helper()
	if !got {
		t.Errorf("got: false; want: true")
	}
}

func False(t *testing.T, got bool) {
	t.Helper()
	if got {
		t.Errorf("got: true; want: false")
	}
}

func Nil(t *testing.T, got any) {
	t.Helper()
	if !isNil(got) {
		t.Errorf("got: %v; want: nil", got)
	}
}

func NotNil(t *testing.T, got any) {
	t.Helper()
	if isNil(got) {
		t.Errorf("got: nil; want: non-nil")
	}
}

func ErrorIs(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func isEqual[T any](got, want T) bool {
	if isNil(got) && isNil(want) {
		return true
	}
	if equalable, ok := any(got).(interface{ Equal(T) bool }); ok {
		return equalable.Equal(want)
	}
	return reflect.DeepEqual(got, want)
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
