package config

import "errors"

var (
	ErrNoConfigSource = errors.New("no valid configuration source found")
	ErrParseYAML      = errors.New("failed to parse YAML file")
	ErrParseJSON      = errors.New("failed to parse JSON file")
)
