package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yamataka22/pitboard-cli/internal/client"
)

// 書き込み系。API の意図ごとのエンドポイント（assignee / progress / point / archive）を1コマンドずつ対応させる。
// タスクを直接 update する口は無い。

func newTaskCreateCmd(a *app) *cobra.Command {
	var (
		space, name, document, kind, project, point, assignee, progress string
		labels                                                          []string
		yes                                                             bool
	)
	cmd := &cobra.Command{
		Use:   "create --name NAME [--assignee ID|me] [--progress ID] ... --yes",
		Short: "タスクを作成する（報告者は自分、kind の既定は confirmed）",
		Long: `タスクを作成する。作成時に決められることは1回で受け付ける。
--document FILE はファイルから本文（markdown）を読む。--document - なら標準入力から読む。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireYes(yes, "create a task"); err != nil {
				return err
			}
			if strings.TrimSpace(name) == "" {
				return client.NewError(client.KindArgument, "--name is required", "Run: pitboard task create --name \"...\" --yes")
			}
			spaceID, err := a.resolveSpaceID(space)
			if err != nil {
				return err
			}
			body := map[string]any{"name": name}
			if document != "" {
				text, err := readDocument(document, a.stdin)
				if err != nil {
					return err
				}
				body["document"] = text
			}
			setIf(body, "kind", kind)
			setIf(body, "project", project)
			setIf(body, "point", point)
			setIf(body, "assignee", assignee)
			setIf(body, "progress", progress)
			if len(labels) > 0 {
				body["labels"] = labels
			}
			env, err := a.post("/spaces/"+spaceID+"/tasks", body)
			if err != nil {
				return err
			}
			return a.emitTask(env)
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&space, "space", "", "スペース ID（省略時は既定スペース）")
	fl.StringVar(&name, "name", "", "タスク名（必須）")
	fl.StringVar(&document, "document", "", "本文（markdown）。ファイルパスか、- で標準入力")
	fl.StringVar(&kind, "kind", "", "confirmed（既定）| issue")
	fl.StringVar(&project, "project", "", "プロジェクト ID")
	fl.StringVar(&point, "point", "", "h1 | h4 | d1 .. d5")
	fl.StringSliceVar(&labels, "label", nil, "ラベル ID（繰り返し指定可）")
	fl.StringVar(&assignee, "assignee", "", "メンバー ID | me")
	fl.StringVar(&progress, "progress", "", "進捗カラム ID（そのカラムに置いた状態で作る）")
	fl.BoolVar(&yes, "yes", false, "書き込みを確認したことを示す（必須）")
	return cmd
}

func newTaskAssignCmd(a *app) *cobra.Command {
	var space, to string
	var yes bool
	cmd := &cobra.Command{
		Use:   "assign NUMBER --to ID|me|none --yes",
		Short: "担当者を変更する（1タスク1担当）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireYes(yes, "change the assignee"); err != nil {
				return err
			}
			if to == "" {
				return client.NewError(client.KindArgument, "--to is required", "Use a member id, me or none")
			}
			spaceID, err := a.resolveSpaceID(space)
			if err != nil {
				return err
			}
			env, err := a.put("/spaces/"+spaceID+"/tasks/"+args[0]+"/assignee", map[string]any{"assignee": to})
			if err != nil {
				return err
			}
			return a.emitTask(env)
		},
	}
	cmd.Flags().StringVar(&space, "space", "", "スペース ID（省略時は既定スペース）")
	cmd.Flags().StringVar(&to, "to", "", "メンバー ID | me | none")
	cmd.Flags().BoolVar(&yes, "yes", false, "書き込みを確認したことを示す（必須）")
	return cmd
}

func newTaskMoveCmd(a *app) *cobra.Command {
	var space, to string
	var yes bool
	cmd := &cobra.Command{
		Use:   "move NUMBER --to PROGRESS_ID|none --yes",
		Short: "進捗カラムへ移動する（none で進捗なしへ。アーカイブ済みは解除される）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireYes(yes, "move the task"); err != nil {
				return err
			}
			if to == "" {
				return client.NewError(client.KindArgument, "--to is required", "Run: pitboard space show to see progress IDs")
			}
			spaceID, err := a.resolveSpaceID(space)
			if err != nil {
				return err
			}
			env, err := a.put("/spaces/"+spaceID+"/tasks/"+args[0]+"/progress", map[string]any{"progress": to})
			if err != nil {
				return err
			}
			return a.emitTask(env)
		},
	}
	cmd.Flags().StringVar(&space, "space", "", "スペース ID（省略時は既定スペース）")
	cmd.Flags().StringVar(&to, "to", "", "進捗カラム ID | none")
	cmd.Flags().BoolVar(&yes, "yes", false, "書き込みを確認したことを示す（必須）")
	return cmd
}

func newTaskPointCmd(a *app) *cobra.Command {
	var space, to string
	var yes bool
	cmd := &cobra.Command{
		Use:   "point NUMBER --to h1|h4|d1..d5|none --yes",
		Short: "ポイント（サイズ）を設定する",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireYes(yes, "change the point"); err != nil {
				return err
			}
			if to == "" {
				return client.NewError(client.KindArgument, "--to is required", "Use one of: h1, h4, d1, d2, d3, d4, d5, none")
			}
			spaceID, err := a.resolveSpaceID(space)
			if err != nil {
				return err
			}
			env, err := a.put("/spaces/"+spaceID+"/tasks/"+args[0]+"/point", map[string]any{"point": to})
			if err != nil {
				return err
			}
			return a.emitTask(env)
		},
	}
	cmd.Flags().StringVar(&space, "space", "", "スペース ID（省略時は既定スペース）")
	cmd.Flags().StringVar(&to, "to", "", "h1 | h4 | d1 .. d5 | none")
	cmd.Flags().BoolVar(&yes, "yes", false, "書き込みを確認したことを示す（必須）")
	return cmd
}

func newTaskArchiveCmd(a *app) *cobra.Command {
	var space string
	var undo, yes bool
	cmd := &cobra.Command{
		Use:   "archive NUMBER [--undo] --yes",
		Short: "タスクをアーカイブする（--undo で解除）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireYes(yes, "archive the task"); err != nil {
				return err
			}
			spaceID, err := a.resolveSpaceID(space)
			if err != nil {
				return err
			}
			path := "/spaces/" + spaceID + "/tasks/" + args[0] + "/archive"
			var env client.Envelope
			if undo {
				env, err = a.delete(path)
			} else {
				env, err = a.post(path, map[string]any{})
			}
			if err != nil {
				return err
			}
			return a.emitTask(env)
		},
	}
	cmd.Flags().StringVar(&space, "space", "", "スペース ID（省略時は既定スペース）")
	cmd.Flags().BoolVar(&undo, "undo", false, "アーカイブを解除する")
	cmd.Flags().BoolVar(&yes, "yes", false, "書き込みを確認したことを示す（必須）")
	return cmd
}

// emitTask は書き込み後の応答（task show と同じ形）を1行で表示する
func (a *app) emitTask(env client.Envelope) error {
	var t taskSummary
	_ = env.DecodeData(&t)
	return a.emit(env, func() {
		a.printf("%s\n", env.Summary())
		a.printf("#%d %s  state=%s progress=%s assignee=%s point=%s\n", t.Number, t.Name, t.State, progressName(t.Progress), refName(t.Assignee, "-"), deref(t.Point, "-"))
		a.printf("%s\n", t.URL)
	})
}

func setIf(m map[string]any, key, value string) {
	if value != "" {
		m[key] = value
	}
}

func readDocument(source string, stdin io.Reader) (string, error) {
	var b []byte
	var err error
	if source == "-" {
		b, err = io.ReadAll(stdin)
	} else {
		b, err = os.ReadFile(source)
	}
	if err != nil {
		return "", client.NewError(client.KindArgument, fmt.Sprintf("Cannot read document: %v", err), "")
	}
	return string(b), nil
}
