package auth

import (
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"strings"
	"testing"
)

func TestCredentials_GetPollerAuth(t *testing.T) {
	type test struct {
		name           string
		pollerName     string
		yaml           string
		want           PollerAuth
		wantErr        bool
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
			name:           "default cert_auth",
			pollerName:     "test",
			want:           PollerAuth{Username: "username", Password: "", IsCert: true},
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
			name:           "default credentials_script",
			pollerName:     "test",
			want:           PollerAuth{Username: "username", Password: "pass-from-script", IsCert: false},
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
			name:       "no default",
			pollerName: "test",
			want:       PollerAuth{Username: "username", Password: "pass-from-script", IsCert: false},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.Config.Defaults = nil
			if tt.defaultDefined {
				conf.Config.Defaults = &conf.Poller{}
			}
			err := conf.DecodeConfig(toYaml(tt.yaml))
			if err != nil {
				t.Errorf("expected no error got %+v", err)
				return
			}
			poller, err := conf.PollerNamed(tt.pollerName)
			if err != nil {
				t.Errorf("expected no error got %+v", err)
				return
			}
			c := NewCredentials(poller, logging.Get())
			got, err := c.GetPollerAuth()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPollerAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want.Username != got.Username {
				t.Errorf("got username=[%s], want username=[%s]", got.Username, tt.want.Username)
			}
			if tt.want.Password != got.Password {
				t.Errorf("got password=[%s], want password=[%s]", got.Password, tt.want.Password)
			}
			if tt.want.Username != poller.Username {
				t.Errorf("poller got username=[%s], want username=[%s]", poller.Username, tt.want.Username)
			}
			if tt.want.Password != poller.Password {
				t.Errorf("poller got password=[%s], want password=[%s]", poller.Password, tt.want.Password)
			}
		})
	}
}

func toYaml(s string) []byte {
	all := strings.ReplaceAll(s, "\t", " ")
	return []byte(all)
}
