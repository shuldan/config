package config

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func deepCopyMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = deepCopyValue(v)
	}
	return dst
}

func deepCopyValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return deepCopyMap(val)
	case []any:
		cp := make([]any, len(val))
		for i, item := range val {
			cp[i] = deepCopyValue(item)
		}
		return cp
	case []string:
		cp := make([]string, len(val))
		copy(cp, val)
		return cp
	default:
		return v
	}
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

func normalizeMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = normalizeValue(v)
	}
	return out
}

func normalizeValue(v any) any {
	switch val := v.(type) {
	case map[any]any:
		m := make(map[string]any, len(val))
		for k, v := range val {
			m[fmt.Sprintf("%v", k)] = normalizeValue(v)
		}
		return m
	case map[string]any:
		return normalizeMap(val)
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = normalizeValue(item)
		}
		return out
	default:
		return v
	}
}

func setNested(m map[string]any, key string, value any) {
	keys := strings.Split(key, ".")
	last := len(keys) - 1

	current := m
	for i, k := range keys {
		if i == last {
			current[k] = value
			return
		}

		if _, ok := current[k]; !ok {
			current[k] = make(map[string]any)
		}

		if next, ok := current[k].(map[string]any); ok {
			current = next
		} else {
			next := make(map[string]any)
			current[k] = next
			current = next
		}
	}
}

func expandDotKeys(flat map[string]any) map[string]any {
	out := make(map[string]any)
	for k, v := range flat {
		setNested(out, k, v)
	}
	return out
}

func resolveSecurePath(path string, basePath string) (string, error) {
	if basePath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot determine working directory: %w", err)
		}
		basePath = wd
	}

	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve base path: %w", err)
	}
	absBase = filepath.Clean(absBase)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}
	absPath = filepath.Clean(absPath)

	if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside allowed base %q", path, absBase)
	}

	return absPath, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func getFirst[T any](values []T) T {
	var zero T
	if len(values) > 0 {
		return values[0]
	}
	return zero
}

func autoParseString(s string) any {
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	}

	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		if i >= math.MinInt && i <= math.MaxInt {
			return int(i)
		}
		return i
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	return s
}

func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		if val < int64(math.MinInt) || val > int64(math.MaxInt) {
			return 0, false
		}
		return int(val), true
	case uint64:
		if val > uint64(math.MaxInt) {
			return 0, false
		}
		return int(val), true
	case float64:
		if val < float64(math.MinInt) || val > float64(math.MaxInt) {
			return 0, false
		}
		return int(val), true
	case bool:
		if val {
			return 1, true
		}
		return 0, true
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i, true
		}
	}
	return 0, false
}

func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}
