package util

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func Curl(addr string) (string, error) {

	var (
		req    *http.Request
		client *http.Client
		resp   *http.Response
		data   []byte
		err    error
	)

	if !strings.HasPrefix("http", addr) {
		addr = "http://" + addr
	}

	if req, err = http.NewRequest("GET", addr, nil); err != nil {
		return "", err
	}

	client = &http.Client{}

	if resp, err = client.Do(req); err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return "", err
	}

	return string(data), nil
}
