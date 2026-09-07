// Package config は CLI の設定。優先順位は 環境変数 > 設定ファイル > 既定値。
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultAPIURL = "https://app.pitboard.dev/api/v1"

var envKeys = map[string]string{
	"token":    "PITBOARD_TOKEN",
	"api_url":  "PITBOARD_API_URL",
	"space_id": "PITBOARD_SPACE_ID",
	"ca_file":  "PITBOARD_CA_FILE",
}

type Config struct {
	Path string
	file map[string]any
}

// DefaultPath は $XDG_CONFIG_HOME/pitboard/config.yml。無ければ ~/.config/pitboard/config.yml
// （macOS も CLI の慣習に合わせて ~/.config。Windows だけ %AppData%）
func DefaultPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		if runtime.GOOS == "windows" {
			if dir, err := os.UserConfigDir(); err == nil {
				base = dir
			}
		}
	}
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "pitboard", "config.yml")
}

func Load() (*Config, error) {
	return LoadPath(DefaultPath())
}

func LoadPath(path string) (*Config, error) {
	c := &Config{Path: path, file: map[string]any{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &c.file); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if c.file == nil {
		c.file = map[string]any{}
	}
	return c, nil
}

func (c *Config) Exists() bool {
	_, err := os.Stat(c.Path)
	return err == nil
}

func (c *Config) Token() string   { return c.value("token") }
func (c *Config) SpaceID() string { return c.value("space_id") }
func (c *Config) CAFile() string  { return c.value("ca_file") }

func (c *Config) APIURL() string {
	v := c.value("api_url")
	if v == "" {
		v = DefaultAPIURL
	}
	return strings.TrimRight(v, "/")
}

// Insecure は開発用。証明書検証を無効にする（PITBOARD_INSECURE=1）
func (c *Config) Insecure() bool {
	return os.Getenv("PITBOARD_INSECURE") == "1"
}

// Source は値の出所（"env" / "file" / "default" / ""）。auth status と doctor で使う
func (c *Config) Source(key string) string {
	if os.Getenv(envKeys[key]) != "" {
		return "env"
	}
	if s := stringify(c.file[key]); s != "" {
		return "file"
	}
	if key == "api_url" {
		return "default"
	}
	return ""
}

// Save は既存の値にマージして書き込む。ディレクトリは 0700、ファイルは 0600
func (c *Config) Save(attrs map[string]any) error {
	for k, v := range attrs {
		if v == nil {
			continue
		}
		c.file[k] = v
	}
	return c.write()
}

func (c *Config) Delete(keys ...string) error {
	for _, k := range keys {
		delete(c.file, k)
	}
	if !c.Exists() {
		return nil
	}
	return c.write()
}

func (c *Config) write() error {
	if err := os.MkdirAll(filepath.Dir(c.Path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c.file)
	if err != nil {
		return err
	}
	if err := os.WriteFile(c.Path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(c.Path, 0o600)
}

func (c *Config) value(key string) string {
	if v := os.Getenv(envKeys[key]); v != "" {
		return v
	}
	return stringify(c.file[key])
}

func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	default:
		return fmt.Sprint(x)
	}
}
