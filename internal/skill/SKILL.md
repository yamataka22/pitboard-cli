---
name: pitboard
description: Use the pitboard CLI to read and write tasks (list, show, create, assign, move, point, archive, comment with mentions) and read a space's vocabulary. Read this before running any `pitboard` command.
---

# pitboard

pitboard はプレイングマネージャー向けのカンバン型タスク管理。日付ではなく優先度とポイント（サイズ）で管理する。

## ルール（AI も守る）
- 1タスク1担当。複数担当は付けられない
- タスクは5日以内に終わる粒度（PR 相当）。大きすぎるタスクは分割を提案する
- 期限日は無い。「いつまで」は聞かない
- 進捗は担当者がタスクに置く。マネージャーは取りに行かない

## 語彙
- space: テナント。ユーザーは複数に所属できる。既定スペースは `pitboard space use ID` で設定
- progress: 進捗カラム。名前はスペースごとに違うので `space show` で確認し、絞り込みは `--progress <カラム名>` が基本。`done: true` のカラムは複数あることがある（例: リリース待ち / リリース完了）
- state: backlog（担当なし進捗なし。Web の「担当未定」）/ inbox（担当あり進捗なし。Web の「インボックス」）/ in_progress / done（done カラムのどれかにいる）
- progress_changed_at: 今のカラムに置かれた日時。「今週リリース完了に入った」は progress=リリース完了 かつこれが今週。「詰まり」はこれが古い in_progress
- point: h1(1時間) / h4(半日) / d1〜d5(日)
- kind: task（タスク）/ issue（未確定の課題）。--kind 省略時は両方だが、--state backlog だけは task のみ（issue も見るなら --kind all）

## 使い方
- 最初に `pitboard space show` で、そのスペースの進捗カラム・メンバー・プロジェクトの名前と ID を確認する
- progress / project / label / assignee / mention は ID でも名前でも指定できる（数字だけなら ID、それ以外は名前。同名が複数なら 422 になるので ID を使う）
- 日付は今日の日付から自分で計算し、ISO 形式（2026-09-01）で渡す。相対指定（7d, thisweek）は無い
- 「完了」と言われたら、どのカラムを指すか `space show` で確認する。特定のカラムなら `--progress <カラム名>`、done カラム全部をまとめてなら `--state done`

## コマンド
- `pitboard space show`: 進捗カラム（done がどれか）、メンバー、プロジェクト、ラベルの名前と ID
- `pitboard task list [--state S] [--progress ID|NAME] [--assignee ID|NAME|me|none] [--mine] [--project ID|NAME|none] [--label ID|NAME] [--point P] [--kind K] [--archived none|only|all] [--progress-changed-since DATE] [--progress-changed-until DATE] [-q TEXT] [--sort code|updated|created|progress_changed|position] [--order asc|desc] [--all]`
- `pitboard task show NUMBER`: 本文・TODO・コメント・関連タスク
- `pitboard auth status` / `pitboard doctor`: 認証と接続の確認

## 書き込み（読み書きトークンが必要。すべて --yes が必須）
- `pitboard task create --name NAME [--document FILE|-] [--kind task|issue] [--project ID|NAME] [--point P] [--label ID|NAME] [--assignee ID|NAME|me] [--progress ID|NAME] --yes`
- `pitboard task assign NUMBER --to ID|NAME|me|none --yes`: 担当は常に1人。none で解除
- `pitboard task move NUMBER --to PROGRESS_ID|NAME|none --yes`: 進捗カラムの移動。none で進捗なしへ。アーカイブ済みなら解除される
- `pitboard task point NUMBER --to h1|h4|d1..d5|none --yes`
- `pitboard task archive NUMBER --yes` / `pitboard task unarchive NUMBER --yes`: アーカイブと解除。削除は無いのでアーカイブで代替する
- `pitboard comment add NUMBER --body TEXT|- [--reply-to COMMENT_ID] [--mention MEMBER_ID|NAME] --yes`: コメント。返信は task show に出るコメント ID を --reply-to に渡す（返信への返信は不可）
- メンション: `--mention ID|NAME` で本文の先頭に `[@名前](mention:ID)` が付き相手に通知される。本文に直接この記法を書いてもよい。`@名前` だけでは通知されない
- タスク名や本文を直接書き換える口は無い。プロジェクトやラベルの付け替え、コメントの編集・削除も無い
- 書き込む前にユーザーの意図を確認し、同じタスクを二重に作らないよう `task list -q` で既存を確認する

## レシピ（ここから派生させる）
- 特定のカラム（例: リリース完了）に今週入ったもの: 今日の日付から週の始まりを計算し `pitboard task list --progress リリース完了 --progress-changed-since 2026-09-01`
- done カラム全部を今週分まとめて: `pitboard task list --state done --progress-changed-since 2026-09-01`
- 自分の担当で進行中: `pitboard task list --mine --state in_progress`
- 担当が決まっていない: `pitboard task list --state backlog`（issue も含めるなら `--kind all`）
- 詰まり候補: `pitboard task list --state in_progress --json --fields number,name,assignee,progress_changed_at` を取り、progress_changed_at が古いものを挙げる
- 詳細: `pitboard task show 42`

## 出力
- 全コマンドは `--json` で同じ形（ok / data / summary / context）。パイプ時は自動で JSON
- summary を読んでから必要なら data を見る。`--fields a,b` で data の列を絞れる
- エラー時は hint のコマンドを実行する。終了コード: 1 引数 / 2 not found / 3 認証 / 4 権限 / 6 ネットワーク / 7 サーバー
- 気に入った組み合わせは、ユーザーの Claude Code に個人コマンド（.claude/commands/）として保存してよい
