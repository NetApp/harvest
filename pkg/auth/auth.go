package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/third_party/mergo"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defaultSchedule = "24h"
	defaultTimeout  = "10s"
	certType        = "CERTIFICATE"
	keyType         = "PRIVATE KEY"
)

func NewCredentials(p *conf.Poller, logger *logging.Logger) *Credentials {
	return &Credentials{
		poller: p,
		logger: logger,
		authMu: &sync.Mutex{},
	}
}

type Credentials struct {
	poller         *conf.Poller
	nextUpdate     time.Time
	logger         *logging.Logger
	authMu         *sync.Mutex
	cachedResponse ScriptResponse
}

// Expire will reset the credential schedule if the receiver has a CredentialsScript
// Otherwise it will do nothing.
// Resetting the schedule will cause the next call to Password to fetch the credentials
func (c *Credentials) Expire() {
	auth, err := c.GetPollerAuth()
	if err != nil {
		return
	}
	if !auth.HasCredentialScript {
		return
	}
	c.authMu.Lock()
	defer c.authMu.Unlock()
	c.nextUpdate = time.Time{}
}

func (c *Credentials) certs(poller *conf.Poller) (string, error) {
	if poller.CertificateScript.Path == "" {
		return "", nil
	}
	c.authMu.Lock()
	defer c.authMu.Unlock()
	return c.fetchCerts(poller)
}

func (c *Credentials) password(poller *conf.Poller) (ScriptResponse, error) {
	if poller.CredentialsScript.Path == "" {
		return ScriptResponse{
			Data:     poller.Password,
			Username: poller.Username,
		}, nil
	}

	var response ScriptResponse
	var err error
	c.authMu.Lock()
	defer c.authMu.Unlock()
	if time.Now().After(c.nextUpdate) {
		response, err = c.fetchPassword(poller)
		if err != nil {
			return ScriptResponse{}, err
		}
		// Cache the new response and update the next update time.
		c.cachedResponse = response
		c.setNextUpdate()
	}
	return c.cachedResponse, nil
}

func (c *Credentials) fetchPassword(p *conf.Poller) (ScriptResponse, error) {
	response, err := c.execScript(p.CredentialsScript.Path, "credential", p.CredentialsScript.Timeout, func(ctx context.Context, path string) *exec.Cmd {
		return exec.CommandContext(ctx, path, p.Addr, p.Username) // #nosec
	})
	if err != nil {
		return ScriptResponse{}, err
	}
	// If username is empty, use harvest config poller username
	if response.Username == "" {
		response.Username = p.Username
	}
	return response, nil
}

func (c *Credentials) fetchCerts(p *conf.Poller) (string, error) {
	response, err := c.execScript(p.CertificateScript.Path, "certificate", p.CertificateScript.Timeout, func(ctx context.Context, path string) *exec.Cmd {
		return exec.CommandContext(ctx, path, p.Addr) // #nosec
	})
	if err != nil {
		return "", err
	}

	// The script is expected to return only the certificate data, so we don't need to check for a username.
	return response.Data, nil
}

type ScriptResponse struct {
	Username  string `yaml:"username"`
	Data      string `yaml:"password"`
	AuthToken string `yaml:"authToken"`
}

func (c *Credentials) execScript(cmdPath string, kind string, timeout string, e func(ctx context.Context, path string) *exec.Cmd) (ScriptResponse, error) {
	response := ScriptResponse{}
	lookPath, err := exec.LookPath(cmdPath)
	if err != nil {
		return response, fmt.Errorf("script lookup failed kind=%s err=%w", kind, err)
	}
	if timeout == "" {
		timeout = defaultTimeout
	}
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		c.logger.Error().Err(err).
			Str("timeout", timeout).
			Str("default", defaultTimeout).
			Msg("Failed to parse timeout. Using default")
		duration, _ = time.ParseDuration(defaultTimeout)
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), duration)
	defer cancelFunc()
	cmd := e(ctx, lookPath)

	// Create process group - so we can kill any forked processes
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.WaitDelay = duration
	err = cmd.Start()
	defer func() {
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	}()
	if err != nil {
		c.logger.Error().Err(err).
			Str("script", lookPath).
			Str("timeout", duration.String()).
			Str("stderr", stderr.String()).
			Str("stdout", stdout.String()).
			Str("kind", kind).
			Msg("Failed to start script")
		return response, fmt.Errorf("script start failed script=%s kind=%s err=%w", lookPath, kind, err)
	}
	err = cmd.Wait()
	if err != nil {
		c.logger.Error().Err(err).
			Str("script", lookPath).
			Str("timeout", duration.String()).
			Str("stderr", stderr.String()).
			Str("stdout", stdout.String()).
			Str("kind", kind).
			Msg("Failed to execute script")
		return response, fmt.Errorf("script execute failed script=%s kind=%s err=%w", lookPath, kind, err)
	}

	err = yaml.Unmarshal(stdout.Bytes(), &response)
	if err != nil {
		// Log the error but do not return it, we will try to use the output as plain text next.
		c.logger.Debug().Err(err).
			Str("script", lookPath).
			Str("timeout", duration.String()).
			Str("stderr", stderr.String()).
			Str("stdout", stdout.String()).
			Str("kind", kind).
			Msg("Failed to parse YAML output. Treating as plain text.")
	}

	if err == nil && (response.Data != "" || response.AuthToken != "") {
		// If parsing is successful and data is not empty, return the response.
		// Username is optional, so it's okay if it's not present.
		return response, nil
	}

	// If YAML parsing fails or the data is empty,
	// assume the output is the data (password or certificate) in plain text for backward compatibility.
	response.Data = strings.TrimSpace(stdout.String())
	return response, nil
}

