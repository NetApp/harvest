package collector

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/third_party/go-version"
	"sort"
	"testing"
)

func setupVersions(availableVersions []string) []*version.Version {
	versions := make([]*version.Version, len(availableVersions))
	for i, raw := range availableVersions {
		versions[i] = buildVersion(raw)
	}
	sort.Sort(version.Collection(versions))
	return versions
}

func buildVersion(ver string) *version.Version {
	v, _ := version.NewVersion(ver)
	return v
}
func Test_getClosestIndex(t *testing.T) {
	type args struct {
		versions []*version.Version
		version  *version.Version
	}

	type test struct {
		name string
		args args
		want int
	}

	tests := []test{
		{"MatchCaseEqual", args{setupVersions([]string{"9.8.0", "9.8.1", "9.9.0", "10.10.10"}), buildVersion("9.8.1")}, 1},
		{"MatchCaseLast", args{setupVersions([]string{"9.8.0", "9.8.1", "9.9.0", "10.10.10"}), buildVersion("11.11.11")}, 3},
		{"MatchCaseLast_2", args{setupVersions([]string{"9.8.0", "9.8.1", "9.9.0", "9.10.1"}), buildVersion("9.10.2")}, 3},
		{"NonMatchCase", args{setupVersions([]string{"9.8.0", "9.8.1", "9.9.0", "10.10.10"}), buildVersion("9.7.1")}, 0},
		{"ClosestMatchCase", args{setupVersions([]string{"9.8.0", "9.8.1", "9.9.0", "10.10.10"}), buildVersion("9.9.2")}, 2},
		{"EmptyCase", args{setupVersions([]string{}), buildVersion("9.8.1")}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getClosestIndex(tt.args.versions, tt.args.version)
			assert.Equal(t, got, tt.want)
		})
	}
}

func Test_HARVEST_CONF(t *testing.T) {
	t.Setenv(conf.HomeEnvVar, "testdata")
	template, err := ImportTemplate([]string{"conf"}, "test.yaml", "test")
	assert.Nil(t, err)
	name := template.GetChildContentS("collector")
	assert.Equal(t, name, "Test")
}

func TestParseTemplateRef(t *testing.T) {
	tests := []struct {
		name                string
		input               string
		expectedClass       string
		expectedTemplate    string
		expectedIsDelegated bool
	}{
		{
			name:                "valid_keyperf_dsl",
			input:               "KeyPerf:volume.yaml",
			expectedClass:       "KeyPerf",
			expectedTemplate:    "volume.yaml",
			expectedIsDelegated: true,
		},
		{
			name:                "valid_with_spaces",
			input:               " KeyPerf : volume.yaml ",
			expectedClass:       "KeyPerf",
			expectedTemplate:    "volume.yaml",
			expectedIsDelegated: true,
		},
		{
			name:                "invalid_no_colon",
			input:               "KeyPerfvolume.yaml",
			expectedClass:       "",
			expectedTemplate:    "",
			expectedIsDelegated: false,
		},
		{
			name:                "invalid_empty_class",
			input:               ":volume.yaml",
			expectedClass:       "",
			expectedTemplate:    "",
			expectedIsDelegated: false,
		},
		{
			name:                "invalid_too_many_colons",
			input:               "KeyPerf:volume:yaml",
			expectedClass:       "",
			expectedTemplate:    "",
			expectedIsDelegated: false,
		},
		{
			name:                "invalid_empty_string",
			input:               "",
			expectedClass:       "",
			expectedTemplate:    "",
			expectedIsDelegated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			class, template, isDelegated := ParseTemplateRef(tt.input)

			assert.Equal(t, class, tt.expectedClass)
			assert.Equal(t, template, tt.expectedTemplate)
			assert.Equal(t, isDelegated, tt.expectedIsDelegated)
		})
	}
}
