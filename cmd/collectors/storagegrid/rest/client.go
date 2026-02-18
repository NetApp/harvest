package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/go-version"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultTimeout    = "1m"
	DefaultAPIVersion = "3"
)

var NewClientFunc = NewClient

type Client struct {
	client   *http.Client
	request  *http.Request
	buffer   *bytes.Buffer
	Logger   *slog.Logger
	baseURL  string
	Remote   conf.Remote
	token    string
	Timeout  time.Duration
	logRest  bool // used to log Rest request/response
	APIPath  string
	auth     *auth.Credentials
	Metadata *collector.Metadata
}

func NewClient(pollerName string, clientTimeout string, c *auth.Credentials) (*Client, error) {
	var (
		poller  *conf.Poller
		err     error
		client  *Client
		timeout time.Duration
	)

	if poller, err = conf.PollerNamed(pollerName); err != nil {
		return nil, fmt.Errorf("poller [%s] does not exist. err: %w", pollerName, err)
	}
	if poller.Addr == "" {
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	timeout, err = time.ParseDuration(clientTimeout)
	if err != nil {
		timeout, _ = time.ParseDuration(DefaultTimeout)
	}
	if client, err = New(poller, timeout, c); err != nil {
		return nil, fmt.Errorf("unable to create poller [%s]. err: %w", pollerName, err)
	}

	return client, err
}

func New(poller *conf.Poller, timeout time.Duration, c *auth.Credentials) (*Client, error) {
	var (
		client     Client
		httpclient *http.Client
		transport  http.RoundTripper
		addr       string
		href       string
		err        error
	)

	client = Client{
		auth:     c,
		Metadata: &collector.Metadata{},
	}
	client.Logger = slog.Default().With(slog.String("StorageGrid", "Client"))

	if addr = poller.Addr; addr == "" {
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	href = "https://" + addr + "/"

	client.baseURL = href
	client.Timeout = timeout

	transport, err = c.Transport(nil, poller)
	if err != nil {
		return nil, err
	}

	httpclient = &http.Client{Transport: transport, Timeout: timeout}
	client.client = httpclient

	return &client, nil
}

func (c *Client) TraceLogSet(collectorName string, config *node.Node) {
	// check for log sets and enable Rest request logging if collectorName is in the set
	if llogs := config.GetChildS("log"); llogs != nil {
		for _, log := range llogs.GetAllChildContentS() {
			if strings.EqualFold(log, collectorName) {
				c.logRest = true
			}
		}
	}
}

func (c *Client) printRequestAndResponse(response []byte) {
	if c.logRest {
		res := "<nil>"
		if response != nil {
			res = string(response)
		}
		c.Logger.Info(
			"",
			slog.String("Request", c.request.URL.String()),
			slog.String("Response", res),
		)
	}
}

// Fetch makes a REST request to the cluster and stores the parsed JSON in result
func (c *Client) Fetch(request string, result *[]gjson.Result) error {
	var (
		data    gjson.Result
		err     error
		fetched []byte
	)
	fetched, err = c.GetGridRest(request)
	if err != nil {
		return fmt.Errorf("error making request %w", err)
	}

	output := gjson.ParseBytes(fetched)
	data = output.Get("data")
	for _, r := range data.Array() {
		*result = append(*result, r.Array()...)
	}
	return nil
}

// GetGridRest makes a grid API request to the cluster and returns a json response as a []byte
// see also Fetch
func (c *Client) GetGridRest(request string) ([]byte, error) {
	u, err := url.JoinPath(c.baseURL, c.APIPath, request)
	if err != nil {
		return nil, fmt.Errorf("failed to join URL %s err: %w", request, err)
	}
	return c.getRest(u)
}

// GetMetricQuery makes a metrics API request to the cluster and fills the result argument
func (c *Client) GetMetricQuery(metric string, result *[]gjson.Result) error {
	u, err := url.JoinPath(c.baseURL, "/metrics/api/v1/query?query="+metric)
	if err != nil {
		return fmt.Errorf("failed to query metric %s err: %w", metric, err)
	}

	fetched, err := c.getRest(u)
	if err != nil {
		return err
	}
	output := gjson.ParseBytes(fetched)
	data := output.Get("data")
	for _, r := range data.Array() {
		*result = append(*result, r.Array()...)
	}
	return nil
}

// getRest makes a request to the cluster and returns a json response as a []byte
// see also Fetch
func (c *Client) getRest(request string) ([]byte, error) {
	u, err := url.QueryUnescape(request)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape %s err: %w", request, err)
	}

	c.request, err = requests.New("GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.request.Header.Set("Accept", "application/json")
	c.request.Header.Set("Authorization", "Bearer "+c.token)
	return c.invoke()
}

func (c *Client) invoke() ([]byte, error) {
	var (
		resp []byte
		err  error
	)
	resp, err = c.fetch()
	if err != nil {
		// check that the auth token has not expired
		if storageGridErr, ok := errors.AsType[errs.StorageGridError](err); ok {
			if storageGridErr.IsAuthErr() {
				// If using authToken from credential script, expire cache before retry
				// so fetchTokenWithAuthRetry gets fresh token instead of cached expired one
				pollerAuth, authErr := c.auth.GetPollerAuth()
				if authErr != nil {
					return nil, authErr
				}
				if pollerAuth.HasCredentialScript && pollerAuth.AuthToken != "" {
					c.Logger.Debug("Expiring cached credential script token after 401 response")
					c.auth.Expire()
				}

				err2 := c.fetchTokenWithAuthRetry()
				if err2 != nil {
					return nil, err2
				}
				return c.fetch()
			}
		}
		return nil, err
	}
	return resp, nil
}

func (c *Client) fetch() ([]byte, error) {
	var (
		response *http.Response
		body     []byte
		err      error
	)

	if c.request.Body != nil {
		//goland:noinspection GoUnhandledErrorResult
		c.request.Body.Close() //nolint:gosec
	}
	if c.buffer != nil {
		defer c.buffer.Reset()
	}

	// send request to server
	if response, err = c.client.Do(c.request); err != nil {
		return nil, fmt.Errorf("connection error %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if body, err = io.ReadAll(response.Body); err == nil {
			return nil, errs.NewStorageGridErr(response.StatusCode, body)
		}
		api := requests.GetURLWithoutHost(c.request)
		return nil, errs.NewRest().
			StatusCode(response.StatusCode).
			Error(err).
			API(api).
			Build()
	}

	// read response body
	if body, err = io.ReadAll(response.Body); err != nil {
		return nil, err
	}
	defer c.printRequestAndResponse(body)

	c.Metadata.BytesRx += uint64(len(body))
	c.Metadata.NumCalls++

	return body, nil
}

// Init is responsible for determining the StorageGrid server version, API version, hostname, and systemId
func (c *Client) Init(retries int, remote conf.Remote) error {
	var (
		err     error
		content []byte
	)

	c.Remote = remote

	if !remote.IsZero() && c.APIPath != "" {
		return nil
	}

	for range retries {
		// Determine which API versions are supported and then request
		// product version and cluster name - both of which are separate endpoints

		err = c.sniffAPIVersion(retries)
		if err != nil {
			continue
		}

		if content, err = c.GetGridRest("grid/config/product-version"); err != nil {
			continue
		}
		results := gjson.ParseBytes(content)
		err = c.SetVersion(results.Get("data.productVersion").ClonedString())
		if err != nil {
			return err
		}

		if content, err = c.GetGridRest("grid/health/topology?depth=grid"); err != nil {
			continue
		}

		results = gjson.ParseBytes(content)
		c.Remote.Name = results.Get("data.name").ClonedString()

		if content, err = c.GetGridRest("grid/license"); err != nil {
			continue
		}
		results = gjson.ParseBytes(content)
		c.Remote.UUID = results.Get("data.systemId").ClonedString()
		return nil
	}

	return err
}

func (c *Client) SetVersion(v string) error {
	newVersion, err := version.NewVersion(v)
	if err != nil {
		return fmt.Errorf("failed to parse version %s err: %w", v, err)
	}
	// e.g 11.6.0.3-20220802.2201.f58633a
	segments := newVersion.Segments()
	if len(segments) >= 3 {
		v0 := check(segments[0])
		v1 := check(segments[1])
		v2 := check(segments[2])
		c.Remote.Version = strconv.Itoa(v0) + "." + strconv.Itoa(v1) + "." + strconv.Itoa(v2)
	} else {
		return fmt.Errorf("failed to parse version %s", v)
	}
	return nil
}

func check(i int) int {
	if i > 0 && i <= math.MaxInt32 {
		return i
	}
	return 0
}

type authBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *Client) fetchTokenWithAuthRetry() error {
	fetchToken := func() error {
		var (
			err      error
			req      *http.Request
			response *http.Response
			body     []byte
		)
		u, err := url.JoinPath(c.baseURL, c.APIPath, "authorize")
		if err != nil {
			return fmt.Errorf("failed to create auth URL err: %w", err)
		}
		pollerAuth, err := c.auth.GetPollerAuth()
		if err != nil {
			return err
		}
		// If the credential script returns an authToken, use it without re-fetching
		if pollerAuth.AuthToken != "" {
			c.token = pollerAuth.AuthToken
			c.request.Header.Set("Authorization", "Bearer "+c.token)
			c.Logger.Debug("Using authToken from credential script")
			return nil
		}
		authB := authBody{
			Username: pollerAuth.Username,
			Password: pollerAuth.Password,
		}
		postBody, err := json.Marshal(authB)
		if err != nil {
			return err
		}

		req, err = requests.New("POST", u, bytes.NewBuffer(postBody))
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		// send request to server
		client := &http.Client{
			Transport: c.client.Transport,
			Timeout:   c.client.Timeout,
		}
		if response, err = client.Do(req); err != nil {
			return fmt.Errorf("connection error %w", err)
		}

		//goland:noinspection GoUnhandledErrorResult
		defer response.Body.Close()

		// read response body
		if body, err = io.ReadAll(response.Body); err != nil {
			return err
		}

		if response.StatusCode != http.StatusOK {
			return errs.NewStorageGridErr(response.StatusCode, body)
		}

		results := gjson.ParseBytes(body)
		token := results.Get("data")
		errorMsg := results.Get("message.text")

		if token.Exists() {
			c.token = token.ClonedString()
			c.request.Header.Set("Authorization", "Bearer "+c.token)
		} else {
			return errs.New(errs.ErrAuthFailed, errorMsg.ClonedString())
		}
		return nil
	}

	err := fetchToken()
	if err != nil {
		if storageGridErr, ok := errors.AsType[errs.StorageGridError](err); ok {
			// If this is an auth failure and the client is using a credential script,
			// expire the current credentials, call the script again, and try again
			if storageGridErr.IsAuthErr() {
				pollerAuth, err2 := c.auth.GetPollerAuth()
				if err2 != nil {
					return err2
				}
				if pollerAuth.HasCredentialScript {
					c.auth.Expire()
					return fetchToken()
				}
			}
		}
	}
	return err
}

func (c *Client) sniffAPIVersion(retries int) error {
	// This endpoint does not require auth and uses the /api/ endpoint instead of a versioned one

	var (
		apiVersion = DefaultAPIVersion
		u          string
		err        error
	)

	u, err = url.JoinPath(c.baseURL, "/api/versions")
	if err != nil {
		return fmt.Errorf("failed to join getApiVersions %s err: %w", c.baseURL, err)
	}
	for range retries {
		result, err := c.getRest(u)
		if err != nil {
			continue
		}
		versionB := gjson.GetBytes(result, "data")
		if versionB.Exists() && versionB.IsArray() {
			versions := versionB.Array()
			if len(versions) > 0 {
				apiVersion = versions[len(versions)-1].ClonedString()
			}
		}
		c.APIPath = "/api/v" + apiVersion
		return nil
	}
	return err
}
