package config

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"
)

func newTestConfig(m map[string]any) *Config {
	return &Config{values: m}
}

func TestNew_Success(t *testing.T) {
	t.Parallel()
	cfg, err := New(WithLoader(&staticLoader{data: map[string]any{"a": "b"}}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.GetString("a") != "b" {
		t.Errorf("expected %q, got %q", "b", cfg.GetString("a"))
	}
}

func TestNew_LoaderError(t *testing.T) {
	t.Parallel()
	_, err := New(WithLoader(&failLoader{err: errors.New("boom")}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNew_TemplateRenderError(t *testing.T) {
	t.Parallel()
	bad := map[string]any{"k": "{{ end }}"}
	_, err := New(WithLoader(&staticLoader{data: bad}))
	if err == nil {
		t.Fatal("expected template error, got nil")
	}
}

func TestNew_NoLoaders(t *testing.T) {
	t.Parallel()
	cfg, err := New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.All()) != 0 {
		t.Errorf("expected empty config, got %v", cfg.All())
	}
}

func TestFromMap(t *testing.T) {
	t.Parallel()
	orig := map[string]any{"x": "y"}
	cfg := FromMap(orig)
	orig["x"] = "changed"
	if cfg.GetString("x") != "y" {
		t.Errorf("expected %q, got %q", "y", cfg.GetString("x"))
	}
}

func TestWithOverrides(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"a": "1", "b": "2"})
	ov := cfg.WithOverrides(map[string]any{"a": "10", "c.d": "3"})
	if ov.GetString("a") != "10" {
		t.Errorf("expected %q, got %q", "10", ov.GetString("a"))
	}
	if ov.GetString("b") != "2" {
		t.Errorf("expected %q, got %q", "2", ov.GetString("b"))
	}
	if ov.GetString("c.d") != "3" {
		t.Errorf("expected %q, got %q", "3", ov.GetString("c.d"))
	}
}

func TestHas(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"a": 1})
	if !cfg.Has("a") {
		t.Error("expected Has(a) = true")
	}
	if cfg.Has("z") {
		t.Error("expected Has(z) = false")
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"a": 42})
	if cfg.Get("a") != 42 {
		t.Errorf("expected 42, got %v", cfg.Get("a"))
	}
	if cfg.Get("missing") != nil {
		t.Errorf("expected nil, got %v", cfg.Get("missing"))
	}
}

func TestAll(t *testing.T) {
	t.Parallel()
	orig := map[string]any{"k": "v"}
	cfg := newTestConfig(orig)
	all := cfg.All()
	all["k"] = "mutated"
	if cfg.GetString("k") != "v" {
		t.Error("All() did not return a deep copy")
	}
}

