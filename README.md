# pitboard-cli

pitboard をターミナルと AI エージェント（Claude Code, Codex 等）から操作するための CLI です。単一バイナリで、Ruby や Node は要りません。

全コマンドが同じ形の JSON（`ok` / `data` / `summary` / `context`）を返すので、AI が自分の使い方を組み立てられます。pitboard 側に「今週完了したタスクをまとめる」のような機能はありません。材料だけを提供し、組み合わせはユーザーと AI に任せています。

> 状態: 読み取り（list / show / space）と書き込み（create / assign / move / point / archive / comment add）が実装済み。

## インストール

```sh
# macOS / Linux（Homebrew。公開後に有効）
brew install yamataka22/tap/pitboard

# Go が入っていれば
go install github.com/yamataka22/pitboard-cli/cmd/pitboard@latest

# それ以外は GitHub Releases から OS に合ったアーカイブを取り、pitboard を PATH の通った場所に置く
```

## セットアップ

1. pitboard の Web で **プロフィール > アクセストークン** を開き、トークンを発行する（`pb_` で始まる。一度だけ表示）。AI に渡すなら「読み取り専用」で十分です
2. ログインする。所属スペースが1つならそれが既定になります

   ```sh
   pitboard auth login          # プロンプトでトークンを貼り付ける
   pitboard space list          # 複数あるなら pitboard space use ID
   ```

3. AI エージェントに登録する

   ```sh
   pitboard skill install       # Claude Code（~/.claude/skills/pitboard/SKILL.md）
   ```

4. `pitboard doctor` で設定・認証・接続・スキルを確認する

## 使い方

```sh
pitboard space show                                          # 進捗カラム・メンバー・プロジェクトの ID
pitboard task list --mine --state in_progress                # 自分の担当で進行中
pitboard task list --state done --progress-changed-since 2026-09-01   # 9/1 以降に完了したもの（アーカイブ済みも含む）
pitboard task show 42
pitboard task create --name "週報を書く" --assignee me --progress 7 --yes
pitboard task move 42 --to 9 --yes
pitboard comment add 42 --body "レビューお願いします" --mention 13 --yes   # メンション付きコメント
pitboard task list --json --fields number,name,state         # JSON（パイプ時は自動）
```

AI には自然言語で頼みます。

```
自分が担当のタスクのうち、8/31〜9/4 に完了になったものの番号とタイトルを教えて
```

## ドキュメント

- [コマンドリファレンス](docs/commands/README.md): 全コマンドとオプション（`--help` と同じ内容）
- [pitboard の語彙](docs/concepts.md): state、progress_changed_at、ポイント、kind の意味
- [AI エージェントと使う](docs/ai.md): 頼み方、個人コマンド、cron からの定期実行、書き込みの安全装置
- [設定](docs/configuration.md): 設定ファイル、環境変数、トークン、開発環境への向け方、API を直接使うことについて
- [開発](docs/development.md): ビルド、テスト、リリース

## 設計の要点

- **タスクを直接 update する口はありません。** 更新は「担当を変える」「カラムを動かす」のような意図ごとのコマンドだけで、Web の操作と同じ処理を通ります。1タスク1担当、期限日なし、といった制約は API 側で強制されます
- **書き込みは `--yes` が必須**です。AI が確認なしに書き込むのを防ぎます
- **「完了とは何か」は pitboard が決め、「今週とは何か」は決めません。** 日付は AI（またはあなた）が計算して ISO 形式で渡します
- pitboard が公式に対応するのはこの CLI だけです。HTTP API を直接使うのは自己責任です
