## pitboard comment add

タスクにコメントを追加する（返信・メンション可）

### Synopsis

タスクにコメントを追加する。--body - なら本文を標準入力から読む。
--reply-to にコメント ID を渡すと返信になる（ID は task show で確認。返信への返信はできない）。
--mention にメンバー ID を渡すと本文の先頭に [@名前](mention:ID) が付き、相手に通知が届く。本文に直接この記法を書いてもよい。

```
pitboard comment add NUMBER --body TEXT|- [--reply-to COMMENT_ID] [--mention ID]... --yes [flags]
```

### Options

```
      --body string       本文（markdown）。- で標準入力
  -h, --help              help for add
      --mention strings   メンションするメンバー ID または名前（繰り返し指定可）
      --reply-to string   返信先のコメント ID
      --space string      スペース ID（省略時は既定スペース）
      --yes               書き込みを確認したことを示す（必須）
```

### Options inherited from parent commands

```
      --fields string   data に含める列をカンマ区切りで指定（JSON のみ）
      --json            JSON で出力する（パイプ時は自動）
```

### SEE ALSO

* [pitboard comment](pitboard_comment.md)	 - コメント（add）

