// Copyright NetApp Inc, 2021 All rights reserved

// Package zapi provides type Client for connecting to a C-dot or 7-mode
// ONTAP cluster and sending API requests using the ZAPI protocol.
package zapi

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/requests"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultAPIVersion = "1.3"
	DefaultTimeout    = "30s"
)

type Client struct {
	client     *http.Client
	request    *http.Request
	buffer     *bytes.Buffer
	apiVersion string
	vfiler     string
	Logger     *slog.Logger
	logZapi    bool // used to log ZAPI request/response
	auth       *auth.Credentials
	Metadata   *collector.Metadata
	remote     conf.Remote
}

type Response struct {
	Result *node.Node
	Tag    string
	Rd     time.Duration
	Pd     time.Duration
}

func New(poller *conf.Poller, c *auth.Credentials) (*Client, error) {
	var (
		client     Client
		httpclient *http.Client
		request    *http.Request
		transport  http.RoundTripper
		timeout    time.Duration
		url, addr  string
		err        error
	)

	client = Client{
		auth:     c,
		Metadata: &collector.Metadata{},
	}
	client.Logger = slog.Default().With(slog.String("Zapi", "Client"))

	// check required & optional parameters
	if client.apiVersion = poller.APIVersion; client.apiVersion == "" {
		client.apiVersion = DefaultAPIVersion
		client.Logger.Debug("using default API version", slog.String("version", DefaultAPIVersion))
	}

	if client.vfiler = poller.APIVfiler; client.vfiler != "" {
		client.Logger.Debug("using vfiler tunneling", slog.String("vfiler", client.vfiler))
	}

	if addr = poller.Addr; addr == "" {
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	if poller.IsKfs {
		url = "https://" + addr + ":8443/servlets/netapp.servlets.admin.XMLrequest_filer"
	} else {
		url = "https://" + addr + "/servlets/netapp.servlets.admin.XMLrequest_filer"
	}

	if poller.LogSet != nil {
		for _, name := range *poller.LogSet {
			if name == "Zapi" || name == "ZapiPerf" {
				client.logZapi = true
				break
			}
		}
	}

	// create a request object that will be used for later requests
	if request, err = requests.New("POST", url, nil); err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "text/xml")
	request.Header.Set("Charset", "utf-8")

	transport, err = c.Transport(request, poller)
	if err != nil {
		return nil, err
	}

	client.request = request

	// initialize http client
	httpclient = &http.Client{Transport: transport, Timeout: timeout}

	client.client = httpclient
	client.SetTimeout(poller.ClientTimeout)
	// ensure that we can change body dynamically
	request.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(client.buffer.Bytes())
		return io.NopCloser(r), nil
	}

	return &client, nil
}

// parseClientTimeout converts clientTimeout to a duration
// two formats are converted:
// 1. a normal Go duration. e.g., 123m -> 123m
// 2. sequences of numbers are converted to that many seconds. e.g., 123 -> 123s
// If both conversions fail, return the defaultTimeout and an error
func parseClientTimeout(clientTimeout string) (time.Duration, error) {
	// Assume clientTimeout is a normal Go duration
	duration, err := time.ParseDuration(clientTimeout)
	if err == nil { // is a normal Go duration
		return duration, nil
	}
	digits, err2 := strconv.Atoi(clientTimeout)
	if err2 != nil {
		defaultDuration, _ := time.ParseDuration(DefaultTimeout)
		return defaultDuration, err2
	}
	return time.Duration(digits) * time.Second, nil
}

// Init connects to the cluster and retrieves system info
// it will give up after retries
func (c *Client) Init(retries int, remote conf.Remote) error {
	var err error

	c.remote = remote

	if !remote.IsZero() {
		return nil
	}

	for range retries {
		if err = c.getSystem(); err == nil {
			break
		}
	}
	return err
}

// Name returns the name of the Cluster
func (c *Client) Name() string {
	return c.remote.Name
}

// IsClustered returns true if ONTAP is clustered or false if it's a 7-mode system
func (c *Client) IsClustered() bool {
	return c.remote.IsClustered
}

// Version returns version of the ONTAP server (generation, major and minor)
func (c *Client) Version() string {
	return c.remote.Version
}

// Release returns string with long release info of the ONTAP system
func (c *Client) Release() string {
	return c.remote.Release
}

// Serial returns the serial number of the ONTAP system
func (c *Client) Serial() string {
	return c.remote.Serial
}

