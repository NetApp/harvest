package errors

import (
	"strings"
)
const (
	MISSING_PARAM = "missing parameter"
	INVALID_PARAM = "invalid parameter"
	ERR_CONNECTION = "connection error"
	ERR_CONFIG = "configuration error"
	ERR_NO_METRIC = "no metrics"
	ERR_NO_INSTANCE = "no instances"
	MATRIX_HASH = "matrix error"
	MATRIX_EMPTY = "empty cache"
	MATRIX_INV_PARAM = "matrix invalid parameter"
	MATRIX_PARSE_STR = "parse numeric value from string"
	API_RESPONSE = "error reading api response"
	API_REQ_REJECTED = "api request rejected"
	ERR_DLOAD = "dynamic module"
	ERR_IMPLEMENT = "implementation error"
	ERR_SCHEDULE = "schedule error"
)

type Error struct {
	err string
	msg string
}

func (e Error) Error() string {
	return e.err + ": " + e.msg
}

func New(name, msg string) Error {
	return Error{err:name, msg:msg}
}

func IsErr(err error, code string) bool {
	// dirty solution, temporarily
	return strings.Contains(err.Error(), code)
}