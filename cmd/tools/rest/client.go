// Copyright NetApp Inc, 2021 All rights reserved

package rest

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"slices"
	"strings"
	"time"
)

const (
	// DefaultTimeout should be > than ONTAP's default REST timeout, which is 15 seconds for GET requests
	DefaultTimeout = "30s"
	Message        = "message"
	Code           = "code"
	Target         = "target"
)

type Client struct {
	client   *http.Client
	request  *http.Request
	buffer   *bytes.Buffer
	Logger   *slog.Logger
	baseURL  string
	remote   conf.Remote
	token    string
	Timeout  time.Duration
	logRest  bool // used to log Rest request/response
	auth     *auth.Credentials
	Metadata *util.Metadata
}

func New(poller *conf.Poller, timeout time.Duration, credentials *auth.Credentials) (*Client, error) {
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
		Metadata: &util.Metadata{},
	}
	client.Logger = slog.Default().With(slog.String("REST", "Client"))

	if addr = poller.Addr; addr == "" {
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	if poller.IsKfs {
		url = "https://" + addr + ":8443/"
	} else {
		url = "https://" + addr + "/"
	}
	client.baseURL = url
	client.Timeout = timeout

	transport, err = credentials.Transport(nil, poller)
	if err != nil {
		return nil, err
	}

	httpclient = &http.Client{Transport: transport, Timeout: timeout}
	client.client = httpclient

	if poller.LogSet != nil {
		client.logRest = slices.Contains(*poller.LogSet, "Rest")
	}

	return &client, nil
}

func (c *Client) TraceLogSet(collectorName string, config *node.Node) {
	// check for log sets and enable Rest request logging if collectorName is in the set
	if llogs := config.GetChildS("log"); llogs != nil {
		c.logRest = slices.Contains(llogs.GetAllChildContentS(), collectorName)
	}
}

func (c *Client) printRequestAndResponse(req string, response []byte) {
	if c.logRest {
		c.Logger.Info(
			"",
			slog.String("Request", req),
			slog.String("Response", string(response)),
		)
	}
}

// GetPlainRest makes a REST request to the cluster and returns a json response as a []byte
func (c *Client) GetPlainRest(request string, encodeURL bool, headers ...map[string]string) ([]byte, error) {
	var err error
	if strings.Index(request, "/") == 0 {
		request = request[1:]
	}
	if encodeURL {
		request, err = util.EncodeURL(request)
		if err != nil {
			return nil, err
		}
	}

	u := c.baseURL + request
	c.request, err = requests.New("GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.request.Header.Set("Accept", "application/json")

	for _, hs := range headers {
		for k, v := range hs {
			c.request.Header.Set(k, v)
		}
	}

	pollerAuth, err := c.auth.GetPollerAuth()
	if err != nil {
		return nil, err
	}
	if pollerAuth.AuthToken != "" {
		c.request.Header.Set("Authorization", "Bearer "+pollerAuth.AuthToken)
		c.Logger.Debug("Using authToken from credential script")
	} else if pollerAuth.Username != "" {
		c.request.SetBasicAuth(pollerAuth.Username, pollerAuth.Password)
	}

	// ensure that we can change body dynamically
	c.request.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(c.buffer.Bytes())
		return io.NopCloser(r), nil
	}

	result, err := c.invokeWithAuthRetry()
	c.Metadata.BytesRx += uint64(len(result))
	c.Metadata.NumCalls++

	return result, err
}

// GetRest makes a REST request to the cluster and returns a json response as a []byte
func (c *Client) GetRest(request string, headers ...map[string]string) ([]byte, error) {
	return c.GetPlainRest(request, true, headers...)
}

