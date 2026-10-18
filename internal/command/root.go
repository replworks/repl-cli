package command

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

type App struct {
	Version string
	Commit  string
	Date    string
	Stdout  io.Writer
	Stderr  io.Writer
}

func (a App) Execute(args []string) int {
	root := a.newRootCommand()
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		return 2
	}
	return 0
}

func (a App) newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:     "repl",
		Short:   "Command-line program for REPL Works projects",
		Long:    "repl operates on a REPL Works project directory: it installs the canonical AGENTS.md, reports task progress, verifies project rules, and updates AGENTS.md.",
		Version: a.versionString(),
	}
	root.SetOut(a.Stdout)
	root.SetErr(a.Stderr)
	return root
}

func (a App) versionString() string {
	if a.Commit == "" || a.Commit == "none" {
		return a.Version
	}
	return fmt.Sprintf("%s (%s, %s)", a.Version, a.Commit, a.Date)
}
