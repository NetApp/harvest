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

// NewDummyClientWithBaseURL creates a dummy client that talks to baseURL, for
// tests that serve responses from an httptest server.
func NewDummyClientWithBaseURL(baseURL string) *Client {
	c := NewDummyClient()
	c.baseURL = baseURL
	return c
}

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
		client:  httpClient,
		request: httpRequest,
		buffer:  buffer,
		Logger:  slog.Default(),
		baseURL: "http://example.com",
		Remote:  remote,
		token:   "TestToken",
		Timeout: time.Second * 10,
		logRest: true,
		APIPath: "/api/v1",
		// Real Credentials rather than &auth.Credentials{}: a zero-value
		// Credentials has a nil authMu mutex, so any code path that reaches
		// GetPollerAuth -- the 401 retry in invoke, for one -- panics.
		auth: auth.NewCredentials(&conf.Poller{
			Name:     "test",
			Addr:     "127.0.0.1",
			Username: "admin",
			Password: "secret",
		}, slog.Default()),
		Metadata: &collector.Metadata{},
	}

	return client
}
