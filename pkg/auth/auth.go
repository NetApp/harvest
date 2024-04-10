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
	cachedPassword string
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

func (c *Credentials) password(poller *conf.Poller) (string, error) {
	if poller.CredentialsScript.Path == "" {
		return poller.Password, nil
	}
	c.authMu.Lock()
	defer c.authMu.Unlock()
	if time.Now().After(c.nextUpdate) {
		var err error
		c.cachedPassword, err = c.fetchPassword(poller)
		if err != nil {
			return "", err
		}
		c.setNextUpdate()
	}
	return c.cachedPassword, nil
}

func (c *Credentials) fetchPassword(p *conf.Poller) (string, error) {
	return c.execScript(p.CredentialsScript.Path, "credential", p.CredentialsScript.Timeout, func(ctx context.Context, path string) *exec.Cmd {
		return exec.CommandContext(ctx, path, p.Addr, p.Username) // #nosec
	})
}

func (c *Credentials) fetchCerts(p *conf.Poller) (string, error) {
	return c.execScript(p.CertificateScript.Path, "certificate", p.CertificateScript.Timeout, func(ctx context.Context, path string) *exec.Cmd {
		return exec.CommandContext(ctx, path, p.Addr) // #nosec
	})
}

func (c *Credentials) execScript(cmdPath string, kind string, timeout string, e func(ctx context.Context, path string) *exec.Cmd) (string, error) {
	lookPath, err := exec.LookPath(cmdPath)
	if err != nil {
		return "", fmt.Errorf("script lookup failed kind=%s err=%w", kind, err)
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
		return "", fmt.Errorf("script start failed script=%s kind=%s err=%w", lookPath, kind, err)
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
		return "", fmt.Errorf("script execute failed script=%s kind=%s err=%w", lookPath, kind, err)
	}
	return strings.TrimSpace(stdout.String()), nil
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

func (a PollerAuth) NewCertPool() (*x509.CertPool, error) {
	// Create a CA certificate pool and add certificate if specified
	caCertPool := x509.NewCertPool()
	if a.CaCertPath != "" {
		caCert, err := os.ReadFile(a.CaCertPath)
		if err != nil {
			return caCertPool, err
		}
		if caCert != nil {
			ok := caCertPool.AppendCertsFromPEM(caCert)
			if !ok {
				return caCertPool, fmt.Errorf("failed to append ca cert path=%s", a.CaCertPath)
			}
		}
	}
	return caCertPool, nil
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
		return PollerAuth{Username: poller.Username, Password: poller.Password, insecureTLS: insecureTLS}, nil
	}
	if poller.CredentialsScript.Path != "" {
		pass, err := c.password(poller)
		if err != nil {
			return PollerAuth{}, err
		}
		return PollerAuth{
			Username:            poller.Username,
			Password:            pass,
			HasCredentialScript: true,
			Schedule:            poller.CredentialsScript.Schedule,
			insecureTLS:         insecureTLS,
		}, nil
	}
	if poller.CredentialsFile != "" {
		err := conf.ReadCredentialFile(poller.CredentialsFile, poller)
		if err != nil {
			return PollerAuth{}, err
		}
		return PollerAuth{Username: poller.Username, Password: poller.Password, insecureTLS: insecureTLS}, nil
	}
	return PollerAuth{Username: poller.Username, insecureTLS: insecureTLS}, nil
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
		caCertPool, err := pollerAuth.NewCertPool()
		if err != nil {
			return nil, err
		}

		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
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
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: pollerAuth.insecureTLS}, //nolint:gosec
		}
	}
	return transport, err
}
