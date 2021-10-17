package doctor

import (
	"goharvest2/pkg/conf"
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
	redacted = strings.TrimSpace(redacted)
	input = printRedactedConfig("test", []byte(input))
	input = strings.TrimSpace(input)
	if input != redacted {
		t.Fatalf(`input=[%s] != redacted=[%s]`, input, redacted)
	}
}

func TestConfigToStruct(t *testing.T) {
	loadTestConfig()
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

	influxyURL := conf.Config.Exporters["influxz"].Url
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
	loadTestConfig()
	valid := checkUniquePromPorts(conf.Config)
	if valid.isValid {
		t.Fatal(`expected isValid to be false since there are duplicate prom ports, actual was isValid=true`)
	}
	if len(valid.invalid) != 2 {
		t.Fatalf(`expected checkUniquePromPorts to return 2 invalid results, actual was %s`, valid.invalid)
	}
}

func TestExporterTypesAreValid(t *testing.T) {
	loadTestConfig()
	valid := checkExporterTypes(conf.Config)
	if valid.isValid {
		t.Fatalf(`expected isValid to be false since there are invalid exporter types, actual was %+v`, valid)
	}
	if len(valid.invalid) != 3 {
		t.Fatalf(`expected three invalid exporters, got %d`, len(valid.invalid))
	}
}

func loadTestConfig() {
	path := "testdata/testConfig.yml"
	err := conf.LoadHarvestConfig(path)
	if err != nil {
		panic(err)
	}
}
