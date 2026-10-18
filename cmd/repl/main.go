package main

import (
	"os"

	"github.com/replworks/repl-cli/internal/command"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := command.App{
		Version: version,
		Commit:  commit,
		Date:    date,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
	os.Exit(app.Execute(os.Args[1:]))
}
