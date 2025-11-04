package config

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestFromMap_Success(t *testing.T) {
	t.Parallel()
	values := map[string]any{
		"key": "value",
	}
	cfg, err := FromMap(values)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Error("expected config, got nil")
	}
}

func TestNew_NoLoaders(t *testing.T) {
	t.Parallel()
	cfg, err := New()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Error("expected config, got nil")
	}
}

func TestNew_WithLoaderError(t *testing.T) {
	t.Parallel()
	loader := &mockLoader{err: errors.New("load error")}
	_, err := New(loader)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestNew_WithLoaders(t *testing.T) {
	t.Parallel()
	loader := &mockLoader{
		data: map[string]any{
			"key": "value",
		},
	}
	cfg, err := New(loader)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Error("expected config, got nil")
	}
}

func TestConfig_Has_Exists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "value",
		},
	}
	if !cfg.Has("key") {
		t.Error("expected key to exist")
	}
}

func TestConfig_Has_NotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "value",
		},
	}
	if cfg.Has("missing") {
		t.Error("expected key to not exist")
	}
}

func TestConfig_Has_NestedExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": map[string]any{
				"child": "value",
			},
		},
	}
	if !cfg.Has("parent.child") {
		t.Error("expected nested key to exist")
	}
}

func TestConfig_Has_NestedNotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": map[string]any{
				"child": "value",
			},
		},
	}
	if cfg.Has("parent.missing") {
		t.Error("expected nested key to not exist")
	}
}

func TestConfig_Has_NestedInvalidType(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": "not a map",
		},
	}
	if cfg.Has("parent.child") {
		t.Error("expected false for non-map parent")
	}
}

func TestConfig_Get_Exists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "value",
		},
	}
	value := cfg.Get("key")
	if value != "value" {
		t.Errorf("expected value, got %v", value)
	}
}

func TestConfig_Get_NotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "value",
		},
	}
	value := cfg.Get("missing")
	if value != nil {
		t.Errorf("expected nil, got %v", value)
	}
}

func TestConfig_Get_Nested(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": map[string]any{
				"child": "value",
			},
		},
	}
	value := cfg.Get("parent.child")
	if value != "value" {
		t.Errorf("expected value, got %v", value)
	}
}

func TestConfig_GetString_Exists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "value",
		},
	}
	value := cfg.GetString("key")
	if value != "value" {
		t.Errorf("expected value, got %v", value)
	}
}

func TestConfig_GetString_NotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "value",
		},
	}
	value := cfg.GetString("missing", "default")
	if value != "default" {
		t.Errorf("expected default, got %v", value)
	}
}

func TestConfig_GetString_Nil(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": nil,
		},
	}
	value := cfg.GetString("key")
	if value != "" {
		t.Errorf("expected empty string, got %v", value)
	}
}

func TestConfig_GetString_TypeMismatch(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": 123,
		},
	}
	value := cfg.GetString("key")
	if value != "123" {
		t.Errorf("expected 123 as string, got %v", value)
	}
}

func TestConfig_GetInt_Exists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": 42,
		},
	}
	value := cfg.GetInt("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt_NotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": 42,
		},
	}
	value := cfg.GetInt("missing", 99)
	if value != 99 {
		t.Errorf("expected 99, got %v", value)
	}
}

func TestConfig_GetInt_FromInt64(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": int64(42),
		},
	}
	value := cfg.GetInt("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt_FromInt64Overflow(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": int64(math.MaxInt64),
		},
	}
	value := cfg.GetInt("key", 99)
	if value == 99 {
		t.Errorf("expected MaxInt64 converted to int, got default 99")
	}
}

func TestConfig_GetInt_FromUint64(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": uint64(42),
		},
	}
	value := cfg.GetInt("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt_FromUint64Overflow(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": uint64(math.MaxInt64 + 1),
		},
	}
	value := cfg.GetInt("key", 99)
	if value != 99 {
		t.Errorf("expected 99, got %v", value)
	}
}

func TestConfig_GetInt_FromFloat64(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": float64(42.0),
		},
	}
	value := cfg.GetInt("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt_FromFloat64Overflow(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": float64(math.MaxInt64) * 2,
		},
	}
	value := cfg.GetInt("key", 99)
	if value != 99 {
		t.Errorf("expected 99, got %v", value)
	}
}

func TestConfig_GetInt_FromBool(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": true,
		},
	}
	value := cfg.GetInt("key")
	if value != 1 {
		t.Errorf("expected 1, got %v", value)
	}
}

func TestConfig_GetInt_FromBoolFalse(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": false,
		},
	}
	value := cfg.GetInt("key")
	if value != 0 {
		t.Errorf("expected 0, got %v", value)
	}
}

