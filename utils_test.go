package config

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDeepCopyMap(t *testing.T) {
	t.Parallel()
	src := map[string]any{
		"a": map[string]any{"b": 1},
		"c": []any{2, "x"},
		"d": []string{"s1"},
		"e": 42,
	}
	dst := deepCopyMap(src)
	src["a"].(map[string]any)["b"] = 999
	if dst["a"].(map[string]any)["b"] == 999 {
		t.Error("deep copy failed for nested map")
	}
}

func TestMergeMaps(t *testing.T) {
	t.Parallel()
	dst := map[string]any{
		"a": map[string]any{"b": 1, "c": 2},
		"d": "old",
	}
	src := map[string]any{
		"a": map[string]any{"b": 10, "e": 3},
		"d": "new",
		"f": 4,
	}
	mergeMaps(dst, src)
	a := dst["a"].(map[string]any)
	if a["b"] != 10 || a["c"] != 2 || a["e"] != 3 {
		t.Errorf("unexpected merge result: %v", a)
	}
	if dst["d"] != "new" || dst["f"] != 4 {
		t.Errorf("unexpected dst: %v", dst)
	}
}

func TestMergeMaps_OverwriteNonMap(t *testing.T) {
	t.Parallel()
	dst := map[string]any{"a": "string"}
	src := map[string]any{"a": map[string]any{"b": 1}}
	mergeMaps(dst, src)
	if _, ok := dst["a"].(map[string]any); !ok {
		t.Error("expected map to overwrite string")
	}
}

func TestNormalizeMap(t *testing.T) {
	t.Parallel()
	m := map[string]any{
		"a": map[any]any{1: "one"},
		"b": map[string]any{"c": []any{map[any]any{"d": "e"}}},
		"f": 42,
	}
	out := normalizeMap(m)
	a := out["a"].(map[string]any)
	if a["1"] != "one" {
		t.Errorf("expected one, got %v", a["1"])
	}
}

func TestSetNested_Deep(t *testing.T) {
	t.Parallel()
	m := make(map[string]any)
	setNested(m, "a.b.c", "val")
	a := m["a"].(map[string]any)
	b := a["b"].(map[string]any)
	if b["c"] != "val" {
		t.Errorf("expected val, got %v", b["c"])
	}
}

func TestSetNested_OverwriteNonMap(t *testing.T) {
	t.Parallel()
	m := map[string]any{"a": "string"}
	setNested(m, "a.b", "val")
	a := m["a"].(map[string]any)
	if a["b"] != "val" {
		t.Errorf("expected val, got %v", a["b"])
	}
}

func TestExpandDotKeys(t *testing.T) {
	t.Parallel()
	flat := map[string]any{"a.b": 1, "c": 2}
	out := expandDotKeys(flat)
	a := out["a"].(map[string]any)
	if a["b"] != 1 || out["c"] != 2 {
		t.Errorf("unexpected result: %v", out)
	}
}

func TestResolveSecurePath_Valid(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "file.txt")
	os.WriteFile(p, []byte("x"), 0644)
	abs, err := resolveSecurePath(p, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if abs != p {
		t.Errorf("expected %s, got %s", p, abs)
	}
}

func TestResolveSecurePath_Traversal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, err := resolveSecurePath("../../../etc/passwd", dir)
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestResolveSecurePath_DefaultBasePath(t *testing.T) {
	wd, _ := os.Getwd()
	p := filepath.Join(wd, "config.go")
	abs, err := resolveSecurePath(p, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if abs == "" {
		t.Error("expected non-empty abs path")
	}
}

func TestFileExists(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.txt")
	os.WriteFile(p, []byte("x"), 0644)
	if !fileExists(p) {
		t.Error("expected file to exist")
	}
	if fileExists(filepath.Join(dir, "nope.txt")) {
		t.Error("expected file to not exist")
	}
	if fileExists(dir) {
		t.Error("expected false for directory")
	}
}

func TestGetFirst(t *testing.T) {
	t.Parallel()
	if got := getFirst([]int{5, 6}); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
	if got := getFirst[int](nil); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestAutoParseString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  any
	}{
		{"true", true},
		{"false", false},
		{"42", 42},
		{"3.14", 3.14},
		{"hello", "hello"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := autoParseString(tc.input)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %v (%T), got %v (%T)", tc.want, tc.want, got, got)
			}
		})
	}
}

func TestAutoParseString_MaxInt64(t *testing.T) {
	t.Parallel()
	s := "9223372036854775807"
	got := autoParseString(s)
	switch got.(type) {
	case int, int64:
	default:
		t.Errorf("expected int or int64, got %T", got)
	}
}

func TestAutoParseString_OverflowInt64(t *testing.T) {
	t.Parallel()
	s := "99999999999999999999"
	got := autoParseString(s)
	if _, ok := got.(float64); !ok {
		t.Errorf("expected float64 for overflow, got %T", got)
	}
}

func TestToInt_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		val  any
		want int
		ok   bool
	}{
		{"int", 5, 5, true},
		{"int64", int64(6), 6, true},
		{"uint64", uint64(7), 7, true},
		{"uint64_overflow", uint64(math.MaxUint64), 0, false},
		{"float64", 8.0, 8, true},
		{"bool_true", true, 1, true},
		{"bool_false", false, 0, true},
		{"string_ok", "9", 9, true},
		{"string_bad", "abc", 0, false},
		{"unknown", []int{}, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := toInt(tc.val)
			if ok != tc.ok || (ok && got != tc.want) {
				t.Errorf("toInt(%v) = (%d,%v), want (%d,%v)", tc.val, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestToFloat64_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		val  any
		want float64
		ok   bool
	}{
		{"float64", 1.5, 1.5, true},
		{"int", 2, 2.0, true},
		{"int64", int64(3), 3.0, true},
		{"uint64", uint64(4), 4.0, true},
		{"string_ok", "5.5", 5.5, true},
		{"string_bad", "abc", 0, false},
		{"unknown", true, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := toFloat64(tc.val)
			if ok != tc.ok || (ok && got != tc.want) {
				t.Errorf("toFloat64(%v) = (%v,%v), want (%v,%v)", tc.val, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestDeepCopyValue_AllTypes(t *testing.T) {
	t.Parallel()
	m := deepCopyValue(map[string]any{"k": "v"})
	if m.(map[string]any)["k"] != "v" {
		t.Error("map copy failed")
	}
	sl := deepCopyValue([]any{1, "two"})
	if len(sl.([]any)) != 2 {
		t.Error("slice copy failed")
	}
	ss := deepCopyValue([]string{"a"})
	if len(ss.([]string)) != 1 {
		t.Error("string slice copy failed")
	}
	if deepCopyValue(42) != 42 {
		t.Error("scalar copy failed")
	}
}
