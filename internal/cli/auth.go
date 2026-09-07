package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/yamataka22/pitboard-cli/internal/client"
)

func newAuthCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "認証（login / status / logout）"}
	cmd.AddCommand(newAuthLoginCmd(a), newAuthStatusCmd(a), newAuthLogoutCmd(a))
	return cmd
}

func newAuthLoginCmd(a *app) *cobra.Command {
	var apiURL string
	cmd := &cobra.Command{
		Use:   "login [TOKEN]",
		Short: "アクセストークンを保存する（TOKEN 省略時はプロンプトで入力）",
		Long: `Web の「プロフィール > アクセストークン」でトークンを発行してから実行する。
所属スペースが1つならそれが既定スペースになる。
--api-url を付けるとローカルやセルフホストの pitboard に向けられる（設定ファイルに保存）。`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token := ""
			if len(args) == 1 {
				token = args[0]
			} else {
				token = a.askToken()
			}
			token = strings.TrimSpace(token)
			if token == "" {
				return client.NewError(client.KindArgument, "Token is empty.", "Run: pitboard auth login <token>")
			}
			cfg, err := a.config()
			if err != nil {
				return err
			}
			if apiURL != "" {
				if err := cfg.Save(map[string]any{"api_url": strings.TrimRight(apiURL, "/")}); err != nil {
					return client.NewError(client.KindArgument, fmt.Sprintf("Cannot write config: %v", err), "")
				}
			}
			env, err := client.New(cfg, token, a.version).Get("/me", nil)
			if err != nil {
				return err
			}
			var me meData
			if err := env.DecodeData(&me); err != nil {
				return client.NewError(client.KindServer, "Unexpected response from /me", "")
			}
			attrs := map[string]any{"token": token}
			if len(me.Spaces) == 1 {
				attrs["space_id"] = me.Spaces[0].ID
			}
			if err := cfg.Save(attrs); err != nil {
				return client.NewError(client.KindArgument, fmt.Sprintf("Cannot write config: %v", err), "")
			}
			env.SetContext("config", cfg.Path)
			env.SetContext("space_id", cfg.SpaceID())
			return a.emit(env, func() {
				a.printf("Logged in as %s (%s)\n", me.Name, me.Email)
				a.printf("Token:    %s (%s)\n", me.Token.Name, me.Token.Scope)
				a.printf("API:      %s\n", cfg.APIURL())
				a.printf("Config:   %s\n", cfg.Path)
				switch len(me.Spaces) {
				case 0:
					a.printf("Space:    none (you do not belong to any space)\n")
				case 1:
					a.printf("Space:    %d %s (default)\n", me.Spaces[0].ID, me.Spaces[0].Name)
				default:
					names := make([]string, len(me.Spaces))
					for i, s := range me.Spaces {
						names[i] = fmt.Sprintf("%d %s", s.ID, s.Name)
					}
					a.printf("Spaces:   %s\n", strings.Join(names, ", "))
					a.printf("Run: pitboard space use ID  to set the default space\n")
				}
			})
		},
	}
	cmd.Flags().StringVar(&apiURL, "api-url", "", "設定ファイルに保存する API の URL（例: https://localhost:3000/api/v1）")
	return cmd
}

func newAuthStatusCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "ログイン中のユーザー・トークン・既定スペースを表示する",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.config()
			if err != nil {
				return err
			}
			env, err := a.get("/me", nil)
			if err != nil {
				return err
			}
			var me meData
			_ = env.DecodeData(&me)
			env.SetContext("api_url", cfg.APIURL())
			env.SetContext("config", cfg.Path)
			env.SetContext("space_id", cfg.SpaceID())
			env.SetContext("token_source", cfg.Source("token"))
			return a.emit(env, func() {
				a.printf("User:     %s (%s)\n", me.Name, me.Email)
				a.printf("Token:    %s (%s, %s, from %s)\n", me.Token.Name, maskToken(cfg.Token()), me.Token.Scope, cfg.Source("token"))
				a.printf("API:      %s\n", cfg.APIURL())
				id := cfg.SpaceID()
				if id == "" {
					a.printf("Space:    (not set)\n")
					return
				}
				for _, s := range me.Spaces {
					if fmt.Sprint(s.ID) == id {
						a.printf("Space:    %d %s\n", s.ID, s.Name)
						return
					}
				}
				a.printf("Space:    %s (not found or you are not a member; run: pitboard space use ID)\n", id)
			})
		},
	}
}

func newAuthLogoutCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "設定ファイルからトークンを削除する（無効化は Web で失効させる）",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.config()
			if err != nil {
				return err
			}
			if err := cfg.Delete("token", "space_id"); err != nil {
				return client.NewError(client.KindArgument, fmt.Sprintf("Cannot write config: %v", err), "")
			}
			env := client.Envelope{"ok": true, "data": nil, "summary": "Logged out", "context": map[string]any{"config": cfg.Path}}
			return a.emit(env, func() {
				a.printf("Removed token from %s\n", cfg.Path)
				a.printf("To disable the token itself, revoke it on the web: Profile > Access tokens\n")
			})
		},
	}
}

// askToken はプロンプトでトークンを読む。端末なら非表示入力
func (a *app) askToken() string {
	if fd, ok := isTerminal(a.stdin); ok && term.IsTerminal(fd) {
		fmt.Fprint(a.stderr, "Paste your token (input hidden): ")
		b, err := term.ReadPassword(fd)
		fmt.Fprintln(a.stderr)
		if err != nil {
			return ""
		}
		return string(b)
	}
	line, _ := bufio.NewReader(a.stdin).ReadString('\n')
	return line
}
