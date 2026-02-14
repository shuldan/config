package config

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

type basicTarget struct {
	Name    string  `cfg:"name"`
	Port    int     `cfg:"port"`
	Rate    float64 `cfg:"rate"`
	Enabled bool    `cfg:"enabled"`
}

func TestUnmarshal_Basic(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{
		"app": map[string]any{
			"name": "svc", "port": 8080, "rate": 1.5, "enabled": true,
		},
	})
	var target basicTarget
	if err := cfg.Unmarshal("app", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Name != "svc" || target.Port != 8080 {
		t.Errorf("unexpected result: %+v", target)
	}
}

func TestUnmarshal_NonPointer(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	err := cfg.Unmarshal("", basicTarget{})
	if err == nil {
		t.Fatal("expected error for non-pointer")
	}
}

func TestUnmarshal_NilPointer(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	err := cfg.Unmarshal("", (*basicTarget)(nil))
	if err == nil {
		t.Fatal("expected error for nil pointer")
	}
}

func TestUnmarshal_NotStruct(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	var s string
	err := cfg.Unmarshal("", &s)
	if err == nil || !strings.Contains(err.Error(), "pointer to struct") {
		t.Fatalf("expected struct error, got %v", err)
	}
}

func TestUnmarshal_KeyNotFound(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	var target basicTarget
	err := cfg.Unmarshal("missing", &target)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestUnmarshal_KeyNotMap(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"k": "string"})
	var target basicTarget
	err := cfg.Unmarshal("k", &target)
	if err == nil {
		t.Fatal("expected error for non-map key")
	}
}

func TestUnmarshal_EmptyKey(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"name": "root"})
	var target basicTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Name != "root" {
		t.Errorf("expected root, got %s", target.Name)
	}
}

type tagTarget struct {
	Ignored  string `cfg:"-"`
	Explicit string `cfg:"ex"`
	Auto     string
}

func TestUnmarshal_Tags(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"ex": "yes", "auto": "val"})
	var target tagTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Explicit != "yes" {
		t.Errorf("expected yes, got %s", target.Explicit)
	}
	if target.Auto != "val" {
		t.Errorf("expected val, got %s", target.Auto)
	}
}

type defaultTarget struct {
	Host    string        `cfg:"host" default:"localhost"`
	Port    int           `cfg:"port" default:"3000"`
	Timeout time.Duration `cfg:"timeout" default:"5s"`
}

func TestUnmarshal_Defaults(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	var target defaultTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Host != "localhost" {
		t.Errorf("expected localhost, got %s", target.Host)
	}
	if target.Port != 3000 {
		t.Errorf("expected 3000, got %d", target.Port)
	}
	if target.Timeout != 5*time.Second {
		t.Errorf("expected 5s, got %v", target.Timeout)
	}
}

type badDefaultTarget struct {
	Port int `cfg:"port" default:"abc"`
}

func TestUnmarshal_BadDefault(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{})
	var target badDefaultTarget
	err := cfg.Unmarshal("", &target)
	if err == nil {
		t.Fatal("expected error for bad default")
	}
}

type nestedTarget struct {
	DB struct {
		Host string `cfg:"host"`
	} `cfg:"db"`
}

func TestUnmarshal_Nested(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{
		"db": map[string]any{"host": "pg"},
	})
	var target nestedTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.DB.Host != "pg" {
		t.Errorf("expected pg, got %s", target.DB.Host)
	}
}

type nestedPtrTarget struct {
	DB *struct {
		Port int `cfg:"port"`
	} `cfg:"db"`
}

func TestUnmarshal_NestedPtr(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{
		"db": map[string]any{"port": 5432},
	})
	var target nestedPtrTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.DB == nil || target.DB.Port != 5432 {
		t.Errorf("unexpected result: %+v", target.DB)
	}
}

func TestUnmarshal_NestedNotMap(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"db": "notmap"})
	var target nestedTarget
	err := cfg.Unmarshal("", &target)
	if err == nil {
		t.Fatal("expected error for nested non-map")
	}
}