func (c *Client) invokeWithAuthRetry() ([]byte, error) {
	var (
		body []byte
		err  error
	)

	doInvoke := func() ([]byte, error) {
		var (
			response  *http.Response
			innerBody []byte
			innerErr  error
		)

		if c.request.Body != nil {
			//goland:noinspection GoUnhandledErrorResult
			defer response.Body.Close()
		}
		if c.buffer != nil {
			defer c.buffer.Reset()
		}

		restReq := c.request.URL.String()
		api := util.GetURLWithoutHost(c.request)

		// send request to server
		if response, innerErr = c.client.Do(c.request); innerErr != nil {
			return nil, fmt.Errorf("connection error %w", innerErr)
		}
		//goland:noinspection GoUnhandledErrorResult
		defer response.Body.Close()
		innerBody, innerErr = io.ReadAll(response.Body)
		if innerErr != nil {
			return nil, errs.NewRest().
				StatusCode(response.StatusCode).
				Error(innerErr).
				API(api).
				Build()
		}

		if response.StatusCode != http.StatusOK {

			if response.StatusCode == http.StatusUnauthorized {
				return nil, errs.NewRest().
					StatusCode(response.StatusCode).
					Error(errs.ErrAuthFailed).
					Message(response.Status).
					API(api).
					Build()
			}

			result := gjson.GetBytes(innerBody, "error")

			if response.StatusCode == http.StatusForbidden {
				message := result.Get(Message).String()
				return nil, errs.NewRest().
					StatusCode(response.StatusCode).
					Error(errs.ErrPermissionDenied).
					Message(message).
					API(api).
					Build()
			}

			if result.Exists() {
				message := result.Get(Message).String()
				code := result.Get(Code).Int()
				target := result.Get(Target).String()
				return nil, errs.NewRest().
					StatusCode(response.StatusCode).
					Message(message).
					Code(code).
					Target(target).
					API(api).
					Build()
			}
			return nil, errs.NewRest().
				StatusCode(response.StatusCode).
				API(api).
				Build()
		}

		defer c.printRequestAndResponse(restReq, innerBody)

		return innerBody, nil
	}

	body, err = doInvoke()

	if err != nil {
		var he errs.HarvestError
		if errors.As(err, &he) {
			// If this is an auth failure and the client is using a credential script,
			// expire the current credentials, call the script again, update the client's password,
			// and try again
			if errors.Is(he, errs.ErrAuthFailed) {
				pollerAuth, err2 := c.auth.GetPollerAuth()
				if err2 != nil {
					return nil, err2
				}
				if pollerAuth.HasCredentialScript {
					c.auth.Expire()
					pollerAuth2, err2 := c.auth.GetPollerAuth()
					if err2 != nil {
						return nil, err2
					}
					// If the credential script returns an authToken, use it without re-fetching
					if pollerAuth.AuthToken != "" {
						c.token = pollerAuth.AuthToken
						c.request.Header.Set("Authorization", "Bearer "+c.token)
						c.Logger.Debug("Using authToken from credential script")
						return doInvoke()
					}
					c.request.SetBasicAuth(pollerAuth2.Username, pollerAuth2.Password)
					return doInvoke()
				}
			}
		}
	}
	return body, err
}

func downloadSwagger(poller *conf.Poller, path string, url string, verbose bool) (int64, error) {
	out, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("unable to create %s to save swagger.yaml", path)
	}
	defer func(out *os.File) { _ = out.Close() }(out)
	request, err := requests.New("GET", url, nil)
	if err != nil {
		return 0, err
	}

	timeout, _ := time.ParseDuration(DefaultTimeout)
	credentials := auth.NewCredentials(poller, slog.Default())
	transport, err := credentials.Transport(request, poller)
	if err != nil {
		return 0, err
	}
	httpclient := &http.Client{Transport: transport, Timeout: timeout}

	if verbose {
		requestOut, _ := httputil.DumpRequestOut(request, false)
		fmt.Printf("REQUEST: %s\n%s\n", url, requestOut)
	}
	response, err := httpclient.Do(request)
	if err != nil {
		return 0, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()

	if verbose {
		debugResp, _ := httputil.DumpResponse(response, false)
		fmt.Printf("RESPONSE: \n%s", debugResp)
	}
	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error making request. server response statusCode=[%d]", response.StatusCode)
	}
	n, err := io.Copy(out, response.Body)
	if err != nil {
		return 0, fmt.Errorf("error while downloading %s err=%w", url, err)
	}
	return n, nil
}

func (c *Client) UpdateClusterInfo(retries int) error {
	var (
		err     error
		content []byte
	)

	for range retries {
		href := NewHrefBuilder().
			APIPath("api/cluster").
			Fields([]string{"*"}).
			Build()
		content, err = c.GetRest(href)
		if err != nil {
			if errors.Is(err, errs.ErrPermissionDenied) {
				return err
			}
			continue
		}

		results := gjson.ParseBytes(content)
		c.remote.Name = results.Get("name").String()
		c.remote.UUID = results.Get("uuid").String()
		c.remote.Version =
			results.Get("version.generation").String() + "." +
				results.Get("version.major").String() + "." +
				results.Get("version.minor").String()
		return nil
	}
	return err
}

func (c *Client) Init(retries int) error {
	return c.UpdateClusterInfo(retries)
}

func (c *Client) Remote() conf.Remote {
	return c.remote
}
