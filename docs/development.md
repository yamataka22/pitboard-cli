# 開発

- Go。cobra + 標準ライブラリの `net/http`。ビジネスロジックは持たない
- HTTP クライアントは `internal/client` に閉じ、コマンドはそれを呼んで `internal/output` に渡すだけ
- API の仕様は pitboard 本体側（非公開）。CLI はその薄い皮
- `--help` の文言とコマンドリファレンス（`docs/commands/`）は日本語。エラーメッセージは API 側と揃えて英語

```sh
make test        # gofmt / vet / test。API は httptest でスタブ。実サーバーは不要
make build       # ./pitboard を作る
make install     # go の bin に入れる（mise 管理の Go なら shim も作り直す）
make docs        # docs/commands/ を cobra の定義から生成し直す
```

## ローカルで `pitboard` コマンドとして使う

`make install` で `go install` されます。入る場所は `go env GOBIN`（未設定なら `~/go/bin`）です。mise で Go を入れている場合は mise の shim 経由で `pitboard` が使えます（`make install` が `mise reshim` まで行います）。コードを変えたら `make install` し直します。

ローカルの pitboard（`bin/dev` で起動した HTTPS）に向ける方法は [configuration.md](configuration.md) を参照してください。

## ドキュメント

- `README.md`: 入口。インストールと最低限の使い方はここで完結させる。コマンド1つずつの説明は `docs/commands/`
- `docs/commands/`: コマンドリファレンス。`make docs` で生成するので手で編集しない。説明を変えるときは Go 側の `Short` / `Long` / フラグの説明文を直す
- `docs/concepts.md`: pitboard の語彙
- `docs/ai.md`: AI エージェントとの使い方
- `docs/configuration.md`: 設定、トークン、開発環境への向け方
- `internal/skill/SKILL.md`: AI が実行時に読む短い指示。`go:embed` でバイナリに入る。docs へのリンクは置かない（AI はリンク先を持っていない）

## リリース

`v*` タグを push すると GitHub Actions の goreleaser が各 OS 向けのバイナリを GitHub Releases に出します。

```sh
make test
goreleaser check                       # 設定と deprecation の確認
goreleaser release --snapshot --clean  # 任意。dist/ に出るだけで push しない
sed -i 's/0\.1\.0/0.1.1/g' README.md   # インストール手順のファイル名を新しい版に
git commit -am "v0.1.1" && git tag -a v0.1.1 -m "v0.1.1" && git push origin main v0.1.1
```

README のインストール手順は Releases の実際のファイル名（`pitboard_0.1.0_darwin_arm64.tar.gz` など）を書いているので、タグを打つ前に版を更新してコミットしておく。

タグと Release の消し直しは事故りやすいので、失敗したらパッチバージョンを上げて出し直します。

配布は当面 Releases のアーカイブと `go install` だけです。Homebrew tap（`brew install --cask yamataka22/tap/pitboard`）は `.goreleaser.yaml` の `homebrew_casks` のコメントを外すと有効になりますが、その前に yamataka22/homebrew-tap を作り、そこへ commit するための PAT を Actions secrets（`HOMEBREW_TAP_TOKEN`）に登録して `release.yml` の env に渡す必要があります。goreleaser 2 系では `brews`（formula）は deprecated なので cask を使います。

## 構成

```
cmd/pitboard/main.go        エントリポイント（version は ldflags で埋める）
cmd/gendocs/main.go         docs/commands/ の生成
internal/cli/               cobra のコマンド。auth / space / task / skill / doctor / version
internal/config/            設定（環境変数 > 設定ファイル > 既定値）
internal/client/            HTTP クライアント。エンベロープとエラー（終了コード）の変換
internal/output/            JSON / テーブル出力、--fields
internal/skill/SKILL.md     skill install が書き出す本体（go:embed）
.goreleaser.yaml            リリース設定
```

## 終了コード

```
0 成功 / 1 引数エラー・--yes 無し / 2 not found / 3 認証失敗 / 4 権限不足 / 5 CLI が古い（要更新） / 6 ネットワーク / 7 サーバーエラー・レート制限
```

## API とのバージョンずれ

- CLI は `User-Agent: pitboard-cli/x.y.z` を送る。API は受け付ける CLI の下限バージョンを持っていて、それより古い CLI には `426 cli_outdated` を返す（終了コード 5、`hint` に更新方法）
- API は全レスポンスに `X-Pitboard-Api-Version` と `X-Pitboard-Min-Cli-Version` を付ける。`pitboard doctor` の `version` 項目がこれを見て、古ければ NG にする
- 開発ビルド（`dev-...`）は判定の対象外
