# pitboard-cli

pitboard を AI エージェント（Claude Code, Codex 等）とターミナルから操作するための CLI です。単一のバイナリで、Ruby や Node は要りません。

ほとんどの人は「AI に pitboard を触らせるため」に入れます。その場合、覚えるコマンドは最初のセットアップの3つだけで、あとは日本語で頼みます。

```
自分が担当のタスクのうち、8/31〜9/4 にリリース完了になったものの番号とタイトルを教えて
```

全コマンドが同じ形の JSON（`ok` / `data` / `summary` / `context`）を返すので、AI が自分で使い方を組み立てられます。

## インストール

配布しているのは GitHub Releases のアーカイブと `go install` です。Homebrew や Scoop にはまだ対応していません。

### macOS / Linux（ターミナルで入れる）

次の3行を貼れば `/usr/local/bin/pitboard` に入ります（`sudo` のパスワードを聞かれます）。この方法なら macOS の警告も出ません。

```sh
VERSION=$(curl -fsSL https://api.github.com/repos/yamataka22/pitboard-cli/releases/latest | sed -n 's/.*"tag_name": *"v\([^"]*\)".*/\1/p')
OS=$([ "$(uname -s)" = Darwin ] && echo darwin || echo linux); ARCH=$([ "$(uname -m)" = x86_64 ] && echo amd64 || echo arm64)
curl -fsSL "https://github.com/yamataka22/pitboard-cli/releases/download/v${VERSION}/pitboard_${VERSION}_${OS}_${ARCH}.tar.gz" | tar xz pitboard && sudo mv pitboard /usr/local/bin/
```

`pitboard version` が動けば成功です。

### macOS（ブラウザからダウンロードする）

