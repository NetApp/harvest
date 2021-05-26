/*
 * Copyright NetApp Inc, 2021 All rights reserved

Package zapi provides type Client for connecting to a C-dot or 7-mode
Ontap cluster and sending API requests using the ZAPI protocol.
*/
package zapi

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logging"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	DefaultApiVersion = "1.3"
	DefaultTimeout    = 10
)

type Client struct {
	client     *http.Client
	request    *http.Request
	buffer     *bytes.Buffer
	system     *system
	apiVersion string
	vfiler     string
	Logger     *logging.Logger // logger used for logging
}

func New(config *node.Node) (*Client, error) {
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

	client = Client{}
	client.Logger = logging.GetInstanceSubLogger("Zapi", "Client")

	// check required & optional parameters
	if client.apiVersion = config.GetChildContentS("api_version"); client.apiVersion == "" {
		client.apiVersion = DefaultApiVersion
		client.Logger.Debug().Msgf("using default API version [%s]", DefaultApiVersion)
	}

	if client.vfiler = config.GetChildContentS("api_vfiler"); client.vfiler != "" {
		client.Logger.Debug().Msgf("using vfiler tunneling [%s]", client.vfiler)
	}

	if addr = config.GetChildContentS("addr"); addr == "" {
		return nil, errors.New(errors.MISSING_PARAM, "addr")
	}

	url = "https://" + addr + ":443/servlets/netapp.servlets.admin.XMLrequest_filer"

	// create a request object that will be used for later requests
	if request, err = http.NewRequest("POST", url, nil); err != nil {
		return nil, err
	}

	request.Header.Set("Content-type", "text/xml")
	request.Header.Set("Charset", "utf-8")

	// by default, encorce secure TLS, if not requested otherwise by user
	if x := config.GetChildContentS("use_insecure_tls"); x != "" {
		if useInsecureTLS, err = strconv.ParseBool(x); err != nil {
			client.Logger.Error().Stack().Err(err).Msgf("use_insecure_tls:")
		}
	} else {
		useInsecureTLS = false
	}

	// set authentication method
	if config.GetChildContentS("auth_style") == "certificate_auth" {

		certPath := config.GetChildContentS("ssl_cert")
		keyPath := config.GetChildContentS("ssl_key")

		if certPath == "" {
			return nil, errors.New(errors.MISSING_PARAM, "ssl_cert")
		} else if keyPath == "" {
			return nil, errors.New(errors.MISSING_PARAM, "ssl_key")
		} else if cert, err = tls.LoadX509KeyPair(certPath, keyPath); err != nil {
			return nil, err
		}

		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: useInsecureTLS},
		}
	} else {

		if !useInsecureTLS {
			return nil, errors.New(errors.INVALID_PARAM, "use_insecure_tls is false, but no certificates")
		}

		username := config.GetChildContentS("username")
		password := config.GetChildContentS("password")

		if username == "" {
			return nil, errors.New(errors.MISSING_PARAM, "username")
		} else if password == "" {
			return nil, errors.New(errors.MISSING_PARAM, "password")
		}

		request.SetBasicAuth(username, password)
		transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: useInsecureTLS}}
	}
	client.request = request

	// initialize http client
	if t, err := strconv.Atoi(config.GetChildContentS("client_timeout")); err == nil {
		timeout = time.Duration(t) * time.Second
		client.Logger.Debug().Msgf("using timeout [%d] s", t)
	} else {
		// default timeout
		timeout = time.Duration(DefaultTimeout) * time.Second
		client.Logger.Debug().Msgf("using default timeout [%d] s", DefaultTimeout)
	}

	httpclient = &http.Client{Transport: transport, Timeout: timeout}

	client.client = httpclient

	// ensure that we can change body dynamically
	request.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(client.buffer.Bytes())
		return ioutil.NopCloser(r), nil
	}

	return &client, nil
}

// init connects to the cluster and retrieves system info
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

// Version returns version of the ONTAP server (generation, manjor and minor)
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

// BuildRequestString builds an API request from the string request
// request is usually the API name (e.g. "system-get-info") without any attributes
func (c *Client) BuildRequestString(request string) error {
	return c.buildRequestString(request, false)
}

