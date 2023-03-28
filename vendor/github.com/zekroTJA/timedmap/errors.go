package timedmap

import (
	"errors"
)

var (
	// ErrKeyNotFound is returned when a key was
	// requested which is not present in the map.
	ErrKeyNotFound = errors.New("key not found")

	// ErrValueNoMap is returned when a value passed
	// expected was of another type.
	ErrValueNoMap = errors.New("value is not of type map")
)
