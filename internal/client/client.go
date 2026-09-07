// Package client は API クライアント。HTTP はここに閉じる（コマンドはこれを呼ぶだけ）。
package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/yamataka22/pitboard-cli/internal/config"
)

// Kind は終了コードに対応するエラー種別
type Kind string

const (
	KindArgument     Kind = "argument"     // 1
	KindNotFound     Kind = "not_found"    // 2
	KindUnauthorized Kind = "unauthorized" // 3
	KindForbidden    Kind = "forbidden"    // 4
	KindOutdated     Kind = "outdated"     // 5: CLI が古い（API が 426 を返した）
	KindNetwork      Kind = "network"      // 6
	KindServer       Kind = "server"       // 7
)

var exitCodes = map[Kind]int{
	KindArgument: 1, KindNotFound: 2, KindUnauthorized: 3, KindForbidden: 4, KindOutdated: 5, KindNetwork: 6, KindServer: 7,
}

// Error は CLI が返すエラー。API のエンベロープ（error / message / hint）を引き継ぐ
type Error struct {
	Kind    Kind
	Code    string
	Message string
	Hint    string
}

func (e *Error) Error() string { return e.Message }
func (e *Error) ExitCode() int { return exitCodes[e.Kind] }

func (e *Error) Envelope() map[string]any {
	code := e.Code
	if code == "" {
		code = string(e.Kind)
	}
	m := map[string]any{"ok": false, "error": code, "message": e.Message}
	if e.Hint != "" {
		m["hint"] = e.Hint
	}
	return m
}

func NewError(kind Kind, message, hint string) *Error {
	return &Error{Kind: kind, Message: message, Hint: hint}
}

// Envelope は API のレスポンス（ok / data / summary / context）
type Envelope map[string]any

func (e Envelope) Summary() string { s, _ := e["summary"].(string); return s }

// DecodeData は data を任意の構造体に読み直す
func (e Envelope) DecodeData(v any) error {
	raw, err := json.Marshal(e["data"])
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

// SetContext は context にキーを足す（CLI 側の情報: config の場所など）
func (e Envelope) SetContext(key string, value any) {
	ctx, _ := e["context"].(map[string]any)
	if ctx == nil {
		ctx = map[string]any{}
	}
	ctx[key] = value
	e["context"] = ctx
}

// Versions は API が返すバージョン情報（X-Pitboard-Api-Version / X-Pitboard-Min-Cli-Version）
type Versions struct {
	API    string
	MinCLI string
}

type Client struct {
	cfg     *config.Config
	token   string
	version string
	http    *http.Client
	// LastVersions は直近のレスポンスのバージョンヘッダー（doctor が見る）
	LastVersions Versions
}

// New はクライアントを作る。token が空なら設定から読む
func New(cfg *config.Config, token, version string) *Client {
	return &Client{cfg: cfg, token: token, version: version}
}

func (c *Client) Get(path string, params url.Values) (Envelope, error) {
	return c.do(http.MethodGet, path, params, nil)
}

// Post / Put / Delete は JSON ボディで送る。書き込み API 用
func (c *Client) Post(path string, body map[string]any) (Envelope, error) {
	return c.do(http.MethodPost, path, nil, body)
}

func (c *Client) Put(path string, body map[string]any) (Envelope, error) {
	return c.do(http.MethodPut, path, nil, body)
}

func (c *Client) Delete(path string) (Envelope, error) {
	return c.do(http.MethodDelete, path, nil, nil)
}

func (c *Client) do(method, path string, params url.Values, body map[string]any) (Envelope, error) {
	token := c.token
	if token == "" {
		token = c.cfg.Token()
	}
	if token == "" {
		return nil, NewError(KindUnauthorized, "No token configured.", "Run: pitboard auth login")
	}

	u := c.cfg.APIURL() + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, NewError(KindArgument, err.Error(), "")
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, u, reader)
	if err != nil {
		return nil, NewError(KindArgument, err.Error(), "")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "pitboard-cli/"+c.version)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	hc, err := c.httpClient()
	if err != nil {
		return nil, err
	}
	res, err := hc.Do(req)
	if err != nil {
		return nil, c.networkError(err)
	}
	defer res.Body.Close()
	c.LastVersions = Versions{API: res.Header.Get("X-Pitboard-Api-Version"), MinCLI: res.Header.Get("X-Pitboard-Min-Cli-Version")}
	raw, _ := io.ReadAll(res.Body)
	return c.handle(res.StatusCode, raw)
}

func (c *Client) httpClient() (*http.Client, error) {
	if c.http != nil {
		return c.http, nil
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if c.cfg.Insecure() {
		tlsConfig.InsecureSkipVerify = true //nolint:gosec // 開発用。PITBOARD_INSECURE=1 のときだけ
	}
	if ca := c.cfg.CAFile(); ca != "" {
		pem, err := os.ReadFile(ca)
		if err != nil {
			return nil, NewError(KindArgument, fmt.Sprintf("Cannot read PITBOARD_CA_FILE: %v", err), "")
		}
		pool, _ := x509.SystemCertPool()
		if pool == nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, NewError(KindArgument, "PITBOARD_CA_FILE is not a valid PEM certificate", "")
		}
		tlsConfig.RootCAs = pool
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	c.http = &http.Client{Timeout: 30 * time.Second, Transport: transport}
	return c.http, nil
}

func (c *Client) handle(status int, body []byte) (Envelope, error) {
	var env Envelope
	parsed := json.Unmarshal(body, &env) == nil && env != nil
	if parsed {
		if ok, _ := env["ok"].(bool); ok {
			return env, nil
		}
	}

	var kind Kind
	switch status {
	case 401:
		kind = KindUnauthorized
	case 403:
		kind = KindForbidden
	case 404:
		kind = KindNotFound
	case 400, 422:
		kind = KindArgument
	case 426:
		kind = KindOutdated
	default:
		kind = KindServer
	}
	e := &Error{Kind: kind, Message: fmt.Sprintf("HTTP %d from %s", status, c.cfg.APIURL())}
	if parsed {
		if m, _ := env["message"].(string); m != "" {
			e.Message = m
		}
		e.Hint, _ = env["hint"].(string)
		e.Code, _ = env["error"].(string)
	}
	return nil, e
}

func (c *Client) networkError(err error) *Error {
	msg := fmt.Sprintf("Cannot connect to %s: %v", c.cfg.APIURL(), err)
	var certErr x509.UnknownAuthorityError
	var hostErr x509.HostnameError
	if errors.As(err, &certErr) || errors.As(err, &hostErr) || strings.Contains(err.Error(), "x509") || strings.Contains(err.Error(), "tls:") {
		return NewError(KindNetwork, msg, "Certificate not trusted. For local development set PITBOARD_CA_FILE to your CA (e.g. mkcert rootCA.pem).")
	}
	var netErr net.Error
	_ = errors.As(err, &netErr)
	return NewError(KindNetwork, msg, fmt.Sprintf("Check PITBOARD_API_URL (current: %s) and run: pitboard doctor", c.cfg.APIURL()))
}
