package api

import (
    "fmt"
    "time"
    "net/http"
    "io"
    "io/ioutil"
    "bytes"
    "crypto/tls"
)

type ConnectionParams struct {
    Hostname string
    UseCert bool
    Authorization [2]string
    Timeout int
}

type Connection struct {
    client      *http.Client
    request     *http.Request
    buffer      *bytes.Buffer
}

func NewConnection(p ConnectionParams) (Connection, error) {
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


func (c *Connection) InvokeAPI(api string) ([]byte, error) {
    var err error
    var body []byte
    var data string
    var buffer *bytes.Buffer
    var response *http.Response

    data = `<netapp xmlns="http://www.netapp.com/filer/admin" version="1.3">` + api + `</netapp>`

    buffer = bytes.NewBuffer([]byte(data))
    c.buffer = buffer
    c.request.Body = ioutil.NopCloser(buffer)
    c.request.ContentLength = int64(buffer.Len())

    response, err = c.client.Do(c.request)
    if err != nil {
        fmt.Printf("error reading response: %s\n", err)
        return body, err
    }

    defer response.Body.Close()

    body, err = ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Printf("error reading body: %s\n", err)
    }

    return body, err
}

