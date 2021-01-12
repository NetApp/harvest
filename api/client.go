package api

import (
    "fmt"
    "time"
    "bytes"
    "errors"
    "io"
    "io/ioutil"
    "net/http"
    "crypto/tls"
)

type Params struct {
    Hostname string
    UseCert bool
    Authorization [2]string
    Timeout int
    Path string
    Object string
    Template string
    Subtemplate string
}

type Connection struct {
    client      *http.Client
    request     *http.Request
    buffer      *bytes.Buffer
}

func NewConnection(p Params) (Connection, error) {
    var connection Connection
    var client *http.Client
    var request *http.Request
    var transport *http.Transport
    var cert tls.Certificate
    var url string
    var err error

    err = nil

    url = fmt.Sprintf("https://%s:443/servlets/netapp.servlets.admin.XMLrequest_filer", p.Hostname)
    request, err = http.NewRequest("POST", url, nil)
    if err != nil {
        fmt.Printf("Error creating request: %s\n", err)
        return connection, err
    }
    request.Header.Set("Content-type", "text/xml")
    request.Header.Set("Charset", "utf-8")

    if p.UseCert == true {
        cert, err = tls.LoadX509KeyPair(p.Authorization[0], p.Authorization[1])
        if err != nil {
            fmt.Printf("Error loading key pair: %s\n", err)
            return connection, err
        }
        transport = &http.Transport{ TLSClientConfig : &tls.Config{Certificates : []tls.Certificate{cert}, InsecureSkipVerify : true }, }
    } else {
        request.SetBasicAuth(p.Authorization[0], p.Authorization[1])
        transport = &http.Transport{ TLSClientConfig : &tls.Config{ InsecureSkipVerify : true }, }
    }

    // build client
    client = &http.Client{ Transport : transport, Timeout: time.Duration(p.Timeout) * time.Second }

    connection = Connection{ client: client, request: request }

    // ensure that we can change body dynamically
    request.GetBody = func() (io.ReadCloser, error) {
        r := bytes.NewReader(connection.buffer.Bytes())
        return ioutil.NopCloser(r), nil
    }

    return connection, err
}

func (c *Connection) BuildRequest(node *Node) error {
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

func (c *Connection) InvokeRequest() (*Node, error) {
    var err error
    var body []byte
    var response *http.Response
    var node *Node
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

    node, err = Parse(body)
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