func TestConfig_GetInt_FromString(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "42",
		},
	}
	value := cfg.GetInt("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt_FromInvalidString(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "invalid",
		},
	}
	value := cfg.GetInt("key", 99)
	if value != 99 {
		t.Errorf("expected 99, got %v", value)
	}
}

func TestConfig_GetInt64_Exists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": int64(42),
		},
	}
	value := cfg.GetInt64("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt64_NotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": int64(42),
		},
	}
	value := cfg.GetInt64("missing", 99)
	if value != 99 {
		t.Errorf("expected 99, got %v", value)
	}
}

func TestConfig_GetInt64_FromInt(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": 42,
		},
	}
	value := cfg.GetInt64("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt64_FromUint64(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": uint64(42),
		},
	}
	value := cfg.GetInt64("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt64_FromUint64Overflow(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": uint64(1 << 63),
		},
	}
	value := cfg.GetInt64("key", 99)
	if value != 99 {
		t.Errorf("expected 99, got %v", value)
	}
}

func TestConfig_GetInt64_FromFloat64(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": float64(42.0),
		},
	}
	value := cfg.GetInt64("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt64_FromFloat64Overflow(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": float64(math.MaxInt64) * 2,
		},
	}
	value := cfg.GetInt64("key", 99)
	if value != 99 {
		t.Errorf("expected 99, got %v", value)
	}
}

func TestConfig_GetInt64_FromBool(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": true,
		},
	}
	value := cfg.GetInt64("key")
	if value != 1 {
		t.Errorf("expected 1, got %v", value)
	}
}

func TestConfig_GetInt64_FromString(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "42",
		},
	}
	value := cfg.GetInt64("key")
	if value != 42 {
		t.Errorf("expected 42, got %v", value)
	}
}

func TestConfig_GetInt64_FromInvalidString(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "invalid",
		},
	}
	value := cfg.GetInt64("key", 99)
	if value != 99 {
		t.Errorf("expected 99, got %v", value)
	}
}

func TestConfig_GetFloat64_Exists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": float64(42.5),
		},
	}
	value := cfg.GetFloat64("key")
	if value != 42.5 {
		t.Errorf("expected 42.5, got %v", value)
	}
}

func TestConfig_GetFloat64_NotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": float64(42.5),
		},
	}
	value := cfg.GetFloat64("missing", 99.5)
	if value != 99.5 {
		t.Errorf("expected 99.5, got %v", value)
	}
}

func TestConfig_GetFloat64_FromInt(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": 42,
		},
	}
	value := cfg.GetFloat64("key")
	if value != 42.0 {
		t.Errorf("expected 42.0, got %v", value)
	}
}

func TestConfig_GetFloat64_FromInt64(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": int64(42),
		},
	}
	value := cfg.GetFloat64("key")
	if value != 42.0 {
		t.Errorf("expected 42.0, got %v", value)
	}
}

func TestConfig_GetFloat64_FromString(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "42.5",
		},
	}
	value := cfg.GetFloat64("key")
	if value != 42.5 {
		t.Errorf("expected 42.5, got %v", value)
	}
}

func TestConfig_GetFloat64_FromInvalidString(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "invalid",
		},
	}
	value := cfg.GetFloat64("key", 99.5)
	if value != 99.5 {
		t.Errorf("expected 99.5, got %v", value)
	}
}

func TestConfig_GetBool_Exists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": true,
		},
	}
	value := cfg.GetBool("key")
	if !value {
		t.Error("expected true")
	}
}

func TestConfig_GetBool_NotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": true,
		},
	}
	value := cfg.GetBool("missing", false)
	if value {
		t.Error("expected false")
	}
}

func TestConfig_GetBool_FromStringTrue(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"true", true},
		{"1", true},
		{"on", true},
		{"yes", true},
		{"y", true},
		{"TRUE", true},
		{"True", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{
				values: map[string]any{
					"key": tt.input,
				},
			}
			value := cfg.GetBool("key")
			if value != tt.want {
				t.Errorf("expected %v, got %v", tt.want, value)
			}
		})
	}
}

func TestConfig_GetBool_FromStringFalse(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"false", false},
		{"0", false},
		{"off", false},
		{"no", false},
		{"n", false},
		{"FALSE", false},
		{"False", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{
				values: map[string]any{
					"key": tt.input,
				},
			}
			value := cfg.GetBool("key")
			if value != tt.want {
				t.Errorf("expected %v, got %v", tt.want, value)
			}
		})
	}
}

func TestConfig_GetBool_FromFloat64(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": float64(1.0),
		},
	}
	value := cfg.GetBool("key")
	if !value {
		t.Error("expected true")
	}
}

func TestConfig_GetBool_FromFloat64Zero(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": float64(0.0),
		},
	}
	value := cfg.GetBool("key")
	if value {
		t.Error("expected false")
	}
}

func TestConfig_GetBool_FromInt(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": 1,
		},
	}
	value := cfg.GetBool("key")
	if !value {
		t.Error("expected true")
	}
}

