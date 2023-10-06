package errs

import (
	"errors"
	"fmt"
	"strings"
)

type RestError struct {
	InnerError string
	Message    string
	Target     string
	Code       int64
	StatusCode int
	API        string
}

func (r *RestError) Error() string {
	var parts []string
	if r.StatusCode != 0 {
		parts = append(parts, fmt.Sprintf("StatusCode: %d", r.StatusCode))
	}
	if r.InnerError != "" {
		parts = append(parts, fmt.Sprintf("InnerError: %s", r.InnerError))
	}
	if r.Message != "" {
		parts = append(parts, fmt.Sprintf("Message: %s", r.Message))
	}
	if r.Code != 0 {
		parts = append(parts, fmt.Sprintf("Code: %d", r.Code))
	}
	if r.Target != "" {
		parts = append(parts, fmt.Sprintf("Target: %s", r.Target))
	}
	if r.API != "" {
		parts = append(parts, fmt.Sprintf("API: %s", r.API))
	}
	return strings.Join(parts, ", ")
}

type RestErrorBuilder struct {
	restError RestError
}

func NewRestError() *RestErrorBuilder {
	return &RestErrorBuilder{}
}

func (b *RestErrorBuilder) StatusCode(statusCode int) *RestErrorBuilder {
	b.restError.StatusCode = statusCode
	return b
}

func (b *RestErrorBuilder) InnerError(innerError string) *RestErrorBuilder {
	b.restError.InnerError = innerError
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
