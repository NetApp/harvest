package rest

import (
	"bytes"
	"github.com/netapp/harvest/v2/pkg/util"

	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/logging"
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

	logger := logging.Get()

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
		Logger:   logger,
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
