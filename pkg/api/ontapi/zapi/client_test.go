package zapi

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"testing"
	"time"
)

func TestNew(t *testing.T) {

	type test struct {
		name    string
		config  *node.Node
		wantErr bool
	}

	certificatePollerFail := node.NewS("test")
	certificatePollerFail.NewChildS("datacenter", "cluster-01")
	certificatePollerFail.NewChildS("addr", "localhost")
	certificatePollerFail.NewChildS("auth_style", conf.CertificateAuth)
	certificatePollerFail.NewChildS("use_insecure_tls", "false")

	certificatePollerPass := node.NewS("test")
	certificatePollerPass.NewChildS("datacenter", "cluster-01")
	certificatePollerPass.NewChildS("addr", "localhost")
	certificatePollerPass.NewChildS("auth_style", conf.CertificateAuth)
	certificatePollerPass.NewChildS("use_insecure_tls", "false")
	certificatePollerPass.NewChildS("ssl_cert", "testdata/ubuntu.pem")
	certificatePollerPass.NewChildS("ssl_key", "testdata/ubuntu.key")

	basicAuthPollerFail := node.NewS("test")
	basicAuthPollerFail.NewChildS("datacenter", "cluster-01")
	basicAuthPollerFail.NewChildS("addr", "localhost")
	basicAuthPollerFail.NewChildS("auth_style", conf.BasicAuth)
	basicAuthPollerFail.NewChildS("use_insecure_tls", "false")

	basicAuthPollerPass := node.NewS("test")
	basicAuthPollerPass.NewChildS("datacenter", "cluster-01")
	basicAuthPollerPass.NewChildS("addr", "localhost")
	basicAuthPollerPass.NewChildS("auth_style", conf.BasicAuth)
	basicAuthPollerPass.NewChildS("use_insecure_tls", "false")
	basicAuthPollerPass.NewChildS("username", "username")
	basicAuthPollerPass.NewChildS("password", "password")

	tests := []test{
		{"missing_certificate_keys", certificatePollerFail, true},
		{"correct_certificate_configuration", certificatePollerPass, false},
		{"missing_username_password", basicAuthPollerFail, true},
		{"correct_basic_auth_configuration", basicAuthPollerPass, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poller := conf.ZapiPoller(tt.config)
			_, err := New(poller, auth.NewCredentials(poller, slog.Default()))
			if err != nil {
				assert.True(t, tt.wantErr)
			}
		})
	}
}

func TestClientTimeout(t *testing.T) {

	type test struct {
		name         string
		fromTemplate string
		want         time.Duration
		display      string
		hasErr       bool
	}

	timeout, _ := time.ParseDuration(DefaultTimeout)
	tests := []test{
		{"no units", "180", 180 * time.Second, "180", false},
		{"no units", "123", 123 * time.Second, "123", false},
		{"empty", "", timeout, "30s", true},
		{"zero", "0", 0 * time.Second, "0", false},
		{"with units", "5m4s", 5*time.Minute + 4*time.Second, "5m4s", false},
		{"invalid", "bob", timeout, "30s", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientTimeout, err := parseClientTimeout(tt.fromTemplate)
			if err != nil {
				assert.True(t, tt.hasErr)
				assert.Equal(t, tt.display, clientTimeout.String())
			}
			assert.Equal(t, clientTimeout, tt.want)
		})
	}
}
