package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Netapp/harvest-automation/test/errs"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"io"
	"log/slog"
	http2 "net/http"
	"os"
)

func GetResponse(url string) (string, error) {
	resp, err := http2.Get(url) //nolint:gosec
	if err != nil {
		return "", err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func GetResponseBody(url string) ([]byte, error) {
	resp, err := http2.Get(url) //nolint:gosec
	if err != nil {
		slog.Error("", slogx.Err(err))
		os.Exit(1)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("", slogx.Err(err))
		os.Exit(1)
	}
	//goland:noinspection GoUnhandledErrorResult
	resp.Body.Close() //nolint:gosec
	return body, nil
}

func SendReqAndGetRes(url string, method string,
	buf []byte) map[string]any {
	client := &http2.Client{}
	req, err := requests.New(method, url, bytes.NewBuffer(buf))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	errs.PanicIfNotNil(err)
	//goland:noinspection GoUnhandledErrorResult
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	errs.PanicIfNotNil(err)
	slog.Info(string(body))
	var data map[string]any
	err = json.Unmarshal(body, &data)
	errs.PanicIfNotNil(err)
	return data
}
