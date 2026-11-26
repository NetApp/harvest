package tools

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/netapp/harvest/v2/cmd/tools/rest"
	template2 "github.com/netapp/harvest/v2/cmd/tools/template"
	template3 "github.com/netapp/harvest/v2/pkg/template"
	"github.com/netapp/harvest/v2/pkg/tree"
)

// writeTemplates creates empty placeholder files, and their parent directories, at each
// of paths, for use as visitRestTemplates fixtures.
func writeTemplates(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range paths {
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatalf("failed to create dir for %s: %v", p, err)
		}
		if err := os.WriteFile(p, []byte("placeholder"), 0o600); err != nil {
			t.Fatalf("failed to write %s: %v", p, err)
		}
	}
}

// TestVisitRestTemplates_ModelDirsAreAddOnly verifies a model directory doesn't
// overwrite an existing base metric, and model-only metrics still surface.
func TestVisitRestTemplates_ModelDirsAreAddOnly(t *testing.T) {
	dir := t.TempDir()

	baseLun := filepath.Join(dir, "9.12.0", "lun.yaml")
	modelLun := filepath.Join(dir, "asar2", "9.16.0", "lun.yaml")
	modelStorageUnit := filepath.Join(dir, "asar2", "9.16.0", "storage_unit.yaml")
	modelDefault := filepath.Join(dir, "asar2", "default.yaml")
	writeTemplates(t, baseLun, modelLun, modelStorageUnit, modelDefault)

	eachTemp := func(path string, _ *rest.Client) map[string]Counter {
		if filepath.Base(path) == "default.yaml" {
			t.Fatalf("eachTemp should not be called for default.yaml, got %s", path)
		}
		switch path {
		case baseLun, modelLun:
			// Same endpoint/counter: model must not add a second APIs entry.
			return map[string]Counter{
				"lun_size": {
					Name: "lun_size",
					APIs: []MetricDef{{
						Endpoint:     "api/private/cli/lun",
						ONTAPCounter: "size",
						Template:     path,
					}},
				},
			}
		case modelStorageUnit:
			return map[string]Counter{
				"storage_unit_size": {
					Name: "storage_unit_size",
					APIs: []MetricDef{{
						Endpoint:     "api/storage/storage-units",
						ONTAPCounter: "space.size",
						Template:     path,
					}},
				},
			}
		default:
			t.Fatalf("unexpected path passed to eachTemp: %s", path)
			return nil
		}
	}

	result := visitRestTemplates(dir, nil, eachTemp)

	lun, ok := result["lun_size"]
	if !ok {
		t.Fatalf("expected lun_size to be present in result")
	}
	if len(lun.APIs) != 1 || lun.APIs[0].Template != baseLun {
		t.Errorf("expected lun_size to keep only the base template's API entry (%s), got %+v", baseLun, lun.APIs)
	}

	storageUnit, ok := result["storage_unit_size"]
	if !ok {
		t.Fatalf("expected storage_unit_size (asar2-only metric) to be present in result")
	}
	if len(storageUnit.APIs) != 1 || storageUnit.APIs[0].Template != modelStorageUnit {
		t.Errorf("expected storage_unit_size to point at %s, got %+v", modelStorageUnit, storageUnit.APIs)
	}
}

// TestVisitRestTemplates_ModelDirAppendsDivergentAPI verifies a model metric with a
// different endpoint/counter than the base is appended, not dropped.
func TestVisitRestTemplates_ModelDirAppendsDivergentAPI(t *testing.T) {
	dir := t.TempDir()

	baseLun := filepath.Join(dir, "9.12.0", "lun.yaml")
	modelLun := filepath.Join(dir, "asar2", "9.16.0", "lun.yaml")
	writeTemplates(t, baseLun, modelLun)

	eachTemp := func(path string, _ *rest.Client) map[string]Counter {
		switch path {
		case baseLun:
			return map[string]Counter{
				"lun_size": {
					Name: "lun_size",
					APIs: []MetricDef{{
						Endpoint:     "api/private/cli/lun",
						ONTAPCounter: "size",
						Template:     baseLun,
					}},
				},
			}
		case modelLun:
			return map[string]Counter{
				"lun_size": {
					Name: "lun_size",
					APIs: []MetricDef{{
						Endpoint:     "api/storage/luns",
						ONTAPCounter: "space.size",
						Template:     modelLun,
					}},
				},
			}
		default:
			t.Fatalf("unexpected path passed to eachTemp: %s", path)
			return nil
		}
	}

	result := visitRestTemplates(dir, nil, eachTemp)

	lun, ok := result["lun_size"]
	if !ok {
		t.Fatalf("expected lun_size to be present in result")
	}
	if len(lun.APIs) != 2 {
		t.Fatalf("expected lun_size to have both the base and model API entries, got %+v", lun.APIs)
	}
	if lun.APIs[0].Template != baseLun {
		t.Errorf("expected the base template's API entry to come first, got %+v", lun.APIs[0])
	}
	if lun.APIs[1].Template != modelLun || lun.APIs[1].Endpoint != "api/storage/luns" {
		t.Errorf("expected the model template's divergent API entry to be appended, got %+v", lun.APIs[1])
	}
}

