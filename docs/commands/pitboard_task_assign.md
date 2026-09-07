## pitboard task assign

担当者を変更する（1タスク1担当）

```
pitboard task assign NUMBER --to ID|NAME|me|none --yes [flags]
```

### Options

```
  -h, --help           help for assign
      --space string   スペース ID（省略時は既定スペース）
      --to string      メンバー ID または名前 | me | none
      --yes            書き込みを確認したことを示す（必須）
```

### Options inherited from parent commands

```
      --fields string   data に含める列をカンマ区切りで指定（JSON のみ）
      --json            JSON で出力する（パイプ時は自動）
```

### SEE ALSO

* [pitboard task](pitboard_task.md)	 - タスク（list / show / create / assign / move / point / archive）

