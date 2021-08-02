package collector

import (
	"github.com/hashicorp/go-version"
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
		{"MatchCase", args{setupVersions([]string{"9.8.0", "9.8.1", "9.9.0", "10.10.10"}), buildVersion("9.8.1")}, 1},
		{"MatchFirstYolo", args{setupVersions([]string{"9.8.0", "9.8.1", "9.9.0", "10.10.10"}), buildVersion("9.7.1")}, 0},
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
