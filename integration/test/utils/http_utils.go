package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func GetResponse(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	return string(body), nil
}

func GetResponseBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
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
	defer func(Body io.ReadCloser) { _ = Body.Close() }(res.Body)
	body, err := io.ReadAll(res.Body)
	PanicIfNotNil(err)
	log.Println(string(body))
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	PanicIfNotNil(err)
	return data
}
