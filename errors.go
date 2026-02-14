package config

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNoConfigSource = errors.New("no valid configuration source found")
	ErrParseYAML      = errors.New("failed to parse YAML")
	ErrParseJSON      = errors.New("failed to parse JSON")
)

type LoadErrorDetail struct {
	Path   string
	Reason string
}

type LoadError struct {
	Message string
	Details []LoadErrorDetail
}

func (e *LoadError) Error() string {
	if len(e.Details) == 0 {
		return "config: " + e.Message
	}

	var b strings.Builder
	b.WriteString("config: ")
	b.WriteString(e.Message)
	b.WriteString(":")

	for _, d := range e.Details {
		b.WriteString("\n  - ")
		fmt.Fprintf(&b, "%q: %s", d.Path, d.Reason)
	}

	return b.String()
}

func (e *LoadError) Unwrap() error {
	return ErrNoConfigSource
}

type ValidationError struct {
	Violations []string
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("config: validation failed:")

	for _, v := range e.Violations {
		b.WriteString("\n  - ")
		b.WriteString(v)
	}

	return b.String()
}
