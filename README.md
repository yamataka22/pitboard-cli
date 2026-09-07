# pitboard-cli

pitboard をターミナルと AI エージェント（Claude Code, Codex 等）から操作するための CLI です。

全コマンドが同じ形の JSON（`ok` / `data` / `summary` / `context`）を返すので、AI が自分で使い方を組み立てられます。

## インストール

### macOS / Linux (WSL)

環境に合うものをターミナルで実行すると `/usr/local/bin` に入ります。

macOS（Apple Silicon）

```sh
curl -fsSL https://github.com/yamataka22/pitboard-cli/releases/download/v0.1.1/pitboard_0.1.1_darwin_arm64.tar.gz | tar xz pitboard && sudo mv pitboard /usr/local/bin/
```

macOS（Intel）

```sh
curl -fsSL https://github.com/yamataka22/pitboard-cli/releases/download/v0.1.1/pitboard_0.1.1_darwin_amd64.tar.gz | tar xz pitboard && sudo mv pitboard /usr/local/bin/
```

Linux（x86_64） WSLの場合もこちら

```sh
curl -fsSL https://github.com/yamataka22/pitboard-cli/releases/download/v0.1.1/pitboard_0.1.1_linux_amd64.tar.gz | tar xz pitboard && sudo mv pitboard /usr/local/bin/
```

Linux（arm64）

```sh
curl -fsSL https://github.com/yamataka22/pitboard-cli/releases/download/v0.1.1/pitboard_0.1.1_linux_arm64.tar.gz | tar xz pitboard && sudo mv pitboard /usr/local/bin/
```

[Releases](https://github.com/yamataka22/pitboard-cli/releases) からブラウザでダウンロードした場合は、展開した `pitboard` を `/usr/local/bin` に置きます。macOS では「開発元を検証できません」と出るので、次を実行してください。

```sh
xattr -d com.apple.quarantine /usr/local/bin/pitboard
```

`pitboard version` が動けば完了です。更新は同じ手順で上書きします。

### go install

Go が入っていれば、Releases を使わずにこれで入ります。

```sh
go install github.com/yamataka22/pitboard-cli/cmd/pitboard@latest
```

### Windows

WSL で Claude Code を使っている場合、WSL の中は Linux なので、WSL のターミナルで次を実行します。

```sh
curl -fsSL https://github.com/yamataka22/pitboard-cli/releases/download/v0.1.1/pitboard_0.1.1_linux_amd64.tar.gz | tar xz pitboard && sudo mv pitboard /usr/local/bin/
```

なお、WSL ではなく Windows で直接使うための zip も Releases に置いていますが、動作確認はしていません。

## セットアップ

1. pitboard の Web で **プロフィール > アクセストークン** を開き、トークンを発行する（`pb_` で始まる。一度だけ表示）。AI に渡すなら「読み取り専用」で十分です（[トークンについて](docs/configuration.md#トークンについて)）
2. ログインする。所属スペースが1つならそれが既定になります（[スペースの決まり方](docs/configuration.md#スペースの決まり方)）

   ```sh
   pitboard auth login          # プロンプトでトークンを貼り付ける
   pitboard space list          # 複数あるなら pitboard space use ID
   ```

3. AI エージェントに登録する

   ```sh
   pitboard skill install       # Claude Code（~/.claude/skills/pitboard/SKILL.md）
   ```

4. `pitboard doctor` で設定・認証・接続・スキルを確認する

## AI から使う

`pitboard skill install` は `~/.claude/skills/pitboard/SKILL.md` を書きます。ユーザー全体の設定なので、Claude Code をどのディレクトリで起動しても使えます。Codex は `pitboard skill install --target codex` が表示する1行を `AGENTS.md` に足します。

あとは自然言語で頼みます。Claude Code は SKILL.md を読んで `pitboard` コマンドを組み立て、返ってきた JSON から答えます。

```
自分が担当のタスクのうち、8/31〜9/4 にリリース完了になったものの番号とタイトルを教えて

進行中で1週間以上動いていないタスクは？

担当が決まっていないタスクを一覧して

#42 の内容を要約して
```

書き込みも頼めます。作成、担当変更、進捗カラムの移動、ポイント、アーカイブ、コメントです。

```
「週報を書く」を自分の担当で着手中に作って

#42 を鈴木さんの担当にして、リリース待ちに動かして

#42 に「レビューお願いします」と鈴木さん宛にコメントして
```

最初の依頼で AI が実行するのは、たとえばこうです。カラム名やメンバー名は `space show` で確認してから使います。

```sh
pitboard task list --mine --progress リリース完了 --progress-changed-since 2026-08-31 --progress-changed-until 2026-09-04 --json --fields number,name
```

- AI が pitboard を使わないときは「pitboard で」と添えるか、`/pitboard` とスキル名を明示します
- 「完了」に当たるカラムが複数あるスペースでは、カラム名を指定して頼むと確実です（[進捗カラムと state](docs/concepts.md#進捗カラムと-state)）
- 書き込みコマンドは `--yes` が無いと実行されません。SKILL.md は AI に「書く前に意図を確認する」「二重作成を避ける」と指示しています。読み取り専用トークンを渡しておけば、書き込みは API 側で拒否されます

定型の依頼を[個人コマンドにする方法](docs/ai.md#個人コマンドとして保存する)と、[cron から定期実行する方法](docs/ai.md#定期的に動かす)は別にまとめています。

## コマンドを直接使う

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

日付は ISO 8601（`2026-09-01`）で渡します。`7d` や `thisweek` のような相対指定はありません。全コマンドとオプションは[コマンドリファレンス](docs/commands/README.md)に、`state` や `progress_changed_at` の意味は [pitboard の語彙](docs/concepts.md)にあります。

## ドキュメント

- [コマンドリファレンス](docs/commands/README.md): 全コマンドとオプション（`--help` と同じ内容）
- [pitboard の語彙](docs/concepts.md): state、progress_changed_at、ポイント、kind の意味
- [AI エージェントと使う](docs/ai.md): 頼み方、個人コマンド、cron からの定期実行、書き込みの安全装置
- [設定](docs/configuration.md): 設定ファイル、環境変数、トークン、開発環境への向け方、API を直接使うことについて
- [開発](docs/development.md): ビルド、テスト、リリース

## 設計の要点

- **タスクを直接 update する口はありません。** 更新は「担当を変える」「カラムを動かす」のような意図ごとのコマンドだけで、Web の操作と同じ処理を通ります。1タスク1担当、期限日なし、といった制約は API 側で強制されます
- **書き込みは `--yes` が必須**です。AI が確認なしに書き込むのを防ぎます
- **日付は AI（またはあなた）が計算して ISO 形式で渡します。** 「今週」「7日前」のような相対指定は CLI にはありません
- pitboard が公式に対応するのはこの CLI だけです。HTTP API を直接使うのは自己責任です（[API を直接使うことについて](docs/configuration.md#api-を直接使うことについて)）
