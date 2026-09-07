package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type captured struct {
	method string
	body   map[string]any
}

func captureWrite(api *fakeAPI, path string, number int) *captured {
	c := &captured{}
	api.routes["/api/v1"+path] = func(w http.ResponseWriter, r *http.Request) {
		c.method = r.Method
		b, _ := io.ReadAll(r.Body)
		c.body = nil // 前のリクエストの内容にマージされないように
		_ = json.Unmarshal(b, &c.body)
		if r.Method == http.MethodPost && strings.HasSuffix(path, "/tasks") {
			w.WriteHeader(201)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": taskRow(number, "x", "inbox"), "summary": "done", "context": map[string]any{}})
	}
	return c
}

func TestTaskWritesRequireYes(t *testing.T) {
	newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	for _, args := range [][]string{
		{"task", "create", "--name", "x"},
		{"task", "assign", "1", "--to", "me"},
		{"task", "move", "1", "--to", "9"},
		{"task", "point", "1", "--to", "d1"},
		{"task", "archive", "1"},
	} {
		env, _, code := run(t, args...)
		if code != 1 || env["error"] != "confirmation_required" {
			t.Fatalf("%v: code=%d env=%v", args, code, env)
		}
	}
}

func TestTaskCreateSendsBody(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	c := captureWrite(api, "/spaces/12/tasks", 43)

	var out strings.Builder
	code := Execute("test", []string{"task", "create", "--name", "週報", "--assignee", "me", "--progress", "7", "--point", "h4",
		"--label", "1", "--label", "2", "--kind", "issue", "--document", "-", "--yes", "--json"}, &out, io.Discard, strings.NewReader("# 本文\n"))
	if code != 0 {
		t.Fatalf("code=%d out=%s", code, out.String())
	}
	if c.method != "POST" || c.body["name"] != "週報" || c.body["assignee"] != "me" || c.body["progress"] != "7" ||
		c.body["point"] != "h4" || c.body["kind"] != "issue" || c.body["document"] != "# 本文\n" {
		t.Fatalf("unexpected body %v", c.body)
	}
	if labels, _ := c.body["labels"].([]any); len(labels) != 2 {
		t.Fatalf("labels %v", c.body["labels"])
	}
	var env map[string]any
	_ = json.Unmarshal([]byte(out.String()), &env)
	if env["data"].(map[string]any)["number"] != float64(43) {
		t.Fatalf("unexpected env %v", env)
	}
}

func TestTaskCreateRequiresName(t *testing.T) {
	newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	if _, _, code := run(t, "task", "create", "--yes"); code != 1 {
		t.Fatalf("expected 1, got %d", code)
	}
}

func TestTaskAssignMovePointArchive(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	assign := captureWrite(api, "/spaces/12/tasks/42/assignee", 42)
	move := captureWrite(api, "/spaces/12/tasks/42/progress", 42)
	point := captureWrite(api, "/spaces/12/tasks/42/point", 42)
	archive := captureWrite(api, "/spaces/12/tasks/42/archive", 42)

	if _, _, code := run(t, "task", "assign", "42", "--to", "none", "--yes"); code != 0 || assign.method != "PUT" || assign.body["assignee"] != "none" {
		t.Fatalf("assign: code=%d %v", code, assign)
	}
	if _, _, code := run(t, "task", "move", "42", "--to", "9", "--yes"); code != 0 || move.method != "PUT" || move.body["progress"] != "9" {
		t.Fatalf("move: code=%d %v", code, move)
	}
	if _, _, code := run(t, "task", "point", "42", "--to", "d2", "--yes"); code != 0 || point.method != "PUT" || point.body["point"] != "d2" {
		t.Fatalf("point: code=%d %v", code, point)
	}
	if _, _, code := run(t, "task", "archive", "42", "--yes"); code != 0 || archive.method != "POST" {
		t.Fatalf("archive: code=%d %v", code, archive)
	}
	if _, _, code := run(t, "task", "archive", "42", "--undo", "--yes"); code != 0 || archive.method != "DELETE" {
		t.Fatalf("unarchive: code=%d %v", code, archive)
	}
}

func TestTaskWriteForbiddenExitCode(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	api.fail("/spaces/12/tasks", 403, "forbidden", "This token is read-only.", "Issue a read/write token from your profile page.")
	env, _, code := run(t, "task", "create", "--name", "x", "--yes")
	if code != 4 || env["error"] != "forbidden" {
		t.Fatalf("code=%d env=%v", code, env)
	}
}
