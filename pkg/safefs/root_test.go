package safefs

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestReadAndWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.txt")
	outputPath := filepath.Join(tmpDir, "output.txt")

	if err := os.WriteFile(inputPath, []byte("hello"), 0o600); err != nil {
		t.Fatalf("write input file: %v", err)
	}

	got, err := ReadFile(inputPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("ReadFile()=%q, want %q", got, "hello")
	}

	if err := WriteFile(outputPath, []byte("world"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	written, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if string(written) != "world" {
		t.Fatalf("output=%q, want %q", written, "world")
	}
}

func TestWalkDir(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmpDir, "nested"), 0o750); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "nested", "a.json"), []byte("a"), 0o600); err != nil {
		t.Fatalf("write a.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("b"), 0o600); err != nil {
		t.Fatalf("write b.txt: %v", err)
	}

	seen := make(map[string]string)
	err := WalkDir(tmpDir, func(root *os.Root, path string, d fs.DirEntry) error {
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		data, err := root.ReadFile(path)
		if err != nil {
			return err
		}
		seen[path] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}

	if len(seen) != 1 {
		t.Fatalf("WalkDir() visited %d json files, want 1", len(seen))
	}
	if seen[filepath.Join("nested", "a.json")] != "a" {
		t.Fatalf("WalkDir() data=%q, want %q", seen[filepath.Join("nested", "a.json")], "a")
	}
}
