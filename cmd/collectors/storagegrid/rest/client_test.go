package rest

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
)

// pathRecorder records every path the client requests, so a test can assert on
// the URLs the client built rather than only on the final result.
type pathRecorder struct {
	mu    sync.Mutex
	paths []string
}

func (p *pathRecorder) add(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paths = append(p.paths, path)
}

func (p *pathRecorder) all() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string(nil), p.paths...)
}

func (p *pathRecorder) has(path string) bool {
	return slices.Contains(p.all(), path)
}

// newTestClient builds a Client aimed at srv, with username/password auth and
// the given APIPath. Tests that need a credential script build their own Client
// so they can mutate the poller's script path mid-test.
func newTestClient(t *testing.T, srv *httptest.Server, apiPath string) *Client {
	t.Helper()

	poller := &conf.Poller{
		Name:     "test",
		Addr:     "127.0.0.1",
		Username: "admin",
		Password: "secret",
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest err: %v", err)
	}

	return &Client{
		client:   srv.Client(),
		request:  req,
		buffer:   new(bytes.Buffer),
		Logger:   slog.Default(),
		baseURL:  srv.URL + "/",
		token:    "stale-token",
		Timeout:  10 * time.Second,
		APIPath:  apiPath,
		auth:     auth.NewCredentials(poller, slog.Default()),
		Metadata: &collector.Metadata{},
	}
}

// authorizeServer returns a server that 401s every request except the versioned
// authorize endpoint, and records the paths it is asked for.
func authorizeServer(t *testing.T, versionedAuthPath string) (*httptest.Server, *pathRecorder) {
	t.Helper()
	rec := &pathRecorder{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.add(r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == versionedAuthPath {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":"fresh-token"}`))
			return
		}

		// Anything else is unauthorized. The body must be JSON so that
		// NewStorageGridErr produces a StorageGridError, which is what
		// invoke() keys the retry off.
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":{"text":"unauthorized"}}`))
	}))
	t.Cleanup(srv.Close)

	return srv, rec
}

// TestFetchTokenUsesVersionedAuthorizeWhenAPIPathUnset is the regression test
// for the unversioned-/authorize bug.
//
// Client.New never assigns APIPath, and Init calls sniffAPIVersion before
// anything sets it. sniffAPIVersion fetches through getRest -> invoke, so a 401
// on /api/versions reaches fetchTokenWithAuthRetry while APIPath is still "".
// url.JoinPath drops empty elements, so the authorize URL came out as
// "<base>/authorize" instead of "<base>/api/v3/authorize". Against a target
// that is not StorageGRID, that path returns an HTML error page, which
// NewStorageGridErr cannot unmarshal, surfacing as
// "invalid character '<' looking for beginning of value" -- the symptom
// reported in GitHub issue #4035.
func TestFetchTokenUsesVersionedAuthorizeWhenAPIPathUnset(t *testing.T) {
	srv, rec := authorizeServer(t, "/api/v3/authorize")

	// APIPath is empty, exactly as it is during sniffAPIVersion.
	c := newTestClient(t, srv, "")

	err := c.fetchTokenWithAuthRetry()

	assert.Nil(t, err)
	assert.True(t, rec.has("/api/v3/authorize"))
	// The bare, unversioned path must never be requested.
	assert.False(t, rec.has("/authorize"))
	// The fresh token replaced the stale one.
	assert.Equal(t, c.token, "fresh-token")
}

// TestFetchTokenHonoursDiscoveredAPIPath is the control: once sniffAPIVersion
// has set APIPath, the fallback must not override it.
func TestFetchTokenHonoursDiscoveredAPIPath(t *testing.T) {
	srv, rec := authorizeServer(t, "/api/v4/authorize")

	c := newTestClient(t, srv, "/api/v4")

	err := c.fetchTokenWithAuthRetry()

	assert.Nil(t, err)
	assert.True(t, rec.has("/api/v4/authorize"))
	assert.False(t, rec.has("/api/v3/authorize"))
	assert.Equal(t, c.token, "fresh-token")
}

// TestSniffAPIVersion401DoesNotHitUnversionedAuthorize drives the bug through
// the real entry point: a 401 during version discovery.
func TestSniffAPIVersion401DoesNotHitUnversionedAuthorize(t *testing.T) {
	srv, rec := authorizeServer(t, "/api/v3/authorize")

	c := newTestClient(t, srv, "")

	// The result is expected to fail -- the server 401s /api/versions no matter
	// what. What matters is which authorize URL was attempted on the way.
	_ = c.sniffAPIVersion(1)

	assert.True(t, rec.has("/api/versions"))
	assert.False(t, rec.has("/authorize"))
}

// TestAuthorizeURL exercises the production URL builder directly, for both the
// unset and discovered APIPath cases. It calls Client.authorizeURL rather than
// re-deriving the fallback, so removing that fallback fails this test.
func TestAuthorizeURL(t *testing.T) {
	tests := []struct {
		name    string
		apiPath string
		want    string
	}{
		{name: "unset APIPath falls back to the default version", apiPath: "", want: "https://host/api/v3/authorize"},
		{name: "discovered APIPath is used as-is", apiPath: "/api/v4", want: "https://host/api/v4/authorize"},
		{name: "a future version is not rewritten", apiPath: "/api/v9", want: "https://host/api/v9/authorize"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{baseURL: "https://host/", APIPath: tc.apiPath}
			got, err := c.authorizeURL()
			assert.Nil(t, err)
			assert.Equal(t, got, tc.want)
		})
	}
}

