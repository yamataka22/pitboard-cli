# pitboard-cli

pitboard をターミナルと AI エージェント（Claude Code, Codex 等）から操作するための CLI です。

全コマンドが同じ形の JSON（`ok` / `data` / `summary` / `context`）を返すので、AI が自分で使い方を組み立てられます。

## インストール

macOS / Linux は、次の3行を貼れば `/usr/local/bin/pitboard` に入ります（`sudo` のパスワードを聞かれます）。

```sh
VERSION=$(curl -fsSL https://api.github.com/repos/yamataka22/pitboard-cli/releases/latest | sed -n 's/.*"tag_name": *"v\([^"]*\)".*/\1/p')
OS=$([ "$(uname -s)" = Darwin ] && echo darwin || echo linux); ARCH=$([ "$(uname -m)" = x86_64 ] && echo amd64 || echo arm64)
curl -fsSL "https://github.com/yamataka22/pitboard-cli/releases/download/v${VERSION}/pitboard_${VERSION}_${OS}_${ARCH}.tar.gz" | tar xz pitboard && sudo mv pitboard /usr/local/bin/
```

`pitboard version` が動けば成功です。Go が入っているなら次でも入ります。

```sh
go install github.com/yamataka22/pitboard-cli/cmd/pitboard@latest
```

ブラウザからダウンロードする手順、置き場所（PATH）の説明、macOS で「開発元を検証できません」と出たときの対処は [インストール](docs/installation.md) を参照してください。

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
pitboard space show                                          # 進捗カラム・メンバー・プロジェクトの ID と名前
pitboard task list --mine --state in_progress                # 自分の担当で進行中
pitboard task list --progress リリース完了 --progress-changed-since 2026-09-01   # 9/1 以降にそのカラムに入ったもの（アーカイブ済みも含む）
pitboard task show 42
pitboard task create --name "週報を書く" --assignee me --progress 作業中 --yes   # ID の代わりに名前でもよい
pitboard task move 42 --to リリース待ち --yes
pitboard comment add 42 --body "レビューお願いします" --mention 鈴木 --yes   # メンション付きコメント
pitboard task list --json --fields number,name,state         # JSON（パイプ時は自動）
```

AI には自然言語で頼みます。

```
自分が担当のタスクのうち、8/31〜9/4 にリリース完了になったものの番号とタイトルを教えて
```

## ドキュメント

- [インストール](docs/installation.md): OS ごとの入れ方、置き場所（PATH）、macOS の警告への対処、更新と削除
- [コマンドリファレンス](docs/commands/README.md): 全コマンドとオプション（`--help` と同じ内容）
- [pitboard の語彙](docs/concepts.md): state、progress_changed_at、ポイント、kind の意味
- [AI エージェントと使う](docs/ai.md): 頼み方、個人コマンド、cron からの定期実行、書き込みの安全装置
- [設定](docs/configuration.md): 設定ファイル、環境変数、トークン、開発環境への向け方、API を直接使うことについて
- [開発](docs/development.md): ビルド、テスト、リリース

## 設計の要点

- **タスクを直接 update する口はありません。** 更新は「担当を変える」「カラムを動かす」のような意図ごとのコマンドだけで、Web の操作と同じ処理を通ります。1タスク1担当、期限日なし、といった制約は API 側で強制されます
- **書き込みは `--yes` が必須**です。AI が確認なしに書き込むのを防ぎます
- **日付は AI（またはあなた）が計算して ISO 形式で渡します。** 「今週」「7日前」のような相対指定は CLI にはありません
- pitboard が公式に対応するのはこの CLI だけです。HTTP API を直接使うのは自己責任です
