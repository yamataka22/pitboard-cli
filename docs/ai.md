# AI エージェントと使う

pitboard 側に「今週完了したタスクをまとめる」のような機能はありません。材料（フィルタとフィールド）だけを CLI が提供し、組み合わせはユーザーと AI に任せています。

## 仕組み

1. `pitboard skill install` が `~/.claude/skills/pitboard/SKILL.md` を書きます。ユーザー全体の設定なので、Claude Code をどのディレクトリで起動しても使えます
2. ユーザーが自然言語で頼むと、Claude Code は SKILL.md を読み、`pitboard` コマンドを組み立てて実行し、JSON を読んで答えます
3. 全コマンドは `--json`（パイプ時は自動）で同じ形の JSON を返します。`summary` を読んでから必要なら `data` を見る、エラー時は `hint` のコマンドを実行する、という約束です

Claude Code は依頼文からスキルの要否を判断します。「タスク」や「pitboard」が含まれていれば拾いますが、拾わない場合は「pitboard で」と一言添えるか、`/pitboard` とスキル名を明示してください。

Codex は `pitboard skill install --target codex` が表示する1行を `AGENTS.md` に足します。

## 頼み方の例

```
自分が担当のタスクのうち、8/31〜9/4 に完了になったものの番号とタイトルを教えて
```

AI はこう組み立てます。

```sh
pitboard task list --mine --state done \
  --progress-changed-since 2026-08-31 --progress-changed-until 2026-09-04 \
  --json --fields number,name
```

他の例:

- 「進行中で1週間以上動いていないタスクは？」→ `--state in_progress` で取り、`progress_changed_at` が古いものを挙げる
- 「担当が決まっていないタスクを一覧して」→ `--state backlog`
- 「#42 の内容を要約して」→ `task show 42`
- 「『週報を書く』を自分の担当で着手中に作って」→ `space show` で着手中の ID を確認し、`task create --name "週報を書く" --assignee me --progress 7 --yes`
- 「#42 に『レビューお願いします』と鈴木さん宛にコメントして」→ `space show` で鈴木さんの ID を確認し、`comment add 42 --body "レビューお願いします" --mention 13 --yes`

## 書き込みの安全装置

- 書き込みコマンド（create / assign / move / point / archive / comment add）は `--yes` が無いと実行されません。AI が確認なしに書き込むのを防ぐためです
- SKILL.md は AI に「書く前にユーザーの意図を確認する」「二重作成を避けるため既存を確認する」と指示しています
- 読み取り専用トークンを AI に渡しておけば、書き込みは API 側で拒否されます

## 個人コマンドとして保存する

同じ依頼を繰り返すなら、AI に頼んで Claude Code の個人コマンドにしておけます。AI に「今のやり取りを `/weekly-done` にして」と頼めば作ってくれます。

```markdown
<!-- ~/.claude/commands/weekly-done.md -->
今日の日付から今週の月曜日を求め、
`pitboard task list --state done --progress-changed-since <月曜日> --json` を実行して、
担当者ごとに完了タスクをまとめて報告して。
```

`~/.claude/skills/pitboard/` は pitboard が配る語彙とルール、`~/.claude/commands/` はユーザー自身の頼み方の定型文、という2層です。前者は `skill install` で上書きされるので編集しないでください。

## 定期的に動かす

内容が決まっているタスクなら AI を挟まず、cron や launchd から CLI を直接呼ぶのが確実です。

```sh
# 毎週木曜 9:00 に「週報を書く」を自分の担当・着手中で作る（着手中カラムの id は space show で確認）
0 9 * * 4 pitboard task create --name "週報を書く" --point h4 --assignee me --progress 7 --yes
```

二重作成を避けるなら、先に `pitboard task list -q "週報を書く" --progress-changed-since <今週の月曜>` で確認します。

判断が要る場合（先週の完了を見て今週のレビュータスクを作る等）は、非対話で Claude Code を呼びます。

```sh
claude -p "/weekly-review" --allowedTools "Bash(pitboard:*)"
```

- 書き込みには読み書きトークンが必要です。cron 用に別のトークンを発行するのが安全です
- macOS の cron はスリープ中の実行を飛ばします。launchd の `StartCalendarInterval` なら起動後に実行されます。二重実行に備えて、上の重複確認を入れておいてください
