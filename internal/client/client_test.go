package client

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/yamataka22/pitboard-cli/internal/config"
)

func newTestClient(t *testing.T, handler http.HandlerFunc, token string) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("PITBOARD_API_URL", srv.URL+"/api/v1")
	t.Setenv("PITBOARD_TOKEN", token)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	return New(cfg, "", "test"), srv
}

func TestGetSuccess(t *testing.T) {
	var gotAuth, gotQuery string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":[{"id":1}],"summary":"1 space","context":{}}`))
	}, "pb_test")

	env, err := c.Get("/spaces", url.Values{"page": {"2"}})
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer pb_test" || gotQuery != "page=2" {
		t.Fatalf("unexpected request: %s %s", gotAuth, gotQuery)
	}
	if env.Summary() != "1 space" {
		t.Fatalf("unexpected summary %q", env.Summary())
	}
}

func TestNoToken(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}, "")
	_, err := c.Get("/me", nil)
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindUnauthorized || e.ExitCode() != 3 || e.Hint != "Run: pitboard auth login" {
		t.Fatalf("unexpected error %#v", err)
	}
}

func TestStatusMapping(t *testing.T) {
	cases := []struct {
		status int
		kind   Kind
		code   int
	}{{401, KindUnauthorized, 3}, {403, KindForbidden, 4}, {404, KindNotFound, 2}, {422, KindArgument, 1}, {500, KindServer, 7}}
	for _, tc := range cases {
		c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(`{"ok":false,"error":"x","message":"msg","hint":"do this"}`))
		}, "pb_test")
		_, err := c.Get("/me", nil)
		var e *Error
		if !errors.As(err, &e) {
			t.Fatalf("%d: expected Error, got %v", tc.status, err)
		}
		if e.Kind != tc.kind || e.ExitCode() != tc.code || e.Message != "msg" || e.Hint != "do this" || e.Code != "x" {
			t.Fatalf("%d: unexpected %#v", tc.status, e)
		}
	}
}

func TestNonJSONServerError(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("<html>oops</html>"))
	}, "pb_test")
	_, err := c.Get("/me", nil)
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindServer || e.Message == "" {
		t.Fatalf("unexpected %#v", err)
	}
}

func TestNetworkError(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {}, "pb_test")
	srv.Close()
	_, err := c.Get("/me", nil)
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindNetwork || e.ExitCode() != 6 {
		t.Fatalf("unexpected %#v", err)
	}
}

func TestTLSErrorHint(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("PITBOARD_API_URL", srv.URL+"/api/v1")
	t.Setenv("PITBOARD_TOKEN", "pb_test")
	t.Setenv("PITBOARD_INSECURE", "")
	cfg, _ := config.Load()
	_, err := New(cfg, "", "test").Get("/me", nil)
	var e *Error
	if !errors.As(err, &e) || e.Kind != KindNetwork {
		t.Fatalf("expected network error, got %#v", err)
	}
	if e.Hint == "" || !contains(e.Hint, "PITBOARD_CA_FILE") {
		t.Fatalf("expected CA hint, got %q", e.Hint)
	}

	// PITBOARD_INSECURE=1 なら通る
	t.Setenv("PITBOARD_INSECURE", "1")
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"data":null,"summary":"","context":{}}`))
	})
	if _, err := New(cfg, "", "test").Get("/me", nil); err != nil {
		t.Fatalf("insecure should succeed: %v", err)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0) }
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
