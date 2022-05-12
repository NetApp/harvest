/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package errors

import (
	"fmt"
	"strings"
)

const (
	MissingParam   = "missing parameter"
	InvalidParam   = "invalid parameter"
	ErrConnection  = "connection error"
	ErrConfig      = "configuration error"
	ErrNoMetric    = "no metrics"
	ErrNoInstance  = "no instances"
	ErrTemplate    = "invalid template"
	ErrNoCollector = "no collectors"
	ApiResponse    = "error reading api response"
	ApiReqRejected = "api request rejected"
	ErrImplement   = "implementation error"
	GoRoutinePanic = "goroutine panic"
)

func New(class, msg string) error {
	return fmt.Errorf("%s => %s", class, msg)
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
