package config

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestYamlLoader_Load_Success(t *testing.T) {
	dir := t.TempDir()
	p := writeTestFile(t, dir, "c.yaml", "port: 8080\nhost: localhost\n")
	loader := FromYAML(p).WithBasePath(dir)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg["host"] != "localhost" {
		t.Errorf("expected localhost, got %v", cfg["host"])
	}
}

func TestYamlLoader_Load_FileNotFound_Optional(t *testing.T) {
	dir := t.TempDir()
	loader := FromYAML(filepath.Join(dir, "nope.yaml")).WithBasePath(dir).Optional()
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg) != 0 {
		t.Errorf("expected empty, got %v", cfg)
	}
}

func TestYamlLoader_Load_FileNotFound_Required(t *testing.T) {
	dir := t.TempDir()
	loader := FromYAML(filepath.Join(dir, "nope.yaml")).WithBasePath(dir)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error")
	}
	var le *LoadError
	if !errors.As(err, &le) {
		t.Error("expected LoadError")
	}
}

func TestYamlLoader_Load_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	p := writeTestFile(t, dir, "bad.yaml", ":\n  :\n    : [invalid yaml")
	loader := FromYAML(p).WithBasePath(dir)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for bad YAML")
	}
	if !errors.Is(err, ErrParseYAML) {
		t.Errorf("expected ErrParseYAML, got %v", err)
	}
}

func TestYamlLoader_Load_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	loader := FromYAML("../../../etc/passwd").WithBasePath(dir)
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestYamlLoader_Apply(t *testing.T) {
	t.Parallel()
	b := &builder{}
	l := FromYAML("a.yaml")
	l.apply(b)
	if len(b.loaders) != 1 {
		t.Errorf("expected 1 loader, got %d", len(b.loaders))
	}
}

func TestYamlLoader_MultiplePaths_FirstFails(t *testing.T) {
	dir := t.TempDir()
	p2 := writeTestFile(t, dir, "ok.yaml", "key: val\n")
	loader := FromYAML(filepath.Join(dir, "nope.yaml"), p2).WithBasePath(dir)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg["key"] != "val" {
		t.Errorf("expected val, got %v", cfg["key"])
	}
}
