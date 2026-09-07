package main

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/yamataka22/pitboard-cli/internal/cli"
)

// リリースビルドは goreleaser が ldflags で埋める: -X main.version=1.2.3
var version = "dev"

func main() {
	os.Exit(cli.Execute(resolveVersion(), os.Args[1:], os.Stdout, os.Stderr, os.Stdin))
}

// go install で入れた場合は ldflags が無く "dev" のままになるが、module 経由の
// ビルドには Go が版を埋めているので、そちらを使う。
// version が分からないと API の下限バージョンチェック（User-Agent を見る）が
// 素通りし、古い CLI が 426 で止まらなくなる
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return version
	}
	return strings.TrimPrefix(info.Main.Version, "v")
}
