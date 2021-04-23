//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package zapi

import (
	"bytes"
	"crypto/tls"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logger"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/tree/xml"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var (
	API_VERSION = "1.3"
	API_VFILER  = ""
)

type Client struct {
	client  *http.Client
	request *http.Request
	buffer  *bytes.Buffer
	System  *System
}

func New(config *node.Node) (*Client, error) {
	var client *Client
	var httpclient *http.Client
	var request *http.Request
	var transport *http.Transport
	var cert tls.Certificate
	var timeout time.Duration
	var url, addr string
	var err error

	err = nil

	// check required parameters
	if addr = config.GetChildContentS("addr"); addr == "" {
		return client, errors.New(errors.MISSING_PARAM, "addr")
	}

	if v := config.GetChildContentS("api_version"); v != "" {
		API_VERSION = v
		logger.Debug("(Zapi:Client)", "using custom API version [%s]", v)
	} else {
		logger.Debug("(Zapi:Client)", "using default API version [%s]", API_VERSION)
	}

	if v := config.GetChildContentS("api_vfiler"); v != "" {
		API_VFILER = v
		logger.Debug("(Zapi:Client)", "using vfiler tunneling to [%s]", v)
	}

	url = "https://" + addr + ":443/servlets/netapp.servlets.admin.XMLrequest_filer"

	request, err = http.NewRequest("POST", url, nil)
	if err != nil {
		//fmt.Printf("[Client.New] Error initializing request: %s\n", err)
		return client, err
	}

	request.Header.Set("Content-type", "text/xml")
	request.Header.Set("Charset", "utf-8")

	useInsecureTLS, err := strconv.ParseBool(config.GetChildContentS("useInsecureTLS"))
	if err != nil {
		logger.Error("(Zapi:Client)", "Error %v ", err)
	}

	if config.GetChildContentS("auth_style") == "certificate_auth" {

		cert_path := config.GetChildContentS("ssl_cert")
		key_path := config.GetChildContentS("ssl_key")

		if cert_path == "" {
			return client, errors.New(errors.MISSING_PARAM, "ssl_cert")
		} else if key_path == "" {
			return client, errors.New(errors.MISSING_PARAM, "ssl_key")
		} else if cert, err = tls.LoadX509KeyPair(cert_path, key_path); err != nil {
			return client, err
		}

		transport = &http.Transport{TLSClientConfig: &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: useInsecureTLS}}
	} else {

		username := config.GetChildContentS("username")
		password := config.GetChildContentS("password")

		if username == "" {
			return client, errors.New(errors.MISSING_PARAM, "username")
		} else if password == "" {
			return client, errors.New(errors.MISSING_PARAM, "password")
		}

		request.SetBasicAuth(username, password)
		transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: useInsecureTLS}}

	}

	// initialize http client
	t, err := strconv.Atoi(config.GetChildContentS("client_timeout"))
	if err != nil {
		// default timeout
		timeout = time.Duration(5) * time.Second
	} else {
		timeout = time.Duration(t) * time.Second
	}

	httpclient = &http.Client{Transport: transport, Timeout: timeout}

	client = &Client{client: httpclient, request: request}

	// ensure that we can change body dynamically
	request.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(client.buffer.Bytes())
		return ioutil.NopCloser(r), nil
	}

	return client, nil
}

func (my *Client) IsClustered() bool {
	if my.System != nil {
		return my.System.Clustered
	}
	panic("System is nil")
}

func (my *Client) BuildRequestString(request string) error {
	return my.BuildRequest(node.NewXmlS(request))
}

func (my *Client) build_request_string(request string, force_cluster bool) error {
	return my.build_request(node.NewXmlS(request), force_cluster)
}

func (my *Client) BuildRequest(query *node.Node) error {
	return my.build_request(query, false)
}

func (my *Client) build_request(query *node.Node, force_cluster bool) error {
	var buffer *bytes.Buffer
	var data []byte
	var err error

	request := node.NewXmlS("netapp")
	request.NewAttrS("xmlns", "http://www.netapp.com/filer/admin")
	request.NewAttrS("version", API_VERSION)
	if API_VFILER != "" && !force_cluster {
		request.NewAttrS("vfiler", API_VFILER)
	}
	request.AddChild(query)

	if data, err = tree.DumpXml(request); err == nil {
		buffer = bytes.NewBuffer(data)
		my.buffer = buffer
		my.request.Body = ioutil.NopCloser(buffer)
		my.request.ContentLength = int64(buffer.Len())
	}
	return err
}

func (c *Client) Invoke() (*node.Node, error) {
	result, _, _, err := c.invoke(false)
	return result, err
}

func (c *Client) InvokeBatchRequest(request *node.Node, tag string) (*node.Node, string, error) {
	// wasteful of course, need to rewrite later...
	results, tag, _, _, err := c.InvokeBatchWithTimers(request, tag)
	return results, tag, err
}

func (c *Client) InvokeBatchWithTimers(request *node.Node, tag string) (*node.Node, string, time.Duration, time.Duration, error) {

	var results *node.Node
	var next_tag string
	var err error
	var rd, pd time.Duration // response time, parse time

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

	// avoid ZAPI bug
	if next_tag = results.GetChildContentS("next-tag"); next_tag == tag {
		next_tag = ""
	}

	return results, next_tag, rd, pd, nil
}

func (me *Client) InvokeRequestString(request string) (*node.Node, error) {
	if err := me.BuildRequestString(request); err != nil {
		return nil, err
	}
	return me.Invoke()
}

