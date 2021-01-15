package client

import (
    "fmt"
    "time"
    "bytes"
    "errors"
    "io"
    "io/ioutil"
    "net/http"
    "crypto/tls"
    "local.host/params"
    "local.host/xmltree"
)

type Client struct {
    client      *http.Client
    request     *http.Request
    buffer      *bytes.Buffer
}

func New(p params.Params) (Client, error) {
    var client Client
    var httpclient *http.Client
    var request *http.Request
    var transport *http.Transport
    var cert tls.Certificate
    var url string
    var err error

    err = nil

    url = "https://" + p.Hostname + ":443/servlets/netapp.servlets.admin.XMLrequest_filer"

    request, err = http.NewRequest("POST", url, nil)
    if err != nil {
        fmt.Printf("[Client.New] Error initializing request: %s\n", err)
        return client, err
    }

    request.Header.Set("Content-type", "text/xml")
    request.Header.Set("Charset", "utf-8")

    if p.UseCert {
        cert, err = tls.LoadX509KeyPair(p.Authorization[0], p.Authorization[1])
        if err != nil {
            fmt.Printf("[Client.New] Error loading key pair: %s\n", err)
            return client, err
        }
        transport = &http.Transport{ TLSClientConfig : &tls.Config{Certificates : []tls.Certificate{cert}, InsecureSkipVerify : true }, }
    } else {
        request.SetBasicAuth(p.Authorization[0], p.Authorization[1])
        transport = &http.Transport{ TLSClientConfig : &tls.Config{ InsecureSkipVerify : true }, }
    }

    // initialize http client
    httpclient = &http.Client{ Transport : transport, Timeout: time.Duration(p.Timeout) * time.Second }

    client = Client{ client: httpclient, request: request }

    // ensure that we can change body dynamically
    request.GetBody = func() (io.ReadCloser, error) {
        r := bytes.NewReader(client.buffer.Bytes())
        return ioutil.NopCloser(r), nil
    }

    return client, err
}

func (c *Client) BuildRequest(node *xmltree.Node) error {
    var buffer *bytes.Buffer
    var xml []byte
    var err error

    xml, err = node.Build()

    if err == nil {
        buffer = bytes.NewBuffer(xml)
        c.buffer = buffer
        c.request.Body = ioutil.NopCloser(buffer)
        c.request.ContentLength = int64(buffer.Len())
    }
    return err
}

func (c *Client) InvokeRequest() (*xmltree.Node, error) {
    var err error
    var body []byte
    var response *http.Response
    var node *xmltree.Node
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

    node, err = xmltree.Parse(body)
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

