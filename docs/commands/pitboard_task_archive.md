## pitboard task archive

タスクをアーカイブする（--undo で解除）

```
pitboard task archive NUMBER [--undo] --yes [flags]
```

### Options

```
  -h, --help           help for archive
      --space string   スペース ID（省略時は既定スペース）
      --undo           アーカイブを解除する
      --yes            書き込みを確認したことを示す（必須）
```

### Options inherited from parent commands

```
      --fields string   data に含める列をカンマ区切りで指定（JSON のみ）
      --json            JSON で出力する（パイプ時は自動）
```

### SEE ALSO

* [pitboard task](pitboard_task.md)	 - タスク（list / show / create / assign / move / point / archive）

