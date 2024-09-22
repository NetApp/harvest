package rest

import (
	"bytes"
	"github.com/netapp/harvest/v2/pkg/util"
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

	cluster := Cluster{
		Name:    "TestCluster",
		Info:    "TestInfo",
		UUID:    "TestUUID",
		Version: [3]int{1, 2, 3},
	}

	client := &Client{
		client:   httpClient,
		request:  httpRequest,
		buffer:   buffer,
		Logger:   slog.Default(),
		baseURL:  "http://example.com",
		Cluster:  cluster,
		token:    "TestToken",
		Timeout:  time.Second * 10,
		logRest:  true,
		APIPath:  "/api/v1",
		auth:     &auth.Credentials{},
		Metadata: &util.Metadata{},
	}

	return client
}
