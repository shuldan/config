package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type yamlLoader struct {
	paths []string
}

func FromYaml(paths ...string) Loader {
	return &yamlLoader{paths: paths}
}

func (l *yamlLoader) Load() (map[string]any, error) {
	for _, path := range l.paths {
		absPath, err := filepath.Abs(path)

		if err != nil {
			continue
		}
		absPath = filepath.Clean(absPath)

		wd, err := os.Getwd()
		if err != nil {
			wd = "."
		}
		secureBase, err := filepath.Abs(wd)
		if err != nil {
			secureBase = "/"
		}
		secureBase = filepath.Clean(secureBase)

		if !strings.HasPrefix(absPath, secureBase+string(filepath.Separator)) {
			continue
		}

		if strings.Contains(absPath, "..") {
			continue
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		var cfg map[string]any
		if err = yaml.UnmarshalWithOptions(data, &cfg, yaml.UseJSONUnmarshaler()); err != nil {
			return nil, errors.Join(ErrParseYAML, err)
		}

		return cfg, nil
	}

	return nil, ErrNoConfigSource
}
