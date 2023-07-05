package auth

import (
	"bytes"
	"context"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defaultSchedule = "24h"
	defaultTimeout  = "10s"
)

func NewCredentials(p *conf.Poller, logger *logging.Logger) *Credentials {
	return &Credentials{
		poller: p,
		logger: logger,
		authMu: &sync.Mutex{},
	}
}

type Credentials struct {
	poller     *conf.Poller
	nextUpdate time.Time
	logger     *logging.Logger
	authMu     *sync.Mutex
}

func (c *Credentials) Password() string {
	return c.password(c.poller)
}

// Expire will reset the credential schedule if the receiver has a CredentialsScript
// Otherwise it will do nothing.
// Resetting the schedule will cause the next call to Password to fetch the credentials
func (c *Credentials) Expire() {
	if c.poller.CredentialsScript.Path == "" {
		return
	}
	c.authMu.Lock()
	defer c.authMu.Unlock()
	c.nextUpdate = time.Time{}
	c.poller.Password = ""
}

func (c *Credentials) password(poller *conf.Poller) string {
	if poller.CredentialsScript.Path == "" {
		return poller.Password
	}
	c.authMu.Lock()
	defer c.authMu.Unlock()
	if time.Now().After(c.nextUpdate) {
		poller.Password = c.fetchPassword(poller)
		c.setNextUpdate()
	}
	return poller.Password
}

func (c *Credentials) fetchPassword(p *conf.Poller) string {
	path, err := exec.LookPath(p.CredentialsScript.Path)
	if err != nil {
		c.logger.Error().Err(err).Str("path", p.CredentialsScript.Path).Msg("Credentials script lookup failed")
		return ""
	}
	timeout := p.CredentialsScript.Timeout
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
	cmd := exec.CommandContext(ctx, path, p.Addr, p.Username)

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
			Str("script", path).
			Str("timeout", duration.String()).
			Str("stderr", stderr.String()).
			Str("stdout", stdout.String()).
			Msg("Failed to start credentials script")
		return ""
	}
	err = cmd.Wait()
	if err != nil {
		c.logger.Error().Err(err).
			Str("script", path).
			Str("timeout", duration.String()).
			Str("stderr", stderr.String()).
			Str("stdout", stdout.String()).
			Msg("Failed to execute credentials script")
		return ""
	}
	return strings.TrimSpace(stdout.String())
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
	Username string
	Password string
	IsCert   bool
}

func (c *Credentials) GetPollerAuth() (PollerAuth, error) {
	auth, err := getPollerAuth(c, c.poller)
	if err != nil {
		return PollerAuth{}, err
	}
	if auth.IsCert {
		return auth, nil
	}
	if auth.Password != "" {
		c.poller.Username = auth.Username
		c.poller.Password = auth.Password
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
	if auth.IsCert {
		return auth, nil
	}
	if auth.Username != "" {
		defaultAuth.Username = auth.Username
	}
	c.poller.Username = defaultAuth.Username
	c.poller.Password = defaultAuth.Password
	return defaultAuth, nil
}

func (c *Credentials) HasCredentialScript() bool {
	return c.poller.CredentialsScript.Path != ""
}

func getPollerAuth(c *Credentials, poller *conf.Poller) (PollerAuth, error) {
	if poller.AuthStyle == conf.CertificateAuth {
		return PollerAuth{IsCert: true}, nil
	}
	if poller.Password != "" {
		return PollerAuth{Username: poller.Username, Password: poller.Password}, nil
	}
	if poller.CredentialsScript.Path != "" {
		return PollerAuth{Username: poller.Username, Password: c.password(poller)}, nil
	}
	if poller.CredentialsFile != "" {
		err := conf.ReadCredentialFile(poller.CredentialsFile, poller)
		if err != nil {
			return PollerAuth{}, err
		}
		return PollerAuth{Username: poller.Username, Password: poller.Password}, nil
	}
	return PollerAuth{Username: poller.Username}, nil
}
