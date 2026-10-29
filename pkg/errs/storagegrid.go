package errs

import (
	"encoding/json"
)

type StorageGridError struct {
	Message Message `json:"message"`
	Code    int     `json:"code"`
	Status  string  `json:"status"`
}

type Message struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

func (s StorageGridError) Error() string {
	return s.Message.Text
}

func (s StorageGridError) IsAuthErr() bool {
	return s.Code == 401
}

// Is lets errors.Is(err, ErrAuthFailed) match a StorageGRID 401,
// so callers can treat auth failures the same way across collectors.
func (s StorageGridError) Is(target error) bool {
	return target == ErrAuthFailed && s.IsAuthErr()
}

func NewStorageGridErr(statusCode int, jsonText []byte) error {
	var e StorageGridError
	err := json.Unmarshal(jsonText, &e)
	truncated := string(jsonText[:min(100, len(jsonText))])
	// Check for 401 first so an empty or non-JSON body (e.g., an HTML page from a proxy)
	// is still recognized as an auth failure.
	if statusCode == 401 {
		e.Code = 401
		e.Message.Text = ErrAuthFailed.Error()
		e.Status = truncated
		return e
	}
	if err != nil {
		return New(err, "failed to unmarshal storage grid err because text was "+truncated)
	}
	return e
}
