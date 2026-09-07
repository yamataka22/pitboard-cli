package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yamataka22/pitboard-cli/internal/config"
)

type fakeAPI struct {
	routes map[string]func(w http.ResponseWriter, r *http.Request)
}

func newFakeAPI(t *testing.T) (*fakeAPI, *httptest.Server) {
	t.Helper()
	f := &fakeAPI{routes: map[string]func(w http.ResponseWriter, r *http.Request){}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := f.routes[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			h(w, r)
			return
		}
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"ok":false,"error":"not_found","message":"no route","hint":""}`))
	}))
	t.Cleanup(srv.Close)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "config"))
	t.Setenv("PITBOARD_API_URL", srv.URL+"/api/v1")
	t.Setenv("PITBOARD_TOKEN", "")
	t.Setenv("PITBOARD_SPACE_ID", "")
	return f, srv
}

func (f *fakeAPI) ok(path string, data any, summary string) {
	f.routes["/api/v1"+path] = func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data, "summary": summary, "context": map[string]any{}})
	}
}

func (f *fakeAPI) fail(path string, status int, code, message, hint string) {
	f.routes["/api/v1"+path] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": code, "message": message, "hint": hint})
	}
}

func run(t *testing.T, args ...string) (map[string]any, string, int) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Execute("test", args, &out, &errOut, strings.NewReader(""))
	var env map[string]any
	if out.Len() > 0 {
		if err := json.Unmarshal(out.Bytes(), &env); err != nil {
			t.Fatalf("stdout is not JSON: %s", out.String())
		}
	}
	return env, errOut.String(), code
}

func writeConfig(t *testing.T, attrs map[string]any) {
	t.Helper()
	cfg, _ := config.Load()
	if err := cfg.Save(attrs); err != nil {
		t.Fatal(err)
	}
}

func loadConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

var meOneSpace = map[string]any{
	"id": 3, "name": "山田", "email": "y@example.com",
	"token":  map[string]any{"name": "MacBook", "scope": "read"},
	"spaces": []any{map[string]any{"id": 12, "name": "開発チーム"}},
}

var spaceDetailBody = map[string]any{
	"id": 12, "name": "開発チーム", "role": "normal", "default_kind": "issue",
	"me":         map[string]any{"id": 5, "name": "山田"},
	"progresses": []any{map[string]any{"id": 7, "name": "作業中", "done": false}, map[string]any{"id": 9, "name": "完了", "done": true}},
	"projects":   []any{map[string]any{"id": 1, "name": "認証"}},
	"labels":     []any{map[string]any{"id": 2, "name": "バグ"}},
	"members":    []any{map[string]any{"id": 5, "name": "山田", "role": "normal", "assignee": true, "me": true}},
}

func TestAuthLoginSavesTokenAndSingleSpace(t *testing.T) {
	api, _ := newFakeAPI(t)
	api.routes["/api/v1/me"] = func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer pb_new" {
			w.WriteHeader(401)
			_, _ = w.Write([]byte(`{"ok":false,"error":"unauthorized","message":"bad"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": meOneSpace, "summary": "", "context": map[string]any{}})
	}

	env, _, code := run(t, "auth", "login", "pb_new")
	if code != 0 || env["ok"] != true {
		t.Fatalf("code=%d env=%v", code, env)
	}
	cfg := loadConfig(t)
	if cfg.Token() != "pb_new" || cfg.SpaceID() != "12" {
		t.Fatalf("config not saved: %q %q", cfg.Token(), cfg.SpaceID())
	}
	if env["context"].(map[string]any)["space_id"] != "12" {
		t.Fatalf("context: %v", env["context"])
	}
}

func TestAuthLoginMultipleSpacesKeepsDefaultUnset(t *testing.T) {
	api, _ := newFakeAPI(t)
	api.ok("/me", map[string]any{"id": 3, "name": "山田", "email": "y@x", "token": map[string]any{"name": "x", "scope": "read"},
		"spaces": []any{map[string]any{"id": 12, "name": "A"}, map[string]any{"id": 13, "name": "B"}}}, "")
	_, _, code := run(t, "auth", "login", "pb_new")
	if code != 0 || loadConfig(t).SpaceID() != "" {
		t.Fatalf("code=%d space=%q", code, loadConfig(t).SpaceID())
	}
}

