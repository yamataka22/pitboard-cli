package cli

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yamataka22/pitboard-cli/internal/client"
	"github.com/yamataka22/pitboard-cli/internal/output"
)

func newTaskCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{Use: "task", Short: "タスク（list / show / create / assign / move / point / archive）"}
	cmd.AddCommand(newTaskListCmd(a), newTaskShowCmd(a), newTaskCreateCmd(a), newTaskAssignCmd(a), newTaskMoveCmd(a), newTaskPointCmd(a), newTaskArchiveCmd(a), newTaskUnarchiveCmd(a))
	return cmd
}

// taskListFlags は task list のオプション。API のパラメータ名と一致させる
type taskListFlags struct {
	space, state, progress, assignee, owner, project, point, kind, archived string
	labels                                                                  []string
	since, until, query, sort, order                                        string
	mine, all                                                               bool
	page, perPage                                                           int
}

func (f *taskListFlags) values() url.Values {
	v := url.Values{}
	set := func(k, val string) {
		if val != "" {
			v.Set(k, val)
		}
	}
	set("state", f.state)
	set("progress", f.progress)
	if f.mine {
		set("assignee", "me")
	} else {
		set("assignee", f.assignee)
	}
	set("owner", f.owner)
	set("project", f.project)
	for _, l := range f.labels {
		v.Add("label[]", l)
	}
	set("point", f.point)
	set("kind", f.kind)
	set("archived", f.archived)
	set("progress_changed_since", f.since)
	set("progress_changed_until", f.until)
	set("q", f.query)
	set("sort", f.sort)
	set("order", f.order)
	if f.page > 0 {
		set("page", strconv.Itoa(f.page))
	}
	if f.perPage > 0 {
		set("per_page", strconv.Itoa(f.perPage))
	}
	return v
}

