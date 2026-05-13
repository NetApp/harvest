// Copyright NetApp Inc, 2021 All rights reserved

package rest

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestClientGetPlainRestConcurrent(t *testing.T) {
	credentials := auth.NewCredentials(&conf.Poller{
		Name:     "test",
		Addr:     "cluster.example",
		Username: "user",
		Password: "pass",
	}, slog.Default())

	client := &Client{
		client: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"records":[],"num_records":0}`)),
				Header:     make(http.Header),
			}, nil
		})},
		Logger:  slog.Default(),
		baseURL: "https://cluster.example/",
		auth:    credentials,
	}
	metadata := &collector.Metadata{}

	const workers = 32
	const iterations = 50

	var wg sync.WaitGroup
	for worker := range workers {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := range iterations {
				_, err := client.GetPlainRest(metadata, "api/cluster", false, map[string]string{
					"X-Worker": strconv.Itoa(worker),
					"X-Iter":   strconv.Itoa(i),
				})
				if err != nil {
					t.Errorf("GetPlainRest() error = %v", err)
					return
				}
			}
		}(worker)
	}
	wg.Wait()

	if metadata.NumCalls.Load() != uint64(workers*iterations) {
		t.Fatalf("NumCalls = %d, want %d", metadata.NumCalls.Load(), workers*iterations)
	}
}
