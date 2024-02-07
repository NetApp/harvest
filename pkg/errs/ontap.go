package errs

import (
	"errors"
	"fmt"
	"strings"
)

type RestError struct {
	Err        error
	Message    string
	Target     string
	Code       int64
	StatusCode int
	API        string
}

func (r *RestError) Unwrap() error {
	return r.Err
}

func (r *RestError) Error() string {
	var parts []string
	if r.StatusCode != 0 {
		parts = append(parts, fmt.Sprintf("StatusCode: %d", r.StatusCode))
	}
	if r.Err != nil {
		parts = append(parts, fmt.Sprintf("Error: %s", r.Err))
	}
	if r.Message != "" {
		parts = append(parts, "Message: "+r.Message)
	}
	if r.Code != 0 {
		parts = append(parts, fmt.Sprintf("Code: %d", r.Code))
	}
	if r.Target != "" {
		parts = append(parts, "Target: "+r.Target)
	}
	if r.API != "" {
		parts = append(parts, "API: "+r.API)
	}
	return strings.Join(parts, ", ")
}

type RestErrorBuilder struct {
	restError RestError
}

func NewRest() *RestErrorBuilder {
	return &RestErrorBuilder{}
}

func (b *RestErrorBuilder) StatusCode(statusCode int) *RestErrorBuilder {
	b.restError.StatusCode = statusCode
	return b
}

func (b *RestErrorBuilder) Error(err error) *RestErrorBuilder {
	b.restError.Err = err
	return b
}

func (b *RestErrorBuilder) Message(message string) *RestErrorBuilder {
	b.restError.Message = message
	return b
}

func (b *RestErrorBuilder) Code(code int64) *RestErrorBuilder {
	b.restError.Code = code
	return b
}

func (b *RestErrorBuilder) Target(target string) *RestErrorBuilder {
	b.restError.Target = target
	return b
}

func (b *RestErrorBuilder) API(api string) *RestErrorBuilder {
	b.restError.API = api
	return b
}

func (b *RestErrorBuilder) Build() error {
	return &b.restError
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
