package doctor

import (
	"github.com/netapp/harvest/v2/pkg/conf"
	"strings"
	"testing"
)

func TestRedaction(t *testing.T) {
	assertRedacted(t, `foo: bar`, `foo: bar`)
	assertRedacted(t, `username: pass`, `username: -REDACTED-`)
	assertRedacted(t, `password: f`, `password: -REDACTED-`)
	assertRedacted(t, `grafana_api_token: secret`, `grafana_api_token: -REDACTED-`)
	assertRedacted(t, `token: secret`, `token: -REDACTED-`)
	assertRedacted(t, "# foo\nusername: pass\n#foot", `username: -REDACTED-`)
	assertRedacted(t, `host: 1.2.3.4`, `host: -REDACTED-`)
	assertRedacted(t, `addr: 1.2.3.4`, `addr: -REDACTED-`)
	assertRedacted(t, "auth_style: password\nusername: cat", "auth_style: password\nusername: -REDACTED-")
}

func assertRedacted(t *testing.T, input, redacted string) {
	t.Helper()
	redacted = strings.TrimSpace(redacted)
	input = printRedactedConfig("test", []byte(input))
	input = strings.TrimSpace(input)
	if input != redacted {
		t.Fatalf(`input=[%s] != redacted=[%s]`, input, redacted)
	}
}

func TestConfigToStruct(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/testConfig.yml")
	if conf.Config.Defaults.Password != "123#abc" {
		t.Fatalf(`expected harvestConfig.Defaults.Password to be 123#abc, actual=[%+v]`,
			conf.Config.Defaults.Addr)
	}

	if conf.Config.Defaults.Addr != "" {
		t.Fatalf(`expected harvestConfig.Defaults.addr to be nil, actual=[%+v]`,
			conf.Config.Defaults.Addr)
	}
	if len(conf.Config.Defaults.Collectors) != 2 {
		t.Fatalf(`expected two default collectors, actual=%+v`, conf.Config.Defaults.Collectors)
	}

	allowedRegexes := conf.Config.Exporters["prometheus"].AllowedAddrsRegex
	if (*allowedRegexes)[0] != "^192.168.0.\\d+$" {
		t.Fatalf(`expected allow_addrs_regex to be ^192.168.0.\d+$ actual=%+v`,
			(*allowedRegexes)[0])
	}

	influxyAddr := conf.Config.Exporters["influxy"].Addr
	if (*influxyAddr) != "localhost" {
		t.Fatalf(`expected addr to be "localhost", actual=%+v`, *influxyAddr)
	}

	influxyURL := conf.Config.Exporters["influxz"].URL
	if (*influxyURL) != "www.example.com/influxdb" {
		t.Fatalf(`expected addr to be "www.example.com/influxdb", actual=%+v`, *influxyURL)
	}

	infinity2, _ := conf.PollerNamed("infinity2")
	collectors := infinity2.Collectors
	if collectors[0].Name != "Zapi" {
		t.Fatalf(`expected infinity2 collectors to contain Zapi actual=%+v`, collectors[0])
	}
	if infinity2.IsKfs {
		t.Fatalf(`expected infinity2 is_kfs to be false, but was true`)
	}
	sim1 := conf.Config.Pollers["sim-0001"]
	if !sim1.IsKfs {
		t.Fatalf(`expected sim-0001 is_kfs to be true, but was false`)
	}
}

func TestUniquePromPorts(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/testConfig.yml")
	valid := checkUniquePromPorts(conf.Config)
	if valid.isValid {
		t.Fatal(`expected isValid to be false since there are duplicate prom ports, actual was isValid=true`)
	}
	if len(valid.invalid) != 2 {
		t.Fatalf(`expected checkUniquePromPorts to return 2 invalid results, actual was %s`, valid.invalid)
	}
}

func TestExporterTypesAreValid(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/testConfig.yml")
	valid := checkExporterTypes(conf.Config)
	if valid.isValid {
		t.Fatalf(`expected isValid to be false since there are invalid exporter types, actual was %+v`, valid)
	}
	if len(valid.invalid) != 3 {
		t.Fatalf(`expected three invalid exporters, got %d`, len(valid.invalid))
	}
}

func TestCustomYamlIsValid(t *testing.T) {
	type test struct {
		path        string
		numInvalid  int
		msgContains string
	}

	tests := []test{
		{
			path:        "testdata/conf1/conf",
			numInvalid:  1,
			msgContains: "top-level",
		},
		{
			path:        "testdata/conf2/conf",
			numInvalid:  1,
			msgContains: "should be a map",
		},
		{
			path:        "testdata/conf3/conf",
			numInvalid:  1,
			msgContains: "does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			valid := checkConfTemplates([]string{tt.path})
			if valid.isValid {
				t.Errorf("want isValid=%t, got %t", false, valid.isValid)
			}
			if len(valid.invalid) != tt.numInvalid {
				t.Errorf("want %d invalid, got %d", tt.numInvalid, len(valid.invalid))
			}
			for _, invalid := range valid.invalid {
				if !strings.Contains(invalid, tt.msgContains) {
					t.Errorf("want invalid to contain %s, got %s", tt.msgContains, invalid)
				}
			}
		})
	}
}

func TestCheckCollectorName(t *testing.T) {
	type test struct {
		path string
		want bool
	}

	tests := []test{
		{
			path: "testdata/collector/conf1.yml",
			want: false,
		},
		{
			path: "testdata/collector/conf2.yml",
			want: false,
		},
		{
			path: "testdata/collector/conf3.yml",
			want: false,
		},
		{
			path: "testdata/collector/conf4.yml",
			want: true,
		},
		{
			path: "testdata/collector/conf5.yml",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			conf.TestLoadHarvestConfig(tt.path)
			valid := checkCollectorName(conf.Config)
			if valid.isValid != tt.want {
				t.Errorf("want isValid=%t, got %t", tt.want, valid.isValid)
			}
		})
	}
}
