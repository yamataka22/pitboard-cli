package cli

import (
	"io"
	"strings"
	"testing"
)

func TestCommentAdd(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	c := captureWrite(api, "/spaces/12/tasks/42/comments", 42)

	if _, _, code := run(t, "comment", "add", "42", "--body", "確認しました"); code != 1 {
		t.Fatalf("without --yes should exit 1, got %d", code)
	}
	_, _, code := run(t, "comment", "add", "42", "--body", "確認しました", "--reply-to", "7", "--mention", "5", "--mention", "6", "--yes")
	if code != 0 || c.method != "POST" || c.body["content"] != "確認しました" || c.body["parent"] != "7" {
		t.Fatalf("code=%d body=%v", code, c.body)
	}
	if m, _ := c.body["mentions"].([]any); len(m) != 2 {
		t.Fatalf("mentions %v", c.body["mentions"])
	}

	var out strings.Builder
	code = Execute("test", []string{"comment", "add", "42", "--body", "-", "--yes", "--json"}, &out, io.Discard, strings.NewReader("stdin から\n"))
	if code != 0 || c.body["content"] != "stdin から\n" {
		t.Fatalf("stdin: code=%d body=%v", code, c.body)
	}
	if _, ok := c.body["parent"]; ok {
		t.Fatalf("parent should be omitted: %v", c.body)
	}
}
