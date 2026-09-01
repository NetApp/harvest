package staticcounter

import (
	"github.com/netapp/harvest/v2/assert"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadStaticCounterDefinitions(t *testing.T) {
	path := writeDefinitions(t, `
objects:
  volume:
    counter_definitions:
      - name: operations
        type: rate
      - name: latency
        type: average
        base_counter: operations
      - name: missing_type
      - name: missing_base
        type: average
        base_counter: unknown
`)

	got, err := LoadStaticCounterDefinitions("volume", path, slog.Default())
	assert.Nil(t, err)
	assert.Equal(t, len(got.CounterDefinitions), 2)
	assert.Equal(t, got.CounterDefinitions["operations"].Type, "rate")
	assert.Equal(t, got.CounterDefinitions["latency"].BaseCounter, "operations")

	missing, err := LoadStaticCounterDefinitions("aggregate", path, slog.Default())
	assert.Nil(t, err)
	assert.Equal(t, missing.CounterDefinitions == nil, true)
}

func TestLoadStaticCounterDefinitionsErrors(t *testing.T) {
	_, err := LoadStaticCounterDefinitions("volume", filepath.Join(t.TempDir(), "missing.yaml"), slog.Default())
	if err == nil {
		t.Fatal("expected missing file error")
	}

	path := writeDefinitions(t, "objects: [")
	_, err = LoadStaticCounterDefinitions("volume", path, slog.Default())
	if err == nil {
		t.Fatal("expected malformed YAML error")
	}
}

func writeDefinitions(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "static_counter_definitions.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