func newTaskListCmd(a *app) *cobra.Command {
	f := &taskListFlags{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "タスクの一覧（既定: アーカイブ除外、番号の降順）",
		Args:  cobra.NoArgs,
		Long: `スペースのタスクを一覧する。日付は ISO 8601（2026-09-01）で渡す。"7d" や "thisweek" のような相対指定は無い。
--progress-changed-since / --until を付けるとアーカイブ済みも含まれる（除くなら --archived none）。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceID, err := a.resolveSpaceID(f.space)
			if err != nil {
				return err
			}
			path := "/spaces/" + spaceID + "/tasks"
			env, err := a.get(path, f.values())
			if err != nil {
				return err
			}
			if f.all {
				if env, err = a.fetchAllPages(path, f.values(), env); err != nil {
					return err
				}
			}
			var tasks []taskSummary
			_ = env.DecodeData(&tasks)
			return a.emit(env, func() {
				rows := make([][]string, len(tasks))
				for i, t := range tasks {
					rows[i] = []string{"#" + strconv.Itoa(t.Number), t.State, progressName(t.Progress), deref(t.Point, "-"), refName(t.Assignee, "-"), t.Name}
				}
				output.Table(a.stdout, []string{"#", "State", "Progress", "Point", "Assignee", "Name"}, rows)
				a.printf("\n%s\n", env.Summary())
				if p, ok := envPagination(env); ok && p.HasNext {
					a.printf("(page %d of more; use --all or --page %d)\n", p.Page, p.Page+1)
				}
			})
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&f.space, "space", "", "スペース ID（省略時は既定スペース）")
	fl.StringVar(&f.state, "state", "", "backlog | inbox | in_progress | done（カンマ区切りで複数）")
	fl.StringVar(&f.progress, "progress", "", "進捗カラム ID または名前")
	fl.StringVar(&f.assignee, "assignee", "", "メンバー ID または名前 | me | none")
	fl.BoolVar(&f.mine, "mine", false, "--assignee me の別名")
	fl.StringVar(&f.owner, "owner", "", "メンバー ID または名前 | me")
	fl.StringVar(&f.project, "project", "", "プロジェクト ID または名前 | none")
	fl.StringSliceVar(&f.labels, "label", nil, "ラベル ID または名前（繰り返し指定で AND）")
	fl.StringVar(&f.point, "point", "", "h1 | h4 | d1 .. d5 | none")
	fl.StringVar(&f.kind, "kind", "", "task | issue | all（省略時は両方。ただし --state backlog は task だけ）")
	fl.StringVar(&f.archived, "archived", "", "none（アーカイブ済みを除く）| only（アーカイブ済みだけ）| all（既定 none。--progress-changed-* 指定時は all）")
	fl.StringVar(&f.since, "progress-changed-since", "", "今のカラムに置かれた日時がこれ以降（ISO 8601）")
	fl.StringVar(&f.until, "progress-changed-until", "", "今のカラムに置かれた日時がこれ以前（ISO 8601。日付だけならその日の終わりまで）")
	fl.StringVarP(&f.query, "query", "q", "", "番号・タスク名・本文を検索")
	fl.StringVar(&f.sort, "sort", "", "code（既定）| updated | created | progress_changed | position")
	fl.StringVar(&f.order, "order", "", "asc | desc（既定は --sort による）")
	fl.BoolVar(&f.all, "all", false, "ページングを最後まで辿る")
	fl.IntVar(&f.page, "page", 0, "ページ番号")
	fl.IntVar(&f.perPage, "per-page", 0, "1ページの件数（最大 200）")
	return cmd
}

// fetchAllPages は has_next が false になるまで辿り、data を結合したエンベロープを返す
func (a *app) fetchAllPages(path string, params url.Values, first client.Envelope) (client.Envelope, error) {
	all, _ := first["data"].([]any)
	env := first
	for {
		p, ok := envPagination(env)
		if !ok || !p.HasNext {
			break
		}
		params.Set("page", strconv.Itoa(p.Page+1))
		next, err := a.get(path, params)
		if err != nil {
			return nil, err
		}
		rows, _ := next["data"].([]any)
		all = append(all, rows...)
		env = next
	}
	first["data"] = all
	if p, ok := envPagination(env); ok {
		first.SetContext("pagination", map[string]any{"page": 1, "per_page": len(all), "has_next": false, "total": p.Total})
	}
	return first, nil
}

func envPagination(env client.Envelope) (pagination, bool) {
	ctx, _ := env["context"].(map[string]any)
	raw, ok := ctx["pagination"]
	if !ok {
		return pagination{}, false
	}
	b, _ := json.Marshal(raw)
	var p pagination
	if err := json.Unmarshal(b, &p); err != nil {
		return pagination{}, false
	}
	return p, true
}

func newTaskShowCmd(a *app) *cobra.Command {
	var space string
	cmd := &cobra.Command{
		Use:   "show NUMBER",
		Short: "タスクの詳細（本文・TODO・コメント・関連タスク）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceID, err := a.resolveSpaceID(space)
			if err != nil {
				return err
			}
			env, err := a.get("/spaces/"+spaceID+"/tasks/"+args[0], nil)
			if err != nil {
				return err
			}
			var t taskDetail
			_ = env.DecodeData(&t)
			return a.emit(env, func() {
				a.printf("#%d %s\n", t.Number, t.Name)
				a.printf("State:     %s%s\n", t.State, map[bool]string{true: "  (archived)", false: ""}[t.Archived])
				a.printf("Progress:  %s", progressName(t.Progress))
				if t.ProgressChangedAt != nil {
					a.printf("  (since %s)", *t.ProgressChangedAt)
				}
				a.printf("\n")
				a.printf("Assignee:  %s\n", refName(t.Assignee, "-"))
				a.printf("Owner:     %s\n", refName(t.Owner, "-"))
				a.printf("Project:   %s\n", refName(t.Project, "-"))
				a.printf("Point:     %s\n", deref(t.Point, "-"))
				a.printf("Kind:      %s\n", t.Kind)
				if len(t.Labels) > 0 {
					names := make([]string, len(t.Labels))
					for i, l := range t.Labels {
						names[i] = l.Name
					}
					a.printf("Labels:    %s\n", strings.Join(names, ", "))
				}
				a.printf("URL:       %s\n", t.URL)
				if t.Document != nil && *t.Document != "" {
					a.printf("\n%s\n", *t.Document)
				}
				if len(t.Todos) > 0 {
					a.printf("\nTodos (%d/%d):\n", t.TodoDoneCount, t.TodoCount)
					for _, td := range t.Todos {
						mark := " "
						if td.Done {
							mark = "x"
						}
						a.printf("  [%s] %s\n", mark, td.Name)
					}
				}
				if len(t.Comments) > 0 {
					a.printf("\nComments:\n")
					for _, c := range t.Comments {
						a.printf("  [%d] %s (%s):\n    %s\n", c.ID, c.Author.Name, c.CreatedAt, strings.ReplaceAll(c.Content, "\n", "\n    "))
						for _, r := range c.Replies {
							a.printf("    ↳ [%d] %s (%s):\n      %s\n", r.ID, r.Author.Name, r.CreatedAt, strings.ReplaceAll(r.Content, "\n", "\n      "))
						}
					}
				}
				if len(t.RelatedTasks) > 0 {
					a.printf("\nRelated:\n")
					for _, r := range t.RelatedTasks {
						a.printf("  #%d %s (%s)\n", r.Number, r.Name, r.State)
					}
				}
			})
		},
	}
	cmd.Flags().StringVar(&space, "space", "", "スペース ID（省略時は既定スペース）")
	return cmd
}

func progressName(p *progressRef) string {
	if p == nil {
		return "-"
	}
	return p.Name
}

func refName(r *ref, fallback string) string {
	if r == nil {
		return fallback
	}
	return r.Name
}

func deref(s *string, fallback string) string {
	if s == nil || *s == "" {
		return fallback
	}
	return *s
}
