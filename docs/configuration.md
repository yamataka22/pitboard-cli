# 設定

## macOS でブラウザからダウンロードしたとき

README の `curl` の手順で入れたなら、この節は関係ありません。

[Releases](https://github.com/yamataka22/pitboard-cli/releases) のページからブラウザで tar.gz を落とした場合、macOS がファイルに `com.apple.quarantine` を付けます。pitboard-cli は署名・公証をしていないので、そのまま実行すると「開発元を検証できません」と拒否されます。展開した `pitboard` を `/usr/local/bin` に置いたうえで、そのファイルの属性を外してください。

```sh
xattr -d com.apple.quarantine /usr/local/bin/pitboard
```

`pitboard version` が動けば完了です。`curl` でダウンロードしたファイルにはこの属性が付かないので、この操作は要りません。

## 設定ファイルと環境変数

| 項目 | 環境変数（優先） | 設定ファイル |
|---|---|---|
| トークン | `PITBOARD_TOKEN` | `token` |
| API の URL | `PITBOARD_API_URL` | `api_url`（既定: `https://app.pitboard.dev/api/v1`） |
| 既定スペース | `PITBOARD_SPACE_ID` | `space_id` |
| CA ファイル | `PITBOARD_CA_FILE` | `ca_file` |
| 証明書検証の無効化 | `PITBOARD_INSECURE=1` | （なし。ローカル開発専用） |

設定ファイルは `$XDG_CONFIG_HOME/pitboard/config.yml`、無ければ `~/.config/pitboard/config.yml` です（macOS も同じ。Windows は `%AppData%\\pitboard\\config.yml`）。パーミッション `600` で作成されます。
`pitboard auth login` と `pitboard space use` が書き込み、`pitboard auth logout` がトークンと既定スペースを消します。

環境変数は設定ファイルより優先されるので、「このシェルだけ別の環境に向ける」ときに使います。`pitboard auth status` と `pitboard doctor` は、それぞれの値がどこから来たか（env / file）を表示します。

## スペースの決まり方

スペースを必要とするコマンドは、次の順でスペースを決めます。

1. `--space ID` オプション
2. 環境変数 `PITBOARD_SPACE_ID`
3. 設定ファイルの `space_id`（`pitboard space use` で保存したもの）
4. どれも無ければ API に所属スペースを問い合わせ、1つだけならそれを使う

所属スペースが複数あるのにどれも指定されていない場合は `space_required` エラー（終了コード 1）になります。

`space use ID` は保存する前に所属を確認します。所属していないスペース（存在しない ID を含む）は `not_found`（終了コード 2）で、保存されません。存在するかどうかは区別しません。保存後にスペースから退出した場合も同じエラーになるので、`space use` をやり直してください。

## トークンについて

- トークンは Web の「プロフィール > アクセストークン」で発行します。発行時に一度だけ表示されます
- ユーザー単位で、所属している全スペースにアクセスできます
- 発行時に権限を選びます。**読み取り専用**（既定）か**読み書き**か。AI エージェントに渡すトークンは読み取り専用にしておくと、書き込みコマンドは `403 forbidden` になります
- 同じ画面で失効できます。`pitboard auth logout` はローカルの設定を消すだけで、トークン自体は無効になりません
- 有効期限は既定で無期限です。発行時に 30 / 90 / 365 日を選べます
- トークンは発行したサーバーでしか使えません。ローカルの pitboard で発行したトークンは本番では使えず、逆も同じです

## 開発環境（HTTPS）に向ける

pitboard の開発環境は mkcert の証明書で HTTPS になっています。mkcert のルート CA をシステムに入れてあれば（`mkcert -install` 済み）追加設定は要りません。

```sh
pitboard auth login --api-url https://localhost:3000/api/v1   # api_url を設定ファイルに保存してログイン
pitboard doctor
```

ルート CA を信頼していない環境では CA ファイルを指定します。

```sh
export PITBOARD_CA_FILE=~/.local/share/mkcert/rootCA.pem
```

`PITBOARD_INSECURE=1` で証明書検証を無効にもできますが、ローカル開発以外では使わないでください。

## API を直接使うことについて

pitboard が公式に対応するのはこの CLI だけです。CLI は pitboard の HTTP API を叩いていますが、API 自体は公開仕様ではなく、予告なく変わることがあります。
`--json` の出力を見て API を直接呼ぶことは止めませんが、**自己責任**でお願いします。動かなくなっても対応はしません。

古い CLI は API に拒否されます（`cli_outdated`、終了コード 5）。`pitboard doctor` でバージョンのずれを確認できます。

API にはトークン単位のレート制限（1分あたり 120 回）があります。超えると `rate_limited` エラーになるので、大きなスペースで `--all` を繰り返すより、フィルタで絞ってください。
