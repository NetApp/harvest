package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func GetResponse(url string) (string, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func GetResponseBody(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	return body, nil
}

func SendReqAndGetRes(url string, method string,
	buf []byte) map[string]interface{} {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(buf))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	PanicIfNotNil(err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	PanicIfNotNil(err)
	log.Println(string(body))
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	PanicIfNotNil(err)
	return data
}

func SendPostReqAndGetRes(url string, method string, buf []byte, user string, pass string) map[string]interface{} {
	tlsConfig := &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(buf))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(user, pass)
	res, err := client.Do(req)
	PanicIfNotNil(err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	PanicIfNotNil(err)
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	PanicIfNotNil(err)
	return data
}
