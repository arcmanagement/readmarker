package main

import (
	"os"

	"github.com/arcmanagement/readmarker/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(cli.New(version, commit, date).Run(os.Args[1:], os.Stdout, os.Stderr))
}
