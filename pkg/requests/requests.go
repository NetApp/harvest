package requests

import (
	"github.com/netapp/harvest/v2/cmd/harvest/version"
	"io"
	"net/http"
)

func New(method, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	harvestUserAgent := "Harvest/" + version.VERSION
	request.Header.Set("User-Agent", harvestUserAgent)
	return request, nil
}