func (c *Client) InvokeRequest(request *node.Node) (*node.Node, error) {

	var err error

	if err = c.BuildRequest(request); err == nil {
		return c.Invoke()
	}
	return nil, err
}

func (c *Client) InvokeWithTimers() (*node.Node, time.Duration, time.Duration, error) {
	return c.invoke(true)
}

func (my *Client) InvokeRaw() ([]byte, error) {
	var response *http.Response
	var body []byte
	var err error

	if response, err = my.client.Do(my.request); err != nil {
		return body, errors.New(errors.ERR_CONNECTION, err.Error())
	}

	if response.StatusCode != 200 {
		return body, errors.New(errors.API_RESPONSE, response.Status)
	}

	return ioutil.ReadAll(response.Body)
}

func (my *Client) invoke(with_timers bool) (*node.Node, time.Duration, time.Duration, error) {

	var (
		root, result        *node.Node
		response            *http.Response
		start               time.Time
		response_t, parse_t time.Duration
		body                []byte
		status, reason      string
		found               bool
		err                 error
	)

	defer my.request.Body.Close()
	defer my.buffer.Reset()

	// issue request to server
	if with_timers {
		start = time.Now()
	}
	if response, err = my.client.Do(my.request); err != nil {
		return result, response_t, parse_t, errors.New(errors.ERR_CONNECTION, err.Error())
	}
	if with_timers {
		response_t = time.Since(start)
	}

	if response.StatusCode != 200 {
		return result, response_t, parse_t, errors.New(errors.API_RESPONSE, response.Status)
	}

	// read response body
	defer response.Body.Close()

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return result, response_t, parse_t, err
	}

	// parse xml
	if with_timers {
		start = time.Now()
	}
	if root, err = tree.LoadXml(body); err != nil {
		return result, response_t, parse_t, err
	}
	if with_timers {
		parse_t = time.Since(start)
	}

	// check if request was successful
	if result = root.GetChildS("results"); result == nil {
		return result, response_t, parse_t, errors.New(errors.API_RESPONSE, "missing \"results\"")
	}

	if status, found = result.GetAttrValueS("status"); !found {
		return result, response_t, parse_t, errors.New(errors.API_RESPONSE, "missing status attribute")
	}

	if status != "passed" {
		if reason, found = result.GetAttrValueS("reason"); !found {
			err = errors.New(errors.API_REQ_REJECTED, "no reason")
		} else {
			err = errors.New(errors.API_REQ_REJECTED, reason)
		}
		return result, response_t, parse_t, err
	}

	return result, response_t, parse_t, nil
}

func (c *Client) InvokeBatchWithMoreTimers(request *node.Node, tag string) (*node.Node, string, int64, time.Duration, time.Duration, time.Duration, time.Duration, error) {

	var results *node.Node
	var next_tag string
	var err error
	var cl int64
	var ad, rd, bd, pd time.Duration // api time, read time, parse time

	if tag == "" {
		return nil, "", cl, bd, ad, rd, pd, nil
	}

	if tag != "initial" {
		request.SetChildContentS("tag", tag)
	}

	build_start := time.Now()
	if err = c.BuildRequest(request); err != nil {
		return nil, "", cl, bd, ad, rd, pd, err
	}
	bd = time.Since(build_start)

	if results, cl, ad, rd, pd, err = c.invoke_with_more_timers(); err != nil {
		return nil, "", cl, bd, ad, rd, pd, err
	}

	// avoid ZAPI bug
	if next_tag = results.GetChildContentS("next-tag"); next_tag == tag {
		next_tag = ""
	}

	return results, next_tag, cl, bd, ad, rd, pd, nil
}

func (c *Client) invoke_with_more_timers() (*node.Node, int64, time.Duration, time.Duration, time.Duration, error) {

	var (
		root, result                *node.Node
		response                    *http.Response
		response_d, read_d, parse_d time.Duration
		//	body                []byte
		status, reason string
		found          bool
		err            error
		cl             int64
	)

	// issue request
	response_start := time.Now()
	if response, err = c.client.Do(c.request); err != nil {
		return result, cl, response_d, read_d, parse_d, errors.New(errors.ERR_CONNECTION, err.Error())
	}
	response_d = time.Since(response_start)

	if response.StatusCode != 200 {
		return result, cl, response_d, read_d, parse_d, errors.New(errors.API_RESPONSE, response.Status)
	}

	// read response body
	defer response.Body.Close()
	cl = response.ContentLength

	/*
		read_start := time.Now()
		if body, err = ioutil.ReadAll(response.Body); err != nil {
			return result, response_d, read_d, parse_d, err
		}
		read_d = time.Since(read_start)
	*/

	// parse xml
	parse_start := time.Now()
	//if root, err = tree.LoadXml(body); err != nil {
	if root, err = xml.LoadFromReader(response.Body); err != nil {
		return result, cl, response_d, read_d, parse_d, err
	}
	parse_d = time.Since(parse_start)

	// check if request was successful
	if result = root.GetChildS("results"); result == nil {
		return result, cl, response_d, read_d, parse_d, errors.New(errors.API_RESPONSE, "missing \"results\"")
	}

	if status, found = result.GetAttrValueS("status"); !found {
		return result, cl, response_d, read_d, parse_d, errors.New(errors.API_RESPONSE, "missing status attribute")
	}

	if status != "passed" {
		if reason, found = result.GetAttrValueS("reason"); !found {
			err = errors.New(errors.API_REQ_REJECTED, "no reason")
		} else {
			err = errors.New(errors.API_REQ_REJECTED, reason)
		}
	}

	return result, cl, response_d, read_d, parse_d, err
}
