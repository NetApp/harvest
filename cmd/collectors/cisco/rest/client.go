package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultTimeout = "30s"
)

type Client struct {
	client   *http.Client
	Logger   *slog.Logger
	baseURL  string
	remote   conf.Remote
	Timeout  time.Duration
	auth     *auth.Credentials
	Metadata *collector.Metadata
}
type apiType string

const (
	cliShowArray apiType = "cli_show_array"
	cliShow      apiType = "cli_show"
)

type API struct {
	Version      string  `json:"version"`
	Type         apiType `json:"type"`
	Chunk        string  `json:"chunk"`
	Sid          string  `json:"sid"`
	Input        string  `json:"input"`
	OutputFormat string  `json:"output_format"`
}

type PostCmd struct {
	InsAPI API `json:"ins_api"`
}

// CLIShow uses the cli_show command type when talking to the switch.
// This was needed because cli_show_array does not support sending multiple commands
// (e.g. `show version ; show banner motd`)
// When sending multiple commands with type=cli_show_array, the response is invalid JSON
func (c *Client) CLIShow(command string) (gjson.Result, error) {
	return c.callAPI(command, cliShow)
}

// CLIShowArray uses the cli_show_array command type when talking to the switch.
// This was needed because the cli_show output truncated tx_pwr when calling `show interface transceiver details`
func (c *Client) CLIShowArray(command string) (gjson.Result, error) {
	return c.callAPI(command, cliShowArray)
}

func (c *Client) callAPI(command string, callType apiType) (gjson.Result, error) {
	pollerAuth, err := c.auth.GetPollerAuth()
	if err != nil {
		return gjson.Result{}, err
	}

	result, err := c.callWithAuthRetry(command, callType)

	if err != nil {
		if he, ok := errors.AsType[errs.HarvestError](err); ok {
			// If this is an auth failure and the client is using a credential script,
			// expire the current credentials, call the script again, update the client's password,
			// and try again
			if errors.Is(he, errs.ErrAuthFailed) && pollerAuth.HasCredentialScript {
				c.auth.Expire()
				return c.callWithAuthRetry(command, callType)
			}
		}
	}

	if err != nil {
		return gjson.Result{}, err
	}

	return result, nil
}

func (c *Client) callWithAuthRetry(command string, callType apiType) (gjson.Result, error) {
	cmd := API{
		Version:      "1.0",
		Type:         callType,
		Chunk:        "0",
		Sid:          "sid",
		Input:        command,
		OutputFormat: "json",
	}

	aPost := PostCmd{
		InsAPI: cmd,
	}

	jsonBytes, err := json.Marshal(aPost)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to create request: %w", err)
	}

	pollerAuth, err := c.auth.GetPollerAuth()
	if err != nil {
		return gjson.Result{}, err
	}

	req.SetBasicAuth(pollerAuth.Username, pollerAuth.Password)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cache-Control", "no-cache")

	resp, err := c.client.Do(req)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to do request: %w", err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return gjson.Result{}, errs.New(errs.ErrAuthFailed, resp.Status, errs.WithStatus(resp.StatusCode))
		}
		return gjson.Result{}, fmt.Errorf("failed to do request: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to read response: %w", err)
	}

	result := gjson.GetBytes(body, "ins_api.outputs")

	// Check if the result is an error, e.g.
	// {
	//  "output": {
	//    "clierror": "Internal Server error",
	//    "input": "show version",
	//    "msg": "Max sessions reached. If you are a new user/client, please try again later.",
	//    "code": "429"
	//  }
	// }
	code := result.Get("output.code")
	if code.Exists() && code.Int() != 200 {
		errMsg := result.Get("output.msg").String()
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return gjson.Result{}, fmt.Errorf("API call failed with code %s: %s", code.String(), errMsg)
	}
	return result, nil
}

func (c *Client) Init(retries int, remote conf.Remote) error {
	c.remote = remote

	if !remote.IsZero() {
		return nil
	}

	var (
		err             error
		output, content gjson.Result
	)

	for range retries {
		output, err = c.CLIShowArray("show version")
		if err != nil {
			if errors.Is(err, errs.ErrPermissionDenied) {
				return err
			}
			continue
		}
		content = output.Get("output.body")
		header := content.Get("header_str").ClonedString()
		if strings.Contains(header, "NX-OS") {
			c.remote.Model = "nxos"
			version := content.Get("nxos_ver_str").String()
			version = strings.Replace(version, "(", ".", 1)
			version = strings.Replace(version, ")", "", 1)
			c.remote.Version = version
		} else {
			before, _, found := strings.Cut(header, "\n")
			if found {
				return fmt.Errorf("unknown OS: %s", before)
			}
			return fmt.Errorf("unknown OS: %s", header)

		}

		c.remote.Name = content.Get("host_name").ClonedString()
		c.remote.Serial = content.Get("chassis_id").ClonedString()

		return nil
	}

	return err
}

func (c *Client) Remote() conf.Remote {
	return c.remote
}

func New(poller *conf.Poller, timeout time.Duration, credentials *auth.Credentials) (*Client, error) {
	var (
		client     Client
		httpclient *http.Client
		transport  http.RoundTripper
		addr       string
		err        error
	)

	client = Client{
		auth:     credentials,
		Metadata: &collector.Metadata{},
	}
	client.Logger = slog.Default().With(slog.String("REST", "Client"))

	if addr = poller.Addr; addr == "" {
		return nil, errs.New(errs.ErrMissingParam, "addr")
	}

	client.baseURL = "https://" + addr + "/ins"
	client.Timeout = timeout

	transport, err = credentials.Transport(nil, poller)
	if err != nil {
		return nil, err
	}

	httpclient = &http.Client{Transport: transport, Timeout: timeout}
	client.client = httpclient

	return &client, nil
}
