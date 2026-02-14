package config

import (
	"os"
	"strings"
)

type EnvLoader struct {
	prefix        string
	autoTypeParse bool
}

func FromEnv(prefix string) *EnvLoader {
	return &EnvLoader{prefix: prefix}
}

func (l *EnvLoader) WithAutoTypeParse() *EnvLoader {
	l.autoTypeParse = true
	return l
}

func (l *EnvLoader) apply(b *builder) {
	b.loaders = append(b.loaders, l)
}

func (l *EnvLoader) Load() (map[string]any, error) {
	cfg := make(map[string]any)

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, l.prefix) {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		configKey := strings.ToLower(strings.TrimPrefix(key, l.prefix))
		configKey = strings.ReplaceAll(configKey, "__", ".")

		var parsed any = value
		if l.autoTypeParse {
			parsed = autoParseString(value)
		}

		setNested(cfg, configKey, parsed)
	}

	return cfg, nil
}
