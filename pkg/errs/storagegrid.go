package errs

import (
	"encoding/json"
)

type StorageGridError struct {
	Message Message `json:"message"`
	Code    int     `json:"code"`
	Status  string  `json:"status"`
}

func (s StorageGridError) Error() string {
	return s.Message.Text
}

type Message struct {
	Text string `json:"text"`
	Key  string `json:"key"`
}

func NewStorageGridErr(jsonText []byte) error {
	var e StorageGridError
	err := json.Unmarshal(jsonText, &e)
	if err != nil {
		return New(err, "failed to unmarshal storage grid err")
	}
	return e
}
