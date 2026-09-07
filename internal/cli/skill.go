package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yamataka22/pitboard-cli/internal/client"
	"github.com/yamataka22/pitboard-cli/internal/skill"
)

func newSkillCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{Use: "skill", Short: "AI エージェントへのスキル登録（install / show）"}
	cmd.AddCommand(newSkillInstallCmd(a), newSkillShowCmd(a))
	return cmd
}

func newSkillInstallCmd(a *app) *cobra.Command {
	var target string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "pitboard スキルを AI エージェントに登録する",
		RunE: func(cmd *cobra.Command, args []string) error {
			if target != "claude" && target != "codex" && target != "all" {
				return client.NewError(client.KindArgument, "Unknown target: "+target, "Use one of: claude, codex, all")
			}
			var results []map[string]any
			if target == "claude" || target == "all" {
				path, err := skill.InstallClaude()
				if err != nil {
					return client.NewError(client.KindArgument, fmt.Sprintf("Cannot install skill: %v", err), "")
				}
				results = append(results, map[string]any{"target": "claude", "path": path, "message": "Installed SKILL.md to " + path})
			}
			if target == "codex" || target == "all" {
				snippet := "Run `pitboard skill show` and follow it before using the pitboard CLI."
				results = append(results, map[string]any{"target": "codex", "snippet": snippet,
					"message": "Codex: add the following line to your AGENTS.md\n  " + snippet})
			}
			env := client.Envelope{"ok": true, "data": results, "summary": "skill install (" + target + ")", "context": map[string]any{}}
			return a.emit(env, func() {
				for _, r := range results {
					a.printf("%s\n", r["message"])
				}
			})
		},
	}
	cmd.Flags().StringVar(&target, "target", "claude", "claude | codex | all")
	return cmd
}

func newSkillShowCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "同梱の SKILL.md を表示する",
		RunE: func(cmd *cobra.Command, args []string) error {
			env := client.Envelope{"ok": true, "data": map[string]any{"content": skill.Content, "claude_path": skill.ClaudePath()}, "summary": "SKILL.md", "context": map[string]any{}}
			return a.emit(env, func() { a.printf("%s", skill.Content) })
		},
	}
}