func TestIsModelTemplate(t *testing.T) {
	restDir := filepath.Join("conf", "rest")
	tests := []struct {
		name string
		dir  string
		path string
		want bool
	}{
		{"base version dir", restDir, filepath.Join(restDir, "9.12.0", "lun.yaml"), false},
		{"model dir", restDir, filepath.Join(restDir, "asar2", "9.16.0", "lun.yaml"), true},
		// visitRestTemplates filters out default.yaml before calling isModelTemplate;
		// this only pins down isModelTemplate's own behavior in isolation.
		{"model default", restDir, filepath.Join(restDir, "asar2", "default.yaml"), true},
		{"asar2 in a filename, not a dir", restDir, filepath.Join(restDir, "9.16.0", "asar2_lun.yaml"), false},
		{"asar2 nested, not top-level", restDir, filepath.Join(restDir, "9.16.0", "asar2", "lun.yaml"), false},
		{"similar dir name", restDir, filepath.Join(restDir, "asar2x", "9.16.0", "lun.yaml"), false},
		{"path equals dir", restDir, restDir, false},
		{"keyperf model dir", filepath.Join("conf", "keyperf"), filepath.Join("conf", "keyperf", "asar2", "9.16.0", "storage_unit.yaml"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isModelTemplate(tt.dir, tt.path); got != tt.want {
				t.Errorf("isModelTemplate(%q, %q) = %v, want %v", tt.dir, tt.path, got, tt.want)
			}
		})
	}
}

// captureStdout returns what fn prints to stdout
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close pipe writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read captured stdout: %v", err)
	}
	return string(out)
}

func TestHandlePerfLookupFailure(t *testing.T) {
	cause := errors.New(`Object "workload_queue_combined" was not found`)
	path := "conf/statperf/9.20.1/workload_queue_combined.yaml"

	tests := []struct {
		name      string
		query     string
		version   string
		wantWarn  bool
		wantError bool
		wantExit  bool
	}{
		{"older 9.16.1", "workload_queue_combined", "9.16.1", true, false, false},
		{"older 9.19.1", "workload_queue_combined", "9.19.1", true, false, false},
		{"minimum 9.20.0", "workload_queue_combined", "9.20.0", false, true, true},
		{"newer 9.20.1", "workload_queue_combined", "9.20.1", false, true, true},
		{"newer 9.21.0", "workload_queue_combined", "9.21.0", false, true, true},
		{"empty version", "workload_queue_combined", "", false, true, true},
		{"unparsable version", "workload_queue_combined", "bogus", false, true, true},
		{"not gated", "volume", "9.16.1", false, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := exitFn
			defer func() { exitFn = orig }()
			exited := false
			exitFn = func(code int) { exited = code == 1 }

			out := captureStdout(t, func() {
				handlePerfLookupFailure(tt.query, path, tt.version, cause)
			})

			if got := strings.HasPrefix(out, "warning:"); got != tt.wantWarn {
				t.Errorf("warning = %v, want %v, output %q", got, tt.wantWarn, out)
			}
			if got := strings.HasPrefix(out, "error"); got != tt.wantError {
				t.Errorf("error = %v, want %v, output %q", got, tt.wantError, out)
			}
			if exited != tt.wantExit {
				t.Errorf("exit = %v, want %v", exited, tt.wantExit)
			}
			if tt.wantWarn && !strings.Contains(out, path) {
				t.Errorf("expected warning to name the template, got %q", out)
			}
		})
	}
}

// qosQueueCounters returns the qos_queue counters after merging the real counter.yaml into counters
func qosQueueCounters(t *testing.T, counters map[string]Counter) map[string]Counter {
	t.Helper()
	counters = ProcessExternalCounters(filepath.Join("..", ".."), counters, nil)
	if counters == nil {
		t.Fatalf("failed to process counter.yaml")
	}
	result := make(map[string]Counter)
	for name, c := range counters {
		if strings.HasPrefix(name, "qos_queue_") {
			result[name] = c
		}
	}
	return result
}

