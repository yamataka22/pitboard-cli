package cli

import (
	"github.com/spf13/cobra"

	"github.com/yamataka22/pitboard-cli/internal/client"
)

func newVersionCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "バージョンを表示する",
		RunE: func(cmd *cobra.Command, args []string) error {
			env := client.Envelope{"ok": true, "data": map[string]any{"version": a.version}, "summary": "pitboard-cli " + a.version, "context": map[string]any{}}
			return a.emit(env, func() { a.printf("pitboard-cli %s\n", a.version) })
		},
	}
}
