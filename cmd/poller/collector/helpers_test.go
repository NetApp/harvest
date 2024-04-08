package collector

import (
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
			if got := getClosestIndex(tt.args.versions, tt.args.version); got != tt.want {
				t.Errorf("getClosestIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_HARVEST_CONF(t *testing.T) {
	t.Setenv(conf.HomeEnvVar, "testdata")
	template, err := ImportTemplate([]string{"conf"}, "test.yaml", "test")
	if err != nil {
		t.Errorf(`got err="%v", want no err`, err)
		return
	}
	name := template.GetChildContentS("collector")
	if name != "Test" {
		t.Errorf("collectorName got=%s, want=Test", name)
	}
}
