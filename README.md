# pitboard-cli

pitboard をターミナルと AI エージェント（Claude Code, Codex 等）から操作するための CLI です。

## インストール

### macOS / Linux

環境に合うものをターミナルで実行すると `/usr/local/bin` に入ります。

macOS（Apple Silicon）

```sh
curl -fsSL https://github.com/yamataka22/pitboard-cli/releases/download/v0.1.1/pitboard_0.1.1_darwin_arm64.tar.gz | tar xz pitboard && sudo mkdir -p /usr/local/bin && sudo mv pitboard /usr/local/bin/
```

Intel Mac の場合は、URL の `darwin_arm64` を `darwin_amd64` に置き換えてください。

Linux（x86_64）

```sh
curl -fsSL https://github.com/yamataka22/pitboard-cli/releases/download/v0.1.1/pitboard_0.1.1_linux_amd64.tar.gz | tar xz pitboard && sudo mkdir -p /usr/local/bin && sudo mv pitboard /usr/local/bin/
```

Linux（arm64）

```sh
curl -fsSL https://github.com/yamataka22/pitboard-cli/releases/download/v0.1.1/pitboard_0.1.1_linux_arm64.tar.gz | tar xz pitboard && sudo mkdir -p /usr/local/bin && sudo mv pitboard /usr/local/bin/
```

### Windows（WSL）

WSL で Claude Code を使っている場合、WSL の中は Linux なので、WSL のターミナルで次を実行します。

```sh
curl -fsSL https://github.com/yamataka22/pitboard-cli/releases/download/v0.1.1/pitboard_0.1.1_linux_amd64.tar.gz | tar xz pitboard && sudo mkdir -p /usr/local/bin && sudo mv pitboard /usr/local/bin/
```

### go install

Go が入っていれば、Releases を使わずにこれで入ります。

```sh
go install github.com/yamataka22/pitboard-cli/cmd/pitboard@latest
```

## インストールの確認

```sh
pitboard version
# pitboard-cli 0.1.1 などのバージョンが表示されたら正常にインストールされています

pitboard doctor
# 設定・トークン・API 接続・既定スペース・スキルをまとめて診断します
```

`doctor` はこの時点では `token` と `skill` が NG になります（まだログインもスキルの登録もしていないため）。次の手順を終えてからもう一度実行すると、すべて OK になります。あとで「動かない」ときも、まずこれを実行すると原因が分かります。

## セットアップ

1. pitboard の Web で **プロフィール > アクセストークン** を開き、トークンを発行する（`pb_` で始まる。一度だけ表示）。[トークンについて](docs/configuration.md#トークンについて)
2. ログインする。所属スペースが1つならそれが既定になります。

   ```sh
   pitboard auth login
   # プロンプトでトークンを貼り付ける

   pitboard space list

   # スペースが複数ある場合、利用するスペースを指定する
   pitboard space use <ID>
   ```

## AI（Claude Code）から使う

AI に pitboard-cli の利用スキルを登録することで、以下のような自然言語での利用ができます。

```
自分が担当のタスクのうち、8/31〜9/4 にリリース完了になったものの番号とタイトルを教えて

進行中で1週間以上動いていないタスクは？

#42 の内容を要約して
```

### スキルのインストール

```sh
pitboard skill install       # Claude Code（~/.claude/skills/pitboard/SKILL.md）
```

`pitboard skill install` は `~/.claude/skills/pitboard/SKILL.md` を書きます。ユーザー全体の設定なので、Claude Code をどのディレクトリで起動しても使えます。Codex は `pitboard skill install --target codex` が表示する1行を `AGENTS.md` に足します。

### 使い方

任意のディレクトリで `claude` を起動し、あとは自然言語で依頼します。AI が pitboard を使わないときは「pitboard で」と添えるか、`/pitboard` とスキル名を明示します。

```
pitboard で自分が担当のタスクのうち、8/31〜9/4 にリリース完了になったものの番号とタイトルを教えて

/pitboard 担当が決まっていないタスクを一覧して
```

書き込みも頼めます。作成、担当変更、進捗カラムの移動、ポイント、アーカイブ、コメントです。

```
pitboard #42 を鈴木さんの担当にして、リリース待ちに動かして

pitboard #42 に「確認してください」と鈴木さん宛にコメントして
```

定型の依頼を[個人コマンドにする方法](docs/ai.md#個人コマンドとして保存する)と、[cron から定期実行する方法](docs/ai.md#定期的に動かす)は別にまとめています。

## コマンドを直接使う

上記のように AI のスキルで利用するのではなく、CLI コマンドを直接利用することも可能です。

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

全コマンドとオプションは[コマンドリファレンス](docs/commands/README.md)に、`state` や `progress_changed_at` の意味は [pitboard の語彙](docs/concepts.md)にあります。

## ドキュメント

- [コマンドリファレンス](docs/commands/README.md): 全コマンドとオプション（`--help` と同じ内容）
- [pitboard の語彙](docs/concepts.md): state、progress_changed_at、ポイント、kind の意味
- [AI エージェントと使う](docs/ai.md): 頼み方、個人コマンド、cron からの定期実行、書き込みの安全装置
- [設定](docs/configuration.md): 設定ファイル、環境変数、トークン、macOS の Gatekeeper、開発環境への向け方、API を直接使うことについて
- [開発](docs/development.md): ビルド、テスト、リリース

## 設計の要点

- **タスクを直接 update する口はありません。** 更新は「担当を変える」「カラムを動かす」のような意図ごとのコマンドだけで、Web の操作と同じ処理を通ります。1タスク1担当、期限日なし、といった制約は API 側で強制されます
- **書き込みは `--yes` が必須**です。AI が確認なしに書き込むのを防ぎます
- **日付は AI（またはあなた）が計算して ISO 形式で渡します。** 「今週」「7日前」のような相対指定は CLI にはありません
- pitboard が公式に対応するのはこの CLI だけです。HTTP API を直接使うのは自己責任です（[API を直接使うことについて](docs/configuration.md#api-を直接使うことについて)）
