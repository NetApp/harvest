/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package errs

import (
	"fmt"
)

type harvestError string

func (e harvestError) Error() string { return string(e) }

const (
	ErrAPIRequestRejected = harvestError("API request rejected")
	ErrAPIResponse        = harvestError("error reading api response")
	ErrAttributeNotFound  = harvestError("attribute not found")
	ErrAuthFailed         = harvestError("auth failed")
	ErrConfig             = harvestError("configuration error")
	ErrConnection         = harvestError("connection error")
	ErrImplement          = harvestError("implementation error")
	ErrInvalidItem        = harvestError("invalid item")
	ErrInvalidParam       = harvestError("invalid parameter")
	ErrMissingParam       = harvestError("missing parameter")
	ErrMissingParams      = harvestError("missing parameter")
	ErrNoCollector        = harvestError("no collectors")
	ErrNoInstance         = harvestError("no instances")
	ErrNoMetric           = harvestError("no metrics")
	ErrPanic              = harvestError("goroutine panic")
	ErrWrongTemplate      = harvestError("wrong template")
)

const (
	ErrNumZAPISuspended = "61253"
)

type HarvestError struct {
	Message    string
	Inner      error
	ErrNum     string
	StatusCode int
}

func (e HarvestError) Error() string {
	if e.Inner == nil {
		return e.Message
	}
	if e.Message == "" {
		return e.Inner.Error()
	}
	if e.ErrNum == "" && e.StatusCode == 0 {
		return fmt.Sprintf("%s => %s", e.Inner.Error(), e.Message)
	}
	return fmt.Sprintf(`%s => %s errNum="%s" statusCode="%d"`, e.Inner.Error(), e.Message, e.ErrNum, e.StatusCode)
}

func (e HarvestError) Unwrap() error {
	return e.Inner
}

func New(err error, message string) error {
	return HarvestError{Message: message, Inner: err}
}

func NewWithStatus(err error, message string, statusCode int) error {
	return HarvestError{Message: message, Inner: err, StatusCode: statusCode}
}

func NewWithErrorNum(err error, message string, errNum string) error {
	return HarvestError{Message: message, Inner: err, ErrNum: errNum}
}