// ClusterUUID returns the cluster UUID of a c-mode system and system-id for 7-mode
func (c *Client) ClusterUUID() string {
	return c.remote.UUID
}

// Info returns a string with details about the ONTAP system identity
func (c *Client) Info() string {
	var model, version string
	if c.IsClustered() {
		model = "CDOT"
	} else {
		model = "7MODE"
	}
	version = fmt.Sprintf("(%s version %s)", model, c.remote.Version)
	return fmt.Sprintf("%s %s (serial %s) (%s)", c.Name(), version, c.Serial(), c.Release())
}

// BuildRequestString builds an API request from the request.
// request is usually the API name (e.g. "system-get-info") without any attributes
func (c *Client) BuildRequestString(request string) error {
	return c.buildRequestString(request, false)
}

// BuildRequest builds an API request from the node query
// root element of the request is usually the API name (e.g. "volume-get-iter") and
// its children are the attributes requested
func (c *Client) BuildRequest(request *node.Node) error {
	return c.buildRequest(request, false)
}

func (c *Client) buildRequestString(request string, forceCluster bool) error {
	return c.buildRequest(node.NewXMLS(request), forceCluster)
}

// build API request from the given node object.
func (c *Client) buildRequest(query *node.Node, forceCluster bool) error {
	var (
		request *node.Node
		buffer  *bytes.Buffer
		data    []byte
		err     error
	)

	request = node.NewXMLS("netapp")
	//goland:noinspection HttpUrlsUsage
	request.NewAttrS("xmlns", "http://www.netapp.com/filer/admin")
	request.NewAttrS("version", c.apiVersion)
	// optionally use fviler-tunneling, this option is never used in Harvest
	if !forceCluster && c.vfiler != "" {
		request.NewAttrS("vfiler", c.vfiler)
	}
	request.AddChild(query)

	if data, err = tree.DumpXML(request); err != nil {
		return err
	}

	buffer = bytes.NewBuffer(data)
	c.buffer = buffer
	c.request.Body = io.NopCloser(buffer)
	c.request.ContentLength = int64(buffer.Len())
	return nil
}

// invokeZapi will issue API requests with batching
// The method bails on the first error
func (c *Client) invokeZapi(request *node.Node, handle func([]*node.Node) error) error {
	tag := "initial"

	for {
		var (
			result   *node.Node
			response []*node.Node
			err      error
		)

		responseData, err := c.InvokeBatchRequest(request, tag, "")
		if err != nil {
			return err
		}
		result = responseData.Result
		tag = responseData.Tag

		if result == nil {
			break
		}

		// for 7mode, the output will be the zapi response since 7mode does not support pagination
		if !c.IsClustered() {
			response = append(response, result)
			// 7mode does not support pagination. set the tag an empty string to break the for loop
			tag = ""
		} else if x := result.GetChildS("attributes-list"); x != nil {
			response = x.GetChildren()
		} else if y := result.GetChildS("attributes"); y != nil {
			// Check for non-list response
			response = y.GetChildren()
		}

		if len(response) == 0 {
			break
		}
		err = handle(response)
		if err != nil {
			return err
		}
	}

	return nil
}

