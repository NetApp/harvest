package rest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultTimeout    = "1m"
	DefaultAPIVersion = "3"
)

type Client struct {
	client   *http.Client
	request  *http.Request
	buffer   *bytes.Buffer
	Logger   *logging.Logger
	baseURL  string
	Cluster  Cluster
	username string
	token    string
	Timeout  time.Duration
	logRest  bool // used to log Rest request/response
	APIPath  string
}

type Cluster struct {
	Name    string
	Info    string
	UUID    string
	Version [3]int
}

func NewClient(pollerName string, clientTimeout string) (*Client, error) {
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
	if client, err = New(poller, timeout); err != nil {
		return nil, fmt.Errorf("uanble to create poller [%s]. err: %w", pollerName, err)
	}

	return client, err
}

func New(poller *conf.Poller, timeout time.Duration) (*Client, error) {
	var (
		client         Client
		httpclient     *http.Client
		transport      *http.Transport
		cert           tls.Certificate
		addr           string
		href           string
		useInsecureTLS bool
		err            error
	)

	client = Client{}
	client.Logger = logging.Get().SubLogger("StorageGrid", "Client")

	if addr = poller.Addr; addr == "" {
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	href = "https://" + addr + "/"

	client.baseURL = href
	client.Timeout = timeout

	// by default, enforce secure TLS, if not requested otherwise by user
	if x := poller.UseInsecureTLS; x != nil {
		useInsecureTLS = *poller.UseInsecureTLS
	} else {
		useInsecureTLS = false
	}

	// check if a credentials file is being used and if so, parse and use the values from it
	if poller.CredentialsFile != "" {
		err := conf.ReadCredentialsFile(poller.CredentialsFile, poller)
		if err != nil {
			client.Logger.Error().
				Err(err).
				Str("credPath", poller.CredentialsFile).
				Str("poller", poller.Name).
				Msg("Unable to read credentials file")
			return nil, err
		}
	}
	// set authentication method
	if poller.AuthStyle == "certificate_auth" {
		certPath := poller.SslCert
		keyPath := poller.SslKey
		if certPath == "" {
			return nil, errs.New(errs.ErrMissingParam, "ssl_cert")
		} else if keyPath == "" {
			return nil, errs.New(errs.ErrMissingParam, "ssl_key")
		} else if cert, err = tls.LoadX509KeyPair(certPath, keyPath); err != nil {
			return nil, err
		}

		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: useInsecureTLS}, //nolint:gosec
		}
	} else {
		username := poller.Username
		password := auth.Get().Password()
		client.username = username
		if username == "" {
			return nil, errs.New(errs.ErrMissingParam, "username")
		} else if password == "" {
			return nil, errs.New(errs.ErrMissingParam, "password")
		}

		transport = &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: useInsecureTLS}, //nolint:gosec
		}
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
		c.Logger.Info().
			Str("Request", c.request.URL.String()).
			Str("Response", res).
			Msg("")
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

	output := gjson.GetManyBytes(fetched, "data")
	data = output[0]
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
	output := gjson.GetManyBytes(fetched, "data")
	data := output[0]
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

	c.request, err = http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.request.Header.Set("accept", "application/json")
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
		var storageGridErr errs.StorageGridError
		if errors.As(err, &storageGridErr) {
			if storageGridErr.IsAuthErr() {
				err2 := c.fetchToken()
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
		defer func(Body io.ReadCloser) { _ = Body.Close() }(response.Body)
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

	if response.StatusCode != 200 {
		if body, err = io.ReadAll(response.Body); err == nil {
			return nil, errs.NewStorageGridErr(response.StatusCode, body)
		}
		return nil, errs.Rest(response.StatusCode, err.Error(), 0, "")
	}

	// read response body
	if body, err = io.ReadAll(response.Body); err != nil {
		return nil, err
	}
	defer c.printRequestAndResponse(body)

	if err != nil {
		return nil, err
	}
	return body, nil
}

// Init is responsible for determining the StorageGrid server version, API version, hostname, and systemId
func (c *Client) Init(retries int) error {
	var (
		err     error
		content []byte
		i       int
	)

	for i = 0; i < retries; i++ {
		// Determine which API versions are supported and then request
		// product version and cluster name - both of which are separate endpoints

		err = c.sniffAPIVersion(retries)
		if err != nil {
			continue
		}

		if content, err = c.GetGridRest("grid/config/product-version"); err != nil {
			continue
		}
		results := gjson.GetManyBytes(content, "data.productVersion")
		err = c.SetVersion(results[0].String())
		if err != nil {
			return err
		}

		if content, err = c.GetGridRest("grid/health/topology?depth=grid"); err != nil {
			continue
		}
		results = gjson.GetManyBytes(content, "data.name")
		c.Cluster.Name = strings.ReplaceAll(results[0].String(), " ", "_")

		if content, err = c.GetGridRest("grid/license"); err != nil {
			continue
		}
		results = gjson.GetManyBytes(content, "data.systemId")
		c.Cluster.UUID = results[0].String()
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
		c.Cluster.Version[0] = segments[0]
		c.Cluster.Version[1] = segments[1]
		c.Cluster.Version[2] = segments[2]
	} else {
		return fmt.Errorf("failed to parse version %s", v)
	}
	return nil
}

type authBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *Client) fetchToken() error {
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
	authB := authBody{
		Username: c.username,
		Password: auth.Get().Password(),
	}
	postBody, err := json.Marshal(authB)
	if err != nil {
		return err
	}

	req, err = http.NewRequest("POST", u, bytes.NewBuffer(postBody))
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/json")

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

	if response.StatusCode != 200 {
		return errs.NewStorageGridErr(response.StatusCode, body)
	}

	results := gjson.GetManyBytes(body, "data", "message.text")
	token := results[0]
	errorMsg := results[1]

	if token.Exists() {
		c.token = token.String()
		c.request.Header.Set("Authorization", "Bearer "+c.token)
	} else {
		return errs.New(errs.ErrAuthFailed, errorMsg.String())
	}
	return nil
}

func (c *Client) sniffAPIVersion(retries int) error {
	// This endpoint does not require auth and uses the /api/ endpoint instead of a versioned one

	var (
		apiVersion = DefaultAPIVersion
		u          string
		err        error
		i          int
	)

	u, err = url.JoinPath(c.baseURL, "/api/versions")
	if err != nil {
		return fmt.Errorf("failed to join getApiVersions %s err: %w", c.baseURL, err)
	}
	for i = 0; i < retries; i++ {
		result, err := c.getRest(u)
		if err != nil {
			continue
		}
		versionB := gjson.GetBytes(result, "data")
		if versionB.Exists() && versionB.IsArray() {
			versions := versionB.Array()
			if len(versions) > 0 {
				apiVersion = versions[len(versions)-1].String()
			}
		}
		c.APIPath = "/api/v" + apiVersion
		return nil
	}
	return err
}
