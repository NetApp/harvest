package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"io"
	"log/slog"
	"net/http"
	"os"
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
		slog.Error("", slogx.Err(err))
		os.Exit(1)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("", slogx.Err(err))
		os.Exit(1)
	}
	resp.Body.Close()
	return body, nil
}

func SendReqAndGetRes(url string, method string,
	buf []byte) map[string]interface{} {
	client := &http.Client{}
	req, err := requests.New(method, url, bytes.NewBuffer(buf))
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
	slog.Info(string(body))
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	PanicIfNotNil(err)
	return data
}
