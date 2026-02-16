// Copyright NetApp Inc, 2021 All rights reserved

package rest

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"io"
	"log/slog"
	"net/http"
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
	Metadata *collector.Metadata
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
		Metadata: &collector.Metadata{},
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
		request, err = requests.EncodeURL(request)
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

// PostRest makes a REST POST request using the provided JSON body
// and returns the JSON response as a []byte.
func (c *Client) PostRest(endpoint string, body []byte, headers ...map[string]string) ([]byte, error) {
	endpoint = strings.TrimPrefix(endpoint, "/")
	u := c.baseURL + endpoint

	var err error
	c.request, err = requests.New("POST", u, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	c.request.Header.Set("Content-Type", "application/json")
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
		c.Logger.Debug("Using authToken from credential script for POST")
	} else if pollerAuth.Username != "" {
		c.request.SetBasicAuth(pollerAuth.Username, pollerAuth.Password)
	}

	c.buffer = bytes.NewBuffer(body)
	c.request.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(c.buffer.Bytes())
		return io.NopCloser(r), nil
	}

	result, err := c.invokeWithAuthRetry()
	c.Metadata.BytesRx += uint64(len(result))
	c.Metadata.NumCalls++

	return result, err
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

		// If c.buffer exists, defer its Reset
		if c.buffer != nil {
			defer c.buffer.Reset()
		}
		restReq := c.request.URL.String()
		api := requests.GetURLWithoutHost(c.request)

		// Send request to the server.
		response, innerErr = c.client.Do(c.request)
		if innerErr != nil {
			return nil, fmt.Errorf("connection error: %w", innerErr)
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

			errorMsg := gjson.GetBytes(innerBody, "error")

			if response.StatusCode == http.StatusForbidden {
				message := errorMsg.Get(Message).ClonedString()
				return nil, errs.NewRest().
					StatusCode(response.StatusCode).
					Error(errs.ErrPermissionDenied).
					Message(message).
					API(api).
					Build()
			}

			if errorMsg.Exists() {
				message := errorMsg.Get(Message).ClonedString()
				code := errorMsg.Get(Code).Int()
				target := errorMsg.Get(Target).ClonedString()
				outputMsg := gjson.GetBytes(innerBody, "output")

				fullMessage := message + " " + outputMsg.ClonedString()
				fullMessage = strings.TrimSpace(fullMessage)

				return nil, errs.NewRest().
					StatusCode(response.StatusCode).
					Message(fullMessage).
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

		// Print for logging if enabled.
		defer c.printRequestAndResponse(restReq, innerBody)

		return innerBody, nil
	}

	body, err = doInvoke()

	if err != nil {
		var re *errs.RestError
		if errors.As(err, &re) {
			// If this is an auth failure and the client is using a credential script,
			// expire the current credentials, call the script again, update the client's password,
			// and try again
			if errors.Is(re, errs.ErrAuthFailed) {
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
		c.remote = conf.NewRemote(results)

		return nil
	}

	return err
}

func (c *Client) Init(retries int, remote conf.Remote) error {
	c.remote = remote

	if !remote.IsZero() {
		return nil
	}

	return c.UpdateClusterInfo(retries)
}

func (c *Client) Remote() conf.Remote {
	return c.remote
}