// TestProcessExternalCounters_QosQueueAddedWhenMissing covers clusters older than 9.20
func TestProcessExternalCounters_QosQueueAddedWhenMissing(t *testing.T) {
	got := qosQueueCounters(t, make(map[string]Counter))
	if len(got) == 0 {
		t.Fatalf("expected qos_queue counters to be added from counter.yaml")
	}
	for name, c := range got {
		if c.Description == "" {
			t.Errorf("%s: expected a description", name)
		}
		if len(c.APIs) != 2 || c.APIs[0].API != "StatPerf" || c.APIs[1].API != "ZapiPerf" {
			t.Errorf("%s: expected StatPerf then ZapiPerf APIs, got %+v", name, c.APIs)
			continue
		}
		if c.APIs[0].Endpoint != "workload_queue_combined" || c.APIs[0].Template != "conf/statperf/9.20.1/workload_queue_combined.yaml" {
			t.Errorf("%s: unexpected StatPerf API %+v", name, c.APIs[0])
		}
		if c.APIs[1].Endpoint != "perf-object-get-instances workload_queue_combined" || c.APIs[1].Template != "conf/zapiperf/cdot/9.20.1/workload_queue_combined.yaml" {
			t.Errorf("%s: unexpected ZapiPerf API %+v", name, c.APIs[1])
		}
	}
}

// TestProcessExternalCounters_QosQueueMergesWithLiveCounters covers 9.20+ clusters, where live counters exist
func TestProcessExternalCounters_QosQueueMergesWithLiveCounters(t *testing.T) {
	live := map[string]Counter{
		"qos_queue_iops_total": {
			Name:        "qos_queue_iops_total",
			Description: "live description",
			APIs: []MetricDef{
				{API: "StatPerf", Endpoint: "workload_queue_combined", ONTAPCounter: "iops_total", Unit: "none", Type: "rate",
					Template: "conf/statperf/9.20.1/workload_queue_combined.yaml"},
				{API: "ZapiPerf", Endpoint: "perf-object-get-instances workload_queue_combined", ONTAPCounter: "iops_total", Unit: "none", Type: "rate",
					Template: "conf/zapiperf/cdot/9.20.1/workload_queue_combined.yaml"},
			},
		},
	}
	got := qosQueueCounters(t, live)
	c, ok := got["qos_queue_iops_total"]
	if !ok {
		t.Fatalf("expected qos_queue_iops_total")
	}
	if c.Description == "" || c.Description == "live description" {
		t.Errorf("expected counter.yaml description to override the live one, got %q", c.Description)
	}
	if len(c.APIs) != 2 {
		t.Fatalf("expected 2 APIs, got %+v", c.APIs)
	}
	for _, api := range c.APIs {
		if api.Unit != "none" || api.Type != "rate" || api.ONTAPCounter != "iops_total" {
			t.Errorf("unexpected API values after merge: %+v", api)
		}
	}
}

// TestQosQueueCountersMatchTemplates checks every workload_queue_combined template counter is in counter.yaml
func TestQosQueueCountersMatchTemplates(t *testing.T) {
	defined := qosQueueCounters(t, make(map[string]Counter))

	templates := map[string]string{
		"StatPerf": filepath.Join("..", "..", "conf", "statperf", "9.20.1", "workload_queue_combined.yaml"),
		"ZapiPerf": filepath.Join("..", "..", "conf", "zapiperf", "cdot", "9.20.1", "workload_queue_combined.yaml"),
	}
	for api, path := range templates {
		model, err := template2.ReadTemplate(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		tmpl, err := tree.ImportYaml(path)
		if err != nil || tmpl == nil {
			t.Fatalf("failed to import %s: %v", path, err)
		}
		counters := tmpl.GetChildS("counters")
		if counters == nil {
			t.Fatalf("%s has no counters", path)
		}
		found := 0
		for _, c := range counters.GetAllChildContentS() {
			if c == "" {
				continue
			}
			name, display, kind, _ := template3.ParseMetric(c)
			// ZapiPerf templates list the instance identifiers as plain names, but ONTAP reports them as strings
			if kind != "float" || name == "instance_name" || name == "instance_uuid" {
				continue
			}
			found++
			harvestName := model.Object + "_" + display
			def, ok := defined[harvestName]
			if !ok {
				t.Errorf("%s: %s is exported by %s but missing from counter.yaml", api, harvestName, path)
				continue
			}
			idx := -1
			for i, a := range def.APIs {
				if a.API == api {
					idx = i
				}
			}
			if idx < 0 {
				t.Errorf("%s: counter.yaml entry for %s has no %s API", api, harvestName, api)
				continue
			}
			if def.APIs[idx].ONTAPCounter != name {
				t.Errorf("%s: %s has ONTAPCounter %q in counter.yaml, want %q", api, harvestName, def.APIs[idx].ONTAPCounter, name)
			}
		}
		if found == 0 {
			t.Errorf("%s: no counters found in %s", api, path)
		}
		if found != len(defined) {
			t.Errorf("%s: template exports %d counters but counter.yaml defines %d qos_queue counters", api, found, len(defined))
		}
	}
}
