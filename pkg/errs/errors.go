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
	ErrAPIRequestRejected        = harvestError("API request rejected")
	ErrAPIResponse               = harvestError("error reading api response")
	ErrAttributeNotFound         = harvestError("attribute not found")
	ErrAuthFailed                = harvestError("auth failed")
	ErrConfig                    = harvestError("configuration error")
	ErrConnection                = harvestError("connection error")
	ErrImplement                 = harvestError("implementation error")
	ErrInvalidItem               = harvestError("invalid item")
	ErrInvalidParam              = harvestError("invalid parameter")
	ErrMissingParam              = harvestError("missing parameter")
	ErrNoCollector               = harvestError("no collectors")
	ErrNoInstance                = harvestError("no instances")
	ErrNoMetric                  = harvestError("no metrics")
	ErrPermissionDenied          = harvestError("Permission denied")
	ErrResponseNotFound          = harvestError("response not found")
	ErrWrongTemplate             = harvestError("wrong template")
	ErrMetroClusterNotConfigured = harvestError("MetroCluster is not configured in cluster")
	ErrTemplateNotSupported      = harvestError("template not supported")
)

const (
	ErrNumZAPISuspended  = "61253"
	ZAPIPermissionDenied = "13003"
)

type HarvestError struct {
	Message    string
	Inner      error
	ErrNum     string
	StatusCode int
}

type Option func(*HarvestError)

func WithStatus(statusCode int) Option {
	return func(e *HarvestError) {
		e.StatusCode = statusCode
	}
}

func WithErrorNum(errNum string) Option {
	return func(e *HarvestError) {
		e.ErrNum = errNum
	}
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
	return fmt.Sprintf(`%s => %s errNum=%q statusCode="%d"`, e.Inner.Error(), e.Message, e.ErrNum, e.StatusCode)
}

func (e HarvestError) Unwrap() error {
	return e.Inner
}

func New(innerError error, message string, opts ...Option) error {
	err := HarvestError{Message: message, Inner: innerError}
	for _, opt := range opts {
		opt(&err)
	}
	return err
}
