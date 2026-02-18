package conf

import (
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

var testYml = "../../cmd/tools/doctor/testdata/testConfig.yml"

func TestGetLastPromPort(t *testing.T) {
	TestLoadHarvestConfig(testYml)
	type args struct {
		pollerNames []string
	}

	type test struct {
		name    string
		args    args
		wantErr []bool
	}
	tests := []test{
		{"test", args{pollerNames: []string{"unix-01", "cluster-02", "test1"}}, []bool{false, true, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, v := range tt.args.pollerNames {
				got, err := GetLastPromPort(v, true)
				if (err != nil) != tt.wantErr[i] {
					t.Errorf("GetLastPromPort() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err == nil && got == 0 {
					t.Errorf("GetLastPromPort() got = %v, want %s", got, "non zero value")
				}
			}
		})
	}
}

func TestGetLastPromPortIssue284(t *testing.T) {
	TestLoadHarvestConfig("../../cmd/tools/doctor/testdata/issue-284.yml")
	loadPrometheusExporterPortRangeMapping(false)
	got, _ := GetLastPromPort("issue-284", false)
	assert.Equal(t, got, 0)
}

func TestPollerStructDefaults(t *testing.T) {
	TestLoadHarvestConfig(testYml)
	t.Run("poller exporters", func(t *testing.T) {
		poller, err := PollerNamed("zeros")
		if err != nil {
			panic(err)
		}
		// the poller does not define exporters but defaults does
		assert.NotNil(t, poller.Exporters)
		assert.Equal(t, len(poller.Exporters), 1)
		diff := cmp.Diff(poller.Exporters, []string{"prometheusrange"})
		assert.Equal(t, diff, "")
	})

	t.Run("poller collector", func(t *testing.T) {
		poller, err := PollerNamed("cluster-01")
		if err != nil {
			panic(err)
		}
		// the poller does not define collectors but defaults does
		assert.NotNil(t, poller.Collectors)
		assert.Equal(t, len(poller.Collectors), 2)
		defaultT := []string{"default.yaml", "custom.yaml"}
		want := []Collector{{Name: "Zapi", Templates: &defaultT}, {Name: "ZapiPerf", Templates: &defaultT}}
		diff := cmp.Diff(poller.Collectors, want)
		assert.Equal(t, diff, "")
	})

	t.Run("poller username", func(t *testing.T) {
		poller, err := PollerNamed("zeros")
		if err != nil {
			panic(err)
		}
		// the poller does not define a username but defaults does
		if poller.Username != "myuser" {
			t.Fatalf(`expected username to be [myuser] but was [%v]`, poller.Username)
		}
	})
}

func TestPollerUnion(t *testing.T) {
	TestLoadHarvestConfig(testYml)
	addr := "addr"
	user := "user"
	defaults := Poller{
		Addr:           addr,
		Collectors:     []Collector{{Name: "0"}, {Name: "1"}, {Name: "2"}, {Name: "3"}},
		Username:       user,
		UseInsecureTLS: new(true),
		IsKfs:          true,
	}
	p := Poller{
		UseInsecureTLS: new(false),
		IsKfs:          false,
	}
	p.Union(&defaults)
	if p.Username != "user" {
		t.Fatalf(`expected username to be [user] but was [%v]`, p.Username)
	}
	if p.Addr != "addr" {
		t.Fatalf(`expected addr to be [addr] but was [%v]`, p.Addr)
	}
	if *p.UseInsecureTLS {
		t.Fatalf(`expected UseInsecureTLS to be [false] but was [%v]`, *p.UseInsecureTLS)
	}
	if p.IsKfs {
		t.Fatalf(`expected IsKfs to be [false] but was [%v]`, p.IsKfs)
	}
	if len(p.Collectors) != 4 {
		t.Fatalf(`expected collectors to have four elements but was [%v]`, p.Collectors)
	}
	for i := range len(p.Collectors) {
		actual := p.Collectors[i].Name
		if actual != strconv.Itoa(i) {
			t.Fatalf(`expected element at index=%d to be %d but was [%v]`, i, i, actual)
		}
	}

	maxFiles := 314
	p2 := Poller{
		Username:    "name",
		Collectors:  []Collector{{Name: "10"}, {Name: "11"}, {Name: "12"}, {Name: "13"}},
		IsKfs:       true,
		LogMaxFiles: 314,
	}
	p2.Union(&defaults)
	if p2.Username != "name" {
		t.Fatalf(`expected username to be [name] but was [%v]`, p2.Username)
	}
	if !p2.IsKfs {
		t.Fatalf(`expected isKfs to be [true] but was [%v]`, p2.IsKfs)
	}
	if p2.LogMaxFiles != maxFiles {
		t.Fatalf(`expected LogMaxFiles to be [314] but was [%v]`, p2.LogMaxFiles)
	}
	for i := range len(p2.Collectors) {
		actual := p2.Collectors[i].Name
		if actual != strconv.Itoa(10+i) {
			t.Fatalf(`expected element at index=%d to be %d but was [%v]`, i, i+10, actual)
		}
	}
}

func TestFlowStyle(t *testing.T) {
	TestLoadHarvestConfig(testYml)
	t.Run("poller with flow", func(t *testing.T) {
		poller, err := PollerNamed("flow")
		if err != nil {
			panic(err)
		}
		if len(poller.Collectors) != 1 {
			t.Fatalf(`expected there to be one collector but got %v`, len(poller.Collectors))
		}
		if poller.Collectors[0].Name != "Zapi" {
			t.Fatalf(`expected the first collector to be Zapi but got %v`, poller.Collectors[0])
		}
		if len(poller.Exporters) != 1 {
			t.Fatalf(`expected there to be one exporter but got %v`, len(poller.Exporters))
		}
		if poller.Exporters[0] != "prom" {
			t.Fatalf(`expected the first exporter to be prom but got %v`, poller.Exporters[0])
		}
	})
}

func TestUniqueExportersByType(t *testing.T) {
	path := "../../cmd/tools/doctor/testdata/testConfig.yml"
	TestLoadHarvestConfig(path)
	poller2, _ := PollerNamed("overlapping")
	t.Run("Exporters are unique by type", func(t *testing.T) {
		exporters := GetUniqueExporters(poller2.Exporters)
		sort.Strings(exporters)
		diff := cmp.Diff([]string{"foo1", "foo2", "influxy", "influxz", "prometheus"}, exporters)
		if diff != "" {
			t.Errorf("Mismatch (-got +want):\n%s", diff)
		}
	})
}

func TestIssue271_PollerPanicsWhenExportDoesNotExist(t *testing.T) {
	TestLoadHarvestConfig("../../cmd/tools/doctor/testdata/testConfig.yml")
	poller, err := PollerNamed("issue-271")
	if err != nil {
		panic(err)
	}
	t.Run("Poller panics when exporter does not exist", func(_ *testing.T) {
		exporters := GetUniqueExporters(poller.Exporters)
		if err != nil {
			panic(err)
		}
		if exporters != nil {
			return
		}
	})
}

func TestQuotedPassword(t *testing.T) {
	TestLoadHarvestConfig(testYml)
	t.Run("quoted password", func(t *testing.T) {
		poller, err := PollerNamed("pass-with-escape")
		if err != nil {
			panic(err)
		}
		if poller.Password != "#pass" {
			t.Fatalf(`expected password to be #pass but got %v`, poller.Password)
		}
	})
}

func TestCollectorConfig(t *testing.T) {
	type test struct {
		name string
		path string
		want []string
	}
	tests := []test{
		{name: "normal", path: "testdata/normal.yaml", want: []string{"default.yaml"}},
		{name: "issue_396", path: "testdata/issue_396.yaml", want: []string{"limited1.yaml", "limited2.yaml", "limited3.yaml"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TestLoadHarvestConfig(tt.path)
			poller, err := PollerNamed("DC-01")
			if err != nil {
				panic(err)
			}
			for i, tc := range tt.want {
				if tc != (*poller.Collectors[0].Templates)[i] {
					t.Errorf("want %s collector config, got %s", tt.want[i], tc)
				}
			}
		})
	}
}

func TestNodeToPoller(t *testing.T) {
	t.Helper()
	testArg := func(t *testing.T, want, got string) {
		if got != want {
			t.Errorf("want=[%s] got=[%s]", want, got)
		}
	}

	Config.Defaults = &Poller{
		Username: "bob",
		Password: "bob",
	}

	defaultNode := node.NewS("root")
	defaultNode.NewChildS("password", "pass")
	defaultNode.NewChildS("use_insecure_tls", "true")
	poller := ZapiPoller(defaultNode)

	testArg(t, DefaultAPIVersion, poller.APIVersion)
	testArg(t, "bob", poller.Username)
	testArg(t, "pass", poller.Password)
	testArg(t, "30s", poller.ClientTimeout)
	testArg(t, "true", strconv.FormatBool(*poller.UseInsecureTLS))
}

func TestEmptyPath(t *testing.T) {
	t.Helper()
	resetConfig()
	path := Path("")
	if path != "" {
		t.Errorf("got=%s want=%s", path, "")
	}

	const HomeVar = "HOME"
	t.Setenv(HomeEnvVar, HomeVar)
	path = Path("")
	if path != HomeVar {
		t.Errorf("got=%s want=%s", path, "")
	}
}

func TestPathFromEnvs(t *testing.T) {
	t.Helper()
	resetConfig()

	// Set the environment variable to a relative path
	t.Setenv(HomeEnvVar, "testdata")
	path := ConfigPath(HarvestYML)
	if path != "testdata/harvest.yml" {
		t.Errorf("got=%s want=%s", path, "testdata/harvest.yml")
	}
	path = ConfigPath(path)
	if path != "testdata/harvest.yml" {
		t.Errorf("got=%s want=%s", path, "testdata/harvest.yml")
	}

	// Set the environment variable to an absolute path
	getwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	absPath := getwd + "/testdata"
	t.Setenv(HomeEnvVar, absPath)
	path = ConfigPath(HarvestYML)
	if path != absPath+"/harvest.yml" {
		t.Errorf("got=%s want=%s", path, "testdata/harvest.yml")
	}
	path = ConfigPath(path)
	if path != absPath+"/harvest.yml" {
		t.Errorf("got=%s want=%s", path, "testdata/harvest.yml")
	}
}

func TestReadHarvestConfigFromEnv(t *testing.T) {
	t.Helper()
	resetConfig()
	t.Setenv(HomeEnvVar, "testdata")
	cp, err := LoadHarvestConfig(HarvestYML)
	if err != nil {
		t.Errorf("Failed to load config at=[%s] err=%+v\n", HarvestYML, err)
		return
	}
	wantCp := "testdata/harvest.yml"
	if cp != wantCp {
		t.Errorf("configPath got=%s want=%s", cp, wantCp)
	}
	poller := Config.Pollers["star"]
	if poller == nil {
		t.Errorf("check if star poller exists. got=nil want=poller")
	}
}

func resetConfig() {
	configRead = false
	Config = HarvestConfig{}
}

func TestMultiplePollerFiles(t *testing.T) {
	t.Helper()
	resetConfig()
	configYaml := "testdata/pollerFiles/harvest.yml"
	_, err := LoadHarvestConfig(configYaml)
	if err == nil {
		t.Fatalf("want errors loading config: %s, got no errors", configYaml)
	}

	wantNumErrs := 2
	numErrs := strings.Count(err.Error(), "\n") + 1
	if numErrs != wantNumErrs {
		t.Errorf("got %d errors, want %d", numErrs, wantNumErrs)
	}

	wantNumPollers := 10
	if len(Config.Pollers) != wantNumPollers {
		t.Errorf("got %d pollers, want %d", len(Config.Pollers), wantNumPollers)
	}

	if len(Config.PollersOrdered) != wantNumPollers {
		t.Errorf("got %d ordered pollers, want %d", len(Config.PollersOrdered), wantNumPollers)
	}

	wantToken := "token"
	if Config.Tools.GrafanaAPIToken != wantToken {
		t.Errorf("got token=%s, want token=%s", Config.Tools.GrafanaAPIToken, wantToken)
	}

	orderWanted := []string{
		"star",
		"netapp1",
		"netapp2",
		"netapp3",
		"netapp4",
		"netapp5",
		"netapp6",
		"netapp7",
		"netapp8",
		"moon",
	}

	for i, n := range orderWanted {
		named, err := PollerNamed(n)
		if err != nil {
			t.Errorf("got no poller, want poller named=%s", n)
			continue
		}
		if named.promIndex != i {
			t.Errorf("got promIndex=%d, want promIndex=%d", named.promIndex, i)
		}
	}

	// Ensure that parent's `Defaults` is merged with children pollers
	named, err := PollerNamed("netapp1")
	if err != nil {
		t.Fatalf("got no poller, want poller named=%s", "netapp1")
	}

	if len(named.Collectors) != 1 {
		t.Fatalf("got %d collectors, want 1", len(named.Collectors))
	}

	if named.Collectors[0].Name != "Simple" {
		t.Fatalf("got collector name=%s, want collector name=%s", named.Collectors[0].Name, "Simple")
	}
}

func TestChildPollers(t *testing.T) {
	t.Helper()
	resetConfig()
	configYaml := "testdata/pollerFiles/harvest_parent_nopoller.yml"
	_, err := LoadHarvestConfig(configYaml)
	if err != nil {
		t.Fatalf("got error loading config: %s, want no errors", err)
	}

	// Replace with the expected number of pollers from your child configuration file
	wantNumPollers := 8
	if len(Config.Pollers) != wantNumPollers {
		t.Errorf("got %d pollers, want %d", len(Config.Pollers), wantNumPollers)
	}

	// Replace "childPollerName" with the name of a poller from your child configuration file
	_, err = PollerNamed("netapp1")
	if err != nil {
		t.Errorf("got no poller, want poller named=%s", "childPollerName")
	}
}

func TestEmptyPoller(t *testing.T) {
	t.Helper()
	resetConfig()

	configYaml := "testdata/pollerFiles/empty_poller.yml"
	_, err := LoadHarvestConfig(configYaml)
	if err != nil {
		t.Fatalf("got error loading config: %s, want no errors", err)
	}

	wantPollers := []string{"abc", "empty", "def"}
	for _, pName := range wantPollers {
		_, err = PollerNamed(pName)
		if err != nil {
			t.Errorf("got no poller, want poller named=%s", pName)
		}
	}
}

func TestEmbeddedExporter(t *testing.T) {
	t.Helper()
	resetConfig()

	configYaml := "testdata/issue_2852_direct_exporters.yml"
	_, err := LoadHarvestConfig(configYaml)
	if err != nil {
		t.Fatalf("got error loading config: %s, want no errors", err)
	}

	p, err := PollerNamed("u2")
	if err != nil {
		t.Fatalf("got no poller, want poller named=u2")
	}
	if len(p.Exporters) != 3 {
		t.Errorf("got %d exporters, want 3", len(p.Exporters))
	}

	port, err := GetLastPromPort("u2", false)
	if err != nil {
		t.Fatalf("got error: %v, want no error", err)
	}
	if port != 12990 {
		t.Errorf("got port=%d, want port=12990", port)
	}

	uniqueExporters := GetUniqueExporters(p.Exporters)
	want := []string{"u2-1", "u2-2"}
	if !slices.Equal(uniqueExporters, want) {
		t.Errorf("got %v, want %v", uniqueExporters, want)
	}

	// Test that the last default Prometheus exporter is used when no exporters are defined for a poller
	port, err = GetLastPromPort("u3", false)
	if err != nil {
		t.Fatalf("got error: %v, want no error", err)
	}
	if port != 32990 {
		t.Errorf("got port=%d, want port=32990", port)
	}
}

// TestPromPort tests the prom_port configuration
//   - If there are multiple Prometheus exporters defined for a poller, pick the last one.
//     (see GetUniqueExporters)
//   - If there is an embedded exporter, prom_port wins
//     If the embedded exporter is the last one in the list, it will be picked (per the above rule), but the
//     prom_port will be used instead of any port defined in the embedded exporter.
func TestPromPort(t *testing.T) {
	t.Helper()
	resetConfig()

	configYaml := "testdata/prom_ports.yml"
	_, err := LoadHarvestConfig(configYaml)
	if err != nil {
		t.Fatalf("got error loading config: %s, want no errors", err)
	}

	p, err := PollerNamed("sar")
	if err != nil {
		t.Fatalf("got no poller, want poller named=u2")
	}
	if len(p.Exporters) != 2 {
		t.Errorf("got %d exporters, want 2", len(p.Exporters))
	}

	// Ensure that the last exporter is used
	uniqueExporters := GetUniqueExporters(p.Exporters)
	want := []string{"prometheus1"}
	if !slices.Equal(uniqueExporters, want) {
		t.Errorf("got %v, want %v", uniqueExporters, want)
	}

	port, err := GetLastPromPort("sar", false)
	if err != nil {
		t.Fatalf("got error: %v, want no error", err)
	}
	if port != 3000 {
		t.Errorf("got port=%d, want port=3000", port)
	}

	// Ensure that the prom_port is used instead of the port defined in the embedded exporter
	port, err = GetLastPromPort("u3", false)
	if err != nil {
		t.Fatalf("got error: %v, want no error", err)
	}
	if port != 9999 {
		t.Errorf("got port=%d, want port=9999", port)
	}

	// Ensure that zero is returned if the poller does not have an exporter
	port, err = GetLastPromPort("no-exporter", false)
	if err != nil {
		t.Fatalf("got error: %v, want no error", err)
	}
	if port != 0 {
		t.Errorf("got port=%d, want port=0", port)
	}
}
