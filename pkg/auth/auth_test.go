package auth

import (
	"github.com/netapp/harvest/v2/assert"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/pkg/conf"
)

func TestCredentials_GetPollerAuth(t *testing.T) {
	type test struct {
		name           string
		pollerName     string
		yaml           string
		want           PollerAuth
		wantErr        bool
		wantSchedule   string
		defaultDefined bool
	}
	tests := []test{
		{
			name:           "no default, poller credentials_file",
			pollerName:     "test",
			want:           PollerAuth{Username: "username", Password: "from-secrets-file"},
			defaultDefined: false,
			yaml: `
Pollers:
  test:
    addr: a.b.c
    username: username
    credentials_file: testdata/secrets.yaml`,
		},

		{
			name:           "poller credentials_file",
			pollerName:     "test",
			want:           PollerAuth{Username: "username", Password: "from-secrets-file"},
			defaultDefined: true,
			yaml: `
Defaults:
  auth_style: certificate_auth
  credentials_file: secrets/openlab
  username: me
  password: pass
  credentials_script:
    path: ../get_pass
Pollers:
  test:
    addr: a.b.c
    username: username
    credentials_file: testdata/secrets.yaml`,
		},

		{
			name:           "poller username from credentials_file defaults",
			pollerName:     "test",
			want:           PollerAuth{Username: "default-user", Password: "from-secrets-file"},
			defaultDefined: true,
			yaml: `
Defaults:
  auth_style: certificate_auth
  credentials_file: secrets/openlab
Pollers:
  test:
    addr: a.b.c
    credentials_file: testdata/secrets.yaml`,
		},

		{
			name:           "poller username from credentials_file, password from poller",
			pollerName:     "test",
			want:           PollerAuth{Username: "default-user", Password: "moon"},
			defaultDefined: true,
			yaml: `
Defaults:
  credentials_file: testdata/secrets.yaml
Pollers:
  test:
    addr: a.b.c
    password: moon`,
		},

		{
			name:           "poller username from credentials_file",
			pollerName:     "test2",
			want:           PollerAuth{Username: "test2-user", Password: "from-secrets-file"},
			defaultDefined: true,
			yaml: `
Defaults:
  auth_style: certificate_auth
  credentials_file: secrets/openlab
Pollers:
  test2:
    addr: a.b.c
    credentials_file: testdata/secrets.yaml`,
		},

		{
			name:           "default cert_auth",
			pollerName:     "test",
			want:           PollerAuth{Username: "username", IsCert: true, CertPath: "my_cert", KeyPath: "my_key"},
			defaultDefined: true,
			yaml: `
Defaults:
  auth_style: certificate_auth
  ssl_cert: my_cert
  ssl_key: my_key
  credentials_file: secrets/openlab
  username: me
  password: pass
  credentials_script:
    path: ../get_pass
Pollers:
  test:
    addr: a.b.c
    username: username`,
		},

		{
			name:           "poller user/pass",
			pollerName:     "test",
			want:           PollerAuth{Username: "username", Password: "pass", IsCert: false},
			defaultDefined: true,
			yaml: `
Defaults:
  auth_style: certificate_auth
  credentials_file: secrets/openlab
  username: me
  password: pass
  credentials_script:
    path: ../get_pass
Pollers:
  test:
    addr: a.b.c
    username: username
    password: pass`,
		},

		{
			name:           "default username",
			pollerName:     "test",
			want:           PollerAuth{Username: "me", Password: "pass2"},
			defaultDefined: true,
			yaml: `
Defaults:
  auth_style: certificate_auth
  credentials_file: secrets/openlab
  username: me
  password: pass
  credentials_script:
    path: ../get_pass
Pollers:
  test:
    addr: a.b.c
    password: pass2`,
		},

		{
			name:       "default credentials_script",
			pollerName: "test",
			want: PollerAuth{
				Username:            "username",
				Password:            "addr=a.b.c user=username",
				IsCert:              false,
				HasCredentialScript: true,
			},
			defaultDefined: true,
			yaml: `
Defaults:
  username: me
  credentials_script:
    path: testdata/get_pass
Pollers:
  test:
    addr: a.b.c
    username: username`,
		},

		{
			name:           "credentials_script with default username",
			pollerName:     "test",
			want:           PollerAuth{Username: "me", Password: "addr=a.b.c user=me", HasCredentialScript: true},
			defaultDefined: true,
			yaml: `
Defaults:
  username: me
  credentials_script:
    path: testdata/get_pass
Pollers:
  test:
    addr: a.b.c`,
		},

		{
			name:       "no default",
			pollerName: "test",
			want:       PollerAuth{Username: "username", Password: "addr=a.b.c user=username", HasCredentialScript: true},
			yaml: `
Pollers:
  test:
    addr: a.b.c
    credentials_script:
      path: testdata/get_pass
    username: username`,
		},

		{
			name:       "none",
			pollerName: "test",
			want:       PollerAuth{Username: "", Password: "", IsCert: false},
			yaml: `
Pollers:
  test:
    addr: a.b.c`,
		},

		{
			name:       "credentials_file missing poller",
			pollerName: "missing",
			want:       PollerAuth{Username: "default-user", Password: "default-pass", IsCert: false},
			yaml: `
Pollers:
  missing:
    addr: a.b.c
    credentials_file: testdata/secrets.yaml`,
		},

		{
			name:       "with cred",
			pollerName: "test",
			want:       PollerAuth{IsCert: true, CertPath: "my_cert", KeyPath: "my_key"},
			yaml: `
Defaults:
  use_insecure_tls: true
  prefer_zapi: true
Pollers:
  test:
    addr: a.b.c
    auth_style: certificate_auth
    ssl_cert: my_cert
    ssl_key: my_key`,
		},

		{
			name:       "poller and default credentials_script",
			pollerName: "test",
			want:       PollerAuth{Username: "bat", Password: "addr=a.b.c user=bat", HasCredentialScript: true},
			yaml: `
Defaults:
  use_insecure_tls: true
  prefer_zapi: true
  credentials_script:
    path: testdata/get_pass2
Pollers:
  test:
    addr: a.b.c
    username: bat
    credentials_script:
      path: testdata/get_pass
`,
		},

		{
			name:         "poller schedule",
			pollerName:   "test",
			want:         PollerAuth{Username: "flo", Password: "addr=a.b.c user=flo", HasCredentialScript: true},
			wantSchedule: "15m",
			yaml: `
Defaults:
  use_insecure_tls: true
  prefer_zapi: true
  credentials_script:
    path: testdata/get_pass
    schedule: 45m
Pollers:
  test:
    addr: a.b.c
    username: flo
    credentials_script:
      path: testdata/get_pass
      schedule: 15m
`,
		},

		{
			name:         "defaults schedule",
			pollerName:   "test",
			want:         PollerAuth{Username: "flo", Password: "addr=a.b.c user=flo", HasCredentialScript: true},
			wantSchedule: "42m",
			yaml: `
Defaults:
  use_insecure_tls: true
  prefer_zapi: true
  credentials_script:
    schedule: 42m
Pollers:
  test:
    addr: a.b.c
    username: flo
    credentials_script:
      path: testdata/get_pass
`,
		},

		{
			name:         "password with space",
			pollerName:   "test",
			want:         PollerAuth{Username: "flo", Password: "abc def"},
			wantSchedule: "42m",
			yaml: `
Pollers:
  test:
    addr: a.b.c
    username: flo
    password: abc def
`,
		},

		{
			name:       "certificate_script in poller",
			pollerName: "test",
			want: PollerAuth{
				IsCert: true, PemCert: []byte(`-----BEGIN CERTIFICATE-----
SSA8MyBIYXJ2ZXN0
-----END CERTIFICATE-----`), PemKey: []byte(`-----BEGIN PRIVATE KEY-----
c3VwZXIgc2VjcmV0
-----END PRIVATE KEY-----`),
			},
			yaml: `
Pollers:
  test:
    auth_style: certificate_auth
    addr: a.b.c
    certificate_script:
            path: testdata/get_cert
`,
		},

		{
			name:       "certificate_script in defaults",
			pollerName: "test",
			want: PollerAuth{
				IsCert: true, PemCert: []byte(`-----BEGIN CERTIFICATE-----
SSA8MyBIYXJ2ZXN0
-----END CERTIFICATE-----`), PemKey: []byte(`-----BEGIN PRIVATE KEY-----
c3VwZXIgc2VjcmV0
-----END PRIVATE KEY-----`),
			},
			yaml: `
Defaults:
  certificate_script:
    path: testdata/get_cert
Pollers:
  test:
    auth_style: certificate_auth
    addr: a.b.c
`,
		},

		{
			name:       "certificate_script in both",
			pollerName: "test",
			want: PollerAuth{
				IsCert: true, PemCert: []byte(`-----BEGIN CERTIFICATE-----
SSA8MyBIYXJ2ZXN0
-----END CERTIFICATE-----`), PemKey: []byte(`-----BEGIN PRIVATE KEY-----
c3VwZXIgc2VjcmV0
-----END PRIVATE KEY-----`),
			},
			yaml: `
Defaults:
  certificate_script:
    path: testdata/get_cert2
Pollers:
  test:
    auth_style: certificate_auth
    addr: a.b.c
    certificate_script:
      path: testdata/get_cert
`,
		},

		{
			name:       "ssl_cert and ssl_key defaults",
			pollerName: "test",
			want:       PollerAuth{IsCert: true, CertPath: "ssl_cert", KeyPath: "ssl_key"},
			yaml: `
Defaults:
    ssl_cert: ssl_cert
    ssl_key: ssl_key
Pollers:
    test:
      auth_style: certificate_auth
      addr: a.b.c
`,
		},

		{
			name:       "certificate_auth with ssl_cert in both",
			pollerName: "test",
			want:       PollerAuth{IsCert: true, CertPath: "ssl_cert", KeyPath: "ssl_key"},
			yaml: `
Defaults:
    ssl_cert: default_ssl_cert
    ssl_key: default_ssl_cert
Pollers:
    test:
      auth_style: certificate_auth
      addr: a.b.c
      ssl_cert: ssl_cert
      ssl_key: ssl_key
`,
		},

		{
			name:       "optional ssl_cert and ssl_key",
			pollerName: "test",
			want:       PollerAuth{IsCert: true, CertPath: "cert/cgrindst-mac-0.pem", KeyPath: "cert/cgrindst-mac-0.key"},
			yaml: `
Pollers:
    test:
      auth_style: certificate_auth
      addr: a.b.c
`,
		},

		{
			name:           "poller user/pass with caCert",
			pollerName:     "test",
			want:           PollerAuth{Username: "username", Password: "pass", IsCert: false, CaCertPath: "testdata/ca.pem"},
			defaultDefined: true,
			yaml: `
Pollers:
  test:
    addr: a.b.c
    username: username
    password: pass
    ca_cert: testdata/ca.pem`,
		},
		{
			name:       "credentials_script returns username and password in YAML",
			pollerName: "test",
			want: PollerAuth{
				Username:            "script-username",
				Password:            "script-password",
				HasCredentialScript: true,
			},
			yaml: `
Pollers:
  test:
    addr: a.b.c
    credentials_script:
      path: testdata/get_credentials_yaml
`,
		},

		{
			name:       "credentials_script returns only password in plain text",
			pollerName: "test",
			want: PollerAuth{
				Username:            "username", // Fallback to the username provided in the poller configuration
				Password:            "plain-text-password",
				HasCredentialScript: true,
			},
			yaml: `
Pollers:
  test:
    addr: a.b.c
    username: username
    credentials_script:
      path: testdata/get_password_plain
`,
		},
		{
			name:       "credentials_script returns only password in YAML format",
			pollerName: "test",
			want: PollerAuth{
				Username:            "username", // Fallback to the username provided in the poller configuration
				Password:            "password #\"`!@#$%^&*()-=[]|:'<>/ password",
				HasCredentialScript: true,
			},
			yaml: `
Pollers:
  test:
    addr: a.b.c
    username: username
    credentials_script:
      path: testdata/get_credentials_yaml_password
`,
		},
		{
			name:       "credentials_script returns username and password in YAML, no username in poller config",
			pollerName: "test",
			want: PollerAuth{
				Username:            "script-username",
				Password:            "script-password",
				HasCredentialScript: true,
			},
			yaml: `
Pollers:
  test:
    addr: a.b.c
    credentials_script:
      path: testdata/get_credentials_yaml
`,
		},

		{
			name:       "credentials_script returns only password in plain text, no username in poller config",
			pollerName: "test",
			want: PollerAuth{
				Username:            "", // No username provided, so it should be empty
				Password:            "plain-text-password",
				HasCredentialScript: true,
			},
			yaml: `
Pollers:
  test:
    addr: a.b.c
    credentials_script:
      path: testdata/get_password_plain
`,
		},

		{
			name:       "credentials_script returns username and password in YAML via Heredoc",
			pollerName: "test",
			want: PollerAuth{
				Username:            "myuser",
				Password:            "my # password",
				HasCredentialScript: true,
			},
			yaml: `
Pollers:
  test:
    addr: a.b.c
    username: username
    credentials_script:
      path: testdata/get_credentials_yaml_heredoc
`,
		},

		{
			name:       "credentials_script returns authToken",
			pollerName: "test",
			want: PollerAuth{
				AuthToken:           "abcd",
				HasCredentialScript: true,
			},
			yaml: `
Pollers:
  test:
    addr: a.b.c
    credentials_script:
      path: testdata/get_credentials_authToken
`,
		},

		{
			name:       "credentials_script returns authToken and password",
			pollerName: "test",
			want: PollerAuth{
				AuthToken:           "abcd",
				HasCredentialScript: true,
				Password:            "script-password",
			},
			yaml: `
Pollers:
  test:
    addr: a.b.c
    credentials_script:
      path: testdata/get_credentials_authToken_password
`,
		},
	}

	hostname, err := os.Hostname()
	if err != nil {
		t.Errorf("failed to get hostname err: %v", err)
	}
	hostCertPath := "cert/" + hostname + ".pem"
	hostKeyPath := "cert/" + hostname + ".key"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.Config.Defaults = nil
			if tt.defaultDefined {
				conf.Config.Defaults = &conf.Poller{}
			}
			err := conf.DecodeConfig([]byte(tt.yaml))
			assert.Nil(t, err)
			poller, err := conf.PollerNamed(tt.pollerName)
			assert.Nil(t, err)
			c := NewCredentials(poller, slog.Default())
			got, err := c.GetPollerAuth()
			if err != nil {
				assert.True(t, tt.wantErr)
				return
			}
			assert.Equal(t, got.Username, tt.want.Username)
			assert.Equal(t, got.Password, tt.want.Password)
			assert.Equal(t, got.AuthToken, tt.want.AuthToken)
			assert.Equal(t, got.IsCert, tt.want.IsCert)
			assert.Equal(t, got.HasCredentialScript, tt.want.HasCredentialScript)

			diff1 := cmp.Diff(tt.want.PemCert, got.PemCert)
			if diff1 != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff1)
			}
			diff2 := cmp.Diff(tt.want.PemKey, got.PemKey)
			if diff2 != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff2)
			}
			if tt.want.CertPath != got.CertPath && got.CertPath != hostCertPath {
				assert.Equal(t, got.CertPath, hostCertPath)
			}
			if tt.want.KeyPath != got.KeyPath && got.KeyPath != hostKeyPath {
				assert.Equal(t, got.KeyPath, hostKeyPath)
			}
			assert.Equal(t, got.CaCertPath, tt.want.CaCertPath)
		})
	}
}
