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

func TestClientPostRestAuthRetryReplaysBody(t *testing.T) {
	credentials := auth.NewCredentials(&conf.Poller{
		Name:     "test",
		Addr:     "cluster.example",
		Username: "user",
		CredentialsScript: conf.CredentialsScript{
			Path: "../../../pkg/auth/testdata/get_credentials_yaml",
		},
	}, slog.Default())

	expectedBody := `{"hello":"world"}`
	var bodies []string
	var attempts int

	client := &Client{
		client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}
			bodies = append(bodies, string(payload))
			attempts++

			statusCode := http.StatusOK
			responseBody := `{"records":[]}`
			if attempts == 1 {
				statusCode = http.StatusUnauthorized
				responseBody = `{"error":{"message":"unauthorized"}}`
			}

			return &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(strings.NewReader(responseBody)),
				Header:     make(http.Header),
			}, nil
		})},
		Logger:  slog.Default(),
		baseURL: "https://cluster.example/",
		auth:    credentials,
	}

	_, err := client.PostRest(nil, "api/cluster", []byte(expectedBody))
	if err != nil {
		t.Fatalf("PostRest() error = %v", err)
	}

	if len(bodies) != 2 {
		t.Fatalf("got %d request bodies, want 2", len(bodies))
	}
	if bodies[0] != expectedBody {
		t.Fatalf("first request body = %q, want %q", bodies[0], expectedBody)
	}
	if bodies[1] != expectedBody {
		t.Fatalf("retried request body = %q, want %q", bodies[1], expectedBody)
	}
}
