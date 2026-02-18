/*
 * Copyright NetApp Inc, 2025 All rights reserved
 */

package rest

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

const (
	DefaultTimeout        = "30s"
	DefaultAPIPath        = "/devmgr/v2"
	DefaultESeriesVersion = "11.80.0"
)

// clientPool stores singleton clients per poller
var (
	clientPool   = sync.Map{} // key: pollerName, value: *Client
	clientPoolMu sync.Mutex
	bundleRe     = regexp.MustCompile(`^(\d+)\.(\d+)(?:\.(\d+))?`)
)

type Client struct {
	client   *http.Client
	Logger   *slog.Logger
	baseURL  string
	APIPath  string
	Timeout  time.Duration
	auth     *auth.Credentials
	Metadata *collector.Metadata
	remote   conf.Remote
	cache    *responseCache
}

type responseCache struct {
	data      []gjson.Result
	expiresAt time.Time
	mu        sync.Mutex // Simple mutex for cache operations
	hits      uint64
	misses    uint64
	ttl       time.Duration
}

type CacheConfig struct {
	Name string
	TTL  time.Duration
}

func getOrCreateClient(poller *conf.Poller, timeout time.Duration, credentials *auth.Credentials, cacheName string) (*Client, error) {
	key := poller.Name + ":" + cacheName

	clientPoolMu.Lock()
	defer clientPoolMu.Unlock()

	if existing, ok := clientPool.Load(key); ok {
		return existing.(*Client), nil
	}

	client, err := newClient(poller, timeout, credentials)
	if err != nil {
		return nil, err
	}

	clientPool.Store(key, client)
	client.Logger.Info("created pooled ESeries client",
		slog.String("poller", poller.Name),
		slog.String("cache", cacheName))
	return client, nil
}

// New creates a new E-Series client. If cacheName is provided, returns a pooled singleton client.
func New(poller *conf.Poller, timeout time.Duration, credentials *auth.Credentials, cacheName string) (*Client, error) {
	if cacheName != "" {
		return getOrCreateClient(poller, timeout, credentials, cacheName)
	}
	return newClient(poller, timeout, credentials)
}

