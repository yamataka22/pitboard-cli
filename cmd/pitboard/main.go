package main

import (
	"os"

	"github.com/yamataka22/pitboard-cli/internal/cli"
)

// ldflags で上書きする: -X main.version=1.2.3
var version = "dev"

func main() {
	os.Exit(cli.Execute(version, os.Args[1:], os.Stdout, os.Stderr, os.Stdin))
}
