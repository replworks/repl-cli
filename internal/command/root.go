package command

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/canon"
)

type App struct {
	Version    string
	Commit     string
	Date       string
	Stdout     io.Writer
	Stderr     io.Writer
	HTTPClient *http.Client
}

func (a App) Execute(args []string) int {
	r := &runner{app: a}
	root := r.newRootCommand()
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(a.Stderr, errorLine(err))
		return 2
	}

	return r.status
}

type runner struct {
	app             App
	status          int
	canonicalSource *canon.Provider
}

func (r *runner) canonical() *canon.Provider {
	if r.canonicalSource == nil {
		r.canonicalSource = canon.New(r.app.HTTPClient)
	}

	return r.canonicalSource
}

func (r *runner) newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "repl",
		Short:         "Command-line program for REPL Works projects",
		Long:          "repl operates on a REPL Works project directory: it installs the canonical AGENTS.md, reports task progress, verifies project rules, and updates AGENTS.md.",
		Version:       r.app.versionString(),
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
	root.SetOut(r.app.Stdout)
	root.SetErr(r.app.Stderr)
	root.AddCommand(r.newStatusCommand())
	root.AddCommand(r.newInitCommand())

	return root
}

func (r *runner) projectDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("determining the project directory: %w", err)
	}

	return dir, nil
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
