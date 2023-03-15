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

var once sync.Once
var auth *Credentials

func Get() *Credentials {
	return auth
}

func NewCredentials(p *conf.Poller, logger *logging.Logger) {
	once.Do(func() {
		if auth == nil {
			auth = &Credentials{
				poller: p,
				logger: logger,
				authMu: &sync.Mutex{},
			}
		}
	})
}

type Credentials struct {
	poller     *conf.Poller
	nextUpdate time.Time
	logger     *logging.Logger
	authMu     *sync.Mutex
}

func (c *Credentials) Password() string {
	if c.poller.CredentialsScript.Path == "" {
		return c.poller.Password
	}
	c.authMu.Lock()
	defer c.authMu.Unlock()
	if time.Now().After(c.nextUpdate) {
		c.poller.Password = c.fetchPassword()
		c.setNextUpdate()
	}
	return c.poller.Password
}

func (c *Credentials) fetchPassword() string {
	path, err := exec.LookPath(c.poller.CredentialsScript.Path)
	if err != nil {
		c.logger.Error().Err(err).Str("path", c.poller.CredentialsScript.Path).Msg("Credentials script lookup failed")
		return ""
	}
	timeout := c.poller.CredentialsScript.Timeout
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
	cmd := exec.CommandContext(ctx, path, c.poller.Addr, c.poller.Username)

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

// TestNewCredentials is used by testing code to reload auth
func TestNewCredentials(p *conf.Poller, logger *logging.Logger) *Credentials {
	auth = &Credentials{
		poller: p,
		logger: logger,
		authMu: &sync.Mutex{},
	}
	return auth
}
