/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package errors

import (
	"strings"
)

const (
	MISSING_PARAM    = "missing parameter"
	INVALID_PARAM    = "invalid parameter"
	ERR_CONNECTION   = "connection error"
	ERR_CONFIG       = "configuration error"
	ERR_NO_METRIC    = "no metrics"
	ERR_NO_INSTANCE  = "no instances"
	ERR_NO_COLLECTOR = "no collectors"
	MATRIX_HASH      = "matrix error"
	MATRIX_EMPTY     = "empty cache"
	MATRIX_INV_PARAM = "matrix invalid parameter"
	MATRIX_PARSE_STR = "parse numeric value from string"
	API_RESPONSE     = "error reading api response"
	API_REQ_REJECTED = "api request rejected"
	// @TODO, implement: API response is something like
	// Insufficient privileges: user 'harvest2-user' does not have write access to this resource
	API_INSUF_PRIV   = "api insufficient priviliges"
	ERR_DLOAD        = "dynamic load"
	ERR_IMPLEMENT    = "implementation error"
	ERR_SCHEDULE     = "schedule error"
	GO_ROUTINE_PANIC = "goroutine panic"
)

type Error struct {
	class string
	msg   string
}

func (e Error) Error() string {
	return e.class + " => " + e.msg
}

func New(class, msg string) Error {
	return Error{class: class, msg: msg}
}

func GetClass(err error) string {
	e := strings.Split(err.Error(), " => ")
	if len(e) > 1 {
		return e[0]
	}
	return ""
}

func IsErr(err error, class string) bool {
	// dirty solution, temporarily
	return strings.Contains(GetClass(err), class)
}
