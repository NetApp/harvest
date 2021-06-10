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
	path := "testConfig.yml"
	err := conf.LoadHarvestConfig(path)
	if err != nil {
		return
	}
	if conf.Config.Defaults.Password != "123#abc" {
		t.Fatalf(`expected harvestConfig.Defaults.Password to be 123#abc, actual=[%+v]`,
			conf.Config.Defaults.Addr)
	}

	if conf.Config.Defaults.Addr != nil {
		t.Fatalf(`expected harvestConfig.Defaults.Addr to be nil, actual=[%+v]`,
			conf.Config.Defaults.Addr)
	}
	if len(*conf.Config.Defaults.Collectors) != 2 {
		t.Fatalf(`expected two default collectors, actual=%+v`,
			*conf.Config.Defaults.Collectors)
	}

	allowedRegexes := (*conf.Config.Exporters)["prometheus"].AllowedAddrsRegex
	if (*allowedRegexes)[0] != "^192.168.0.\\d+$" {
		t.Fatalf(`expected allow_addrs_regex to be ^192.168.0.\d+$ actual=%+v`,
			(*allowedRegexes)[0])
	}

	influxyAddr := (*conf.Config.Exporters)["influxy"].Addr
	if (*influxyAddr) != "localhost" {
		t.Fatalf(`expected addr to be "localhost", actual=%+v`, (*influxyAddr))
	}

	influxyURL := (*conf.Config.Exporters)["influxz"].Url
	if (*influxyURL) != "www.example.com/influxdb" {
		t.Fatalf(`expected addr to be "www.example.com/influxdb", actual=%+v`, (*influxyURL))
	}

	collectors := (*conf.Config.Pollers)["infinity2"].Collectors
	if (*collectors)[0] != "Zapi" {
		t.Fatalf(`expected infinity2 collectors to contain Zapi actual=%+v`,
			(*collectors)[0])
	}
}