func (c *Credentials) setNextUpdate() {
	schedule := c.poller.CredentialsScript.Schedule
	if schedule == "" {
		schedule = defaultSchedule
	}
	if strings.EqualFold(schedule, "always") {
		return
	}
	duration, err := time.ParseDuration(schedule)
	if err != nil {
		c.logger.Error().Err(err).
			Str("schedule", schedule).
			Str("default", defaultSchedule).
			Msg("Failed to parse schedule. Using default")
		duration, _ = time.ParseDuration(defaultSchedule)
	}
	c.nextUpdate = time.Now().Add(duration)
}

type PollerAuth struct {
	Username             string
	Password             string
	AuthToken            string
	IsCert               bool
	HasCredentialScript  bool
	HasCertificateScript bool
	Schedule             string
	PemCert              []byte
	PemKey               []byte
	CertPath             string
	KeyPath              string
	CaCertPath           string
	insecureTLS          bool
}

func (a PollerAuth) Certificate() (tls.Certificate, error) {
	if a.HasCertificateScript {
		return tls.X509KeyPair(a.PemCert, a.PemKey)
	}
	if a.CertPath == "" {
		return tls.Certificate{}, errs.New(errs.ErrMissingParam, "ssl_cert")
	}
	if a.KeyPath == "" {
		return tls.Certificate{}, errs.New(errs.ErrMissingParam, "ssl_key")
	}
	return tls.LoadX509KeyPair(a.CertPath, a.KeyPath)
}

// If the CA certificate path is specified, create a CA certificate pool and add the certificate to it.
// Otherwise, return nil so the host's root CA set is used.
func (a PollerAuth) loadCertPool(logger *logging.Logger) *x509.CertPool {
	if a.CaCertPath == "" {
		return nil
	}

	caCert, err := os.ReadFile(a.CaCertPath)
	if err != nil {
		logger.Error().Err(err).Str("caCertPath", a.CaCertPath).Msg("Failed to read CA certificate. Use host's root CA set.")
		return nil
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		logger.Warn().Str("caCertPath", a.CaCertPath).Msg("Failed to append CA certificate to pool. Use host's root CA set.")
		return nil
	}
	logger.Debug().Str("caCertPath", a.CaCertPath).Msg("CA certificate loaded")
	return caCertPool
}

func (c *Credentials) GetPollerAuth() (PollerAuth, error) {
	auth, err := getPollerAuth(c, c.poller)
	if err != nil {
		return PollerAuth{}, err
	}
	if auth.IsCert {
		return auth, nil
	}
	if auth.Username != "" && auth.Password != "" {
		return auth, nil
	}

	if conf.Config.Defaults == nil {
		return auth, nil
	}

	copyDefault := *conf.Config.Defaults
	copyDefault.Name = c.poller.Name
	copyDefault.Addr = c.poller.Addr
	if c.poller.Username != "" {
		copyDefault.Username = c.poller.Username
	}
	defaultAuth, err := getPollerAuth(c, &copyDefault)
	if err != nil {
		return PollerAuth{}, err
	}
	_ = mergo.Merge(&auth, defaultAuth)
	return auth, nil
}

