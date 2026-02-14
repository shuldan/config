package config

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRequired_Present(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"host": "a"})
	rule := Required("host")
	if err := rule(cfg); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRequired_Missing(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	rule := Required("host")
	if err := rule(cfg); err == nil {
		t.Error("expected error for missing key")
	}
}

func TestInRange_Valid(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"port": 8080})
	rule := InRange("port", 1, 65535)
	if err := rule(cfg); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestInRange_OutOfRange(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"port": 0})
	rule := InRange("port", 1, 65535)
	if err := rule(cfg); err == nil {
		t.Error("expected error for out of range")
	}
}

func TestInRange_NotNumber(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"port": "abc"})
	rule := InRange("port", 1, 65535)
	if err := rule(cfg); err == nil {
		t.Error("expected error for non-number")
	}
}

func TestInRange_MissingKey(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	rule := InRange("port", 1, 65535)
	if err := rule(cfg); err != nil {
		t.Errorf("expected no error for missing key, got %v", err)
	}
}

func TestOneOf_Valid(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"env": "prod"})
	rule := OneOf("env", "dev", "staging", "prod")
	if err := rule(cfg); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestOneOf_Invalid(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"env": "test"})
	rule := OneOf("env", "dev", "prod")
	if err := rule(cfg); err == nil {
		t.Error("expected error for invalid value")
	}
}

func TestOneOf_Missing(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	rule := OneOf("env", "dev")
	if err := rule(cfg); err != nil {
		t.Errorf("expected no error for missing, got %v", err)
	}
}

func TestMatchRegex_Valid(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"email": "a@b.c"})
	rule := MatchRegex("email", `^.+@.+\..+$`)
	if err := rule(cfg); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMatchRegex_NoMatch(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"email": "bad"})
	rule := MatchRegex("email", `^.+@.+$`)
	if err := rule(cfg); err == nil {
		t.Error("expected error for no match")
	}
}

func TestMatchRegex_BadPattern(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"k": "v"})
	rule := MatchRegex("k", `[invalid`)
	if err := rule(cfg); err == nil {
		t.Error("expected error for bad regex")
	}
}

func TestMatchRegex_Missing(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	rule := MatchRegex("k", `.*`)
	if err := rule(cfg); err != nil {
		t.Errorf("expected no error for missing, got %v", err)
	}
}

func TestCustom_Pass(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"k": "ok"})
	rule := Custom("k", func(v any) error { return nil })
	if err := rule(cfg); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCustom_Fail(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"k": "bad"})
	rule := Custom("k", func(v any) error { return fmt.Errorf("invalid") })
	if err := rule(cfg); err == nil {
		t.Error("expected error from custom validator")
	}
}

func TestValidate_NoViolations(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"a": 1})
	err := cfg.Validate(Required("a"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_WithViolations(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	err := cfg.Validate(Required("a"), Required("b"))
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if len(ve.Violations) != 2 {
		t.Errorf("expected 2 violations, got %d", len(ve.Violations))
	}
}

func TestValidate_MixedRules(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"port": 99999})
	err := cfg.Validate(
		Required("port"),
		InRange("port", 1, 65535),
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("expected 'out of range' in error, got %s", err.Error())
	}
}
