package config

import (
	"os"
	"strings"
	"testing"
)

func TestProcessValue_String_NoTemplate(t *testing.T) {
	t.Parallel()
	out, err := processValue("hello", "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello" {
		t.Errorf("expected hello, got %v", out)
	}
}

func TestProcessValue_String_WithTemplate(t *testing.T) {
	os.Setenv("TMPL_TEST_VAR", "world")
	t.Cleanup(func() { os.Unsetenv("TMPL_TEST_VAR") })
	out, err := processValue(`{{ env "TMPL_TEST_VAR" }}`, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world" {
		t.Errorf("expected world, got %v", out)
	}
}

func TestProcessValue_String_BadTemplate(t *testing.T) {
	t.Parallel()
	_, err := processValue(`{{ end }}`, "k")
	if err == nil {
		t.Fatal("expected error from bad template")
	}
}

func TestProcessValue_Map(t *testing.T) {
	t.Parallel()
	in := map[string]any{"a": "plain", "b": map[string]any{"c": "inner"}}
	out, err := processValue(in, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]any)
	if m["a"] != "plain" {
		t.Errorf("expected plain, got %v", m["a"])
	}
}

func TestProcessValue_Map_Error(t *testing.T) {
	t.Parallel()
	in := map[string]any{"a": `{{ end }}`}
	_, err := processValue(in, "root")
	if err == nil {
		t.Fatal("expected error from map child")
	}
}

func TestProcessValue_Slice(t *testing.T) {
	t.Parallel()
	in := []any{"hello", 42}
	out, err := processValue(in, "arr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sl := out.([]any)
	if sl[0] != "hello" || sl[1] != 42 {
		t.Error("unexpected slice content")
	}
}

func TestProcessValue_Slice_Error(t *testing.T) {
	t.Parallel()
	in := []any{`{{ end }}`}
	_, err := processValue(in, "arr")
	if err == nil {
		t.Fatal("expected error from slice child")
	}
}

func TestProcessValue_OtherType(t *testing.T) {
	t.Parallel()
	out, err := processValue(42, "k")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != 42 {
		t.Errorf("expected 42, got %v", out)
	}
}

func TestRender_ParseError(t *testing.T) {
	t.Parallel()
	_, err := render(`{{ end }}`)
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestRender_Funcs(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, input, want string
	}{
		{"upper", `{{ upper "abc" }}`, "ABC"},
		{"lower", `{{ lower "ABC" }}`, "abc"},
		{"trimSpace", `{{ trimSpace "  hi  " }}`, "hi"},
		{"default_nonempty", `{{ default "fallback" "val" }}`, "val"},
		{"default_empty", `{{ default "fallback" "" }}`, "fallback"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := render(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.TrimSpace(got) != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestRender_DefaultFunc_NonStringDef(t *testing.T) {
	t.Parallel()
	got, err := render(`{{ default 123 "" }}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(got) != "" {
		t.Errorf("expected empty string for non-string default with empty val, got %q", got)
	}
}

func TestRender_DefaultFunc_NonStringVal(t *testing.T) {
	t.Parallel()
	got, err := render(`{{ default "fb" 0 }}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(got) != "fb" {
		t.Errorf("expected fb, got %q", got)
	}
}
