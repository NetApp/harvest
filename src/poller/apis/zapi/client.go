package zapi

import (
    "fmt"
    "time"
    "bytes"
    "errors"
    "io"
    "strconv"
    "io/ioutil"
    "net/http"
    "crypto/tls"
    "poller/yaml"
    "poller/xml"
)

type Client struct {
    client      *http.Client
    request     *http.Request
    buffer      *bytes.Buffer
}

func New(config *yaml.Node) (Client, error) {
    var client Client
    var httpclient *http.Client
    var request *http.Request
    var transport *http.Transport
    var cert tls.Certificate
    var timeout time.Duration
    var url string
    var err error

    err = nil

    url = "https://" + config.GetChildValue("url") + ":443/servlets/netapp.servlets.admin.XMLrequest_filer"

    request, err = http.NewRequest("POST", url, nil)
    if err != nil {
        fmt.Printf("[Client.New] Error initializing request: %s\n", err)
        return client, err
    }

    request.Header.Set("Content-type", "text/xml")
    request.Header.Set("Charset", "utf-8")

    if config.GetChildValue("auth_style") == "certificate_auth" {
        cert, err = tls.LoadX509KeyPair(config.GetChildValue("ssl_cert"), config.GetChildValue("ssl_key"))
        if err != nil {
            fmt.Printf("[Client.New] Error loading key pair: %s\n", err)
            return client, err
        }
        transport = &http.Transport{ TLSClientConfig : &tls.Config{Certificates : []tls.Certificate{cert}, InsecureSkipVerify : true }, }
    } else {
        request.SetBasicAuth(config.GetChildValue("username"), config.GetChildValue("password"))
        transport = &http.Transport{ TLSClientConfig : &tls.Config{ InsecureSkipVerify : true }, }
    }

    // initialize http client
    t, err := strconv.Atoi(config.GetChildValue("client_timeout"))
    if err != nil {
        timeout = time.Duration(5) * time.Second
        fmt.Printf("Using default timeout [%s]\n", timeout.String())
    } else {
        timeout = time.Duration(t) * time.Second
        fmt.Printf("Using timeout [%s]\n", timeout.String())
    }

    httpclient = &http.Client{ Transport : transport, Timeout: timeout }

    client = Client{ client: httpclient, request: request }

    // ensure that we can change body dynamically
    request.GetBody = func() (io.ReadCloser, error) {
        r := bytes.NewReader(client.buffer.Bytes())
        return ioutil.NopCloser(r), nil
    }

    return client, nil
}

func (c *Client) BuildRequest(node *xml.Node) error {
    var buffer *bytes.Buffer
    var data []byte
    var err error

    data, err = node.Build()

    if err == nil {
        buffer = bytes.NewBuffer(data)
        c.buffer = buffer
        c.request.Body = ioutil.NopCloser(buffer)
        c.request.ContentLength = int64(buffer.Len())
    }
    return err
}

func (c *Client) InvokeRequest() (*xml.Node, error) {
    var err error
    var body []byte
    var response *http.Response
    var node *xml.Node
    var status, reason string
    var found bool

    response, err = c.client.Do(c.request)
    if err != nil {
        fmt.Printf("error reading response: %s\n", err)
        return node, err
    }

    defer response.Body.Close()

    body, err = ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Printf("error reading body: %s\n", err)
        return node, err
    }

    node, err = xml.Parse(body)
    if err != nil {
        fmt.Printf("error parsing body: %s\n", err)
        return node, err
    }

    if status, found = node.GetAttr("status"); !found {
        err = errors.New("Missing status attribute")
    } else if status != "passed" {
        reason, _ = node.GetAttr("reason")
        err = errors.New("Request rejected: " + reason)
    }

    return node, err
}

