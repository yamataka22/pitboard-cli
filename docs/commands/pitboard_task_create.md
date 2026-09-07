## pitboard task create

タスクを作成する（報告者は自分、kind の既定は confirmed）

### Synopsis

タスクを作成する。作成時に決められることは1回で受け付ける。
--document FILE はファイルから本文（markdown）を読む。--document - なら標準入力から読む。

```
pitboard task create --name NAME [--assignee ID|me] [--progress ID] ... --yes [flags]
```

### Options

```
      --assignee string   メンバー ID | me
      --document string   本文（markdown）。ファイルパスか、- で標準入力
  -h, --help              help for create
      --kind string       confirmed（既定）| issue
      --label strings     ラベル ID（繰り返し指定可）
      --name string       タスク名（必須）
      --point string      h1 | h4 | d1 .. d5
      --progress string   進捗カラム ID（そのカラムに置いた状態で作る）
      --project string    プロジェクト ID
      --space string      スペース ID（省略時は既定スペース）
      --yes               書き込みを確認したことを示す（必須）
```

### Options inherited from parent commands

```
      --fields string   data に含める列をカンマ区切りで指定（JSON のみ）
      --json            JSON で出力する（パイプ時は自動）
```

### SEE ALSO

* [pitboard task](pitboard_task.md)	 - タスク（list / show / create / assign / move / point / archive）

