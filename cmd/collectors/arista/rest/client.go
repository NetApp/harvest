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
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"io"
	"log/slog"
	"net/http"
	"regexp"
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

// params are the parameters for an Arista eAPI JSON-RPC runCmds call.
type params struct {
	Version int      `json:"version"`
	Cmds    []string `json:"cmds"`
	Format  string   `json:"format"`
}

// PostCmd is the JSON-RPC 2.0 request envelope sent to the Arista eAPI endpoint.
type PostCmd struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  params `json:"params"`
	ID      string `json:"id"`
}

// RunCmds sends one or more CLI show commands to the switch using the eAPI
// runCmds method and returns the gjson result rooted at the "result" array.
// The result array has one entry per command, in the order the commands were sent.
func (c *Client) RunCmds(commands ...string) (gjson.Result, error) {
	return c.callAPI(commands)
}

// SplitCommands splits a template query string into individual CLI commands.
// Commands are separated by a semicolon, e.g. "show version ; show hostname".
func SplitCommands(query string) []string {
	parts := strings.Split(query, ";")
	commands := make([]string, 0, len(parts))
	for _, p := range parts {
		if cmd := strings.TrimSpace(p); cmd != "" {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// TrimQuotes removes surrounding escaped quotes that EOS sometimes wraps around
// values such as the LLDP interfaceId (e.g. ""Eth103/1/41"").
func TrimQuotes(s string) string {
	return strings.Trim(s, `"`)
}

func (c *Client) callAPI(commands []string) (gjson.Result, error) {
	pollerAuth, err := c.auth.GetPollerAuth()
	if err != nil {
		return gjson.Result{}, err
	}

	result, err := c.callWithAuthRetry(commands)

	if err != nil {
		if he, ok := errors.AsType[errs.HarvestError](err); ok {
			// If this is an auth failure and the client is using a credential script,
			// expire the current credentials, call the script again, update the client's password,
			// and try again
			if errors.Is(he, errs.ErrAuthFailed) && pollerAuth.HasCredentialScript {
				c.auth.Expire()
				return c.callWithAuthRetry(commands)
			}
		}
		return gjson.Result{}, err
	}

	return result, nil
}

func (c *Client) callWithAuthRetry(commands []string) (gjson.Result, error) {
	cmd := PostCmd{
		JSONRPC: "2.0",
		Method:  "runCmds",
		Params: params{
			Version: 1,
			Cmds:    commands,
			Format:  "json",
		},
		ID: "1",
	}

	jsonBytes, err := json.Marshal(cmd)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to marshal data: %w", err)
	}

	doRequest := func() (*http.Response, error) {
		req, reqErr := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewBuffer(jsonBytes))
		if reqErr != nil {
			return nil, fmt.Errorf("failed to create request: %w", reqErr)
		}

		pollerAuth, authErr := c.auth.GetPollerAuth()
		if authErr != nil {
			return nil, authErr
		}

		req.SetBasicAuth(pollerAuth.Username, pollerAuth.Password)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Cache-Control", "no-cache")

		return c.client.Do(req)
	}

	resp, err := doRequest()
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to do request: %w", err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return gjson.Result{}, errs.New(errs.ErrAuthFailed, resp.Status, errs.WithStatus(resp.StatusCode))
		}
		return gjson.Result{}, fmt.Errorf("failed to do request: statusCode=%d status=%s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return gjson.Result{}, fmt.Errorf("failed to read response: %w", err)
	}

	parsed := gjson.ParseBytes(body)

	// Check for a JSON-RPC error object, e.g.
	// {
	//   "jsonrpc": "2.0",
	//   "id": "1",
	//   "error": {
	//     "code": 1002,
	//     "message": "CLI command 1 of 1 'show banner motd' failed: invalid command",
	//     "data": [{"errors": ["Invalid input (privileged mode required)"]}]
	//   }
	// }
	apiErr := parsed.Get("error")
	if apiErr.Exists() {
		code := apiErr.Get("code").Int()
		errMsg := apiErr.Get("message").ClonedString()
		if detail := apiErr.Get("data.#.errors.0").Array(); len(detail) > 0 {
			errMsg = errMsg + ": " + detail[len(detail)-1].ClonedString()
		}
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return gjson.Result{}, fmt.Errorf("API call failed with code %d: %s", code, errMsg)
	}

	return parsed.Get("result"), nil
}

func (c *Client) Init(retries int, remote conf.Remote) error {
	c.remote = remote

	if !remote.IsZero() {
		return nil
	}

	var (
		err     error
		output  gjson.Result
		version gjson.Result
	)

	for range retries {
		// "show hostname" is sent alongside "show version" because the hostname
		// is not part of the "show version" output on EOS.
		output, err = c.RunCmds("show version", "show hostname")
		if err != nil {
			if errors.Is(err, errs.ErrPermissionDenied) {
				return err
			}
			continue
		}

		version = output.Get("0")
		model := version.Get("modelName").ClonedString()
		if model == "" {
			c.Logger.Warn("Unable to determine model from output. No modelName field")
		}

		c.remote.Model = "eos"
		c.remote.Version = aristaVersion(version.Get("version").ClonedString())
		c.remote.Serial = version.Get("serialNumber").ClonedString()
		c.remote.Name = output.Get("1.hostname").ClonedString()
		if c.remote.Name == "" {
			c.remote.Name = output.Get("1.fqdn").ClonedString()
		}

		return nil
	}

	return err
}

// 4.18.5M     => 4.18.5
// 4.30.2F     => 4.30.2
var versionRe = regexp.MustCompile(`^(\d+\.\d+\.\d+)`)

func aristaVersion(raw string) string {
	submatch := versionRe.FindStringSubmatch(raw)
	if len(submatch) < 2 {
		return raw
	}
	return submatch[1]
}

func (c *Client) Remote() conf.Remote {
	return c.remote
}

func New(poller *conf.Poller, credentials *auth.Credentials) (*Client, error) {
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

	client.baseURL = "https://" + addr + "/command-api"

	transport, err = credentials.Transport(nil, poller)
	if err != nil {
		return nil, err
	}

	timeout, _ := time.ParseDuration(DefaultTimeout)
	if poller.ClientTimeout != "" {
		duration, err := time.ParseDuration(poller.ClientTimeout)
		if err == nil {
			timeout = duration
		} else {
			client.Logger.Warn("Invalid client timeout, using default",
				slog.String("configured_timeout", poller.ClientTimeout),
				slog.String("default_timeout", timeout.String()),
				slogx.Err(err),
			)
		}
	}

	client.Timeout = timeout
	httpclient = &http.Client{Transport: transport, Timeout: timeout}
	client.client = httpclient

	return &client, nil
}
