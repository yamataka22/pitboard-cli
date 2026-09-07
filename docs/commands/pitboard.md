## pitboard

人と AI エージェントのための pitboard CLI

### Synopsis

pitboard API の薄い皮としての CLI。全コマンドが同じ形の JSON（ok / data / summary / context）を返すので、AI エージェントが自分の使い方を組み立てられる。

### Options

```
      --fields string   data に含める列をカンマ区切りで指定（JSON のみ）
  -h, --help            help for pitboard
      --json            JSON で出力する（パイプ時は自動）
```

### SEE ALSO

* [pitboard auth](pitboard_auth.md)	 - 認証（login / status / logout）
* [pitboard comment](pitboard_comment.md)	 - コメント（add）
* [pitboard doctor](pitboard_doctor.md)	 - 設定・トークン・API 接続・既定スペース・スキル登録を診断する
* [pitboard skill](pitboard_skill.md)	 - AI エージェントへのスキル登録（install / show）
* [pitboard space](pitboard_space.md)	 - スペース（list / use / show）
* [pitboard task](pitboard_task.md)	 - タスク（list / show / create / assign / move / point / archive）
* [pitboard version](pitboard_version.md)	 - バージョンを表示する

