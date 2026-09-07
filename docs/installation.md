# インストール

pitboard は単一のバイナリです。ダウンロードして、PATH の通った場所に置くだけで動きます。Ruby や Node は要りません。

配布しているのは以下です。Homebrew や Scoop にはまだ対応していません。

- GitHub Releases のアーカイブ（macOS / Linux / Windows）
- `go install`（Go が入っている場合）

## PATH の通った場所とは

`PATH` は、シェルがコマンドを探すディレクトリの一覧です。`pitboard` と打つと、この一覧を上から順に見て、最初に見つかった実行ファイルが起動します。「PATH の通った場所に置く」とは、この一覧に入っているディレクトリに置く、という意味です。

```sh
echo $PATH        # : 区切りで一覧が出る
```

macOS / Linux なら **`/usr/local/bin`** を使ってください。どの環境でも PATH に入っていて、確実です（置くときに `sudo` が要ります）。

`sudo` を使いたくない場合は `~/.local/bin` などでも構いませんが、そのディレクトリが `echo $PATH` に出てこない場合は、`~/.zshrc`（bash なら `~/.bashrc`）に次を足してターミナルを開き直す必要があります。

```sh
export PATH="$HOME/.local/bin:$PATH"
```

## macOS

### どのファイルを落とすか

| Mac | ファイル |
| --- | --- |
| Apple Silicon（M1 以降） | `pitboard_<版>_darwin_arm64.tar.gz` |
| Intel Mac | `pitboard_<版>_darwin_amd64.tar.gz` |

どちらか分からないときは、ターミナルで `uname -m` を実行します。`arm64` と出れば Apple Silicon、`x86_64` と出れば Intel です（アップルメニュー > このMacについて の「チップ」でも分かります）。

### ターミナルで入れる（推奨）

[README](../README.md) の3行を貼るのが一番簡単です。この方法なら、次に説明する Gatekeeper の警告も出ません。

### ブラウザからダウンロードして入れる

1. [Releases](https://github.com/yamataka22/pitboard-cli/releases) の最新版から、上の表のファイルを選んでダウンロードする
2. Finder でダブルクリックして展開する（`pitboard` という名前のファイルが出てきます）
3. ターミナルで `/usr/local/bin` に移動する

   ```sh
   sudo mv ~/Downloads/pitboard /usr/local/bin/
   ```

4. **ダウンロードしたファイルには macOS が隔離の印を付けるので、それを外す**

   ```sh
   xattr -d com.apple.quarantine /usr/local/bin/pitboard
   ```

   これをしないと、実行時に「"pitboard"は開発元を検証できないため開けません」と出ます（アプリに署名していないためです）。既に出てしまった場合も、このコマンドで解決します。

5. 確認する

   ```sh
   pitboard version
   ```

## Linux

CPU が x86_64 なら `linux_amd64`、arm64 なら `linux_arm64` です（`uname -m` で確認できます）。Gatekeeper のような仕組みは無いので、展開して置くだけです。

```sh
tar xzf pitboard_<版>_linux_amd64.tar.gz
sudo mv pitboard /usr/local/bin/
pitboard version
```

## Windows

zip は Releases に置いていますが、動作確認をしていません。展開した `pitboard.exe` を PATH の通ったフォルダに置いてください。

## go install

Go が入っていれば、これだけで PATH（`go env GOBIN`、未設定なら `~/go/bin`）に入ります。

```sh
go install github.com/yamataka22/pitboard-cli/cmd/pitboard@latest
```

## ダウンロードしたファイルの検証（任意）

Releases の `checksums.txt` と照合できます。

```sh
shasum -a 256 pitboard_<版>_darwin_arm64.tar.gz   # Linux なら sha256sum
grep darwin_arm64 checksums.txt                    # 値が一致すること
```

## 更新と削除

更新は、同じ手順で新しいバイナリを上書きするだけです（`go install` なら `@latest` を再実行）。CLI が古くなって API に拒否された場合は `pitboard doctor` が案内します。

削除は次のとおりです。設定ファイル（トークンを含む）は `~/.config/pitboard/config.yml` にあります。

```sh
sudo rm /usr/local/bin/pitboard
rm -rf ~/.config/pitboard
```

## 次に

[README のセットアップ](../README.md#セットアップ)に進み、アクセストークンでログインしてください。
