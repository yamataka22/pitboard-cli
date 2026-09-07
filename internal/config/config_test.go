package config

import (
	"os"
	"path/filepath"
	"testing"
)

func tempConfig(t *testing.T) *Config {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("PITBOARD_TOKEN", "")
	t.Setenv("PITBOARD_SPACE_ID", "")
	t.Setenv("PITBOARD_API_URL", "")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.Path != filepath.Join(dir, "pitboard", "config.yml") {
		t.Fatalf("unexpected path %s", c.Path)
	}
	return c
}

func TestDefaults(t *testing.T) {
	c := tempConfig(t)
	if c.Token() != "" || c.SpaceID() != "" {
		t.Fatal("expected empty token and space")
	}
	if c.APIURL() != DefaultAPIURL || c.Source("api_url") != "default" {
		t.Fatalf("unexpected api url %s (%s)", c.APIURL(), c.Source("api_url"))
	}
}

func TestSaveAndReload(t *testing.T) {
	c := tempConfig(t)
	if err := c.Save(map[string]any{"token": "pb_abc", "space_id": 12}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(c.Path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600, got %o", info.Mode().Perm())
	}
	r, err := LoadPath(c.Path)
	if err != nil {
		t.Fatal(err)
	}
	if r.Token() != "pb_abc" || r.SpaceID() != "12" || r.Source("token") != "file" {
		t.Fatalf("unexpected reload: %s %s %s", r.Token(), r.SpaceID(), r.Source("token"))
	}
}

func TestEnvOverridesFile(t *testing.T) {
	c := tempConfig(t)
	_ = c.Save(map[string]any{"token": "pb_file", "space_id": 1, "api_url": "https://file/"})
	t.Setenv("PITBOARD_TOKEN", "pb_env")
	t.Setenv("PITBOARD_SPACE_ID", "99")
	if c.Token() != "pb_env" || c.SpaceID() != "99" || c.Source("token") != "env" {
		t.Fatalf("env should win: %s %s", c.Token(), c.SpaceID())
	}
	if c.APIURL() != "https://file" {
		t.Fatalf("trailing slash should be trimmed: %s", c.APIURL())
	}
}

func TestDelete(t *testing.T) {
	c := tempConfig(t)
	_ = c.Save(map[string]any{"token": "pb_abc", "space_id": 12, "api_url": "https://x"})
	if err := c.Delete("token", "space_id"); err != nil {
		t.Fatal(err)
	}
	r, _ := LoadPath(c.Path)
	if r.Token() != "" || r.SpaceID() != "" || r.APIURL() != "https://x" {
		t.Fatalf("unexpected after delete: %q %q %q", r.Token(), r.SpaceID(), r.APIURL())
	}
}
