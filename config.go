package config

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"text/template"
)

type Config struct {
	values map[string]any
}

func FromMap(values map[string]any) (*Config, error) {
	return &Config{values: values}, nil
}

func New(loaders ...Loader) (*Config, error) {
	values := make(map[string]any)

	for _, loader := range loaders {
		cfg, err := loader.Load()
		if err != nil {
			return nil, err
		}

		mergeMaps(values, cfg)
	}

	processed := make(map[string]any)
	for k, v := range values {
		processed[k] = processValue(v)
	}

	return FromMap(processed)
}

func (c *Config) Has(key string) bool {
	_, ok := c.find(key)
	return ok
}

func (c *Config) Get(key string) any {
	value, _ := c.find(key)
	return value
}

func (c *Config) GetString(key string, defaultVal ...string) string {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func (c *Config) GetInt(key string, defaultVal ...int) int {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	if i, ok := v.(int); ok {
		return i
	}
	if i, ok := v.(int64); ok {
		if i < int64(math.MinInt) || i > int64(math.MaxInt) {
			return getFirst(defaultVal)
		}
		return int(i)
	}
	if i, ok := v.(uint64); ok {
		if i > uint64(math.MaxInt) {
			return getFirst(defaultVal)
		}
		return int(i)
	}
	if f, ok := v.(float64); ok {
		if f < float64(math.MinInt) || f > float64(math.MaxInt) {
			return getFirst(defaultVal)
		}
		return int(f)
	}
	if b, ok := v.(bool); ok {
		if b {
			return 1
		}
		return 0
	}
	if s, ok := v.(string); ok {
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}
	}
	return getFirst(defaultVal)
}

func (c *Config) GetInt64(key string, defaultVal ...int64) int64 {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	if i, ok := v.(int64); ok {
		return i
	}
	if i, ok := v.(int); ok {
		return int64(i)
	}
	if i, ok := v.(uint64); ok {
		if i > math.MaxInt64 {
			return getFirst(defaultVal)
		}
		return int64(i)
	}
	if f, ok := v.(float64); ok {
		if f < float64(math.MinInt64) || f > float64(math.MaxInt64) {
			return getFirst(defaultVal)
		}
		return int64(f)
	}
	if b, ok := v.(bool); ok {
		return map[bool]int64{true: 1, false: 0}[b]
	}
	if s, ok := v.(string); ok {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
	}
	return getFirst(defaultVal)
}

func (c *Config) GetFloat64(key string, defaultVal ...float64) float64 {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	if f, ok := v.(float64); ok {
		return f
	}
	if i, ok := v.(int); ok {
		return float64(i)
	}
	if i, ok := v.(int64); ok {
		return float64(i)
	}
	if s, ok := v.(string); ok {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}
	return getFirst(defaultVal)
}

func (c *Config) GetBool(key string, defaultVal ...bool) bool {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	if b, ok := v.(bool); ok {
		return b
	}
	if s, ok := v.(string); ok {
		switch strings.ToLower(s) {
		case "true", "1", "on", "yes", "y":
			return true
		case "false", "0", "off", "no", "n":
			return false
		}
	}
	if f, ok := v.(float64); ok {
		return f != 0
	}
	if i, ok := v.(int); ok {
		return i != 0
	}
	return getFirst(defaultVal)
}

func (c *Config) GetStringSlice(key string, separator ...string) []string {
	v, ok := c.find(key)
	if !ok {
		return nil
	}
	if v == nil {
		return nil
	}

	sep := ","
	if len(separator) > 0 {
		sep = separator[0]
	}

	switch val := v.(type) {
	case []string:
		return val
	case []any:
		result := make([]string, len(val))
		for i, item := range val {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	case string:
		parts := strings.Split(val, sep)
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	default:
		return []string{fmt.Sprintf("%v", v)}
	}
}

func (c *Config) GetSub(key string) (*Config, bool) {
	sub, ok := c.find(key)
	if !ok {
		return nil, false
	}
	if subMap, ok := sub.(map[string]any); ok {
		return &Config{values: subMap}, true
	}
	return nil, false
}

func (c *Config) All() map[string]any {
	cp := make(map[string]any, len(c.values))
	for k, v := range c.values {
		cp[k] = v
	}
	return cp
}

func (c *Config) find(path string) (any, bool) {
	keys := strings.Split(path, ".")
	var current any = c.values

	for _, k := range keys {
		if current == nil {
			return nil, false
		}

		switch cur := current.(type) {
		case map[string]any:
			next, exists := cur[k]
			if !exists {
				return nil, false
			}
			current = next
		case map[any]any:
			next, exists := cur[k]
			if !exists {
				return nil, false
			}
			current = next
		default:
			return nil, false
		}
	}

	return current, true
}

func getFirst[T any](values []T) T {
	var zero T
	if len(values) > 0 {
		return values[0]
	}
	return zero
}

func mergeMaps(dst, src map[string]any) {
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if dstV, exists := dst[k]; exists {
				if dstMap, ok := dstV.(map[string]any); ok {
					mergeMaps(dstMap, vMap)
					continue
				}
			}
		}
		dst[k] = v
	}
}

func processValue(v any) any {
	switch val := v.(type) {
	case string:
		if strings.Contains(val, "{{") && strings.Contains(val, "}}") {
			result, _ := render(val)
			return result
		}
		return val
	case map[string]any:
		mapped := make(map[string]any)
		for k, v := range val {
			mapped[k] = processValue(v)
		}
		return mapped
	case []any:
		var result []any
		for _, item := range val {
			result = append(result, processValue(item))
		}
		return result
	default:
		return val
	}
}

func newFuncMap() template.FuncMap {
	return template.FuncMap{
		"default": func(def, val interface{}) string {
			s, ok := val.(string)
			if !ok || s == "" {
				if s, ok := def.(string); ok {
					return s
				}
				return ""
			}
			return s
		},
		"env":   os.Getenv,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}
}

func render(input string) (string, error) {
	tmpl, err := template.New("config").Funcs(newFuncMap()).Parse(input)
	if err != nil {
		return "", err
	}

	data := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			data[parts[0]] = parts[1]
		}
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
