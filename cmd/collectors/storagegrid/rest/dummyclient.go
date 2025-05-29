package rest

import (
	"bytes"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"log/slog"

	"github.com/netapp/harvest/v2/pkg/auth"
	"net/http"
	"time"
)

// NewDummyClient creates a new dummy client
func NewDummyClient() *Client {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}

	httpRequest, _ := http.NewRequest(http.MethodGet, "http://example.com", http.NoBody)

	buffer := new(bytes.Buffer)

	remote := conf.Remote{
		Name:    "TestCluster",
		UUID:    "TestUUID",
		Version: "1.2.3",
	}

	client := &Client{
		client:   httpClient,
		request:  httpRequest,
		buffer:   buffer,
		Logger:   slog.Default(),
		baseURL:  "http://example.com",
		Remote:   remote,
		token:    "TestToken",
		Timeout:  time.Second * 10,
		logRest:  true,
		APIPath:  "/api/v1",
		auth:     &auth.Credentials{},
		Metadata: &collector.Metadata{},
	}

	return client
}
