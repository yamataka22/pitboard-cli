// Package cli はコマンド定義。各コマンドは client を呼んで output に渡すだけ。
package cli

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yamataka22/pitboard-cli/internal/client"
	"github.com/yamataka22/pitboard-cli/internal/config"
	"github.com/yamataka22/pitboard-cli/internal/output"
)

// app は1回の実行で共有する状態
type app struct {
	version string
	stdout  io.Writer
	stderr  io.Writer
	stdin   io.Reader
	opts    output.Options
	cfg     *config.Config
	cli     *client.Client
}

func (a *app) config() (*config.Config, error) {
	if a.cfg == nil {
		cfg, err := config.Load()
		if err != nil {
			return nil, client.NewError(client.KindArgument, err.Error(), "")
		}
		a.cfg = cfg
	}
	return a.cfg, nil
}

func (a *app) client() (*client.Client, error) {
	if a.cli == nil {
		cfg, err := a.config()
		if err != nil {
			return nil, err
		}
		a.cli = client.New(cfg, "", a.version)
	}
	return a.cli, nil
}

func (a *app) get(path string, params url.Values) (client.Envelope, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	return c.Get(path, params)
}

func (a *app) post(path string, body map[string]any) (client.Envelope, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	return c.Post(path, body)
}

func (a *app) put(path string, body map[string]any) (client.Envelope, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	return c.Put(path, body)
}

func (a *app) delete(path string) (client.Envelope, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	return c.Delete(path)
}

// requireYes は書き込み系の共通ガード。AI が誤って書き込まないよう --yes を必須にする
func requireYes(yes bool, action string) error {
	if yes {
		return nil
	}
	return &client.Error{Kind: client.KindArgument, Code: "confirmation_required",
		Message: "Refusing to " + action + " without --yes.", Hint: "Re-run with --yes to confirm the write."}
}

func (a *app) emit(env client.Envelope, human func()) error {
	return output.Emit(a.stdout, env, a.opts, human)
}

func (a *app) printf(format string, args ...any) {
	fmt.Fprintf(a.stdout, format, args...)
}

// resolveSpaceID: --space > 環境変数 > 設定ファイル > 所属が1つならそれ
func (a *app) resolveSpaceID(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	cfg, err := a.config()
	if err != nil {
		return "", err
	}
	if id := cfg.SpaceID(); id != "" {
		return id, nil
	}
	env, err := a.get("/spaces", nil)
	if err != nil {
		return "", err
	}
	var spaces []spaceSummary
	if err := env.DecodeData(&spaces); err != nil {
		return "", client.NewError(client.KindServer, "Unexpected response from /spaces", "")
	}
	if len(spaces) == 1 {
		return fmt.Sprint(spaces[0].ID), nil
	}
	return "", &client.Error{
		Kind:    client.KindArgument,
		Code:    "space_required",
		Message: fmt.Sprintf("You belong to %d spaces. Specify one.", len(spaces)),
		Hint:    "Run: pitboard space list, then: pitboard space use ID (or pass --space ID)",
	}
}

func maskToken(token string) string {
	if len(token) < 8 {
		return ""
	}
	return token[:3] + "****" + token[len(token)-4:]
}

func newRoot(a *app) *cobra.Command {
	root := &cobra.Command{
		Use:           "pitboard",
		Version:       a.version,
		Short:         "人と AI エージェントのための pitboard CLI",
		Long:          "pitboard API の薄い皮としての CLI。全コマンドが同じ形の JSON（ok / data / summary / context）を返すので、AI エージェントが自分の使い方を組み立てられる。",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.DisableAutoGenTag = true
	root.CompletionOptions.HiddenDefaultCmd = true
	root.PersistentFlags().BoolVar(&a.opts.JSON, "json", false, "JSON で出力する（パイプ時は自動）")
	root.PersistentFlags().StringVar(&a.opts.Fields, "fields", "", "data に含める列をカンマ区切りで指定（JSON のみ）")
	root.SetOut(a.stdout)
	root.SetErr(a.stderr)
	root.SetIn(io.NopCloser(a.stdin))

	root.AddCommand(newAuthCmd(a), newSpaceCmd(a), newTaskCmd(a), newCommentCmd(a), newSkillCmd(a), newDoctorCmd(a), newVersionCmd(a))
	return root
}

// NewRootCommand はドキュメント生成用にルートコマンドを返す
func NewRootCommand(version string) *cobra.Command {
	a := &app{version: version, stdout: io.Discard, stderr: io.Discard, stdin: strings.NewReader("")}
	return newRoot(a)
}

// Execute は CLI を実行して終了コードを返す（main とテストの両方から使う）
func Execute(version string, args []string, stdout, stderr io.Writer, stdin io.Reader) int {
	a := &app{version: version, stdout: stdout, stderr: stderr, stdin: stdin}
	root := newRoot(a)
	root.SetArgs(args)
	err := root.Execute()
	if err == nil {
		return 0
	}
	var xe *exitError
	if errors.As(err, &xe) {
		return xe.code
	}
	var e *client.Error
	if !errors.As(err, &e) {
		// cobra の引数エラーなど
		e = client.NewError(client.KindArgument, err.Error(), "Run: pitboard --help")
	}
	output.Error(stdout, stderr, e.Envelope(), output.IsJSON(a.opts, stdout))
	return e.ExitCode()
}

// isTerminal は stdin が端末か（トークンの非表示入力に使う）
func isTerminal(r io.Reader) (int, bool) {
	f, ok := r.(*os.File)
	if !ok {
		return 0, false
	}
	return int(f.Fd()), true
}
