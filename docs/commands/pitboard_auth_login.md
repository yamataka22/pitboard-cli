## pitboard auth login

アクセストークンを保存する（TOKEN 省略時はプロンプトで入力）

### Synopsis

Web の「プロフィール > アクセストークン」でトークンを発行してから実行する。
所属スペースが1つならそれが既定スペースになる。
--api-url を付けるとローカルやセルフホストの pitboard に向けられる（設定ファイルに保存）。

```
pitboard auth login [TOKEN] [flags]
```

### Options

```
      --api-url string   設定ファイルに保存する API の URL（例: https://localhost:3000/api/v1）
  -h, --help             help for login
```

### Options inherited from parent commands

```
      --fields string   data に含める列をカンマ区切りで指定（JSON のみ）
      --json            JSON で出力する（パイプ時は自動）
```

### SEE ALSO

* [pitboard auth](pitboard_auth.md)	 - 認証（login / status / logout）

