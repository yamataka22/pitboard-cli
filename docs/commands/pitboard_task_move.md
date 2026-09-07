## pitboard task move

進捗カラムへ移動する（none で進捗なしへ。アーカイブ済みは解除される）

```
pitboard task move NUMBER --to PROGRESS_ID|none --yes [flags]
```

### Options

```
  -h, --help           help for move
      --space string   スペース ID（省略時は既定スペース）
      --to string      進捗カラム ID | none
      --yes            書き込みを確認したことを示す（必須）
```

### Options inherited from parent commands

```
      --fields string   data に含める列をカンマ区切りで指定（JSON のみ）
      --json            JSON で出力する（パイプ時は自動）
```

### SEE ALSO

* [pitboard task](pitboard_task.md)	 - タスク（list / show / create / assign / move / point / archive）

