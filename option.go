package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Option interface {
	apply(b *builder)
}

type builder struct {
	loaders []Loader
	logger  Logger
}

type optionFunc func(*builder)

func (f optionFunc) apply(b *builder) { f(b) }

func WithLogger(l Logger) Option {
	return optionFunc(func(b *builder) {
		b.logger = l
	})
}

func WithLoader(l Loader) Option {
	return optionFunc(func(b *builder) {
		b.loaders = append(b.loaders, l)
	})
}

func WithProfile(basePath string, profile string) Option {
	return optionFunc(func(b *builder) {
		ext := filepath.Ext(basePath)
		name := strings.TrimSuffix(basePath, ext)
		profilePath := name + "." + profile + ext

		base, override := profileLoaders(ext, basePath, profilePath)
		b.loaders = append(b.loaders, base, override)
	})
}

func WithProfileFromEnv(basePath string, envVar string) Option {
	return optionFunc(func(b *builder) {
		profile := os.Getenv(envVar)
		ext := filepath.Ext(basePath)

		if profile == "" {
			base, _ := profileLoaders(ext, basePath, "")
			b.loaders = append(b.loaders, base)
			return
		}

		name := strings.TrimSuffix(basePath, ext)
		profilePath := name + "." + profile + ext

		base, override := profileLoaders(ext, basePath, profilePath)
		b.loaders = append(b.loaders, base, override)
	})
}

func profileLoaders(ext, basePath, profilePath string) (Loader, Loader) {
	switch strings.ToLower(ext) {
	case ".json":
		base := &jsonLoader{paths: []string{basePath}}
		var override Loader
		if profilePath != "" {
			override = &jsonLoader{paths: []string{profilePath}, optional: true}
		} else {
			override = &nopLoader{}
		}
		return base, override
	default:
		base := &yamlLoader{paths: []string{basePath}}
		var override Loader
		if profilePath != "" {
			override = &yamlLoader{paths: []string{profilePath}, optional: true}
		} else {
			override = &nopLoader{}
		}
		return base, override
	}
}

type nopLoader struct{}

func (nopLoader) Load() (map[string]any, error) { return make(map[string]any), nil }
