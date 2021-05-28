package doctor

import (
	"fmt"
	"goharvest2/pkg/conf"
	"gopkg.in/yaml.v3"
	"io/ioutil"
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
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("error reading config file %+v\n", err)
		return
	}
	harvestConfig := &conf.HarvestConfig{}
	err = yaml.Unmarshal(contents, harvestConfig)
	if err != nil {
		fmt.Printf("error reading config file=[%s] %+v\n", path, err)
		return
	}

	if harvestConfig.Defaults.Addr != nil {
		t.Fatalf(`expected harvestConfig.Defaults.Addr to be nil, actual=[%+v]`,
			harvestConfig.Defaults.Addr)
	}
	if len(*harvestConfig.Defaults.Collectors) != 2 {
		t.Fatalf(`expected two default collectors, actual=%+v`,
			*harvestConfig.Defaults.Collectors)
	}

	allowedRegexes := (*harvestConfig.Exporters)["influxy"].AllowedAddrsRegex
	if (*allowedRegexes)[0] != "^192.168.0.\\d+$" {
		t.Fatalf(`expected allow_addrs_regex to be ^192.168.0.\d+$ actual=%+v`,
			(*allowedRegexes)[0])
	}

	collectors := (*harvestConfig.Pollers)["infinity2"].Collectors
	if (*collectors)[0] != "Zapi" {
		t.Fatalf(`expected infinity2 collectors to contain Zapi actual=%+v`,
			(*collectors)[0])
	}
}