// BuildRequest builds an API request from the node query
// root element of the request is usualy the API name (e.g. "volume-get-iter") and
// its children are the attributes requested
func (c *Client) BuildRequest(request *node.Node) error {
	return c.buildRequest(request, false)
}

func (my *Client) buildRequestString(request string, forceCluster bool) error {
	return my.buildRequest(node.NewXmlS(request), forceCluster)
}

// build API request from the given node object.
func (c *Client) buildRequest(query *node.Node, forceCluster bool) error {
	var (
		request *node.Node
		buffer  *bytes.Buffer
		data    []byte
		err     error
	)

	request = node.NewXmlS("netapp")
	request.NewAttrS("xmlns", "http://www.netapp.com/filer/admin")
	request.NewAttrS("version", c.apiVersion)
	// optionally use fviler-tunneling, this option is never used in Harvest
	if !forceCluster && c.vfiler != "" {
		request.NewAttrS("vfiler", c.vfiler)
	}
	request.AddChild(query)

	if data, err = tree.DumpXml(request); err != nil {
		return err
	}

	buffer = bytes.NewBuffer(data)
	c.buffer = buffer
	c.request.Body = ioutil.NopCloser(buffer)
	c.request.ContentLength = int64(buffer.Len())
	return nil
}

// Invoke will issue the API request and return server response
// this method should only be callled after building the request
func (c *Client) Invoke() (*node.Node, error) {
	result, _, _, err := c.invoke(false)
	return result, err
}

// InvokeBatchRequest will issue API requests in series, once there
// are no more instances returned by the server, returned results will be nill
// Use the returned tag for subsequent calls to this method
func (c *Client) InvokeBatchRequest(request *node.Node, tag string) (*node.Node, string, error) {
	// wasteful of course, need to rewrite later @TODO
	results, tag, _, _, err := c.InvokeBatchWithTimers(request, tag)
	return results, tag, err
}

// InvokeBatchRequestWithTimers does the same as InvokeBatchRequest, but it also
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
func (me *Client) InvokeRequestString(request string) (*node.Node, error) {
	if err := me.BuildRequestString(request); err != nil {
		return nil, err
	}
	return me.Invoke()
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
// This method should only be callled after building the request
func (c *Client) InvokeWithTimers() (*node.Node, time.Duration, time.Duration, error) {
	return c.invoke(true)
}

// InvokeWithTimers invokes the request and returns the raw server response
// This method should only be callled after building the request
func (c *Client) InvokeRaw() ([]byte, error) {
	var (
		response *http.Response
		body     []byte
		err      error
	)

	if response, err = c.client.Do(c.request); err != nil {
		return body, errors.New(errors.ERR_CONNECTION, err.Error())
	}

	if response.StatusCode != 200 {
		return body, errors.New(errors.API_RESPONSE, response.Status)
	}

	return ioutil.ReadAll(response.Body)
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

	defer c.request.Body.Close()
	defer c.buffer.Reset()

	// issue request to server
	if withTimers {
		start = time.Now()
	}
	if response, err = c.client.Do(c.request); err != nil {
		return result, responseT, parseT, errors.New(errors.ERR_CONNECTION, err.Error())
	}
	if withTimers {
		responseT = time.Since(start)
	}

	if response.StatusCode != 200 {
		return result, responseT, parseT, errors.New(errors.API_RESPONSE, response.Status)
	}

	// read response body
	defer response.Body.Close()

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return result, responseT, parseT, err
	}

	// parse xml
	if withTimers {
		start = time.Now()
	}
	if root, err = tree.LoadXml(body); err != nil {
		return result, responseT, parseT, err
	}
	if withTimers {
		parseT = time.Since(start)
	}

	// check if request was successful
	if result = root.GetChildS("results"); result == nil {
		return result, responseT, parseT, errors.New(errors.API_RESPONSE, "missing \"results\"")
	}

	if status, found = result.GetAttrValueS("status"); !found {
		return result, responseT, parseT, errors.New(errors.API_RESPONSE, "missing status attribute")
	}

	if status != "passed" {
		if reason, found = result.GetAttrValueS("reason"); !found {
			err = errors.New(errors.API_REQ_REJECTED, "no reason")
		} else {
			err = errors.New(errors.API_REQ_REJECTED, reason)
		}
		return result, responseT, parseT, err
	}

	return result, responseT, parseT, nil
}
