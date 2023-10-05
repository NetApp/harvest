package errs

import (
	"errors"
	"fmt"
)

type RestError struct {
	Message    string
	Target     string
	Code       int64
	StatusCode int
	API        string
}

func (r *RestError) Error() string {
	return fmt.Sprintf("%d: %s: %s", r.Code, r.Message, r.API)
}

func Rest(statusCode int, message string, code int64, target string, api string) error {
	return &RestError{
		StatusCode: statusCode,
		Message:    message,
		Code:       code,
		Target:     target,
		API:        api,
	}
}

type OntapRestCode struct {
	Name string
	Code int64
}

var (
	APINotFound   = OntapRestCode{"API not found", 3}
	TableNotFound = OntapRestCode{"Table is not found", 8585320}
)

func IsRestErr(err error, sentinel OntapRestCode) bool {
	var restErr *RestError
	if errors.As(err, &restErr) {
		if restErr.Code == sentinel.Code {
			return true
		}
	}
	return false
}
