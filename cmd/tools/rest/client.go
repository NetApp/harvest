// Copyright NetApp Inc, 2021 All rights reserved

package rest

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/logging"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// DefaultTimeout should be > than ONTAP's default REST timeout, which is 15 seconds for GET requests
	DefaultTimeout = 30 * time.Second
)

type Client struct {
	client   *http.Client
	request  *http.Request
	buffer   *bytes.Buffer
	Logger   *logging.Logger
	baseURL  string
	password string
	username string
}

func New(poller *conf.Poller) (*Client, error) {
	var (
		client         Client
		httpclient     *http.Client
		transport      *http.Transport
		cert           tls.Certificate
		timeout        time.Duration
		addr           string
		url            string
		useInsecureTLS bool
		err            error
	)

	client = Client{}
	client.Logger = logging.SubLogger("REST", "Client")

	if addr = poller.Addr; addr == "" {
		return nil, errors.New(errors.MISSING_PARAM, "addr")
	}

	if poller.IsKfs {
		url = "https://" + addr + ":8443/"
	} else {
		url = "https://" + addr + "/"
	}
	client.baseURL = url

	// by default, enforce secure TLS, if not requested otherwise by user
	if x := poller.UseInsecureTls; x != nil {
		useInsecureTLS = *poller.UseInsecureTls
	} else {
		useInsecureTLS = false
	}

	// set authentication method
	if poller.AuthStyle == "certificate_auth" {
		certPath := poller.SslCert
		keyPath := poller.SslKey
		if certPath == "" {
			return nil, errors.New(errors.MISSING_PARAM, "ssl_cert")
		} else if keyPath == "" {
			return nil, errors.New(errors.MISSING_PARAM, "ssl_key")
		} else if cert, err = tls.LoadX509KeyPair(certPath, keyPath); err != nil {
			return nil, err
		}

		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: useInsecureTLS},
		}
	} else {
		username := poller.Username
		password := poller.Password
		client.username = username
		client.password = password
		if username == "" {
			return nil, errors.New(errors.MISSING_PARAM, "username")
		} else if password == "" {
			return nil, errors.New(errors.MISSING_PARAM, "password")
		}

		transport = &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: useInsecureTLS},
		}
	}

	timeout = DefaultTimeout
	if poller.ClientTimeout != "" {
		timeout, err = time.ParseDuration(poller.ClientTimeout)
		if err != nil {
			client.Logger.Error().Msgf("err paring client timeout of=[%s] err=%+v\n", timeout, err)
		}
	}
	client.Logger.Debug().Msgf("using timeout [%d]", timeout)

	httpclient = &http.Client{Transport: transport, Timeout: timeout}
	client.client = httpclient

	return &client, nil
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
		return ioutil.NopCloser(r), nil
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
		defer silentClose(c.request.Body)
	}
	if c.buffer != nil {
		defer c.buffer.Reset()
	}

	// send request to server
	if response, err = c.client.Do(c.request); err != nil {
		return nil, fmt.Errorf("connection error %w", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("server returned status code %d", response.StatusCode)
	}

	// read response body
	defer silentClose(response.Body)

	if body, err = ioutil.ReadAll(response.Body); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	return body, nil
}

func downloadSwagger(poller *conf.Poller, path string, url string) (int64, error) {
	var restClient *Client

	out, err := os.Create(path)
	if err != nil {
		return 0, fmt.Errorf("unable to create %s to save swagger.yaml", path)
	}
	defer silentClose(out)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	if restClient, err = New(poller); err != nil {
		return 0, fmt.Errorf("error creating new client %w\n", err)
	}

	downClient := &http.Client{Transport: restClient.client.Transport, Timeout: restClient.client.Timeout}
	if restClient.username != "" {
		request.SetBasicAuth(restClient.username, restClient.password)
	}
	response, err := downClient.Do(request)
	if err != nil {
		return 0, err
	}
	defer silentClose(response.Body)

	n, err := io.Copy(out, response.Body)
	if err != nil {
		return 0, fmt.Errorf("error while downloading %s err=%w\n", url, err)
	}
	return n, nil
}
