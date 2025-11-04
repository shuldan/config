package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type jsonLoader struct {
	paths []string
}

func FromJSON(paths ...string) Loader {
	return &jsonLoader{paths: paths}
}

func (l *jsonLoader) Load() (map[string]any, error) {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	secureBase, err := filepath.Abs(wd)
	if err != nil {
		secureBase = "/"
	}
	secureBase = filepath.Clean(secureBase)

	for _, path := range l.paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		absPath = filepath.Clean(absPath)

		if !strings.HasPrefix(absPath, secureBase+string(filepath.Separator)) {
			continue
		}

		if strings.Contains(absPath, "..") {
			continue
		}

		if !fileExists(absPath) {
			continue
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		var cfg map[string]any
		if err = json.Unmarshal(data, &cfg); err != nil {
			return nil, errors.Join(ErrParseJSON, err)
		}

		return cfg, nil
	}

	return nil, ErrNoConfigSource
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