type sliceTarget struct {
	Tags  []string `cfg:"tags"`
	Ports []int    `cfg:"ports"`
}

func TestUnmarshal_Slices(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{
		"tags":  []any{"a", "b"},
		"ports": []any{80, 443},
	})
	var target sliceTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(target.Tags, []string{"a", "b"}) {
		t.Errorf("unexpected tags: %v", target.Tags)
	}
}

type mapTarget struct {
	Meta map[string]string `cfg:"meta"`
}

func TestUnmarshal_Map(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{
		"meta": map[string]any{"k": "v"},
	})
	var target mapTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Meta["k"] != "v" {
		t.Errorf("expected v, got %s", target.Meta["k"])
	}
}

func TestConvertToInt_AllTypes(t *testing.T) {
	t.Parallel()
	intType := reflect.TypeOf(int8(0))
	cases := []struct {
		name string
		val  any
		ok   bool
	}{
		{"int", 1, true},
		{"int64", int64(2), true},
		{"uint64", uint64(3), true},
		{"uint64_overflow", uint64(math.MaxUint64), false},
		{"float64", 4.0, true},
		{"bool_true", true, true},
		{"bool_false", false, true},
		{"string_ok", "5", true},
		{"string_bad", "abc", false},
		{"unknown", []int{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := convertToInt(tc.val, intType)
			if tc.ok && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if !tc.ok && err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestConvertToUint_AllTypes(t *testing.T) {
	t.Parallel()
	utype := reflect.TypeOf(uint8(0))
	cases := []struct {
		name string
		val  any
		ok   bool
	}{
		{"uint64", uint64(1), true},
		{"int_pos", 2, true},
		{"int_neg", -1, false},
		{"int64_pos", int64(3), true},
		{"int64_neg", int64(-1), false},
		{"float64_pos", 4.0, true},
		{"float64_neg", -1.0, false},
		{"string_ok", "5", true},
		{"string_bad", "abc", false},
		{"unknown", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := convertToUint(tc.val, utype)
			if tc.ok && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if !tc.ok && err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestConvertToFloat_AllTypes(t *testing.T) {
	t.Parallel()
	ft := reflect.TypeOf(float64(0))
	cases := []struct {
		name string
		val  any
		ok   bool
	}{
		{"float64", 1.1, true},
		{"float32", float32(2.2), true},
		{"int", 3, true},
		{"int64", int64(4), true},
		{"uint64", uint64(5), true},
		{"string_ok", "6.6", true},
		{"string_bad", "abc", false},
		{"unknown", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := convertToFloat(tc.val, ft)
			if tc.ok && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if !tc.ok && err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestConvertToDuration_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		val  any
		ok   bool
	}{
		{"string_ok", "1s", true},
		{"string_bad", "abc", false},
		{"int", 100, true},
		{"int64", int64(200), true},
		{"float64", 300.0, true},
		{"unknown", true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := convertToDuration(tc.val)
			if tc.ok && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if !tc.ok && err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestConvertToTime_Branches(t *testing.T) {
	t.Parallel()
	_, err := convertToTime(42, time.RFC3339)
	if err == nil {
		t.Error("expected error for non-string")
	}
	_, err = convertToTime("bad", time.RFC3339)
	if err == nil {
		t.Error("expected error for bad format")
	}
	_, err = convertToTime("2024-01-01T00:00:00Z", time.RFC3339)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConvertToSlice_StringInput(t *testing.T) {
	t.Parallel()
	st := reflect.TypeOf([]string{})
	v, err := convertToSlice("a,b,c", st, ``)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sl := v.Interface().([]string)
	if len(sl) != 3 || sl[0] != "a" {
		t.Errorf("unexpected result: %v", sl)
	}
}

func TestConvertToSlice_StringSliceInput(t *testing.T) {
	t.Parallel()
	st := reflect.TypeOf([]string{})
	v, err := convertToSlice([]string{"x", "y"}, st, ``)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sl := v.Interface().([]string)
	if !reflect.DeepEqual(sl, []string{"x", "y"}) {
		t.Errorf("unexpected result: %v", sl)
	}
}

func TestConvertToSlice_SingleValue(t *testing.T) {
	t.Parallel()
	st := reflect.TypeOf([]string{})
	v, err := convertToSlice(42, st, ``)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sl := v.Interface().([]string)
	if len(sl) != 1 || sl[0] != "42" {
		t.Errorf("unexpected result: %v", sl)
	}
}

func TestConvertToSlice_SingleVal_Error(t *testing.T) {
	t.Parallel()
	st := reflect.TypeOf([]int{})
	_, err := convertToSlice(complex(1, 2), st, ``)
	if err == nil {
		t.Error("expected error for unsupported single value conversion")
	}
}

func TestConvertToMap_NonMapInput(t *testing.T) {
	t.Parallel()
	mt := reflect.TypeOf(map[string]string{})
	_, err := convertToMap("notmap", mt)
	if err == nil {
		t.Error("expected error for non-map input")
	}
}

func TestConvertToMap_NonStringKey(t *testing.T) {
	t.Parallel()
	mt := reflect.MapOf(reflect.TypeOf(0), reflect.TypeOf(""))
	_, err := convertToMap(map[string]any{"k": "v"}, mt)
	if err == nil {
		t.Error("expected error for non-string key type")
	}
}

func TestConvertToType_Ptr(t *testing.T) {
	t.Parallel()
	pt := reflect.TypeOf((*string)(nil))
	v, err := convertToType("hello", pt, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Elem().String() != "hello" {
		t.Errorf("expected hello, got %v", v.Elem().String())
	}
}

func TestConvertToType_UnsupportedType(t *testing.T) {
	t.Parallel()
	ct := reflect.TypeOf(make(chan int))
	_, err := convertToType("val", ct, "")
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

func TestConvertToType_Bool_BadType(t *testing.T) {
	t.Parallel()
	bt := reflect.TypeOf(true)
	_, err := convertToType([]int{}, bt, "")
	if err == nil {
		t.Error("expected error for bad bool conversion")
	}
}

type timeTarget struct {
	Created time.Time `cfg:"created" layout:"2006-01-02"`
}

func TestUnmarshal_TimeWithLayout(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"created": "2024-06-15"})
	var target timeTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Created.Year() != 2024 {
		t.Errorf("expected 2024, got %d", target.Created.Year())
	}
}

type durationTarget struct {
	Timeout time.Duration `cfg:"timeout"`
}

func TestUnmarshal_Duration(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"timeout": "3s"})
	var target durationTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Timeout != 3*time.Second {
		t.Errorf("expected 3s, got %v", target.Timeout)
	}
}

func TestToBool_Branches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input any
		want  bool
		ok    bool
	}{
		{true, true, true},
		{false, false, true},
		{"true", true, true},
		{"1", true, true},
		{"on", true, true},
		{"yes", true, true},
		{"y", true, true},
		{"false", false, true},
		{"0", false, true},
		{"off", false, true},
		{"no", false, true},
		{"n", false, true},
		{"maybe", false, false},
		{1, true, true},
		{0, false, true},
		{int64(1), true, true},
		{int64(0), false, true},
		{1.0, true, true},
		{0.0, false, true},
		{[]int{}, false, false},
	}
	for _, tc := range cases {
		got, ok := toBool(tc.input)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("toBool(%v): got (%v,%v), want (%v,%v)", tc.input, got, ok, tc.want, tc.ok)
		}
	}
}

func TestParseStringToType_Time_DefaultLayout(t *testing.T) {
	t.Parallel()
	v, err := parseStringToType("2024-01-01T00:00:00Z", timeType, ``)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ti := v.Interface().(time.Time)
	if ti.Year() != 2024 {
		t.Errorf("expected 2024, got %d", ti.Year())
	}
}

func TestParseStringToType_Time_BadValue(t *testing.T) {
	t.Parallel()
	_, err := parseStringToType("bad", timeType, ``)
	if err == nil {
		t.Error("expected error for bad time string")
	}
}

func TestConvertToInt_Overflow(t *testing.T) {
	t.Parallel()
	i8type := reflect.TypeOf(int8(0))
	_, err := convertToInt(int64(200), i8type)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestConvertToUint_Overflow(t *testing.T) {
	t.Parallel()
	u8type := reflect.TypeOf(uint8(0))
	_, err := convertToUint(uint64(300), u8type)
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestConvertToFloat_Overflow32(t *testing.T) {
	t.Parallel()
	f32type := reflect.TypeOf(float32(0))
	_, err := convertToFloat(math.MaxFloat64, f32type)
	if err == nil {
		t.Error("expected overflow error")
	}
}

type nestedPtrErrTarget struct {
	Sub *struct {
		V int `cfg:"v"`
	} `cfg:"sub"`
}

func TestUnmarshal_NestedPtrError(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"sub": map[string]any{"v": "bad"}})
	var target nestedPtrErrTarget
	err := cfg.Unmarshal("", &target)
	if err == nil {
		t.Fatal("expected error for nested ptr unmarshal failure")
	}
}

func TestUnmarshal_NestedStructError(t *testing.T) {
	t.Parallel()
	type inner struct {
		V int `cfg:"v"`
	}
	type outer struct {
		I inner `cfg:"i"`
	}
	cfg := newTestConfig(map[string]any{"i": map[string]any{"v": "bad"}})
	var target outer
	err := cfg.Unmarshal("", &target)
	if err == nil {
		t.Fatal("expected error for nested struct unmarshal failure")
	}
}

func TestConvertToSlice_AnySlice_Error(t *testing.T) {
	t.Parallel()
	st := reflect.TypeOf([]int{})
	_, err := convertToSlice([]any{"bad"}, st, ``)
	if err == nil {
		t.Error("expected error for bad conversion in []any")
	}
}

func TestConvertToSlice_StringSlice_Error(t *testing.T) {
	t.Parallel()
	st := reflect.TypeOf([]int{})
	_, err := convertToSlice([]string{"bad"}, st, ``)
	if err == nil {
		t.Error("expected error")
	}
}

func TestConvertToSlice_String_Error(t *testing.T) {
	t.Parallel()
	st := reflect.TypeOf([]int{})
	_, err := convertToSlice("bad", st, ``)
	if err == nil {
		t.Error("expected error for invalid int parsing from string split")
	}
}

func TestConvertToMap_ValueError(t *testing.T) {
	t.Parallel()
	mt := reflect.TypeOf(map[string]int{})
	_, err := convertToMap(map[string]any{"k": "bad"}, mt)
	if err == nil {
		t.Error("expected error for bad map value conversion")
	}
}

type sepTarget struct {
	Items []string `cfg:"items" separator:"|"`
}

func TestUnmarshal_SliceWithSeparator(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"items": "a|b|c"})
	var target sepTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(target.Items, []string{"a", "b", "c"}) {
		t.Errorf("unexpected items: %v", target.Items)
	}
}

type unexportedTarget struct {
	Name string `cfg:"name"`
	priv int    //nolint
}

func TestUnmarshal_UnexportedField(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"name": "test"})
	var target unexportedTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Name != "test" {
		t.Errorf("expected test, got %s", target.Name)
	}
}

func TestUnmarshal_NilValue(t *testing.T) {
	t.Parallel()
	cfg := newTestConfig(map[string]any{"name": nil})
	var target defaultTarget
	if err := cfg.Unmarshal("", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Host != "localhost" {
		t.Errorf("expected default localhost, got %s", target.Host)
	}
}

func TestConvertToType_TimeDefaultLayout(t *testing.T) {
	t.Parallel()
	v, err := convertToType("2024-01-01T00:00:00Z", timeType, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ti := v.Interface().(time.Time)
	if ti.Year() != 2024 {
		t.Errorf("expected 2024, got %d", ti.Year())
	}
}
