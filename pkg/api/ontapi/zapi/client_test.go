package zapi

import (
	"goharvest2/pkg/tree/node"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		config *node.Node
	}

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
		test{"missing_certificate_keys", certificatePollerFail, true},
		test{"correct_certificate_configuration", certificatePollerPass, false},
		test{"missing_username_password", basicAuthPollerFail, true},
		test{"correct_basic_auth_configuration", basicAuthPollerPass, false},
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
