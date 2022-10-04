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

func NewStorageGridErr(statusCode int, jsonText []byte) error {
	var e StorageGridError
	err := json.Unmarshal(jsonText, &e)
	if err != nil {
		return New(err, "failed to unmarshal storage grid err")
	}
	if statusCode == 401 {
		e.Code = 401
		e.Message.Text = ErrAuthFailed.Error()
		e.Status = string(jsonText)
	}
	return e
}
