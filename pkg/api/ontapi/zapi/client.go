// Copyright NetApp Inc, 2021 All rights reserved

// Package zapi provides type Client for connecting to a C-dot or 7-mode
// ONTAP cluster and sending API requests using the ZAPI protocol.
package zapi

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	DefaultAPIVersion = "1.3"
	DefaultTimeout    = "30s"
)

type Client struct {
	client     *http.Client
	request    *http.Request
	buffer     *bytes.Buffer
	system     *system
	apiVersion string
	vfiler     string
	Logger     *logging.Logger // logger used for logging
	logZapi    bool            // used to log ZAPI request/response
	auth       *auth.Credentials
}

func New(poller *conf.Poller, c *auth.Credentials) (*Client, error) {
	var (
		client         Client
		httpclient     *http.Client
		request        *http.Request
		transport      *http.Transport
		cert           tls.Certificate
		timeout        time.Duration
		url, addr      string
		useInsecureTLS bool
		err            error
	)

	client = Client{
		auth: c,
	}
	client.Logger = logging.Get().SubLogger("Zapi", "Client")

	// check required & optional parameters
	if client.apiVersion = poller.APIVersion; client.apiVersion == "" {
		client.apiVersion = DefaultAPIVersion
		client.Logger.Debug().Msgf("using default API version [%s]", DefaultAPIVersion)
	}

	if client.vfiler = poller.APIVfiler; client.vfiler != "" {
		client.Logger.Debug().Msgf("using vfiler tunneling [%s]", client.vfiler)
	}

	if addr = poller.Addr; addr == "" {
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	if poller.IsKfs {
		url = "https://" + addr + ":8443/servlets/netapp.servlets.admin.XMLrequest_filer"
	} else {
		url = "https://" + addr + ":443/servlets/netapp.servlets.admin.XMLrequest_filer"
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
	if request, err = http.NewRequest("POST", url, nil); err != nil {
		return nil, err
	}

	request.Header.Set("Content-type", "text/xml")
	request.Header.Set("Charset", "utf-8")

	// by default, enforce secure TLS, if not requested otherwise by user
	useInsecureTLS = false
	if poller.UseInsecureTLS != nil {
		useInsecureTLS = *poller.UseInsecureTLS
	}

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

		sslCertPath := poller.SslCert
		keyPath := poller.SslKey
		caCertPath := poller.CaCertPath

		if sslCertPath == "" {
			return nil, errs.New(errs.ErrMissingParam, "ssl_cert")
		} else if keyPath == "" {
			return nil, errs.New(errs.ErrMissingParam, "ssl_key")
		} else if cert, err = tls.LoadX509KeyPair(sslCertPath, keyPath); err != nil {
			return nil, err
		}

		// Create a CA certificate pool and add certificate if specified
		caCertPool := x509.NewCertPool()
		if caCertPath != "" {
			caCert, err := os.ReadFile(caCertPath)
			if err != nil {
				client.Logger.Error().Err(err).Str("cacert", caCertPath).Msg("Failed to read ca cert")
				// continue
			}
			if caCert != nil {
				pem := caCertPool.AppendCertsFromPEM(caCert)
				if !pem {
					client.Logger.Error().Err(err).Str("cacert", caCertPath).Msg("Failed to append ca cert")
					// continue
				}
			}
		}

		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: useInsecureTLS, //nolint:gosec
			},
		}
	} else {
		password := c.Password()
		if poller.Username == "" {
			return nil, errs.New(errs.ErrMissingParam, "username")
		} else if password == "" {
			return nil, errs.New(errs.ErrMissingParam, "password")
		}

		request.SetBasicAuth(poller.Username, password)
		transport = &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: useInsecureTLS}, //nolint:gosec
		}
	}
	if poller.TLSMinVersion != "" {
		tlsVersion := client.tlsVersion(poller.TLSMinVersion)
		if tlsVersion != 0 {
			client.Logger.Info().Uint16("tlsVersion", tlsVersion).Msg("Using TLS version")
			transport.TLSClientConfig.MinVersion = tlsVersion
		}
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

func parseClientTimeout(clientTimeout string) (time.Duration, error) {
	// does clientTimeout contain non digits?
	charIndex := strings.IndexFunc(clientTimeout, func(r rune) bool {
		return !unicode.IsDigit(r)
	})
	if charIndex != -1 {
		duration, err := time.ParseDuration(clientTimeout)
		if err != nil {
			timeout, _ := time.ParseDuration(DefaultTimeout)
			return timeout, err
		}
		return duration, nil
	}
	t, err := strconv.Atoi(clientTimeout)
	if err != nil {
		// when there is an error return the default timeout
		timeout, _ := time.ParseDuration(DefaultTimeout)
		return timeout, nil //nolint:nilerr
	}
	return time.Duration(t) * time.Second, nil
}

// Init connects to the cluster and retrieves system info
// it will give up after retries
func (c *Client) Init(retries int) error {
	var err error
	for i := 0; i < retries; i++ {
		if err = c.getSystem(); err == nil {
			break
		}
	}
	return err
}

// Name returns the name of the Cluster
func (c *Client) Name() string {
	return c.system.name
}

// IsClustered returns true if ONTAP is clustered or false if it's a 7-mode system
func (c *Client) IsClustered() bool {
	return c.system.clustered
}

// Version returns version of the ONTAP server (generation, major and minor)
func (c *Client) Version() [3]int {
	return c.system.version
}

// Release returns string with long release info of the ONTAP system
func (c *Client) Release() string {
	return c.system.release
}

// Serial returns the serial number of the ONTAP system
func (c *Client) Serial() string {
	return c.system.serial
}

// ClusterUUID returns the cluster UUID of a c-mode system and system-id for 7-mode
func (c *Client) ClusterUUID() string {
	return c.system.clusterUUID
}

// Info returns a string with details about the ONTAP system identity
func (c *Client) Info() string {
	var model, version string
	if c.IsClustered() {
		model = "CDOT"
	} else {
		model = "7MODE"
	}
	version = fmt.Sprintf("(%s version %d.%d.%d)", model, c.system.version[0], c.system.version[1], c.system.version[2])
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

// InvokeZapi will issue API requests with batching
// The method bails on the first error
func (c *Client) InvokeZapi(request *node.Node, handle func([]*node.Node) error) error {
	var output []*node.Node
	tag := "initial"

	for {
		var (
			result   *node.Node
			response []*node.Node
			err      error
		)

		if result, tag, err = c.InvokeBatchRequest(request, tag); err != nil {
			return err
		}

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

	c.Logger.Trace().Int("object", len(output)).Msg("fetching")

	return nil
}

// InvokeZapiCall will issue API requests with batching
func (c *Client) InvokeZapiCall(request *node.Node) ([]*node.Node, error) {
	var output []*node.Node
	err := c.InvokeZapi(request, func(response []*node.Node) error {
		output = append(output, response...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Invoke will issue the API request and return server response
// this method should only be called after building the request
func (c *Client) Invoke() (*node.Node, error) {
	result, _, _, err := c.invoke(false)
	return result, err
}

// InvokeBatchRequest will issue API requests in series, once there
// are no more instances returned by the server, returned results will be nil
// Use the returned tag for subsequent calls to this method
func (c *Client) InvokeBatchRequest(request *node.Node, tag string) (*node.Node, string, error) {
	// wasteful of course, need to rewrite later @TODO
	results, tag, _, _, err := c.InvokeBatchWithTimers(request, tag)
	return results, tag, err
}

// InvokeBatchWithTimers does the same as InvokeBatchRequest, but it also
// returns API time and XML parse time
func (c *Client) InvokeBatchWithTimers(request *node.Node, tag string) (*node.Node, string, time.Duration, time.Duration, error) {

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

	if err = c.BuildRequest(request); err != nil {
		return nil, "", rd, pd, err
	}

	if results, rd, pd, err = c.invoke(true); err != nil {
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
	return c.Invoke()
}

// InvokeRequest builds a request from request and invokes it
func (c *Client) InvokeRequest(request *node.Node) (*node.Node, error) {

	var err error

	if err = c.BuildRequest(request); err == nil {
		return c.Invoke()
	}
	return nil, err
}

// InvokeWithTimers invokes the request and returns parsed XML response and timers:
// API wait time and XML parse time.
// This method should only be called after building the request
func (c *Client) InvokeWithTimers() (*node.Node, time.Duration, time.Duration, error) {
	return c.invoke(true)
}

// InvokeRaw invokes the request and returns the raw server response
// This method should only be called after building the request
func (c *Client) InvokeRaw() ([]byte, error) {
	var (
		response *http.Response
		body     []byte
		err      error
	)

	if response, err = c.client.Do(c.request); err != nil {
		return body, errs.New(errs.ErrConnection, err.Error())
	}
	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return body, errs.New(errs.ErrAPIResponse, response.Status)
	}

	return io.ReadAll(response.Body)
}

// invokes the request that has been built with one of the BuildRequest* methods
func (c *Client) invoke(withTimers bool) (*node.Node, time.Duration, time.Duration, error) {

	var (
		root, result      *node.Node
		response          *http.Response
		start             time.Time
		responseT, parseT time.Duration
		body              []byte
		status, reason    string
		found             bool
		err               error
	)

	defer func(Body io.ReadCloser) { _ = Body.Close() }(c.request.Body)
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

	if response, err = c.client.Do(c.request); err != nil {
		return result, responseT, parseT, errs.New(errs.ErrConnection, err.Error())
	}
	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()
	if withTimers {
		responseT = time.Since(start)
	}

	if response.StatusCode != 200 {
		return result, responseT, parseT, errs.New(errs.ErrAPIResponse, response.Status)
	}

	// read response body
	if body, err = io.ReadAll(response.Body); err != nil {
		return result, responseT, parseT, err
	}
	defer c.printRequestAndResponse(zapiReq, body)

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

	// check if request was successful
	if result = root.GetChildS("results"); result == nil {
		return result, responseT, parseT, errs.New(errs.ErrAPIResponse, "missing \"results\"")
	}

	if status, found = result.GetAttrValueS("status"); !found {
		return result, responseT, parseT, errs.New(errs.ErrAPIResponse, "missing status attribute")
	}

	if status != "passed" {
		reason, _ = result.GetAttrValueS("reason")
		if reason == "" {
			reason = "no reason"
		}
		err = errs.New(errs.ErrAPIRequestRejected, reason)
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
	res := "<nil>"
	if response != nil {
		res = string(response)
	}
	if req != "" {
		c.Logger.Info().
			Str("Request", req).
			Str("Response", res).
			Msg("")
	}
}

func (c *Client) SetTimeout(timeout string) {
	if c.client == nil {
		return
	}
	newTimeout, err := parseClientTimeout(timeout)
	if err == nil {
		c.Logger.Debug().Str("timeout", newTimeout.String()).Msg("Using timeout")
	} else {
		c.Logger.Debug().Str("timeout", newTimeout.String()).Msg("Using default timeout")
	}
	c.client.Timeout = newTimeout
}

func (c *Client) tlsVersion(version string) uint16 {
	lower := strings.ToLower(version)
	switch lower {
	case "tls10":
		return tls.VersionTLS10
	case "tls11":
		return tls.VersionTLS11
	case "tls12":
		return tls.VersionTLS12
	case "tls13":
		return tls.VersionTLS13
	default:
		c.Logger.Warn().Str("version", version).Msg("Unknown TLS version, using default")
	}
	return 0
}
