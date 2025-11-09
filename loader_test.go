package config

import (
	"errors"
	"testing"
)

type mockLoader struct {
	data map[string]any
	err  error
}

func (m *mockLoader) Load() (map[string]any, error) {
	return m.data, m.err
}

func TestFromEnv_Load(t *testing.T) {
	t.Setenv("TEST_KEY", "value")
	t.Setenv("TEST_NUM", "42")
	t.Setenv("TEST_BOOL", "true")
	t.Setenv("TEST_PARENT__CHILD", "nested")

	loader := FromEnv("TEST_")
	data, err := loader.Load()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if data["key"] != "value" {
		t.Errorf("expected value, got %v", data["key"])
	}
	if data["num"] != "42" {
		t.Errorf("expected 42, got %v", data["num"])
	}
	if data["bool"] != "true" {
		t.Errorf("expected true, got %v", data["bool"])
	}
	parent := data["parent"].(map[string]any)
	if parent["child"] != "nested" {
		t.Errorf("expected nested, got %v", parent["child"])
	}
}

func TestFromJSON_Load_FileNotFound(t *testing.T) {
	t.Parallel()
	loader := FromJSON("nonexistent.json")
	_, err := loader.Load()
	if !errors.Is(err, ErrNoConfigSource) {
		t.Errorf("expected ErrNoConfigSource, got %v", err)
	}
}

func TestFromJSON_Load_InvalidJSON(t *testing.T) {
	t.Parallel()
	loader := FromJSON("testdata/invalid.json")
	_, err := loader.Load()
	if err == nil {
		t.Error("expected error")
	}
}

func TestFromYaml_Load_FileNotFound(t *testing.T) {
	t.Parallel()
	loader := FromYaml("nonexistent.yaml")
	_, err := loader.Load()
	if !errors.Is(err, ErrNoConfigSource) {
		t.Errorf("expected ErrNoConfigSource, got %v", err)
	}
}

func TestFromYaml_Load_InvalidYAML(t *testing.T) {
	t.Parallel()
	loader := FromYaml("testdata/invalid.yaml")
	_, err := loader.Load()
	if err == nil {
		t.Error("expected error")
	}
}

func TestFileExists_Exists(t *testing.T) {
	t.Parallel()
	result := fileExists("config.go")
	if !result {
		t.Error("expected file to exist")
	}
}

func TestFileExists_NotExists(t *testing.T) {
	t.Parallel()
	result := fileExists("nonexistent")
	if result {
		t.Error("expected file to not exist")
	}
}

func TestFileExists_IsDir(t *testing.T) {
	t.Parallel()
	result := fileExists(".")
	if result {
		t.Error("expected directory to return false")
	}
}
