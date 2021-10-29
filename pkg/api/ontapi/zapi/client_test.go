package zapi

import (
	"goharvest2/pkg/tree/node"
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
	certificatePollerFail.NewChildS("auth_style", "certificate_auth")
	certificatePollerFail.NewChildS("use_insecure_tls", "false")

	certificatePollerPass := node.NewS("test")
	certificatePollerPass.NewChildS("datacenter", "cluster-01")
	certificatePollerPass.NewChildS("addr", "localhost")
	certificatePollerPass.NewChildS("auth_style", "certificate_auth")
	certificatePollerPass.NewChildS("use_insecure_tls", "false")
	certificatePollerPass.NewChildS("ssl_cert", "testdata/ubuntu.pem")
	certificatePollerPass.NewChildS("ssl_key", "testdata/ubuntu.key")

	basicAuthPollerFail := node.NewS("test")
	basicAuthPollerFail.NewChildS("datacenter", "cluster-01")
	basicAuthPollerFail.NewChildS("addr", "localhost")
	basicAuthPollerFail.NewChildS("auth_style", "basic_auth")
	basicAuthPollerFail.NewChildS("use_insecure_tls", "false")

	basicAuthPollerPass := node.NewS("test")
	basicAuthPollerPass.NewChildS("datacenter", "cluster-01")
	basicAuthPollerPass.NewChildS("addr", "localhost")
	basicAuthPollerPass.NewChildS("auth_style", "basic_auth")
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
			_, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestClientTimeout(t *testing.T) {

	type test struct {
		name         string
		fromTemplate string
		want         time.Duration
		hasErr       bool
	}

	tests := []test{
		{"no units", "180", 180 * time.Second, false},
		{"no units", "123", 123 * time.Second, false},
		{"empty", "", DefaultTimeout * time.Second, true},
		{"zero", "0", 0 * time.Second, false},
		{"with units", "5m4s", 5*time.Minute + 4*time.Second, false},
		{"invalid", "bob", DefaultTimeout * time.Second, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout, err := parseClientTimeout(tt.fromTemplate)
			if err != nil && !tt.hasErr {
				t.Errorf("parseClientTimeout() error = %v", err)
			}
			if timeout != tt.want {
				t.Errorf("parseClientTimeout got=[%s] want=[%s]", timeout.String(), tt.want.String())
				return
			}
		})
	}
}