func getPollerAuth(c *Credentials, poller *conf.Poller) (PollerAuth, error) {
	// by default, enforce secure TLS
	insecureTLS := false
	if poller.UseInsecureTLS != nil {
		insecureTLS = *poller.UseInsecureTLS
	}
	if poller.AuthStyle == conf.CertificateAuth {
		return handCertificateAuth(c, poller, insecureTLS)
	}
	if poller.Password != "" {
		return PollerAuth{
			Username:    poller.Username,
			Password:    poller.Password,
			insecureTLS: insecureTLS,
			CaCertPath:  poller.CaCertPath,
		}, nil
	}
	if poller.CredentialsScript.Path != "" {
		response, err := c.password(poller)
		if err != nil {
			return PollerAuth{}, err
		}
		return PollerAuth{
			Username:            response.Username,
			Password:            response.Data,
			AuthToken:           response.AuthToken,
			HasCredentialScript: true,
			Schedule:            poller.CredentialsScript.Schedule,
			insecureTLS:         insecureTLS,
			CaCertPath:          poller.CaCertPath,
		}, nil
	}
	if poller.CredentialsFile != "" {
		err := conf.ReadCredentialFile(poller.CredentialsFile, poller)
		if err != nil {
			return PollerAuth{}, err
		}
		return PollerAuth{
			Username:    poller.Username,
			Password:    poller.Password,
			insecureTLS: insecureTLS,
			CaCertPath:  poller.CaCertPath,
		}, nil
	}
	return PollerAuth{
		Username:    poller.Username,
		insecureTLS: insecureTLS,
		CaCertPath:  poller.CaCertPath,
	}, nil
}

func handCertificateAuth(c *Credentials, poller *conf.Poller, insecureTLS bool) (PollerAuth, error) {
	if poller.CertificateScript.Path != "" {
		certBlob, err := c.certs(poller)
		if err != nil {
			return PollerAuth{}, err
		}
		cert, key, err := extractCertAndKey(certBlob)
		if err != nil {
			return PollerAuth{}, err
		}
		return PollerAuth{
			IsCert:               true,
			HasCertificateScript: true,
			PemCert:              cert,
			PemKey:               key,
			insecureTLS:          insecureTLS,
			CaCertPath:           poller.CaCertPath,
		}, nil
	}

	var pathPrefix string
	certPath := poller.SslCert
	keyPath := poller.SslKey

	if certPath == "" || keyPath == "" {
		o := options.New()
		pathPrefix = path.Join(o.HomePath, "cert/", o.Hostname)
	}

	if certPath == "" {
		certPath = pathPrefix + ".pem"
	}
	if keyPath == "" {
		keyPath = pathPrefix + ".key"
	}
	return PollerAuth{
		IsCert:      true,
		CertPath:    certPath,
		KeyPath:     keyPath,
		CaCertPath:  poller.CaCertPath,
		insecureTLS: insecureTLS,
	}, nil
}

func extractCertAndKey(blob string) ([]byte, []byte, error) {
	block1, rest := pem.Decode([]byte(blob))
	block2, _ := pem.Decode(rest)

	if block1 == nil {
		return nil, nil, errors.New("PEM block1 is nil")
	}
	if block2 == nil {
		return nil, nil, errors.New("PEM block2 is nil")
	}

	if block1.Type == certType && block2.Type == keyType {
		return bytes.TrimSpace(pem.EncodeToMemory(block1)), bytes.TrimSpace(pem.EncodeToMemory(block2)), nil
	}
	if block1.Type == keyType && block2.Type == certType {
		return bytes.TrimSpace(pem.EncodeToMemory(block2)), bytes.TrimSpace(pem.EncodeToMemory(block1)), nil
	}

	return nil, nil, fmt.Errorf("unexpected PEM block1Type=%s block2Type=%s", block1.Type, block2.Type)
}

func (c *Credentials) Transport(request *http.Request) (*http.Transport, error) {
	var (
		cert      tls.Certificate
		transport *http.Transport
	)

	pollerAuth, err := c.GetPollerAuth()
	if err != nil {
		return nil, err
	}

	if pollerAuth.IsCert {
		cert, err = pollerAuth.Certificate()
		if err != nil {
			return nil, err
		}

		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				RootCAs:            pollerAuth.loadCertPool(c.logger),
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: pollerAuth.insecureTLS, //nolint:gosec
			},
		}
	} else {
		password := pollerAuth.Password
		if pollerAuth.Username == "" {
			return nil, errs.New(errs.ErrMissingParam, "username")
		} else if password == "" {
			return nil, errs.New(errs.ErrMissingParam, "password")
		}

		if request != nil {
			request.SetBasicAuth(pollerAuth.Username, password)
		}
		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				RootCAs:            pollerAuth.loadCertPool(c.logger),
				InsecureSkipVerify: pollerAuth.insecureTLS, //nolint:gosec
			},
		}
	}
	return transport, err
}
