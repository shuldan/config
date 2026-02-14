package config

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type ConfigProvider interface {
	Has(key string) bool
	Get(key string) any
	GetString(key string, defaultVal ...string) string
	GetInt(key string, defaultVal ...int) int
	GetInt64(key string, defaultVal ...int64) int64
	GetUint64(key string, defaultVal ...uint64) uint64
	GetFloat64(key string, defaultVal ...float64) float64
	GetBool(key string, defaultVal ...bool) bool
	GetDuration(key string, defaultVal ...time.Duration) time.Duration
	GetTime(key string, layout string, defaultVal ...time.Time) time.Time
	GetStringSlice(key string, separator ...string) []string
	GetIntSlice(key string) []int
	GetFloat64Slice(key string) []float64
	GetMap(key string) (map[string]any, bool)
	GetSub(key string) (ConfigProvider, bool)
	Unmarshal(key string, target any) error
	All() map[string]any
}

var _ ConfigProvider = (*Config)(nil)

type Config struct {
	values map[string]any
}

func New(opts ...Option) (*Config, error) {
	b := &builder{
		logger: nopLogger{},
	}
	for _, opt := range opts {
		opt.apply(b)
	}

	values := make(map[string]any)

	for _, loader := range b.loaders {
		cfg, err := loader.Load()
		if err != nil {
			b.logger.Debug("config: loader failed", "error", err)
			return nil, err
		}
		b.logger.Debug("config: loader succeeded", "keys", len(cfg))
		mergeMaps(values, cfg)
	}

	processed, err := processValue(values, "")
	if err != nil {
		return nil, fmt.Errorf("config: template rendering failed: %w", err)
	}

	processedMap, ok := processed.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("config: unexpected processed type %T", processed)
	}

	b.logger.Debug("config: ready", "total_keys", len(processedMap))

	return &Config{values: processedMap}, nil
}

func FromMap(values map[string]any) *Config {
	return &Config{values: deepCopyMap(values)}
}

func (c *Config) WithOverrides(overrides map[string]any) *Config {
	cp := deepCopyMap(c.values)
	expanded := expandDotKeys(overrides)
	mergeMaps(cp, expanded)
	return &Config{values: cp}
}

func (c *Config) Has(key string) bool {
	_, ok := c.find(key)
	return ok
}

func (c *Config) Get(key string) any {
	v, _ := c.find(key)
	return v
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
	if i, ok := toInt(v); ok {
		return i
	}
	return getFirst(defaultVal)
}

func (c *Config) GetInt64(key string, defaultVal ...int64) int64 {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case uint64:
		if val > math.MaxInt64 {
			return getFirst(defaultVal)
		}
		return int64(val)
	case float64:
		if val < float64(math.MinInt64) || val > float64(math.MaxInt64) {
			return getFirst(defaultVal)
		}
		return int64(val)
	case bool:
		if val {
			return 1
		}
		return 0
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return getFirst(defaultVal)
}

func (c *Config) GetUint64(key string, defaultVal ...uint64) uint64 {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	switch val := v.(type) {
	case uint64:
		return val
	case int:
		if val < 0 {
			return getFirst(defaultVal)
		}
		return uint64(val)
	case int64:
		if val < 0 {
			return getFirst(defaultVal)
		}
		return uint64(val)
	case float64:
		if val < 0 || val > float64(math.MaxUint64) {
			return getFirst(defaultVal)
		}
		return uint64(val)
	case string:
		if i, err := strconv.ParseUint(val, 10, 64); err == nil {
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
	if f, ok := toFloat64(v); ok {
		return f
	}
	return getFirst(defaultVal)
}

func (c *Config) GetBool(key string, defaultVal ...bool) bool {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	if b, ok := toBool(v); ok {
		return b
	}
	return getFirst(defaultVal)
}

func (c *Config) GetDuration(key string, defaultVal ...time.Duration) time.Duration {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	switch val := v.(type) {
	case time.Duration:
		return val
	case string:
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	case int:
		return time.Duration(val) * time.Millisecond
	case int64:
		return time.Duration(val) * time.Millisecond
	case float64:
		return time.Duration(val * float64(time.Millisecond))
	}
	return getFirst(defaultVal)
}

func (c *Config) GetTime(key string, layout string, defaultVal ...time.Time) time.Time {
	v, ok := c.find(key)
	if !ok {
		return getFirst(defaultVal)
	}
	if s, ok := v.(string); ok {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return getFirst(defaultVal)
}

func (c *Config) GetStringSlice(key string, separator ...string) []string {
	v, ok := c.find(key)
	if !ok || v == nil {
		return nil
	}

	sep := ","
	if len(separator) > 0 {
		sep = separator[0]
	}

	switch val := v.(type) {
	case []string:
		cp := make([]string, len(val))
		copy(cp, val)
		return cp
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

func (c *Config) GetIntSlice(key string) []int {
	v, ok := c.find(key)
	if !ok || v == nil {
		return nil
	}

	switch val := v.(type) {
	case []int:
		cp := make([]int, len(val))
		copy(cp, val)
		return cp
	case []any:
		out := make([]int, 0, len(val))
		for _, item := range val {
			if i, ok := toInt(item); ok {
				out = append(out, i)
			}
		}
		return out
	case []float64:
		out := make([]int, len(val))
		for i, f := range val {
			out[i] = int(f)
		}
		return out
	default:
		return nil
	}
}

func (c *Config) GetFloat64Slice(key string) []float64 {
	v, ok := c.find(key)
	if !ok || v == nil {
		return nil
	}

	switch val := v.(type) {
	case []float64:
		cp := make([]float64, len(val))
		copy(cp, val)
		return cp
	case []any:
		out := make([]float64, 0, len(val))
		for _, item := range val {
			if f, ok := toFloat64(item); ok {
				out = append(out, f)
			}
		}
		return out
	case []int:
		out := make([]float64, len(val))
		for i, n := range val {
			out[i] = float64(n)
		}
		return out
	default:
		return nil
	}
}

func (c *Config) GetMap(key string) (map[string]any, bool) {
	v, ok := c.find(key)
	if !ok {
		return nil, false
	}
	if m, ok := v.(map[string]any); ok {
		return deepCopyMap(m), true
	}
	return nil, false
}

func (c *Config) GetSub(key string) (ConfigProvider, bool) {
	sub, ok := c.find(key)
	if !ok {
		return nil, false
	}
	if subMap, ok := sub.(map[string]any); ok {
		return &Config{values: deepCopyMap(subMap)}, true
	}
	return nil, false
}

func (c *Config) All() map[string]any {
	return deepCopyMap(c.values)
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
