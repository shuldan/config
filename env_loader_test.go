package config

import (
	"os"
	"testing"
)

func TestEnvLoader_Load_Basic(t *testing.T) {
	prefix := "TESTENVLDR_"
	t.Cleanup(func() {
		os.Unsetenv(prefix + "HOST")
		os.Unsetenv(prefix + "DB__PORT")
	})
	os.Setenv(prefix+"HOST", "localhost")
	os.Setenv(prefix+"DB__PORT", "5432")

	loader := FromEnv(prefix)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg["host"] != "localhost" {
		t.Errorf("expected localhost, got %v", cfg["host"])
	}
	dbMap, ok := cfg["db"].(map[string]any)
	if !ok {
		t.Fatalf("expected db to be map, got %T", cfg["db"])
	}
	if dbMap["port"] != "5432" {
		t.Errorf("expected 5432, got %v", dbMap["port"])
	}
}

func TestEnvLoader_WithAutoTypeParse(t *testing.T) {
	prefix := "TESTATP_"
	t.Cleanup(func() {
		os.Unsetenv(prefix + "NUM")
		os.Unsetenv(prefix + "FLAG")
		os.Unsetenv(prefix + "F")
	})
	os.Setenv(prefix+"NUM", "42")
	os.Setenv(prefix+"FLAG", "true")
	os.Setenv(prefix+"F", "3.14")

	loader := FromEnv(prefix).WithAutoTypeParse()
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg["num"] != 42 {
		t.Errorf("expected int 42, got %v (%T)", cfg["num"], cfg["num"])
	}
	if cfg["flag"] != true {
		t.Errorf("expected true, got %v", cfg["flag"])
	}
	if cfg["f"] != 3.14 {
		t.Errorf("expected 3.14, got %v", cfg["f"])
	}
}

func TestEnvLoader_NoMatchingPrefix(t *testing.T) {
	t.Parallel()
	loader := FromEnv("ZZZZZZZ_NONEXIST_")
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg) != 0 {
		t.Errorf("expected empty map, got %v", cfg)
	}
}

func TestEnvLoader_Apply(t *testing.T) {
	t.Parallel()
	b := &builder{}
	loader := FromEnv("TEST_")
	loader.apply(b)
	if len(b.loaders) != 1 {
		t.Errorf("expected 1 loader, got %d", len(b.loaders))
	}
}
