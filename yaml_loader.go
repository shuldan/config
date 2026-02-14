package config

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"
)

type yamlLoader struct {
	paths    []string
	basePath string
	optional bool
}

func FromYAML(paths ...string) *yamlLoader {
	return &yamlLoader{paths: paths}
}

func (l *yamlLoader) WithBasePath(path string) *yamlLoader {
	l.basePath = path
	return l
}

func (l *yamlLoader) Optional() *yamlLoader {
	l.optional = true
	return l
}

func (l *yamlLoader) apply(b *builder) {
	b.loaders = append(b.loaders, l)
}

func (l *yamlLoader) Load() (map[string]any, error) {
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
		if err = yaml.UnmarshalWithOptions(data, &cfg, yaml.UseJSONUnmarshaler()); err != nil {
			return nil, errors.Join(ErrParseYAML, err)
		}

		return normalizeMap(cfg), nil
	}

	if l.optional {
		return make(map[string]any), nil
	}

	return nil, &LoadError{Message: "no valid YAML configuration source found", Details: details}
}
