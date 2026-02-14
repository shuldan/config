package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWithLogger_Option(t *testing.T) {
	t.Parallel()
	b := &builder{}
	opt := WithLogger(nopLogger{})
	opt.apply(b)
	if b.logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestWithLoader_Option(t *testing.T) {
	t.Parallel()
	b := &builder{}
	opt := WithLoader(&staticLoader{data: nil})
	opt.apply(b)
	if len(b.loaders) != 1 {
		t.Errorf("expected 1 loader, got %d", len(b.loaders))
	}
}

func TestWithProfile_YAML(t *testing.T) {
	t.Parallel()
	b := &builder{}
	opt := WithProfile("config.yaml", "prod")
	opt.apply(b)
	if len(b.loaders) != 2 {
		t.Errorf("expected 2 loaders, got %d", len(b.loaders))
	}
}

func TestWithProfile_JSON(t *testing.T) {
	t.Parallel()
	b := &builder{}
	opt := WithProfile("config.json", "dev")
	opt.apply(b)
	if len(b.loaders) != 2 {
		t.Errorf("expected 2 loaders, got %d", len(b.loaders))
	}
}

func TestWithProfileFromEnv_WithProfile(t *testing.T) {
	envKey := "TEST_PROFILE_OPT_XYZ"
	os.Setenv(envKey, "staging")
	t.Cleanup(func() { os.Unsetenv(envKey) })
	b := &builder{}
	opt := WithProfileFromEnv("app.yaml", envKey)
	opt.apply(b)
	if len(b.loaders) != 2 {
		t.Errorf("expected 2 loaders, got %d", len(b.loaders))
	}
}

func TestWithProfileFromEnv_EmptyProfile(t *testing.T) {
	envKey := "TEST_PROFILE_EMPTY_XYZ"
	os.Unsetenv(envKey)
	b := &builder{}
	opt := WithProfileFromEnv("app.yaml", envKey)
	opt.apply(b)
	if len(b.loaders) != 1 {
		t.Errorf("expected 1 loader, got %d", len(b.loaders))
	}
}

func TestWithProfileFromEnv_JSON(t *testing.T) {
	envKey := "TEST_PROFILE_JSON_XYZ"
	os.Setenv(envKey, "test")
	t.Cleanup(func() { os.Unsetenv(envKey) })
	b := &builder{}
	opt := WithProfileFromEnv("app.json", envKey)
	opt.apply(b)
	if len(b.loaders) != 2 {
		t.Errorf("expected 2 loaders, got %d", len(b.loaders))
	}
}

func TestProfileLoaders_JSON_NoProfile(t *testing.T) {
	t.Parallel()
	base, override := profileLoaders(".json", "config.json", "")
	if base == nil || override == nil {
		t.Fatal("expected non-nil loaders")
	}
	data, err := override.Load()
	if err != nil || len(data) != 0 {
		t.Errorf("nopLoader expected empty, got %v err=%v", data, err)
	}
}

func TestProfileLoaders_YAML_NoProfile(t *testing.T) {
	t.Parallel()
	base, override := profileLoaders(".yaml", "config.yaml", "")
	if base == nil || override == nil {
		t.Fatal("expected non-nil loaders")
	}
	data, err := override.Load()
	if err != nil || len(data) != 0 {
		t.Errorf("nopLoader expected empty, got %v err=%v", data, err)
	}
}

func TestNopLoader_Load(t *testing.T) {
	t.Parallel()
	l := nopLoader{}
	m, err := l.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

func TestWithProfileFromEnv_EmptyJSON(t *testing.T) {
	envKey := "TEST_PROF_EMPTY_JSON"
	os.Unsetenv(envKey)
	b := &builder{}
	opt := WithProfileFromEnv("app.json", envKey)
	opt.apply(b)
	if len(b.loaders) != 1 {
		t.Errorf("expected 1 loader, got %d", len(b.loaders))
	}
}

func TestWithProfile_ActualFile(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}
	dir := filepath.Join(wd, "testdata_profile_tmp")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("cannot create dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	writeTestFile(t, dir, "config.json", `{"a":1}`)
	writeTestFile(t, dir, "config.dev.json", `{"a":2}`)

	base := filepath.Join(dir, "config.json")
	cfg, errN := New(WithProfile(base, "dev"))
	if errN != nil {
		t.Fatalf("unexpected error: %v", errN)
	}
	if cfg.GetInt("a") != 2 {
		t.Errorf("expected override 2, got %d", cfg.GetInt("a"))
	}
}
