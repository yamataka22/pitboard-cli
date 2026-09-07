package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/yamataka22/pitboard-cli/internal/client"
)

func newCommentCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{Use: "comment", Short: "コメント（add）"}
	cmd.AddCommand(newCommentAddCmd(a))
	return cmd
}

func newCommentAddCmd(a *app) *cobra.Command {
	var (
		space, body, replyTo string
		mentions             []string
		yes                  bool
	)
	cmd := &cobra.Command{
		Use:   "add NUMBER --body TEXT|- [--reply-to COMMENT_ID] [--mention ID]... --yes",
		Short: "タスクにコメントを追加する（返信・メンション可）",
		Long: `タスクにコメントを追加する。--body - なら本文を標準入力から読む。
--reply-to にコメント ID を渡すと返信になる（ID は task show で確認。返信への返信はできない）。
--mention にメンバー ID を渡すと本文の先頭に [@名前](mention:ID) が付き、相手に通知が届く。本文に直接この記法を書いてもよい。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireYes(yes, "add a comment"); err != nil {
				return err
			}
			if body == "" {
				return client.NewError(client.KindArgument, "--body is required", "Pass the text with --body, or --body - to read stdin")
			}
			spaceID, err := a.resolveSpaceID(space)
			if err != nil {
				return err
			}
			text := body
			if body == "-" {
				if text, err = readDocument("-", a.stdin); err != nil {
					return err
				}
			}
			if strings.TrimSpace(text) == "" {
				return client.NewError(client.KindArgument, "Comment body is empty", "")
			}
			payload := map[string]any{"content": text}
			setIf(payload, "parent", replyTo)
			if len(mentions) > 0 {
				payload["mentions"] = mentions
			}
			env, err := a.post("/spaces/"+spaceID+"/tasks/"+args[0]+"/comments", payload)
			if err != nil {
				return err
			}
			return a.emitTask(env)
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&space, "space", "", "スペース ID（省略時は既定スペース）")
	fl.StringVar(&body, "body", "", "本文（markdown）。- で標準入力")
	fl.StringVar(&replyTo, "reply-to", "", "返信先のコメント ID")
	fl.StringSliceVar(&mentions, "mention", nil, "メンションするメンバー ID または名前（繰り返し指定可）")
	fl.BoolVar(&yes, "yes", false, "書き込みを確認したことを示す（必須）")
	return cmd
}
