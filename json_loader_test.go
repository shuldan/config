package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestJsonLoader_Load_Success(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeTestFile(t, dir, "c.json", `{"port":8080}`)
	loader := FromJSON(p).WithBasePath(dir)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg["port"] != float64(8080) {
		t.Errorf("expected 8080, got %v", cfg["port"])
	}
}

func TestJsonLoader_Load_FileNotFound_Optional(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	loader := FromJSON(filepath.Join(dir, "nope.json")).WithBasePath(dir).Optional()
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg) != 0 {
		t.Errorf("expected empty, got %v", cfg)
	}
}

func TestJsonLoader_Load_FileNotFound_Required(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	loader := FromJSON(filepath.Join(dir, "nope.json")).WithBasePath(dir)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error")
	}
	var le *LoadError
	if !errors.As(err, &le) {
		t.Error("expected LoadError")
	}
}

func TestJsonLoader_Load_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeTestFile(t, dir, "bad.json", `{invalid}`)
	loader := FromJSON(p).WithBasePath(dir)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error on bad JSON")
	}
	if !errors.Is(err, ErrParseJSON) {
		t.Errorf("expected ErrParseJSON, got %v", err)
	}
}

func TestJsonLoader_Load_PathTraversal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	loader := FromJSON("../../../etc/passwd").WithBasePath(dir)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestJsonLoader_Apply(t *testing.T) {
	t.Parallel()
	b := &builder{}
	l := FromJSON("a.json")
	l.apply(b)
	if len(b.loaders) != 1 {
		t.Errorf("expected 1 loader, got %d", len(b.loaders))
	}
}

func TestJsonLoader_Load_MultiplePaths_FirstFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p2 := writeTestFile(t, dir, "ok.json", `{"key":"val"}`)
	loader := FromJSON(filepath.Join(dir, "nope.json"), p2).WithBasePath(dir)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg["key"] != "val" {
		t.Errorf("expected val, got %v", cfg["key"])
	}
}