func TestConfig_GetBool_FromIntZero(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": 0,
		},
	}
	value := cfg.GetBool("key")
	if value {
		t.Error("expected false")
	}
}

func TestConfig_GetStringSlice_ExistsString(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "a,b,c",
		},
	}
	value := cfg.GetStringSlice("key")
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(value, expected) {
		t.Errorf("expected %v, got %v", expected, value)
	}
}

func TestConfig_GetStringSlice_ExistsStringWithCustomSeparator(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "a|b|c",
		},
	}
	value := cfg.GetStringSlice("key", "|")
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(value, expected) {
		t.Errorf("expected %v, got %v", expected, value)
	}
}

func TestConfig_GetStringSlice_ExistsStringSlice(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": []string{"a", "b", "c"},
		},
	}
	value := cfg.GetStringSlice("key")
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(value, expected) {
		t.Errorf("expected %v, got %v", expected, value)
	}
}

func TestConfig_GetStringSlice_ExistsAnySlice(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": []any{"a", 1, true},
		},
	}
	value := cfg.GetStringSlice("key")
	expected := []string{"a", "1", "true"}
	if !reflect.DeepEqual(value, expected) {
		t.Errorf("expected %v, got %v", expected, value)
	}
}

func TestConfig_GetStringSlice_NotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": "a,b,c",
		},
	}
	value := cfg.GetStringSlice("missing")
	if value != nil {
		t.Errorf("expected nil, got %v", value)
	}
}

func TestConfig_GetStringSlice_Nil(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": nil,
		},
	}
	value := cfg.GetStringSlice("key")
	if value != nil {
		t.Errorf("expected nil, got %v", value)
	}
}

func TestConfig_GetStringSlice_Default(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"key": 123,
		},
	}
	value := cfg.GetStringSlice("key")
	expected := []string{"123"}
	if !reflect.DeepEqual(value, expected) {
		t.Errorf("expected %v, got %v", expected, value)
	}
}

func TestConfig_GetSub_Exists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": map[string]any{
				"child": "value",
			},
		},
	}
	sub, ok := cfg.GetSub("parent")
	if !ok {
		t.Error("expected sub config")
	}
	if sub == nil {
		t.Error("expected sub config, got nil")
	}
}

func TestConfig_GetSub_NotExists(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": map[string]any{
				"child": "value",
			},
		},
	}
	_, ok := cfg.GetSub("missing")
	if ok {
		t.Error("expected false")
	}
}

func TestConfig_GetSub_InvalidType(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": "not a map",
		},
	}
	_, ok := cfg.GetSub("parent")
	if ok {
		t.Error("expected false")
	}
}

func TestConfig_All(t *testing.T) {
	t.Parallel()
	original := map[string]any{
		"key": "value",
	}
	cfg := &Config{
		values: original,
	}
	all := cfg.All()
	if !reflect.DeepEqual(all, original) {
		t.Errorf("expected %v, got %v", original, all)
	}
	if &all == &original {
		t.Error("expected copy, got same map")
	}
}

func TestConfig_find_NestedMapString(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": map[string]any{
				"child": "value",
			},
		},
	}
	value, ok := cfg.find("parent.child")
	if !ok {
		t.Error("expected to find key")
	}
	if value != "value" {
		t.Errorf("expected value, got %v", value)
	}
}

func TestConfig_find_NestedMapAny(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": map[any]any{
				"child": "value",
			},
		},
	}
	value, ok := cfg.find("parent.child")
	if !ok {
		t.Error("expected to find key")
	}
	if value != "value" {
		t.Errorf("expected value, got %v", value)
	}
}

func TestConfig_find_NestedNotFound(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": map[string]any{
				"child": "value",
			},
		},
	}
	_, ok := cfg.find("parent.missing")
	if ok {
		t.Error("expected not found")
	}
}

func TestConfig_find_NestedInvalidType(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": "not a map",
		},
	}
	_, ok := cfg.find("parent.child")
	if ok {
		t.Error("expected not found")
	}
}

func TestConfig_find_NilCurrent(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		values: map[string]any{
			"parent": nil,
		},
	}
	_, ok := cfg.find("parent.child")
	if ok {
		t.Error("expected not found")
	}
}

func TestGetFirst_WithValues(t *testing.T) {
	t.Parallel()
	values := []int{1, 2, 3}
	result := getFirst(values)
	if result != 1 {
		t.Errorf("expected 1, got %v", result)
	}
}

func TestGetFirst_Empty(t *testing.T) {
	t.Parallel()
	var values []int
	result := getFirst(values)
	if result != 0 {
		t.Errorf("expected 0, got %v", result)
	}
}