func TestAuthLoginInvalidTokenNotSaved(t *testing.T) {
	api, _ := newFakeAPI(t)
	api.fail("/me", 401, "unauthorized", "Token is invalid.", "Run: pitboard auth login")
	env, _, code := run(t, "auth", "login", "pb_bad")
	if code != 3 || env["error"] != "unauthorized" {
		t.Fatalf("code=%d env=%v", code, env)
	}
	if loadConfig(t).Exists() {
		t.Fatal("config should not be written")
	}
}

func TestAuthStatus(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	api.ok("/me", meOneSpace, "")
	env, _, code := run(t, "auth", "status")
	if code != 0 {
		t.Fatalf("code=%d", code)
	}
	ctx := env["context"].(map[string]any)
	if ctx["space_id"] != "12" || ctx["token_source"] != "file" {
		t.Fatalf("context: %v", ctx)
	}
}

func TestAuthStatusWithoutToken(t *testing.T) {
	newFakeAPI(t)
	env, _, code := run(t, "auth", "status")
	if code != 3 || !strings.Contains(env["hint"].(string), "pitboard auth login") {
		t.Fatalf("code=%d env=%v", code, env)
	}
}

func TestAuthLogout(t *testing.T) {
	newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	_, _, code := run(t, "auth", "logout")
	cfg := loadConfig(t)
	if code != 0 || cfg.Token() != "" || cfg.SpaceID() != "" {
		t.Fatalf("code=%d token=%q space=%q", code, cfg.Token(), cfg.SpaceID())
	}
}

func TestSpaceListAndFields(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test"})
	api.ok("/spaces", []any{map[string]any{"id": 12, "name": "開発チーム", "role": "normal", "url": "https://x"}}, "1 space")

	env, _, code := run(t, "space", "list")
	if code != 0 || env["data"].([]any)[0].(map[string]any)["name"] != "開発チーム" {
		t.Fatalf("code=%d env=%v", code, env)
	}
	env, _, _ = run(t, "space", "list", "--fields", "id,name")
	row := env["data"].([]any)[0].(map[string]any)
	if _, ok := row["role"]; ok || len(row) != 2 {
		t.Fatalf("fields not applied: %v", row)
	}
}

func TestSpaceUse(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test"})
	api.ok("/spaces/12", spaceDetailBody, "")
	_, _, code := run(t, "space", "use", "12")
	if code != 0 || loadConfig(t).SpaceID() != "12" {
		t.Fatalf("code=%d space=%q", code, loadConfig(t).SpaceID())
	}
}

func TestSpaceUseNotMember(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test"})
	api.fail("/spaces/99", 404, "not_found", "Space 99 not found or you are not a member.", "Run: pitboard space list")
	env, _, code := run(t, "space", "use", "99")
	if code != 2 || !strings.Contains(env["hint"].(string), "pitboard space list") || loadConfig(t).SpaceID() != "" {
		t.Fatalf("code=%d env=%v", code, env)
	}
}

func TestSpaceShowResolution(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test"})
	api.ok("/spaces/12", spaceDetailBody, "")

	// 所属が1つなら自動
	api.ok("/spaces", []any{map[string]any{"id": 12, "name": "A"}}, "")
	if _, _, code := run(t, "space", "show"); code != 0 {
		t.Fatalf("single space should resolve, code=%d", code)
	}

	// 複数なら space_required
	api.ok("/spaces", []any{map[string]any{"id": 12, "name": "A"}, map[string]any{"id": 13, "name": "B"}}, "")
	env, _, code := run(t, "space", "show")
	if code != 1 || env["error"] != "space_required" {
		t.Fatalf("code=%d env=%v", code, env)
	}

	// 環境変数
	t.Setenv("PITBOARD_SPACE_ID", "12")
	if env, _, code := run(t, "space", "show"); code != 0 || env["data"].(map[string]any)["progresses"].([]any)[1].(map[string]any)["done"] != true {
		t.Fatalf("env var should resolve, code=%d", code)
	}
}

