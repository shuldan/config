package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

func processValue(v any, path string) (any, error) {
	switch val := v.(type) {
	case string:
		if strings.Contains(val, "{{") && strings.Contains(val, "}}") {
			result, err := render(val)
			if err != nil {
				return nil, fmt.Errorf("key %q: %w", path, err)
			}
			return result, nil
		}
		return val, nil

	case map[string]any:
		out := make(map[string]any, len(val))
		for k, item := range val {
			childPath := k
			if path != "" {
				childPath = path + "." + k
			}
			processed, err := processValue(item, childPath)
			if err != nil {
				return nil, err
			}
			out[k] = processed
		}
		return out, nil

	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			processed, err := processValue(item, childPath)
			if err != nil {
				return nil, err
			}
			out[i] = processed
		}
		return out, nil

	default:
		return v, nil
	}
}

func newFuncMap() template.FuncMap {
	return template.FuncMap{
		"env": os.Getenv,
		"default": func(def, val any) string {
			if s, ok := val.(string); ok && s != "" {
				return s
			}
			if s, ok := def.(string); ok {
				return s
			}
			return ""
		},
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"trimSpace": strings.TrimSpace,
	}
}

func render(input string) (string, error) {
	tmpl, err := template.New("config").Funcs(newFuncMap()).Parse(input)
	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}

	return buf.String(), nil
}
