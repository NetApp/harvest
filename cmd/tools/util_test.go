package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/netapp/harvest/v2/cmd/tools/rest"
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