// TestSetVersion covers the version parsing that Init depends on, trimming the
// StorageGRID build suffix down to a semantic version.
func TestSetVersion(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "full build string", in: "11.6.0.3-20220802.2201.f58633a", want: "11.6.0"},
		{name: "plain three-part version", in: "11.7.0", want: "11.7.0"},
		{name: "two-part version", in: "11.8", want: "11.8.0"},
		{name: "unparsable", in: "not-a-version", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{Logger: slog.Default()}
			err := c.SetVersion(tc.in)
			if tc.wantErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, c.Remote.Version, tc.want)
		})
	}
}

// TestInvokeExpiresCredentialScriptTokenOn401 covers GitHub issue #4009
// "StorageGRID: Cached credential script tokens not expired on 401", which
// shipped with the status/untested label.
//
// The bug is specific to the credential-script-with-authToken path. When the
// script supplies an authToken, fetchTokenWithAuthRetry short-circuits and
// reuses the cached token without any network call, so its own inner retry
// cannot help. The only thing that breaks the loop is invoke calling
// auth.Expire() to invalidate the cache so the script runs again.
//
// The two fixtures make the re-run observable: get_token_stale yields the token
// the grid now rejects, get_token_fresh yields the rotated one. The server
// accepts only the fresh token, so the retry can only succeed if the cache was
// expired and the script re-invoked.
func TestInvokeExpiresCredentialScriptTokenOn401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Header.Get("Authorization") == "Bearer fresh-token" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[{"id":"1"}]}`))
			return
		}
		// The stale token is rejected, as the grid would reject an expired one.
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":{"text":"unauthorized"}}`))
	}))
	t.Cleanup(srv.Close)

	// A poller whose credentials come from a script that returns an authToken.
	poller := &conf.Poller{
		Name: "test",
		Addr: "127.0.0.1",
		CredentialsScript: conf.CredentialsScript{
			Path: "testdata/get_token_stale",
		},
	}
	conf.Config.Defaults = nil

	creds := auth.NewCredentials(poller, slog.Default())

	// Prime the cache with the stale token, as a running poller would have.
	pollerAuth, err := creds.GetPollerAuth()
	assert.Nil(t, err)
	assert.True(t, pollerAuth.HasCredentialScript)
	assert.Equal(t, pollerAuth.AuthToken, "stale-token")

	req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	assert.Nil(t, err)

	c := &Client{
		client:   srv.Client(),
		request:  req,
		buffer:   new(bytes.Buffer),
		Logger:   slog.Default(),
		baseURL:  srv.URL + "/",
		token:    "stale-token",
		Timeout:  10 * time.Second,
		APIPath:  "/api/v3",
		auth:     creds,
		Metadata: &collector.Metadata{},
	}

	// The script has since rotated its token. Without Expire, the cached
	// stale-token is handed back and the retry 401s again.
	poller.CredentialsScript.Path = "testdata/get_token_fresh"

	body, err := c.getRest(srv.URL + "/api/v3/grid/config/product-version")

	assert.Nil(t, err)
	assert.True(t, len(body) > 0)
	// The credential cache was invalidated and the script re-run.
	assert.Equal(t, c.token, "fresh-token")
}

// TestInvokeReusesCachedTokenWithoutA401 is the control for the test above: an
// ordinary successful call must not expire the credential cache, or every poll
// would re-invoke the customer's credential script.
func TestInvokeReusesCachedTokenWithoutA401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"id":"1"}]}`))
	}))
	t.Cleanup(srv.Close)

	poller := &conf.Poller{
		Name: "test",
		Addr: "127.0.0.1",
		CredentialsScript: conf.CredentialsScript{
			Path: "testdata/get_token_stale",
		},
	}
	conf.Config.Defaults = nil

	creds := auth.NewCredentials(poller, slog.Default())
	_, err := creds.GetPollerAuth()
	assert.Nil(t, err)

	req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	assert.Nil(t, err)

	c := &Client{
		client:   srv.Client(),
		request:  req,
		buffer:   new(bytes.Buffer),
		Logger:   slog.Default(),
		baseURL:  srv.URL + "/",
		token:    "stale-token",
		Timeout:  10 * time.Second,
		APIPath:  "/api/v3",
		auth:     creds,
		Metadata: &collector.Metadata{},
	}

	// Point the script somewhere else. A successful call must not notice.
	poller.CredentialsScript.Path = "testdata/get_token_fresh"

	_, err = c.getRest(srv.URL + "/api/v3/grid/config/product-version")
	assert.Nil(t, err)

	// Still the cached token: no 401, so no expiry and no script re-run.
	assert.Equal(t, c.token, "stale-token")
	got, err := creds.GetPollerAuth()
	assert.Nil(t, err)
	assert.Equal(t, got.AuthToken, "stale-token")
}

// TestDummyClientCredentialsAreUsable pins that NewDummyClient's Credentials
// can actually be used. It previously held a zero-value &auth.Credentials{},
// whose nil authMu mutex panicked the moment anything reached GetPollerAuth --
// which the 401 retry in invoke does -- so any future 401 test built on this
// helper would have panicked rather than exercising the retry.
func TestDummyClientCredentialsAreUsable(t *testing.T) {
	c := NewDummyClient()

	pollerAuth, err := c.auth.GetPollerAuth()

	assert.Nil(t, err)
	assert.Equal(t, pollerAuth.Username, "admin")
	// Expire goes through the same mutex, and is called on every 401.
	c.auth.Expire()
}

// TestDummyClientWithBaseURLOverridesURL pins the one thing the httptest
// variant changes, so it cannot silently stop pointing at the test server.
func TestDummyClientWithBaseURLOverridesURL(t *testing.T) {
	c := NewDummyClientWithBaseURL("http://127.0.0.1:9999")

	assert.Equal(t, c.baseURL, "http://127.0.0.1:9999")
	// Everything else still comes from NewDummyClient.
	assert.Equal(t, c.APIPath, "/api/v1")
}
