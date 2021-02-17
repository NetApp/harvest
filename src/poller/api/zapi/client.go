package zapi

import (
    //"fmt"
    "time"
    "bytes"
    "io"
    "strconv"
    "io/ioutil"
    "net/http"
    "crypto/tls"
	"goharvest2/share/errors"
    "goharvest2/share/tree"
    "goharvest2/share/tree/node"
)

type Client struct {
    client      *http.Client
    request     *http.Request
    buffer      *bytes.Buffer
}

func New(config *node.Node) (*Client, error) {
    var client *Client
    var httpclient *http.Client
    var request *http.Request
    var transport *http.Transport
    var cert tls.Certificate
    var timeout time.Duration
    var url string
    var err error

    err = nil

    url = "https://" + config.GetChildContentS("url") + ":443/servlets/netapp.servlets.admin.XMLrequest_filer"

    request, err = http.NewRequest("POST", url, nil)
    if err != nil {
        //fmt.Printf("[Client.New] Error initializing request: %s\n", err)
        return client, err
    }

    request.Header.Set("Content-type", "text/xml")
    request.Header.Set("Charset", "utf-8")

    if config.GetChildContentS("auth_style") == "certificate_auth" {
        cert, err = tls.LoadX509KeyPair(config.GetChildContentS("ssl_cert"), config.GetChildContentS("ssl_key"))
        if err != nil {
            //fmt.Printf("[Client.New] Error loading key pair: %s\n", err)
            return client, err
        }
        transport = &http.Transport{ TLSClientConfig : &tls.Config{Certificates : []tls.Certificate{cert}, InsecureSkipVerify : true }, }
    } else {
        request.SetBasicAuth(config.GetChildContentS("username"), config.GetChildContentS("password"))
        transport = &http.Transport{ TLSClientConfig : &tls.Config{ InsecureSkipVerify : true }, }
    }

    // initialize http client
    t, err := strconv.Atoi(config.GetChildContentS("client_timeout"))
    if err != nil {
        timeout = time.Duration(5) * time.Second
        //fmt.Printf("Using default timeout [%s]\n", timeout.String())
    } else {
        timeout = time.Duration(t) * time.Second
        //fmt.Printf("Using timeout [%s]\n", timeout.String())
    }

    httpclient = &http.Client{ Transport : transport, Timeout: timeout }

    client = &Client{ client: httpclient, request: request }

    // ensure that we can change body dynamically
    request.GetBody = func() (io.ReadCloser, error) {
        r := bytes.NewReader(client.buffer.Bytes())
        return ioutil.NopCloser(r), nil
    }

    return client, nil
}


func (c *Client) BuildRequestString(request string) error {
    return c.BuildRequest(node.NewXmlS(request))
}

func (c *Client) BuildRequest(query *node.Node) error {
    var buffer *bytes.Buffer
    var data []byte
    var err error

    request := node.NewXmlS("netapp")
    request.NewAttrS("xmlns", "http://www.netapp.com/filer/admin")
    request.NewAttrS("version", "1.3")
    request.AddChild(query)

    data, err = tree.DumpXml(request)

    if err == nil {
        buffer = bytes.NewBuffer(data)
        c.buffer = buffer
        c.request.Body = ioutil.NopCloser(buffer)
        c.request.ContentLength = int64(buffer.Len())
    }
    return err
}

func (c *Client) Invoke() (*node.Node, error) {
    result, _, _, err := c.invoke(false)
    return result, err
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

func (c *Client) invoke(with_timers bool) (*node.Node, time.Duration, time.Duration, error) {

    var (
        root, result *node.Node
        response *http.Response
        start time.Time
        response_t, parse_t time.Duration
        body []byte
        status, reason string
        found bool
        err error
    )

    // issue request to server
    if with_timers {
        start = time.Now()
    }
    if response, err = c.client.Do(c.request); err != nil {
        return result, response_t, parse_t, errors.New(errors.ERR_CONNECTION, err.Error())
    }
    if with_timers {
        response_t = time.Since(start)
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
        // look for http response (might be unauthorized)
        if title := root.GetChildS("title"); title != nil {
            err = errors.New(errors.API_RESPONSE, title.GetContentS)
        } else {
        err = errors.New(errors.API_RESPONSE, "missing \"results\"")
        }
    } else if status, found = result.GetAttrValueS("status"); !found {
        err = errors.New(errors.API_RESPONSE, "missing status attribute")
    } else if status != "passed" {
        if reason, found = result.GetAttrValueS("reason"); !found {
            err = errors.New(errors.API_REQ_REJECTED, "no reason")
        } else {
            err = errors.New(errors.API_REQ_REJECTED, reason)
        }
    }

    return result, response_t, parse_t, err
}

