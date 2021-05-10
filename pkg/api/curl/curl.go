/*
 * Copyright NetApp Inc, 2021 All rights reserved

Helper function to do simple HTTP Get request (much like the bash curl command).
*/
package curl

import (
	"io/ioutil"
	"net/http"
	"strings"
)

// Curl issues a GET request to addr and returns data if any
func Curl(addr string) ([]byte, error) {

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
		return data, err
	}

	client = &http.Client{}

	if resp, err = client.Do(req); err != nil {
		return data, err
	}

	defer resp.Body.Close()

	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return data, err
	}

	return data, nil
}
