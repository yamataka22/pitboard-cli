## pitboard task list

タスクの一覧（既定: アーカイブ除外、番号の降順）

### Synopsis

スペースのタスクを一覧する。日付は ISO 8601（2026-09-01）で渡す。"7d" や "thisweek" のような相対指定は無い。
--progress-changed-since / --until を付けるとアーカイブ済みも含まれる（除くなら --archived false）。

```
pitboard task list [flags]
```

### Options

```
      --all                             ページングを最後まで辿る
      --archived string                 false | true | all（既定 false。--progress-changed-* 指定時は all）
      --assignee string                 メンバー ID | me | none
  -h, --help                            help for list
      --kind string                     confirmed | issue（省略時は両方）
      --label strings                   ラベル ID（繰り返し指定で AND）
      --mine                            --assignee me の別名
      --order string                    asc | desc（既定は --sort による）
      --owner string                    メンバー ID | me
      --page int                        ページ番号
      --per-page int                    1ページの件数（最大 200）
      --point string                    h1 | h4 | d1 .. d5 | none
      --progress string                 進捗カラム ID
      --progress-changed-since string   今のカラムに置かれた日時がこれ以降（ISO 8601）
      --progress-changed-until string   今のカラムに置かれた日時がこれ以前（ISO 8601。日付だけならその日の終わりまで）
      --project string                  プロジェクト ID | none
  -q, --query string                    番号・タスク名・本文を検索
      --sort string                     code（既定）| updated | created | progress_changed | position
      --space string                    スペース ID（省略時は既定スペース）
      --state string                    backlog | inbox | in_progress | done（カンマ区切りで複数）
```

### Options inherited from parent commands

```
      --fields string   data に含める列をカンマ区切りで指定（JSON のみ）
      --json            JSON で出力する（パイプ時は自動）
```

### SEE ALSO

* [pitboard task](pitboard_task.md)	 - タスク（list / show / create / assign / move / point / archive）

