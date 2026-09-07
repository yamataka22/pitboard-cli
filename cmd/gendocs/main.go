// gendocs は cobra の定義から docs/commands/ の markdown を生成する（make docs）。
// コマンドの説明はコード側（Short / Long / フラグの説明）が正。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra/doc"

	"github.com/yamataka22/pitboard-cli/internal/cli"
)

func main() {
	dir := "docs/commands"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if err := os.RemoveAll(dir); err != nil {
		fail(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fail(err)
	}
	root := cli.NewRootCommand("x.y.z")
	if err := doc.GenMarkdownTree(root, dir); err != nil {
		fail(err)
	}
	// 目次
	entries, err := os.ReadDir(dir)
	if err != nil {
		fail(err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") && e.Name() != "README.md" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	var b strings.Builder
	b.WriteString("# コマンドリファレンス\n\n`make docs` で cobra の定義から生成。手で編集しない。\n\n")
	for _, n := range names {
		title := strings.ReplaceAll(strings.TrimSuffix(n, ".md"), "_", " ")
		fmt.Fprintf(&b, "- [%s](%s)\n", title, n)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(b.String()), 0o644); err != nil {
		fail(err)
	}
	fmt.Printf("generated %d files in %s\n", len(names), dir)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
