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
				Password:            "script-data-username-a.b.c",
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
			want:           PollerAuth{Username: "me", Password: "script-data-me-a.b.c", HasCredentialScript: true},
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
			want:       PollerAuth{Username: "username", Password: "script-data-username-a.b.c", HasCredentialScript: true},
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
			// #nosec G101
			want: PollerAuth{Username: "bat", Password: "script-data-bat-a.b.c", HasCredentialScript: true},
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
			name:       "poller schedule",
			pollerName: "test",
			// #nosec G101
			want:         PollerAuth{Username: "flo", Password: "script-data-flo-a.b.c", HasCredentialScript: true},
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
			name:       "defaults schedule",
			pollerName: "test",
			// #nosec G101
			want:         PollerAuth{Username: "flo", Password: "script-data-flo-a.b.c", HasCredentialScript: true},
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

// credsFromYAML decodes a harvest config and returns Credentials for the named
// poller, matching how TestCredentials_GetPollerAuth sets up its cases.
func credsFromYAML(t *testing.T, yaml string) (*Credentials, *conf.Poller) {
	t.Helper()
	conf.Config.Defaults = nil
	err := conf.DecodeConfig([]byte(yaml))
	assert.Nil(t, err)
	poller, err := conf.PollerNamed("test")
	assert.Nil(t, err)
	return NewCredentials(poller, slog.Default()), poller
}

// TestExpireForcesCredentialScriptRefetch covers GitHub issues #4140
// "The credential script for the REST Collector does not refresh the password on
// authentication errors" and #4009 "StorageGRID: Cached credential script tokens
// not expired on 401" (the latter shipped with the status/untested label).
//
// Expire is the hook every client calls after a 401. It resets the credential
// schedule so the next Password lookup re-runs the script instead of handing
// back the rejected value. Seven clients depend on it -- cmd/tools/rest,
// pkg/api/ontapi/zapi, and the storagegrid, eseries, cisco and arista REST
// clients -- and none of them could detect a regression here on their own.
//
// The two fixtures make the refetch observable: get_pass echoes "script-data-",
// get_pass2 echoes "script-alt-".
func TestExpireForcesCredentialScriptRefetch(t *testing.T) {
	c, poller := credsFromYAML(t, `
Pollers:
  test:
    addr: a.b.c
    username: flo
    credentials_script:
      path: testdata/get_pass
`)

	first, err := c.GetPollerAuth()
	assert.Nil(t, err)
	assert.Equal(t, first.Password, "script-data-flo-a.b.c")
	assert.True(t, first.HasCredentialScript)

	// Point the poller at a script that returns a different value. Without
	// expiring, the cached response must still be served -- that is the whole
	// point of the schedule.
	poller.CredentialsScript.Path = "testdata/get_pass2"

	cached, err := c.GetPollerAuth()
	assert.Nil(t, err)
	assert.Equal(t, cached.Password, "script-data-flo-a.b.c")

	// After Expire, the next lookup must re-run the script.
	c.Expire()

	refetched, err := c.GetPollerAuth()
	assert.Nil(t, err)
	assert.Equal(t, refetched.Password, "script-alt-flo-a.b.c")
}

// TestExpireIsANoOpWithoutCredentialScript pins that Expire does nothing for a
// poller using a static password. Clearing the schedule there would be
// meaningless, and Expire is called on the 401 path regardless of auth style.
func TestExpireIsANoOpWithoutCredentialScript(t *testing.T) {
	c, _ := credsFromYAML(t, `
Pollers:
  test:
    addr: a.b.c
    username: flo
    password: static-pass
`)

	first, err := c.GetPollerAuth()
	assert.Nil(t, err)
	assert.Equal(t, first.Password, "static-pass")
	assert.False(t, first.HasCredentialScript)

	c.Expire()

	after, err := c.GetPollerAuth()
	assert.Nil(t, err)
	assert.Equal(t, after.Password, "static-pass")
}

// TestExpireRefetchesAuthToken is the #4009 shape specifically: a credential
// script that returns an authToken rather than a password. The StorageGrid
// client short-circuits on a non-empty AuthToken, so a stale token has to be
// cleared through Expire or the 401 retry reuses it and fails again.
func TestExpireRefetchesAuthToken(t *testing.T) {
	c, poller := credsFromYAML(t, `
Pollers:
  test:
    addr: a.b.c
    username: flo
    credentials_script:
      path: testdata/get_credentials_authToken
`)

	first, err := c.GetPollerAuth()
	assert.Nil(t, err)
	assert.NotEqual(t, first.AuthToken, "")
	assert.True(t, first.HasCredentialScript)

	// Swap in a script that yields a password instead of a token, so the
	// refetch is observable.
	poller.CredentialsScript.Path = "testdata/get_pass"

	cachedAuth, err := c.GetPollerAuth()
	assert.Nil(t, err)
	assert.Equal(t, cachedAuth.AuthToken, first.AuthToken)

	c.Expire()

	refetched, err := c.GetPollerAuth()
	assert.Nil(t, err)
	assert.Equal(t, refetched.AuthToken, "")
	assert.Equal(t, refetched.Password, "script-data-flo-a.b.c")
}

// TestExpireIsRepeatable pins that Expire can be called more than once, and
// while no fetch is in flight, without deadlocking on authMu. The 401 paths can
// call it on consecutive polls.
func TestExpireIsRepeatable(t *testing.T) {
	c, _ := credsFromYAML(t, `
Pollers:
  test:
    addr: a.b.c
    username: flo
    credentials_script:
      path: testdata/get_pass
`)

	_, err := c.GetPollerAuth()
	assert.Nil(t, err)

	c.Expire()
	c.Expire()

	got, err := c.GetPollerAuth()
	assert.Nil(t, err)
	assert.Equal(t, got.Password, "script-data-flo-a.b.c")
}
