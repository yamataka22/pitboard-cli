package cli

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func taskRow(number int, name, state string) map[string]any {
	return map[string]any{
		"number": number, "id": number * 10, "name": name, "state": state,
		"progress": map[string]any{"id": 7, "name": "作業中", "done": false},
		"kind":     "confirmed", "point": "d2", "point_days": 2.0,
		"assignee": map[string]any{"id": 5, "name": "山田"}, "owner": map[string]any{"id": 5, "name": "山田"},
		"project": nil, "labels": []any{}, "archived": false, "progress_changed_at": "2026-09-01T10:00:00+09:00",
		"todo_count": 1, "todo_done_count": 0, "created_at": "2026-09-01T09:00:00+09:00", "updated_at": "2026-09-01T09:00:00+09:00",
		"url": "https://x/12/tasks/420",
	}
}

func TestTaskListSendsFiltersAsAPIParams(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	var gotQuery string
	api.routes["/api/v1/spaces/12/tasks"] = func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": []any{taskRow(42, "ログイン", "in_progress")}, "summary": "1 task",
			"context": map[string]any{"pagination": map[string]any{"page": 1, "per_page": 50, "has_next": false, "total": 1}}})
	}

	env, _, code := run(t, "task", "list", "--mine", "--state", "done,in_progress", "--label", "1", "--label", "2",
		"--progress-changed-since", "2026-09-01", "--sort", "updated", "--order", "asc", "-q", "ログイン")
	if code != 0 {
		t.Fatalf("code=%d env=%v", code, env)
	}
	for _, want := range []string{"assignee=me", "state=done%2Cin_progress", "label%5B%5D=1", "label%5B%5D=2", "progress_changed_since=2026-09-01", "sort=updated", "order=asc", "q=%E3%83%AD%E3%82%B0%E3%82%A4%E3%83%B3"} {
		if !strings.Contains(gotQuery, want) {
			t.Fatalf("query %q should contain %q", gotQuery, want)
		}
	}
	if env["data"].([]any)[0].(map[string]any)["number"] != float64(42) {
		t.Fatalf("unexpected data %v", env["data"])
	}
}

func TestTaskListAllPages(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	api.routes["/api/v1/spaces/12/tasks"] = func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		data := []any{taskRow(2, "b", "inbox")}
		hasNext := true
		if page == "2" {
			data = []any{taskRow(1, "a", "backlog")}
			hasNext = false
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data, "summary": "2 tasks",
			"context": map[string]any{"pagination": map[string]any{"page": map[string]int{"": 1, "2": 2}[page], "per_page": 1, "has_next": hasNext, "total": 2}}})
	}

	env, _, code := run(t, "task", "list", "--all")
	if code != 0 || len(env["data"].([]any)) != 2 {
		t.Fatalf("code=%d data=%v", code, env["data"])
	}
	p := env["context"].(map[string]any)["pagination"].(map[string]any)
	if p["has_next"] != false || p["total"] != float64(2) {
		t.Fatalf("pagination %v", p)
	}
}

func TestTaskListUsesExplicitSpace(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	api.ok("/spaces/34/tasks", []any{}, "0 tasks")
	if _, _, code := run(t, "task", "list", "--space", "34"); code != 0 {
		t.Fatalf("code=%d", code)
	}
}

func TestTaskShow(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	row := taskRow(42, "ログイン", "in_progress")
	row["document"] = "# 本文"
	row["todos"] = []any{map[string]any{"id": 1, "name": "a", "done": true}}
	row["comments"] = []any{map[string]any{"id": 1, "author": map[string]any{"id": 5, "name": "山田"}, "content": "c", "created_at": "2026-09-01T09:00:00+09:00", "replies": []any{}}}
	row["related_tasks"] = []any{}
	api.ok("/spaces/12/tasks/42", row, "#42 ログイン (in_progress)")

	env, _, code := run(t, "task", "show", "42")
	if code != 0 || env["data"].(map[string]any)["document"] != "# 本文" {
		t.Fatalf("code=%d env=%v", code, env)
	}
	api.fail("/spaces/12/tasks/999", 404, "not_found", "Not found.", "")
	if _, _, code := run(t, "task", "show", "999"); code != 2 {
		t.Fatalf("expected 2, got %d", code)
	}
}