func TestMergeMaps_Simple(t *testing.T) {
	t.Parallel()
	dst := map[string]any{
		"key1": "value1",
	}
	src := map[string]any{
		"key2": "value2",
	}
	mergeMaps(dst, src)
	if dst["key1"] != "value1" {
		t.Error("expected key1 to remain")
	}
	if dst["key2"] != "value2" {
		t.Error("expected key2 to be added")
	}
}

func TestMergeMaps_Override(t *testing.T) {
	t.Parallel()
	dst := map[string]any{
		"key": "value1",
	}
	src := map[string]any{
		"key": "value2",
	}
	mergeMaps(dst, src)
	if dst["key"] != "value2" {
		t.Error("expected key to be overridden")
	}
}

func TestMergeMaps_Nested(t *testing.T) {
	t.Parallel()
	dst := map[string]any{
		"parent": map[string]any{
			"child1": "value1",
		},
	}
	src := map[string]any{
		"parent": map[string]any{
			"child2": "value2",
		},
	}
	mergeMaps(dst, src)
	parent := dst["parent"].(map[string]any)
	if parent["child1"] != "value1" {
		t.Error("expected child1 to remain")
	}
	if parent["child2"] != "value2" {
		t.Error("expected child2 to be added")
	}
}

func TestMergeMaps_NestedOverride(t *testing.T) {
	t.Parallel()
	dst := map[string]any{
		"parent": map[string]any{
			"child": "value1",
		},
	}
	src := map[string]any{
		"parent": map[string]any{
			"child": "value2",
		},
	}
	mergeMaps(dst, src)
	parent := dst["parent"].(map[string]any)
	if parent["child"] != "value2" {
		t.Error("expected child to be overridden")
	}
}

func TestProcessValue_StringNoTemplate(t *testing.T) {
	t.Parallel()
	result := processValue("plain string")
	if result != "plain string" {
		t.Errorf("expected plain string, got %v", result)
	}
}

func TestProcessValue_StringWithTemplate(t *testing.T) {
	t.Setenv("TEST_KEY", "test_value")
	result := processValue("value is {{env \"TEST_KEY\"}}")
	if result != "value is test_value" {
		t.Errorf("expected rendered template, got %v", result)
	}
}

func TestProcessValue_Map(t *testing.T) {
	t.Parallel()
	input := map[string]any{
		"key": "value",
	}
	result := processValue(input)
	if !reflect.DeepEqual(result, input) {
		t.Errorf("expected same map, got %v", result)
	}
}

func TestProcessValue_Slice(t *testing.T) {
	t.Parallel()
	input := []any{"value"}
	result := processValue(input)
	if !reflect.DeepEqual(result, input) {
		t.Errorf("expected same slice, got %v", result)
	}
}

func TestProcessValue_Other(t *testing.T) {
	t.Parallel()
	input := 42
	result := processValue(input)
	if result != 42 {
		t.Errorf("expected same value, got %v", result)
	}
}

func TestRender_Valid(t *testing.T) {
	t.Setenv("TEST_KEY", "test_value")
	result, err := render("value is {{env \"TEST_KEY\"}}")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "value is test_value" {
		t.Errorf("expected rendered template, got %v", result)
	}
}

func TestRender_Invalid(t *testing.T) {
	t.Parallel()
	_, err := render("{{invalid")
	if err == nil {
		t.Error("expected error")
	}
}

func TestNewFuncMap(t *testing.T) {
	t.Parallel()
	funcMap := newFuncMap()
	if len(funcMap) == 0 {
		t.Error("expected func map to have functions")
	}
}

func TestSetNested_Simple(t *testing.T) {
	t.Parallel()
	m := make(map[string]any)
	setNested(m, "key", "value")
	if m["key"] != "value" {
		t.Errorf("expected value, got %v", m["key"])
	}
}

func TestSetNested_Nested(t *testing.T) {
	t.Parallel()
	m := make(map[string]any)
	setNested(m, "parent.child", "value")
	parent := m["parent"].(map[string]any)
	if parent["child"] != "value" {
		t.Errorf("expected value, got %v", parent["child"])
	}
}

func TestSetNested_ExistingParent(t *testing.T) {
	t.Parallel()
	m := map[string]any{
		"parent": map[string]any{
			"existing": "value1",
		},
	}
	setNested(m, "parent.child", "value2")
	parent := m["parent"].(map[string]any)
	if parent["existing"] != "value1" {
		t.Error("expected existing key to remain")
	}
	if parent["child"] != "value2" {
		t.Errorf("expected new key, got %v", parent["child"])
	}
}

func TestSetNested_InvalidParentType(t *testing.T) {
	t.Parallel()
	m := map[string]any{
		"parent": "not a map",
	}
	setNested(m, "parent.child", "value")
	parent := m["parent"].(map[string]any)
	if parent["child"] != "value" {
		t.Errorf("expected value, got %v", parent["child"])
	}
}
