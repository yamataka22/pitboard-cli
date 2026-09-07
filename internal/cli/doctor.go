package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yamataka22/pitboard-cli/internal/client"
	"github.com/yamataka22/pitboard-cli/internal/skill"
)

type check struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

// 設定・トークン・API 接続・既定スペース・スキルを順に診断する
func newDoctorCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "設定・トークン・API 接続・既定スペース・スキル登録を診断する",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.config()
			if err != nil {
				return err
			}
			var checks []check

			if cfg.Exists() {
				checks = append(checks, check{"config", true, fmt.Sprintf("%s (api_url: %s, from %s)", cfg.Path, cfg.APIURL(), cfg.Source("api_url")), ""})
			} else {
				checks = append(checks, check{"config", true, fmt.Sprintf("no config file yet (%s); using env/defaults (api_url: %s)", cfg.Path, cfg.APIURL()), ""})
			}

			if cfg.Token() != "" {
				checks = append(checks, check{"token", true, "token from " + cfg.Source("token"), ""})

				env, err := a.get("/me", nil)
				if err != nil {
					var e *client.Error
					if ce, ok := err.(*client.Error); ok {
						e = ce
					} else {
						e = client.NewError(client.KindServer, err.Error(), "")
					}
					checks = append(checks, check{"api", false, e.Message, e.Hint})
				} else {
					var me meData
					_ = env.DecodeData(&me)
					checks = append(checks, check{"api", true, fmt.Sprintf("%s ok, logged in as %s (token: %s)", cfg.APIURL(), me.Name, me.Token.Scope), ""})
					if c, _ := a.client(); c != nil {
						checks = append(checks, versionCheck(a.version, c.LastVersions))
					}
					checks = append(checks, spaceCheck(cfg.SpaceID(), cfg.Source("space_id"), me.Spaces))
				}
			} else {
				checks = append(checks, check{"token", false, "no token", "Run: pitboard auth login"})
			}

			if _, err := os.Stat(skill.ClaudePath()); err == nil {
				checks = append(checks, check{"skill", true, "Claude Code skill installed (" + skill.ClaudePath() + ")", ""})
			} else {
				checks = append(checks, check{"skill", false, "Claude Code skill not installed", "Run: pitboard skill install"})
			}

			failed := 0
			for _, c := range checks {
				if !c.OK {
					failed++
				}
			}
			summary := fmt.Sprintf("%d checks passed", len(checks))
			if failed > 0 {
				summary = fmt.Sprintf("%d of %d checks failed", failed, len(checks))
			}
			data := make([]any, len(checks))
			for i, c := range checks {
				m := map[string]any{"name": c.Name, "ok": c.OK, "message": c.Message}
				if c.Hint != "" {
					m["hint"] = c.Hint
				}
				data[i] = m
			}
			env := client.Envelope{"ok": failed == 0, "data": data, "summary": summary, "context": map[string]any{}}
			if err := a.emit(env, func() {
				for _, c := range checks {
					status := "OK"
					if !c.OK {
						status = "NG"
					}
					a.printf("[%s] %-7s %s\n", status, c.Name, c.Message)
					if !c.OK && c.Hint != "" {
						a.printf("         hint: %s\n", c.Hint)
					}
				}
				a.printf("\n%s\n", summary)
			}); err != nil {
				return err
			}
			if failed > 0 {
				return &exitError{code: 1}
			}
			return nil
		},
	}
}

func spaceCheck(spaceID, source string, spaces []spaceSummary) check {
	if spaceID != "" {
		for _, s := range spaces {
			if fmt.Sprint(s.ID) == spaceID {
				return check{"space", true, fmt.Sprintf("default space %d %s (from %s)", s.ID, s.Name, source), ""}
			}
		}
		return check{"space", false, fmt.Sprintf("default space %s not found or you are not a member", spaceID), "Run: pitboard space list, then: pitboard space use ID"}
	}
	switch len(spaces) {
	case 0:
		return check{"space", false, "you do not belong to any space", "Create or join a space on the web"}
	case 1:
		return check{"space", true, fmt.Sprintf("no default set; the only space %d %s will be used", spaces[0].ID, spaces[0].Name), ""}
	default:
		return check{"space", false, fmt.Sprintf("no default space set (you belong to %d)", len(spaces)), "Run: pitboard space use ID"}
	}
}

// exitError は出力済みで終了コードだけ返したいときに使う
type exitError struct{ code int }

func (e *exitError) Error() string { return "" }

// versionCheck は CLI と API のバージョンのずれを見る。
// API が受け付ける CLI の下限（X-Pitboard-Min-Cli-Version）より古ければ NG。開発ビルドは判定しない
func versionCheck(cliVersion string, v client.Versions) check {
	info := fmt.Sprintf("cli %s, api v%s (min cli %s)", cliVersion, orUnknown(v.API), orUnknown(v.MinCLI))
	cur, ok := parseSemver(cliVersion)
	if !ok {
		return check{"version", true, info + " — dev build, not checked", ""}
	}
	min, ok := parseSemver(v.MinCLI)
	if !ok {
		return check{"version", true, info, ""}
	}
	if compareSemver(cur, min) < 0 {
		return check{"version", false, info + " — this CLI is too old", "Upgrade: brew upgrade pitboard (or download the latest release)"}
	}
	return check{"version", true, info, ""}
}

func orUnknown(s string) string {
	if s == "" {
		return "?"
	}
	return s
}

func parseSemver(s string) ([3]int, bool) {
	var v [3]int
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return v, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(strings.SplitN(p, "-", 2)[0])
		if err != nil {
			return v, false
		}
		v[i] = n
	}
	return v, true
}

func compareSemver(a, b [3]int) int {
	for i := range a {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}
