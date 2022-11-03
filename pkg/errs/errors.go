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
	ErrTemplate           = harvestError("invalid template")
)

type HarvestError struct {
	Message string
	Inner   error
}

func (e HarvestError) Error() string {
	if e.Inner == nil {
		return e.Message
	}
	if e.Message == "" {
		return e.Inner.Error()
	}
	return fmt.Sprintf("%s => %s", e.Inner.Error(), e.Message)
}

func (e HarvestError) Unwrap() error {
	return e.Inner
}

func New(err error, message string) error {
	return HarvestError{Message: message, Inner: err}
}
