// Package output は出力。TTY なら人間向け、パイプか --json なら JSON（エンベロープそのまま）。
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/term"
)

// Options は全コマンド共通の出力オプション
type Options struct {
	JSON   bool
	Fields string
}

// IsJSON は --json か、出力先が端末でないときに真
func IsJSON(opts Options, w io.Writer) bool {
	if opts.JSON {
		return true
	}
	f, ok := w.(*os.File)
	if !ok {
		return true
	}
	return !term.IsTerminal(int(f.Fd()))
}

// Emit はエンベロープを出力する。人間向けの場合は human を呼ぶ
func Emit(w io.Writer, env map[string]any, opts Options, human func()) error {
	if IsJSON(opts, w) {
		if opts.Fields != "" {
			env = WithFields(env, opts.Fields)
		}
		return writeJSON(w, env)
	}
	if human != nil {
		human()
	}
	return nil
}

// Error はエラーを出力する。JSON ならエンベロープを stdout、人間向けなら stderr
func Error(stdout, stderr io.Writer, envelope map[string]any, jsonMode bool) {
	if jsonMode {
		_ = writeJSON(stdout, envelope)
		return
	}
	fmt.Fprintf(stderr, "error: %v\n", envelope["message"])
	if hint, ok := envelope["hint"].(string); ok && hint != "" {
		fmt.Fprintf(stderr, "hint:  %s\n", hint)
	}
}

// preferredOrder はエンベロープのキー順。Go の map はソートされるので明示的に並べる
var preferredOrder = []string{"ok", "data", "summary", "context", "error", "message", "hint"}

func writeJSON(w io.Writer, v any) error {
	m, ok := v.(map[string]any)
	if !ok {
		return encode(w, v)
	}
	var buf bytes.Buffer
	buf.WriteString("{\n")
	keys := orderedKeys(m)
	for i, k := range keys {
		val, err := json.Marshal(m[k])
		if err != nil {
			return err
		}
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, val, "  ", "  "); err != nil {
			return err
		}
		fmt.Fprintf(&buf, "  %q: %s", k, pretty.String())
		if i < len(keys)-1 {
			buf.WriteString(",")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("}\n")
	_, err := w.Write(buf.Bytes())
	return err
}

func orderedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	seen := map[string]bool{}
	for _, k := range preferredOrder {
		if _, ok := m[k]; ok {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	rest := make([]string, 0, len(m))
	for k := range m {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	return append(keys, rest...)
}

func encode(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WithFields は --fields a,b で data の列を絞る（トップレベルのキーのみ）
func WithFields(env map[string]any, fields string) map[string]any {
	var keys []string
	for _, k := range strings.Split(fields, ",") {
		if k = strings.TrimSpace(k); k != "" {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return env
	}
	pick := func(row any) any {
		m, ok := row.(map[string]any)
		if !ok {
			return row
		}
		out := map[string]any{}
		for _, k := range keys {
			if v, ok := m[k]; ok {
				out[k] = v
			}
		}
		return out
	}
	out := map[string]any{}
	for k, v := range env {
		out[k] = v
	}
	switch data := env["data"].(type) {
	case []any:
		rows := make([]any, len(data))
		for i, r := range data {
			rows[i] = pick(r)
		}
		out["data"] = rows
	case map[string]any:
		out["data"] = pick(data)
	}
	return out
}

// Table は簡易テーブル。全角文字は幅2として揃える
func Table(w io.Writer, headers []string, rows [][]string) {
	all := append([][]string{headers}, rows...)
	widths := make([]int, len(headers))
	for _, row := range all {
		for i := range headers {
			if i < len(row) {
				if dw := DisplayWidth(row[i]); dw > widths[i] {
					widths[i] = dw
				}
			}
		}
	}
	for idx, row := range all {
		cells := make([]string, len(headers))
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			cells[i] = Pad(cell, widths[i])
		}
		fmt.Fprintln(w, strings.TrimRight(strings.Join(cells, "  "), " "))
		if idx == 0 {
			seps := make([]string, len(headers))
			for i, wd := range widths {
				seps[i] = strings.Repeat("-", wd)
			}
			fmt.Fprintln(w, strings.Join(seps, "  "))
		}
	}
}

func Pad(s string, width int) string {
	if d := width - DisplayWidth(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

// DisplayWidth は東アジアの全角文字を 2、それ以外を 1 として数える
func DisplayWidth(s string) int {
	n := 0
	for _, r := range s {
		if isWide(r) {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func isWide(r rune) bool {
	return (r >= 0x1100 && r <= 0x115F) ||
		(r >= 0x2E80 && r <= 0xA4CF) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE30 && r <= 0xFE4F) ||
		(r >= 0xFF00 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x20000 && r <= 0x3FFFD)
}
