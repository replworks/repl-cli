package command

import (
	"fmt"
	"io"
	"strings"

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
		_, _ = fmt.Fprintln(a.Stderr, errorLine(err))
		return 2
	}
	return 0
}

func (a App) newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "repl",
		Short:         "Command-line program for REPL Works projects",
		Long:          "repl operates on a REPL Works project directory: it installs the canonical AGENTS.md, reports task progress, verifies project rules, and updates AGENTS.md.",
		Version:       a.versionString(),
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
			}
			return cmd.Help()
		},
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

func errorLine(err error) string {
	return "repl: " + strings.Join(strings.Fields(err.Error()), " ")
}
