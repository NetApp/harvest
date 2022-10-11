// Copyright NetApp Inc, 2021 All rights reserved

package rest

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

const (
	// DefaultTimeout should be > than ONTAP's default REST timeout, which is 15 seconds for GET requests
	DefaultTimeout = "30s" // in seconds
)

type Client struct {
	client   *http.Client
	request  *http.Request
	buffer   *bytes.Buffer
	Logger   *logging.Logger
	baseURL  string
	cluster  Cluster
	password string
	username string
	Timeout  time.Duration
	logRest  bool // used to log Rest request/response

}

type Cluster struct {
	Name    string
	Info    string
	UUID    string
	Version [3]int
}

func New(poller conf.Poller, timeout time.Duration) (*Client, error) {
	var (
		client         Client
		httpclient     *http.Client
		transport      *http.Transport
		cert           tls.Certificate
		addr           string
		url            string
		useInsecureTLS bool
		err            error
	)

	client = Client{}
	client.Logger = logging.Get().SubLogger("REST", "Client")

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

	// by default, enforce secure TLS, if not requested otherwise by user
	if x := poller.UseInsecureTLS; x != nil {
		useInsecureTLS = *poller.UseInsecureTLS
	} else {
		useInsecureTLS = false
	}

	// check if a credentials file is being used and if so, parse and use the values from it
	if poller.CredentialsFile != "" {
		err := conf.ReadCredentialsFile(poller.CredentialsFile, &poller)
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
		password := poller.Password
		client.username = username
		client.password = password
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

func (c *Client) printRequestAndResponse(req string, response []byte) {
	if c.logRest {
		res := "<nil>"
		if response != nil {
			res = string(response)
		}
		c.Logger.Info().
			Str("Request", req).
			Str("Response", res).
			Msg("")
	}
}

// GetRest makes a REST request to the cluster and returns a json response as a []byte
func (c *Client) GetRest(request string) ([]byte, error) {
	var err error
	if strings.Index(request, "/") == 0 {
		request = request[1:]
	}
	u := c.baseURL + request
	c.request, err = http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.request.Header.Set("accept", "application/json")
	if c.username != "" {
		c.request.SetBasicAuth(c.username, c.password)
	}
	// ensure that we can change body dynamically
	c.request.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(c.buffer.Bytes())
		return io.NopCloser(r), nil
	}
	if err != nil {
		return nil, err
	}

	result, err := c.invoke()
	return result, err
}

func (c *Client) invoke() ([]byte, error) {
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

	restReq := c.request.URL.String()

	// send request to server
	if response, err = c.client.Do(c.request); err != nil {
		return nil, fmt.Errorf("connection error %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()

	if response.StatusCode != 200 {
		if body, err = io.ReadAll(response.Body); err == nil {
			result := gjson.GetBytes(body, "error")
			if result.Exists() {
				message := result.Get("message").String()
				code := result.Get("code").Int()
				target := result.Get("target").String()
				return nil, errs.Rest(response.StatusCode, message, code, target)
			}
			return nil, errs.Rest(response.StatusCode, "", 0, "")
		}
		return nil, errs.Rest(response.StatusCode, err.Error(), 0, "")
	}

	// read response body
	if body, err = io.ReadAll(response.Body); err != nil {
		return nil, err
	}
	defer c.printRequestAndResponse(restReq, body)

	if err != nil {
		return nil, err
	}
	return body, nil
}

func downloadSwagger(poller *conf.Poller, path string, url string, verbose bool) (int64, error) {
	var restClient *Client

	out, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("unable to create %s to save swagger.yaml", path)
	}
	defer func(out *os.File) { _ = out.Close() }(out)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	timeout, _ := time.ParseDuration(DefaultTimeout)
	if restClient, err = New(*poller, timeout); err != nil {
		return 0, fmt.Errorf("error creating new client %w", err)
	}

	downClient := &http.Client{Transport: restClient.client.Transport, Timeout: restClient.client.Timeout}
	if restClient.username != "" {
		request.SetBasicAuth(restClient.username, restClient.password)
	}
	if verbose {
		requestOut, _ := httputil.DumpRequestOut(request, false)
		fmt.Printf("REQUEST: %s BY: %s\n%s\n", url, restClient.username, requestOut)
	}
	response, err := downClient.Do(request)
	if err != nil {
		return 0, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer response.Body.Close()

	if verbose {
		debugResp, _ := httputil.DumpResponse(response, false)
		fmt.Printf("RESPONSE: \n%s", debugResp)
	}
	if response.StatusCode != 200 {
		return 0, fmt.Errorf("error making request. server response statusCode=[%d]", response.StatusCode)
	}
	n, err := io.Copy(out, response.Body)
	if err != nil {
		return 0, fmt.Errorf("error while downloading %s err=%w", url, err)
	}
	return n, nil
}

func (c *Client) Init(retries int) error {

	var (
		err     error
		content []byte
		i       int
	)

	for i = 0; i < retries; i++ {

		if content, err = c.GetRest(BuildHref("cluster", "*", nil, "", "", "", "", "")); err != nil {
			continue
		}

		results := gjson.GetManyBytes(content, "name", "uuid", "version.full", "version.generation", "version.major", "version.minor")
		c.cluster.Name = results[0].String()
		c.cluster.UUID = results[1].String()
		c.cluster.Info = results[2].String()
		c.cluster.Version[0] = int(results[3].Int())
		c.cluster.Version[1] = int(results[4].Int())
		c.cluster.Version[2] = int(results[5].Int())
		return nil
	}
	return err
}

func BuildHref(apiPath string, fields string, field []string, queryFields string, queryValue string, maxRecords string, returnTimeout string, endpoint string) string {
	href := strings.Builder{}
	if endpoint == "" {
		href.WriteString("api/")
		href.WriteString(apiPath)
	} else {
		href.WriteString(endpoint)
	}
	href.WriteString("?return_records=true")
	addArg(&href, "&fields=", fields)
	for _, f := range field {
		addArg(&href, "&", f)
	}
	addArg(&href, "&query_fields=", queryFields)
	addArg(&href, "&query=", queryValue)
	addArg(&href, "&max_records=", maxRecords)
	addArg(&href, "&return_timeout=", returnTimeout)
	return href.String()
}

func addArg(href *strings.Builder, field string, value string) {
	if value == "" {
		return
	}
	href.WriteString(field)
	href.WriteString(value)
}

func (c *Client) Cluster() Cluster {
	return c.cluster
}
