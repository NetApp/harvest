package timedmap

import "errors"

var (
	// ErrKeyNotFound is returned when a key was
	// requested which is not present in the map.
	ErrKeyNotFound = errors.New("key not found")
)
