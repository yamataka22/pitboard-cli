package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTableAlignsWideChars(t *testing.T) {
	var buf bytes.Buffer
	Table(&buf, []string{"ID", "Name"}, [][]string{{"1", "山田"}, {"22", "ab"}})
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if lines[0] != "ID  Name" || lines[2] != "1   山田" || lines[3] != "22  ab" {
		t.Fatalf("unexpected table:\n%s", buf.String())
	}
}

func TestWithFields(t *testing.T) {
	env := map[string]any{"ok": true, "data": []any{map[string]any{"a": 1, "b": 2}}}
	got := WithFields(env, "a")["data"].([]any)[0].(map[string]any)
	if _, ok := got["b"]; ok || got["a"] != 1 {
		t.Fatalf("unexpected %v", got)
	}
	if WithFields(env, " ")["data"].([]any)[0].(map[string]any)["b"] != 2 {
		t.Fatal("empty fields should keep everything")
	}
}

func TestIsJSONForNonFile(t *testing.T) {
	if !IsJSON(Options{}, &bytes.Buffer{}) {
		t.Fatal("non-terminal writer should be JSON")
	}
}

func TestJSONKeyOrder(t *testing.T) {
	var buf bytes.Buffer
	env := map[string]any{"context": map[string]any{}, "summary": "s", "data": []any{1}, "ok": true}
	if err := Emit(&buf, env, Options{JSON: true}, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Index(out, `"ok"`) > strings.Index(out, `"data"`) || strings.Index(out, `"data"`) > strings.Index(out, `"summary"`) || strings.Index(out, `"summary"`) > strings.Index(out, `"context"`) {
		t.Fatalf("unexpected order:\n%s", out)
	}
}
