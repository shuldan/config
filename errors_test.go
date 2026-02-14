package config

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadError_Error_NoDetails(t *testing.T) {
	t.Parallel()
	e := &LoadError{Message: "test"}
	got := e.Error()
	if got != "config: test" {
		t.Errorf("expected %q, got %q", "config: test", got)
	}
}

func TestLoadError_Error_WithDetails(t *testing.T) {
	t.Parallel()
	e := &LoadError{
		Message: "failed",
		Details: []LoadErrorDetail{
			{Path: "/a", Reason: "not found"},
			{Path: "/b", Reason: "denied"},
		},
	}
	got := e.Error()
	if !strings.Contains(got, "failed") {
		t.Error("expected message in output")
	}
	if !strings.Contains(got, "/a") || !strings.Contains(got, "/b") {
		t.Error("expected paths in output")
	}
}

func TestLoadError_Unwrap(t *testing.T) {
	t.Parallel()
	e := &LoadError{Message: "x"}
	if !errors.Is(e, ErrNoConfigSource) {
		t.Error("expected Unwrap to return ErrNoConfigSource")
	}
}

func TestValidationError_Error(t *testing.T) {
	t.Parallel()
	e := &ValidationError{Violations: []string{"a is missing", "b out of range"}}
	got := e.Error()
	if !strings.Contains(got, "validation failed") {
		t.Error("expected header")
	}
	if !strings.Contains(got, "a is missing") || !strings.Contains(got, "b out of range") {
		t.Error("expected violations")
	}
}

func TestValidationError_Empty(t *testing.T) {
	t.Parallel()
	e := &ValidationError{}
	got := e.Error()
	if !strings.Contains(got, "validation failed") {
		t.Error("expected header even with no violations")
	}
}
