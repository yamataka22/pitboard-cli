package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yamataka22/pitboard-cli/internal/client"
	"github.com/yamataka22/pitboard-cli/internal/output"
)

func newSpaceCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{Use: "space", Short: "スペース（list / use / show）"}
	cmd.AddCommand(newSpaceListCmd(a), newSpaceUseCmd(a), newSpaceShowCmd(a))
	return cmd
}

func newSpaceListCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "所属スペースの一覧",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.config()
			if err != nil {
				return err
			}
			env, err := a.get("/spaces", nil)
			if err != nil {
				return err
			}
			var spaces []spaceSummary
			_ = env.DecodeData(&spaces)
			return a.emit(env, func() {
				rows := make([][]string, len(spaces))
				for i, s := range spaces {
					mark := ""
					if fmt.Sprint(s.ID) == cfg.SpaceID() {
						mark = "*"
					}
					rows[i] = []string{fmt.Sprint(s.ID), s.Name, s.Role, mark}
				}
				output.Table(a.stdout, []string{"ID", "Name", "Role", "Default"}, rows)
				a.printf("\n%s\n", env.Summary())
			})
		},
	}
}

func newSpaceUseCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "use ID",
		Short: "既定スペースを設定する（所属を確認してから保存）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := a.config()
			if err != nil {
				return err
			}
			env, err := a.get("/spaces/"+args[0], nil)
			if err != nil {
				return err
			}
			var space spaceDetail
			if err := env.DecodeData(&space); err != nil {
				return client.NewError(client.KindServer, "Unexpected response from /spaces/:id", "")
			}
			if err := cfg.Save(map[string]any{"space_id": space.ID}); err != nil {
				return client.NewError(client.KindArgument, fmt.Sprintf("Cannot write config: %v", err), "")
			}
			env.SetContext("config", cfg.Path)
			env.SetContext("space_id", cfg.SpaceID())
			return a.emit(env, func() {
				a.printf("Default space: %d %s (saved to %s)\n", space.ID, space.Name, cfg.Path)
			})
		},
	}
}

func newSpaceShowCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "show [ID]",
		Short: "スペースの語彙を表示する（進捗カラム・プロジェクト・ラベル・メンバー）",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			explicit := ""
			if len(args) == 1 {
				explicit = args[0]
			}
			id, err := a.resolveSpaceID(explicit)
			if err != nil {
				return err
			}
			env, err := a.get("/spaces/"+id, nil)
			if err != nil {
				return err
			}
			var d spaceDetail
			_ = env.DecodeData(&d)
			return a.emit(env, func() {
				a.printf("Space: %d %s  (you: %s / %s)\n\n", d.ID, d.Name, d.Me.Name, d.Role)
				a.printf("Progresses:\n")
				for _, p := range d.Progresses {
					suffix := ""
					if p.Done {
						suffix = "  (done)"
					}
					a.printf("  %-5d %s%s\n", p.ID, p.Name, suffix)
				}
				a.printf("\nProjects:\n")
				for _, p := range d.Projects {
					a.printf("  %-5d %s\n", p.ID, p.Name)
				}
				a.printf("\nLabels:\n")
				for _, l := range d.Labels {
					a.printf("  %-5d %s\n", l.ID, l.Name)
				}
				a.printf("\nMembers:\n")
				for _, m := range d.Members {
					var flags []string
					if m.Me {
						flags = append(flags, "me")
					}
					if m.Role == "admin" {
						flags = append(flags, "admin")
					}
					if !m.Assignee {
						flags = append(flags, "not assignable")
					}
					suffix := ""
					if len(flags) > 0 {
						suffix = "  (" + strings.Join(flags, ", ") + ")"
					}
					a.printf("  %-5d %s%s\n", m.ID, m.Name, suffix)
				}
			})
		},
	}
}
