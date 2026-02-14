package config

import (
	"encoding/json"
	"errors"
	"os"
)

type jsonLoader struct {
	paths    []string
	basePath string
	optional bool
}

func FromJSON(paths ...string) *jsonLoader {
	return &jsonLoader{paths: paths}
}

func (l *jsonLoader) WithBasePath(path string) *jsonLoader {
	l.basePath = path
	return l
}

func (l *jsonLoader) Optional() *jsonLoader {
	l.optional = true
	return l
}

func (l *jsonLoader) apply(b *builder) {
	b.loaders = append(b.loaders, l)
}

func (l *jsonLoader) Load() (map[string]any, error) {
	var details []LoadErrorDetail

	for _, path := range l.paths {
		absPath, err := resolveSecurePath(path, l.basePath)
		if err != nil {
			details = append(details, LoadErrorDetail{Path: path, Reason: err.Error()})
			continue
		}

		if !fileExists(absPath) {
			details = append(details, LoadErrorDetail{Path: path, Reason: "file not found"})
			continue
		}

		data, err := os.ReadFile(absPath) // #nosec G304 -- path validated by resolveSecurePath
		if err != nil {
			details = append(details, LoadErrorDetail{Path: path, Reason: err.Error()})
			continue
		}

		var cfg map[string]any
		if err = json.Unmarshal(data, &cfg); err != nil {
			return nil, errors.Join(ErrParseJSON, err)
		}

		return normalizeMap(cfg), nil
	}

	if l.optional {
		return make(map[string]any), nil
	}

	return nil, &LoadError{Message: "no valid JSON configuration source found", Details: details}
}