// InvokeZapiCall will issue API requests with batching
func (c *Client) InvokeZapiCall(request *node.Node) ([]*node.Node, error) {
	var output []*node.Node
	err := c.invokeZapi(request, func(response []*node.Node) error {
		output = append(output, response...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

// InvokeZapiCallStream will issue API requests with batching and process results in streaming fashion
// The processBatch function is called for each batch of records as they are received
func (c *Client) InvokeZapiCallStream(request *node.Node, processBatch func([]*node.Node) error) error {
	return c.invokeZapi(request, processBatch)
}

// Invoke is used for two purposes
// If testFilePath is non-empty -> Used only from unit test
// Else -> will issue the API request and return server response
// this method should only be called after building the request
func (c *Client) Invoke(testFilePath string) (*node.Node, error) {
	if testFilePath != "" {
		testData, err := tree.ImportXML(testFilePath)
		if err != nil {
			return nil, err
		}
		return testData, nil
	}
	result, _, _, err := c.invokeWithAuthRetry(false)
	return result, err
}

// InvokeBatchRequest is used for two purposes
// If testFilePath is non-empty -> Used only from unit test
// Else -> will issue API requests in series, once there
// are no more instances returned by the server, returned results will be nil
// Use the returned tag for subsequent calls to this method
func (c *Client) InvokeBatchRequest(request *node.Node, tag string, testFilePath string, headers ...map[string]string) (Response, error) {
	if testFilePath != "" && tag != "" {
		testData, err := tree.ImportXML(testFilePath)
		if err != nil {
			return Response{}, err
		}
		return Response{Result: testData, Tag: "", Rd: time.Second, Pd: time.Second}, nil
	}
	// wasteful of course, need to rewrite later @TODO
	results, tag, rd, pd, err := c.InvokeBatchWithTimers(request, tag, headers...)
	return Response{Result: results, Tag: tag, Rd: rd, Pd: pd}, err
}

// InvokeBatchWithTimers does the same as InvokeBatchRequest, but it also
// returns API time and XML parse time
func (c *Client) InvokeBatchWithTimers(request *node.Node, tag string, headers ...map[string]string) (*node.Node, string, time.Duration, time.Duration, error) {

	var (
		results *node.Node
		nextTag string
		rd, pd  time.Duration // response time, parse time
		err     error
	)

	if tag == "" {
		return nil, "", rd, pd, nil
	}

	if tag != "initial" {
		request.SetChildContentS("tag", tag)
	}

	if err := c.BuildRequest(request); err != nil {
		return nil, "", rd, pd, err
	}

	if results, rd, pd, err = c.invokeWithAuthRetry(true, headers...); err != nil {
		return nil, "", rd, pd, err
	}

	// avoid ZAPI bug, see:
	// https://community.netapp.com/t5/Software-Development-Kit-SDK-and-API-Discussions/Ontap-SDK-volume-get-iter-ZAPI-returns-erroneous-next-tag/m-p/153957/highlight/true#M2995
	if nextTag = results.GetChildContentS("next-tag"); nextTag == tag {
		nextTag = ""
	}

	return results, nextTag, rd, pd, nil
}

// InvokeRequestString builds a request from request and invokes it
func (c *Client) InvokeRequestString(request string) (*node.Node, error) {
	if err := c.BuildRequestString(request); err != nil {
		return nil, err
	}
	return c.Invoke("")
}

// InvokeRequest builds a request from request and invokes it
func (c *Client) InvokeRequest(request *node.Node) (*node.Node, error) {

	var err error

	if err = c.BuildRequest(request); err == nil {
		return c.Invoke("")
	}
	return nil, err
}

// InvokeWithTimers is used for two purposes
// If testFilePath is non-empty -> Used only from unit test
// Else -> invokes the request and returns parsed XML response and timers:
// API wait time and XML parse time.
// This method should only be called after building the request
func (c *Client) InvokeWithTimers(testFilePath string, headers ...map[string]string) (*node.Node, time.Duration, time.Duration, error) {
	if testFilePath != "" {
		testData, err := tree.ImportXML(testFilePath)
		if err != nil {
			return nil, 0, 0, err
		}
		return testData, 0, 0, nil
	}
	return c.invokeWithAuthRetry(true, headers...)
}

func (c *Client) invokeWithAuthRetry(withTimers bool, headers ...map[string]string) (*node.Node, time.Duration, time.Duration, error) {
	var buffer bytes.Buffer
	pollerAuth, err := c.auth.GetPollerAuth()
	if err != nil {
		return nil, 0, 0, err
	}
	if pollerAuth.HasCredentialScript {
		// Save the buffer in case it needs to be replayed after an auth failure
		// This is required because Go clears the buffer when making a POST request
		buffer = *c.buffer
	}

	resp, t1, t2, err := c.invoke(withTimers, headers...)

	if err != nil {
		if he, ok := errors.AsType[errs.HarvestError](err); ok {
			// If this is an auth failure and the client is using a credential script,
			// expire the current credentials, call the script again, update the client's password,
			// and try again
			if errors.Is(he, errs.ErrAuthFailed) && pollerAuth.HasCredentialScript {
				c.auth.Expire()
				pollerAuth2, err2 := c.auth.GetPollerAuth()
				if err2 != nil {
					return nil, 0, 0, err2
				}
				c.request.SetBasicAuth(pollerAuth2.Username, pollerAuth2.Password)
				c.request.Body = io.NopCloser(&buffer)
				c.request.ContentLength = int64(buffer.Len())
				result2, s1, s2, err3 := c.invoke(withTimers)
				u1 := t1.Nanoseconds() + s1.Nanoseconds()
				u2 := t2.Nanoseconds() + s2.Nanoseconds()
				return result2, time.Duration(u1) * time.Nanosecond, time.Duration(u2) * time.Nanosecond, err3
			}
		}
	}
	return resp, t1, t2, err
}

// invokes the request that has been built with one of the BuildRequest* methods
func (c *Client) invoke(withTimers bool, headers ...map[string]string) (*node.Node, time.Duration, time.Duration, error) {

	var (
		root, result      *node.Node
		response          *http.Response
		start             time.Time
		responseT, parseT time.Duration
		body              []byte
		status            string
		reason            string
		errNum            string
		found             bool
		err               error
	)

	//goland:noinspection GoUnhandledErrorResult
	defer c.request.Body.Close()
	defer c.buffer.Reset()

	// issue request to server
	if withTimers {
		start = time.Now()
	}

	// ZAPI request needs to be saved before calling client.Do because client.Do will zero out the buffer
	zapiReq := ""
	if c.logZapi {
		zapiReq = c.buffer.String()
	}

	for _, hs := range headers {
		for k, v := range hs {
			c.request.Header.Set(k, v)
		}
	}

	if response, err = c.client.Do(c.request); err != nil {
		return result, responseT, parseT, errs.New(errs.ErrConnection, err.Error())
	}
	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()
	if withTimers {
		responseT = time.Since(start)
	}

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusUnauthorized {
			return result, responseT, parseT, errs.New(errs.ErrAuthFailed, response.Status, errs.WithStatus(response.StatusCode))
		}
		return result, responseT, parseT, errs.New(errs.ErrAPIResponse, response.Status, errs.WithStatus(response.StatusCode))
	}

	// read response body
	if body, err = io.ReadAll(response.Body); err != nil {
		return result, responseT, parseT, err
	}
	defer c.printRequestAndResponse(zapiReq, body)
	if withTimers {
		responseT = time.Since(start)
	}

	// parse xml
	if withTimers {
		start = time.Now()
	}
	if root, err = tree.LoadXML(body); err != nil {
		return result, responseT, parseT, err
	}
	if withTimers {
		parseT = time.Since(start)
	}

	c.Metadata.BytesRx += uint64(len(body))
	c.Metadata.NumCalls++

	// check if the request was successful
	if result = root.GetChildS("results"); result == nil {
		return nil, responseT, parseT, errs.New(errs.ErrAPIResponse, "missing \"results\"")
	}

	if status, found = result.GetAttrValueS("status"); !found {
		return result, responseT, parseT, errs.New(errs.ErrAPIResponse, "missing status attribute")
	}

	if status != "passed" {
		reason, _ = result.GetAttrValueS("reason")
		if reason == "" {
			reason = "no reason"
		}
		errNum, _ = result.GetAttrValueS("errno")
		if errNum == errs.ZAPIPermissionDenied {
			err = errs.New(errs.ErrPermissionDenied, reason, errs.WithErrorNum(errNum))
		} else {
			err = errs.New(errs.ErrAPIRequestRejected, reason, errs.WithErrorNum(errNum))
		}
		return result, responseT, parseT, err
	}

	return result, responseT, parseT, nil
}

func (c *Client) TraceLogSet(collectorName string, config *node.Node) {
	// check for log sets and enable zapi request logging if collectorName is in the set
	if llogs := config.GetChildS("log"); llogs != nil {
		for _, log := range llogs.GetAllChildContentS() {
			if strings.EqualFold(log, collectorName) {
				c.logZapi = true
			}
		}
	}
}

func (c *Client) printRequestAndResponse(req string, response []byte) {
	if req != "" {
		c.Logger.Info("", slog.String("Request", req), slog.String("Response", string(response)))
	}
}

func (c *Client) SetTimeout(timeout string) {
	if c.client == nil {
		return
	}
	newTimeout, err := parseClientTimeout(timeout)
	if err == nil {
		c.Logger.Debug("Using timeout", slog.String("timeout", newTimeout.String()))
	} else {
		c.Logger.Debug("Using default timeout", slog.String("timeout", newTimeout.String()))
	}
	c.client.Timeout = newTimeout
}

func (c *Client) Remote() conf.Remote {
	return c.remote
}

// NewTestClient It's used for unit test only
func NewTestClient() *Client {
	return &Client{
		remote:   conf.Remote{Name: "testCluster", IsClustered: true},
		request:  &http.Request{},
		Metadata: &collector.Metadata{},
	}
}