// newClient is the internal constructor
func newClient(poller *conf.Poller, timeout time.Duration, credentials *auth.Credentials) (*Client, error) {
	var (
		client     Client
		httpclient *http.Client
		transport  http.RoundTripper
		addr       string
		url        string
		err        error
	)

	client = Client{
		auth:     credentials,
		APIPath:  DefaultAPIPath,
		Metadata: &collector.Metadata{},
	}
	client.Logger = slog.Default().With(slog.String("ESeries", "Client"))

	if addr = poller.Addr; addr == "" {
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	url = "https://" + addr

	client.baseURL = url
	client.Timeout = timeout

	transport, err = credentials.Transport(nil, poller)
	if err != nil {
		return nil, err
	}

	httpclient = &http.Client{Transport: transport, Timeout: timeout}
	client.client = httpclient

	return &client, nil
}

// GetStorageSystems retrieves all storage systems from the web services proxy
func (c *Client) GetStorageSystems() ([]gjson.Result, error) {
	endpoint := c.APIPath + "/storage-systems"
	return c.get(endpoint)
}

// Fetch makes a REST GET request with optional caching support
func (c *Client) Fetch(fullPath string, cacheConfig *CacheConfig, headers ...map[string]string) ([]gjson.Result, error) {
	if cacheConfig == nil {
		return c.get(fullPath, headers...)
	}

	// Initialize cache if needed
	if c.cache == nil {
		c.cache = &responseCache{}
	}

	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()

	if c.cache.ttl == 0 {
		c.cache.ttl = cacheConfig.TTL
		c.Logger.Debug("cache TTL initialized",
			slog.String("ttl", c.cache.ttl.String()))
	} else if cacheConfig.TTL < c.cache.ttl {
		c.Logger.Info("cache TTL updated to minimum",
			slog.String("old_ttl", c.cache.ttl.String()),
			slog.String("new_ttl", cacheConfig.TTL.String()))
		c.cache.ttl = cacheConfig.TTL

		// Recalculate expiration time based on new TTL if cache exists
		if c.cache.data != nil && !c.cache.expiresAt.IsZero() {
			newExpiresAt := time.Now().Add(c.cache.ttl)
			if newExpiresAt.Before(c.cache.expiresAt) {
				c.cache.expiresAt = newExpiresAt
			}
		}
	}

	// Check if cache is still valid
	if c.cache.data != nil && time.Now().Before(c.cache.expiresAt) {
		c.cache.hits++
		c.Logger.Debug("cache hit",
			slog.String("cache", cacheConfig.Name),
			slog.String("endpoint", fullPath),
			slog.Uint64("hits", c.cache.hits))
		return c.cache.data, nil
	}

	// Cache miss - fetch fresh data
	data, err := c.get(fullPath, headers...)
	if err != nil {
		return nil, err
	}

	// Populate cache with minimum TTL
	c.cache.data = data
	c.cache.expiresAt = time.Now().Add(c.cache.ttl)
	c.cache.misses++

	c.Logger.Debug("cache miss - fetched fresh data",
		slog.String("cache", cacheConfig.Name),
		slog.String("endpoint", fullPath),
		slog.Uint64("misses", c.cache.misses),
		slog.String("ttl", c.cache.ttl.String()),
		slog.Int("records", len(data)))

	return data, nil
}

func (c *Client) get(endpoint string, headers ...map[string]string) ([]gjson.Result, error) {
	var (
		err     error
		results []gjson.Result
	)

	doInvoke := func() ([]gjson.Result, error) {
		var (
			req       *http.Request
			res       *http.Response
			innerBody []byte
			innerErr  error
			innerRes  []gjson.Result
		)

		url := c.baseURL + endpoint

		if req, innerErr = http.NewRequest(http.MethodGet, url, http.NoBody); innerErr != nil {
			return nil, innerErr
		}

		for _, hs := range headers {
			for k, v := range hs {
				req.Header.Set(k, v)
			}
		}

		pollerAuth, innerErr := c.auth.GetPollerAuth()
		if innerErr != nil {
			c.Logger.Error("failed to get auth credentials", slog.String("url", url), slogx.Err(innerErr))
			return nil, innerErr
		}
		req.SetBasicAuth(pollerAuth.Username, pollerAuth.Password)

		if res, innerErr = c.client.Do(req); innerErr != nil {
			c.Logger.Error("request failed", slog.String("url", url), slog.String("err", innerErr.Error()))
			return nil, innerErr
		}
		defer res.Body.Close()

		if innerBody, innerErr = io.ReadAll(res.Body); innerErr != nil {
			return nil, innerErr
		}

		// Track metadata
		c.Metadata.NumCalls++
		c.Metadata.BytesRx += uint64(len(innerBody))

		if res.StatusCode == http.StatusUnauthorized {
			c.Logger.Warn(
				"Authentication failed",
				slog.Int("status", res.StatusCode),
				slog.String("url", url),
			)
			return nil, errs.NewRest().
				StatusCode(res.StatusCode).
				Error(errs.ErrAuthFailed).
				Message(res.Status).
				API(endpoint).
				Build()
		}

		if res.StatusCode != http.StatusOK {
			c.Logger.Error(
				"API request failed",
				slog.Int("status", res.StatusCode),
				slog.String("url", url),
				slog.String("body", string(innerBody)),
			)
			return nil, errs.NewRest().
				StatusCode(res.StatusCode).
				API(endpoint).
				Build()
		}

		// Check if response is an array or object
		parsed := gjson.ParseBytes(innerBody)
		switch {
		case parsed.IsArray():
			innerRes = parsed.Array()
		case parsed.IsObject():
			// Single object response - wrap in array
			innerRes = []gjson.Result{parsed}
		default:
			return nil, fmt.Errorf("unexpected response format from %s", url)
		}

		return innerRes, nil
	}

	results, err = doInvoke()

	if err != nil {
		if re, ok := errors.AsType[*errs.RestError](err); ok {
			if errors.Is(re, errs.ErrAuthFailed) {
				pollerAuth, err2 := c.auth.GetPollerAuth()
				if err2 != nil {
					return nil, err2
				}
				// If this is an auth failure and the client is using a credential script,
				// expire the current credentials, call the script again, update the credentials,
				// and try again
				if pollerAuth.HasCredentialScript {
					c.Logger.Debug("Expiring cached credential script credentials after 401 response")
					c.auth.Expire()
					c.Logger.Debug("Retrying request with refreshed credentials from script")
					results, err = doInvoke()
					return results, err
				}
			}
		}
	}

	return results, err
}

func (c *Client) Init(retries int, remote conf.Remote) error {
	var (
		err     error
		systems []gjson.Result
	)

	c.remote = remote

	if !remote.IsZero() {
		return nil
	}

	for range retries {
		systems, err = c.GetStorageSystems()
		if err != nil {
			continue
		}

		if len(systems) > 0 {
			firstSystem := systems[0]
			systemID := firstSystem.Get("id").ClonedString()

			bundleVersion, err := c.getBundleDisplayVersion(systemID)
			if err != nil {
				c.Logger.Warn(
					"Failed to get bundleDisplay version, using default version",
					slogx.Err(err),
					slog.String("arrayID", systemID),
					slog.String("bundleVersion", bundleVersion),
					slog.String("defaultVersion", DefaultESeriesVersion),
				)
				c.remote.Version = DefaultESeriesVersion
			} else {
				c.remote.Version = bundleVersion
				c.Logger.Debug(
					"Using bundleDisplay version",
					slog.String("version", bundleVersion),
				)
			}

			c.remote.Name = firstSystem.Get("name").ClonedString()
			c.remote.Model = firstSystem.Get("model").ClonedString()
			c.remote.Serial = firstSystem.Get("serialNumber").ClonedString()
		}
		return nil
	}

	return err
}

// Remote returns the remote information
func (c *Client) Remote() conf.Remote {
	return c.remote
}

// Returns normalized version like "11.70.4" from bundleDisplay values like "11.70.4R1"
func (c *Client) getBundleDisplayVersion(systemID string) (string, error) {
	endpoint := c.APIPath + "/firmware/embedded-firmware/" + systemID + "/versions"
	results, err := c.get(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to get firmware versions: %w", err)
	}

	for _, result := range results {
		codeVersions := result.Get("codeVersions")
		if !codeVersions.Exists() {
			continue
		}

		for _, version := range codeVersions.Array() {
			if version.Get("codeModule").ClonedString() == "bundleDisplay" {
				versionString := version.Get("versionString").ClonedString()
				if versionString != "" {
					normalized := c.normalizeBundleVersion(versionString)
					if normalized == "" {
						return "", errors.New("failed to parse bundleDisplay version")
					}
					return normalized, nil
				}
			}
		}
	}

	return "", errors.New("bundleDisplay not found in firmware versions")
}

// normalizeBundleVersion converts bundleDisplay format to template-matchable version
// Examples:
//
//	"11.70.4R1" -> "11.70.4"
//	"11.90.R4"  -> "11.90.0"
//	"12.00GA"   -> "12.00.0"
//	"11.30"     -> "11.30.0"
//
// Returns empty string if parsing fails
func (c *Client) normalizeBundleVersion(bundleDisplay string) string {
	matches := bundleRe.FindStringSubmatch(bundleDisplay)

	if len(matches) == 0 {
		// No match, return empty string to trigger default version fallback
		return ""
	}

	major := matches[1]
	minor := matches[2]
	patch := "0"
	if matches[3] != "" {
		patch = matches[3]
	}

	return fmt.Sprintf("%s.%s.%s", major, minor, patch)
}