func TestDoctor(t *testing.T) {
	api, _ := newFakeAPI(t)
	env, _, code := run(t, "doctor")
	if code != 1 || env["ok"] != false {
		t.Fatalf("code=%d env=%v", code, env)
	}

	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	api.ok("/me", meOneSpace, "")
	skillPath := filepath.Join(os.Getenv("HOME"), ".claude", "skills", "pitboard", "SKILL.md")
	_ = os.MkdirAll(filepath.Dir(skillPath), 0o755)
	_ = os.WriteFile(skillPath, []byte("x"), 0o644)
	env, _, code = run(t, "doctor")
	if code != 0 || env["ok"] != true {
		t.Fatalf("code=%d env=%v", code, env)
	}
}

func TestSkillInstall(t *testing.T) {
	newFakeAPI(t)
	env, _, code := run(t, "skill", "install")
	path := filepath.Join(os.Getenv("HOME"), ".claude", "skills", "pitboard", "SKILL.md")
	b, err := os.ReadFile(path)
	if code != 0 || err != nil || !strings.Contains(string(b), "# pitboard") {
		t.Fatalf("code=%d err=%v env=%v", code, err, env)
	}
	if _, _, code := run(t, "skill", "install", "--target", "foo"); code != 1 {
		t.Fatalf("unknown target should exit 1, got %d", code)
	}
}

func TestUnknownCommandExitCode(t *testing.T) {
	newFakeAPI(t)
	env, _, code := run(t, "nope")
	if code != 1 || env["ok"] != false {
		t.Fatalf("code=%d env=%v", code, env)
	}
}

func TestAuthLoginSavesAPIURL(t *testing.T) {
	api, srv := newFakeAPI(t)
	t.Setenv("PITBOARD_API_URL", "")
	api.ok("/me", meOneSpace, "")
	_, _, code := run(t, "auth", "login", "pb_new", "--api-url", srv.URL+"/api/v1/")
	if code != 0 {
		t.Fatalf("code=%d", code)
	}
	cfg := loadConfig(t)
	if cfg.APIURL() != srv.URL+"/api/v1" || cfg.Source("api_url") != "file" {
		t.Fatalf("api_url not saved: %s (%s)", cfg.APIURL(), cfg.Source("api_url"))
	}
}

func TestDoctorVersionCheck(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test", "space_id": 12})
	api.routes["/api/v1/me"] = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Pitboard-Api-Version", "1")
		w.Header().Set("X-Pitboard-Min-Cli-Version", "0.2.0")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": meOneSpace, "summary": "", "context": map[string]any{}})
	}
	find := func(env map[string]any) map[string]any {
		for _, c := range env["data"].([]any) {
			if m := c.(map[string]any); m["name"] == "version" {
				return m
			}
		}
		return nil
	}

	var out bytes.Buffer
	Execute("0.1.0", []string{"doctor"}, &out, io.Discard, strings.NewReader(""))
	var env map[string]any
	_ = json.Unmarshal(out.Bytes(), &env)
	if v := find(env); v == nil || v["ok"] != false || !strings.Contains(v["hint"].(string), "Upgrade") {
		t.Fatalf("old cli should be NG: %v", env)
	}

	out.Reset()
	Execute("0.2.0", []string{"doctor"}, &out, io.Discard, strings.NewReader(""))
	_ = json.Unmarshal(out.Bytes(), &env)
	if v := find(env); v == nil || v["ok"] != true {
		t.Fatalf("same version should be OK: %v", env)
	}

	out.Reset()
	Execute("dev-local", []string{"doctor"}, &out, io.Discard, strings.NewReader(""))
	_ = json.Unmarshal(out.Bytes(), &env)
	if v := find(env); v == nil || v["ok"] != true {
		t.Fatalf("dev build should be OK: %v", env)
	}
}

func TestOutdatedCLIExitCode(t *testing.T) {
	api, _ := newFakeAPI(t)
	writeConfig(t, map[string]any{"token": "pb_test"})
	api.fail("/spaces", 426, "cli_outdated", "pitboard-cli 0.1.0 is too old (minimum: 0.2.0).", "Upgrade: get the latest release")
	env, _, code := run(t, "space", "list")
	if code != 5 || env["error"] != "cli_outdated" {
		t.Fatalf("code=%d env=%v", code, env)
	}
}