func TestGetString_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		values map[string]any
		key    string
		def    []string
		want   string
	}{
		{"found_string", map[string]any{"k": "val"}, "k", nil, "val"},
		{"found_nil", map[string]any{"k": nil}, "k", nil, ""},
		{"found_int", map[string]any{"k": 42}, "k", nil, "42"},
		{"not_found_no_default", map[string]any{}, "k", nil, ""},
		{"not_found_with_default", map[string]any{}, "k", []string{"def"}, "def"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := newTestConfig(tc.values)
			got := cfg.GetString(tc.key, tc.def...)
			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestGetInt_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		values map[string]any
		key    string
		def    []int
		want   int
	}{
		{"found_int", map[string]any{"k": 5}, "k", nil, 5},
		{"found_string", map[string]any{"k": "7"}, "k", nil, 7},
		{"not_found_default", map[string]any{}, "k", []int{99}, 99},
		{"invalid_type", map[string]any{"k": []int{1}}, "k", []int{-1}, -1},
		{"found_bool_true", map[string]any{"k": true}, "k", nil, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := newTestConfig(tc.values)
			got := cfg.GetInt(tc.key, tc.def...)
			if got != tc.want {
				t.Errorf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestGetInt64_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		values map[string]any
		key    string
		def    []int64
		want   int64
	}{
		{"int64", map[string]any{"k": int64(10)}, "k", nil, 10},
		{"int", map[string]any{"k": 20}, "k", nil, 20},
		{"uint64_ok", map[string]any{"k": uint64(30)}, "k", nil, 30},
		{"uint64_overflow", map[string]any{"k": uint64(math.MaxUint64)}, "k", []int64{-1}, -1},
		{"float64_ok", map[string]any{"k": 40.0}, "k", nil, 40},
		{"float64_underflow", map[string]any{"k": -1e20}, "k", []int64{-1}, -1},
		{"float64_overflow", map[string]any{"k": 1e20}, "k", []int64{-1}, -1},
		{"bool_true", map[string]any{"k": true}, "k", nil, 1},
		{"bool_false", map[string]any{"k": false}, "k", nil, 0},
		{"string_ok", map[string]any{"k": "50"}, "k", nil, 50},
		{"string_bad", map[string]any{"k": "abc"}, "k", []int64{-1}, -1},
		{"unknown_type", map[string]any{"k": []int{}}, "k", []int64{-1}, -1},
		{"not_found", map[string]any{}, "k", []int64{99}, 99},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := newTestConfig(tc.values)
			got := cfg.GetInt64(tc.key, tc.def...)
			if got != tc.want {
				t.Errorf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestGetUint64_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		values map[string]any
		key    string
		def    []uint64
		want   uint64
	}{
		{"uint64", map[string]any{"k": uint64(1)}, "k", nil, 1},
		{"int_pos", map[string]any{"k": 5}, "k", nil, 5},
		{"int_neg", map[string]any{"k": -1}, "k", []uint64{99}, 99},
		{"int64_pos", map[string]any{"k": int64(7)}, "k", nil, 7},
		{"int64_neg", map[string]any{"k": int64(-1)}, "k", []uint64{99}, 99},
		{"float64_pos", map[string]any{"k": 3.0}, "k", nil, 3},
		{"float64_neg", map[string]any{"k": -1.0}, "k", []uint64{99}, 99},
		{"float64_huge", map[string]any{"k": 1e30}, "k", []uint64{99}, 99},
		{"string_ok", map[string]any{"k": "42"}, "k", nil, 42},
		{"string_bad", map[string]any{"k": "abc"}, "k", []uint64{99}, 99},
		{"unknown", map[string]any{"k": true}, "k", []uint64{99}, 99},
		{"not_found", map[string]any{}, "k", []uint64{7}, 7},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := newTestConfig(tc.values)
			got := cfg.GetUint64(tc.key, tc.def...)
			if got != tc.want {
				t.Errorf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestGetFloat64_Branches(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"a": 1.5, "b": "bad"})
	if got := cfg.GetFloat64("a"); got != 1.5 {
		t.Errorf("expected 1.5, got %v", got)
	}
	if got := cfg.GetFloat64("b", 9.9); got != 9.9 {
		t.Errorf("expected 9.9, got %v", got)
	}
	if got := cfg.GetFloat64("missing", 3.3); got != 3.3 {
		t.Errorf("expected 3.3, got %v", got)
	}
}

func TestGetBool_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		values map[string]any
		key    string
		def    []bool
		want   bool
	}{
		{"true", map[string]any{"k": true}, "k", nil, true},
		{"false_str", map[string]any{"k": "false"}, "k", nil, false},
		{"not_found", map[string]any{}, "k", []bool{true}, true},
		{"bad_type", map[string]any{"k": []int{}}, "k", []bool{true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := newTestConfig(tc.values)
			got := cfg.GetBool(tc.key, tc.def...)
			if got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGetDuration_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		values map[string]any
		key    string
		def    []time.Duration
		want   time.Duration
	}{
		{"duration", map[string]any{"k": 5 * time.Second}, "k", nil, 5 * time.Second},
		{"string", map[string]any{"k": "2s"}, "k", nil, 2 * time.Second},
		{"int", map[string]any{"k": 100}, "k", nil, 100 * time.Millisecond},
		{"int64", map[string]any{"k": int64(200)}, "k", nil, 200 * time.Millisecond},
		{"float64", map[string]any{"k": 300.0}, "k", nil, 300 * time.Millisecond},
		{"unknown_type", map[string]any{"k": true}, "k", []time.Duration{time.Hour}, time.Hour},
		{"not_found", map[string]any{}, "k", []time.Duration{time.Minute}, time.Minute},
		{"bad_string", map[string]any{"k": "xxx"}, "k", []time.Duration{time.Second}, time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := newTestConfig(tc.values)
			got := cfg.GetDuration(tc.key, tc.def...)
			if got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGetTime_Branches(t *testing.T) {
	t.Parallel()
	layout := "2006-01-02"
	ref := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	def := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := newTestConfig(map[string]any{
		"good": "2024-01-15", "bad": "nope", "num": 42,
	})
	if got := cfg.GetTime("good", layout); !got.Equal(ref) {
		t.Errorf("expected %v, got %v", ref, got)
	}
	if got := cfg.GetTime("bad", layout, def); !got.Equal(def) {
		t.Errorf("expected default, got %v", got)
	}
	if got := cfg.GetTime("num", layout, def); !got.Equal(def) {
		t.Errorf("expected default for non-string, got %v", got)
	}
	if got := cfg.GetTime("miss", layout, def); !got.Equal(def) {
		t.Errorf("expected default for missing, got %v", got)
	}
}

func TestGetStringSlice_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		vals map[string]any
		key  string
		sep  []string
		want []string
	}{
		{"nil_val", map[string]any{"k": nil}, "k", nil, nil},
		{"not_found", map[string]any{}, "k", nil, nil},
		{"string_slice", map[string]any{"k": []string{"a", "b"}}, "k", nil, []string{"a", "b"}},
		{"any_slice", map[string]any{"k": []any{1, "x"}}, "k", nil, []string{"1", "x"}},
		{"csv_string", map[string]any{"k": "a, b, c"}, "k", nil, []string{"a", "b", "c"}},
		{"custom_sep", map[string]any{"k": "a|b"}, "k", []string{"|"}, []string{"a", "b"}},
		{"default_type", map[string]any{"k": 42}, "k", nil, []string{"42"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := newTestConfig(tc.vals)
			got := cfg.GetStringSlice(tc.key, tc.sep...)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGetIntSlice_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		vals map[string]any
		key  string
		want []int
	}{
		{"nil_val", map[string]any{"k": nil}, "k", nil},
		{"not_found", map[string]any{}, "k", nil},
		{"int_slice", map[string]any{"k": []int{1, 2}}, "k", []int{1, 2}},
		{"any_slice", map[string]any{"k": []any{3, "bad", 4}}, "k", []int{3, 4}},
		{"float_slice", map[string]any{"k": []float64{5.1, 6.9}}, "k", []int{5, 6}},
		{"other", map[string]any{"k": "no"}, "k", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := newTestConfig(tc.vals)
			got := cfg.GetIntSlice(tc.key)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGetFloat64Slice_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		vals map[string]any
		key  string
		want []float64
	}{
		{"nil_val", map[string]any{"k": nil}, "k", nil},
		{"not_found", map[string]any{}, "k", nil},
		{"float_slice", map[string]any{"k": []float64{1.1}}, "k", []float64{1.1}},
		{"any_slice", map[string]any{"k": []any{2.2, "bad"}}, "k", []float64{2.2}},
		{"int_slice", map[string]any{"k": []int{3, 4}}, "k", []float64{3.0, 4.0}},
		{"other", map[string]any{"k": "no"}, "k", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := newTestConfig(tc.vals)
			got := cfg.GetFloat64Slice(tc.key)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGetMap_Branches(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{
		"m": map[string]any{"a": 1}, "s": "nope",
	})
	m, ok := cfg.GetMap("m")
	if !ok || m["a"] != 1 {
		t.Errorf("expected map with a=1, got %v, %v", m, ok)
	}
	_, ok = cfg.GetMap("s")
	if ok {
		t.Error("expected false for non-map")
	}
	_, ok = cfg.GetMap("missing")
	if ok {
		t.Error("expected false for missing")
	}
}

func TestGetSub_Branches(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{
		"db": map[string]any{"host": "localhost"}, "x": 42,
	})
	sub, ok := cfg.GetSub("db")
	if !ok {
		t.Fatal("expected sub config")
	}
	if sub.GetString("host") != "localhost" {
		t.Error("expected localhost")
	}
	_, ok = cfg.GetSub("x")
	if ok {
		t.Error("expected false for non-map")
	}
	_, ok = cfg.GetSub("missing")
	if ok {
		t.Error("expected false for missing")
	}
}

func TestFind_MapAnyAny(t *testing.T) {
	t.Parallel()
	inner := map[any]any{"b": "val"}
	cfg := newTestConfig(map[string]any{"a": inner})
	v, ok := cfg.find("a.b")
	if !ok || v != "val" {
		t.Errorf("expected val, got %v ok=%v", v, ok)
	}
	_, ok = cfg.find("a.missing")
	if ok {
		t.Error("expected not found in map[any]any")
	}
}

func TestFind_NilCurrent(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"a": nil})
	_, ok := cfg.find("a.b")
	if ok {
		t.Error("expected false with nil node")
	}
}

func TestFind_NonMapType(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"a": 42})
	_, ok := cfg.find("a.b")
	if ok {
		t.Error("expected false when current is not a map")
	}
}

type staticLoader struct {
	data map[string]any
}

func (s *staticLoader) Load() (map[string]any, error) { return s.data, nil }

type failLoader struct {
	err error
}

func (f *failLoader) Load() (map[string]any, error) { return nil, f.err }

func TestNew_WithLogger(t *testing.T) {
	t.Parallel()
	l := &capLogger{}
	_, err := New(WithLogger(l), WithLoader(&staticLoader{data: map[string]any{"a": 1}}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(l.msgs) == 0 {
		t.Error("expected log messages")
	}
}

type capLogger struct {
	msgs []string
}

func (c *capLogger) Debug(msg string, args ...any) {
	c.msgs = append(c.msgs, msg)
}