[Releases](https://github.com/yamataka22/pitboard-cli/releases) の最新版から、お使いの Mac に合うファイルを選びます。

| Mac | ファイル |
| --- | --- |
| Apple Silicon（M1 以降） | `pitboard_<版>_darwin_arm64.tar.gz` |
| Intel Mac | `pitboard_<版>_darwin_amd64.tar.gz` |

どちらか分からないときは、アップルメニュー > このMacについて の「チップ」を見るか、ターミナルで `uname -m` を実行します（`arm64` なら Apple Silicon、`x86_64` なら Intel）。

1. ダウンロードしたファイルを Finder でダブルクリックして展開する（`pitboard` という名前のファイルが出てきます）
2. ターミナルで `/usr/local/bin` に移動する

   ```sh
   sudo mv ~/Downloads/pitboard /usr/local/bin/
   ```

3. **macOS が付ける隔離の印を外す**

   ```sh
   xattr -d com.apple.quarantine /usr/local/bin/pitboard
   ```

   これをしないと「"pitboard"は開発元を検証できないため開けません」と出ます（バイナリに署名していないためです）。既に出てしまった場合も、このコマンドで解決します。

4. `pitboard version` で確認する

### Linux（ブラウザからダウンロードする）

CPU が x86_64 なら `linux_amd64`、arm64 なら `linux_arm64` です（`uname -m` で確認できます）。展開して置くだけで、macOS のような警告はありません。

```sh
tar xzf pitboard_<版>_linux_amd64.tar.gz
sudo mv pitboard /usr/local/bin/
pitboard version
```

### Windows

zip は Releases に置いていますが、動作確認をしていません。展開した `pitboard.exe` を PATH の通ったフォルダに置いてください。

### go install

Go が入っていれば、これだけで `go env GOBIN`（未設定なら `~/go/bin`）に入ります。

```sh
go install github.com/yamataka22/pitboard-cli/cmd/pitboard@latest
```

### 置き場所（PATH）について

`PATH` は、シェルがコマンドを探すディレクトリの一覧です（`echo $PATH` で見られます）。`pitboard` と打つと、この一覧を上から順に見て、最初に見つかった実行ファイルが起動します。上の手順で `/usr/local/bin` を使っているのは、どの環境でも PATH に入っていて確実だからです。

`sudo` を使いたくない場合は `~/.local/bin` などでも構いませんが、そのディレクトリが `echo $PATH` に出てこないときは、`~/.zshrc`（bash なら `~/.bashrc`）に次を足してターミナルを開き直してください。

```sh
export PATH="$HOME/.local/bin:$PATH"
```

### 更新と削除

更新は、同じ手順で新しいバイナリを上書きするだけです（`go install` なら `@latest` を再実行）。CLI が古くて API に拒否された場合は `pitboard doctor` が案内します。

```sh
sudo rm /usr/local/bin/pitboard     # 削除
rm -rf ~/.config/pitboard           # 設定ごと消す（トークンもここにあります）
```

ダウンロードしたファイルは、Releases の `checksums.txt` と照合できます。

```sh
shasum -a 256 pitboard_<版>_darwin_arm64.tar.gz   # Linux なら sha256sum
grep darwin_arm64 checksums.txt                    # 値が一致すること
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

## AI から使う

`pitboard skill install` は `~/.claude/skills/pitboard/SKILL.md` を書きます。ユーザー全体の設定なので、Claude Code をどのディレクトリで起動しても使えます。あとは自然言語で頼むと、Claude Code がそれを読んで `pitboard` コマンドを組み立て、返ってきた JSON から答えます。

Codex の場合は `pitboard skill install --target codex` が表示する1行を `AGENTS.md` に足します。

### 頼めること

読み取りは一覧・検索・詳細です。

```
自分が担当のタスクのうち、8/31〜9/4 にリリース完了になったものの番号とタイトルを教えて
進行中で1週間以上動いていないタスクは？
担当が決まっていないタスクを一覧して
#42 の内容を要約して
```

1つ目なら、AI はまず `pitboard space show` でそのスペースのカラム名を確認してから、こう組み立てます。

```sh
pitboard task list --mine --progress リリース完了 \
  --progress-changed-since 2026-08-31 --progress-changed-until 2026-09-04 \
  --json --fields number,name
```

書き込みは作成・担当変更・カラム移動・ポイント・アーカイブ・コメントです。

```
「週報を書く」を自分の担当で着手中に作って
#42 を鈴木さんの担当にして、リリース待ちに動かして
#42 に「レビューお願いします」と鈴木さん宛にコメントして
```

進捗カラム・プロジェクト・ラベル・メンバーは、ID でも名前でも指定できます（同名が複数あるときだけ ID が要ります）。

### うまく動かないとき

- **AI が pitboard を使ってくれない**: 依頼文に「タスク」や「pitboard」が入っていれば拾いますが、拾わないときは「pitboard で」と一言添えるか、`/pitboard` とスキル名を明示してください
- **「完了」の解釈がずれる**: done 扱いのカラムが複数あるスペースでは、AI がどのカラムか確認するか、`--state done` でまとめて取ります。カラム名を指定して頼むのが確実です
- **設定や接続を疑うとき**: `pitboard doctor` が設定・認証・接続・スキルをまとめて確認します

### 書き込みの安全装置

- 書き込みコマンド（create / assign / move / point / archive / unarchive / comment add）は `--yes` が無いと実行されません。AI が確認なしに書き込むのを防ぐためです
- SKILL.md は AI に「書く前にユーザーの意図を確認する」「二重作成を避けるため既存を確認する」と指示しています
- 読み取り専用トークンを渡しておけば、書き込みは API 側で拒否されます

同じ依頼の定型化や、cron から定期実行する方法は [AI エージェントと使う](docs/ai.md) を参照してください。

## 自分でコマンドを打つ

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

日付は ISO 8601（`2026-09-01`）で渡します。`7d` や `thisweek` のような相対指定はありません。全コマンドとオプションは `--help` か [コマンドリファレンス](docs/commands/README.md) にあります。

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
- pitboard が公式に対応するのはこの CLI だけです。HTTP API を直接使うのは自己責任です
