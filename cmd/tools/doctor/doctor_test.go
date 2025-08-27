package doctor

import (
	"github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/conf"
	"os"
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

	inputNode, err := printRedactedConfig("test", []byte(input))
	assert.Nil(t, err)
	inputBytes, err := yaml.Marshal(inputNode)
	assert.Nil(t, err)
	input = strings.TrimSpace(string(inputBytes))

	assert.Equal(t, input, redacted)
}

func TestDoDoctor(t *testing.T) {
	type test struct {
		parentPath string
		outPath    string
	}

	tests := []test{
		{"testdata/merge/merge1/parent.yml", "testdata/merge/merge1/out.yml"},
		{"testdata/merge/merge2/parent.yml", "testdata/merge/merge2/out.yml"},
		{"testdata/merge/merge3/parent.yml", "testdata/merge/merge3/out.yml"},
	}
	for _, tt := range tests {

		output := doDoctor(tt.parentPath)

		outBytes, err := os.ReadFile(tt.outPath)
		assert.Nil(t, err)

		expectedOutput := string(outBytes)

		diff := cmp.Diff(output, expectedOutput)
		assert.Equal(t, diff, "")
	}
}

func TestConfigToStruct(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/testConfig.yml")
	assert.Equal(t, conf.Config.Defaults.Password, "123#abc")
	assert.Equal(t, conf.Config.Defaults.Addr, "")
	assert.Equal(t, len(conf.Config.Defaults.Collectors), 2)

	allowedRegexes := conf.Config.Exporters["prometheus"].AllowedAddrsRegex
	assert.Equal(t, (*allowedRegexes)[0], "^192.168.0.\\d+$")

	influxyAddr := conf.Config.Exporters["influxy"].Addr
	assert.Equal(t, *influxyAddr, "localhost")

	influxyURL := conf.Config.Exporters["influxz"].URL
	assert.Equal(t, *influxyURL, "www.example.com/influxdb")

	infinity2, _ := conf.PollerNamed("infinity2")
	collectors := infinity2.Collectors
	assert.Equal(t, collectors[0].Name, "Zapi")
	assert.False(t, infinity2.IsKfs)
	sim1 := conf.Config.Pollers["sim-0001"]
	assert.True(t, sim1.IsKfs)
}

func TestUniquePromPorts(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/testConfig.yml")
	valid := checkUniquePromPorts(conf.Config)
	assert.False(t, valid.isValid)
	assert.Equal(t, len(valid.invalid), 4)
}

func TestExporterTypesAreValid(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/testConfig.yml")
	valid := checkExporterTypes(conf.Config)
	assert.False(t, valid.isValid)
	assert.Equal(t, len(valid.invalid), 3)
}

func TestCustomYamlIsValid(t *testing.T) {
	type test struct {
		path        string
		isValid     bool
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
		{
			path:        "testdata/nabox/conf",
			numInvalid:  1,
			msgContains: "does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			valid := checkConfTemplates([]string{tt.path})
			assert.Equal(t, valid.isValid, tt.isValid)
			assert.Equal(t, len(valid.invalid), tt.numInvalid)
			for _, invalid := range valid.invalid {
				assert.True(t, strings.Contains(invalid, tt.msgContains))
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
			assert.Equal(t, valid.isValid, tt.want)
		})
	}
}

func TestExportersExist(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/noExporters.yml")
	valid := checkExportersExist(conf.Config)
	assert.False(t, valid.isValid)
}

func TestPollerPromPorts(t *testing.T) {
	conf.TestLoadHarvestConfig("testdata/promPortNoPromExporters.yml")
	valid := checkPollerPromPorts(conf.Config)
	assert.False(t, valid.isValid)
	assert.Equal(t, len(valid.invalid), 2)
}
