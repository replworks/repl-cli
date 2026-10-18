package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/diff"
	"github.com/replworks/repl-cli/internal/syncproject"
	"github.com/replworks/repl-cli/internal/workspace"
)

func (r *runner) newSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Replace AGENTS.md with the canonical AGENTS.md",
		Long:  "sync compares AGENTS.md with the canonical AGENTS.md, prints the line-by-line difference, and replaces AGENTS.md with the canonical content when the contents differ.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			document, err := selectDocument(cmd)
			if err != nil {
				return err
			}

			dir, err := r.projectDir()
			if err != nil {
				return err
			}

			outcome := syncproject.Coordinator{
				Source:     r.canonical(document),
				Workspace:  workspace.GrantWrite(workspace.New(dir)),
				Difference: diff.Calculator{},
			}.Run()
			r.status = renderSync(cmd, outcome)

			return nil
		},
	}
}

func renderSync(cmd *cobra.Command, outcome syncproject.Outcome) int {
	out := cmd.OutOrStdout()

	if outcome.UpToDate {
		_, _ = fmt.Fprintln(out, "AGENTS.md is up to date")

		return 0
	}

	for _, entry := range outcome.Difference {
		marker := "-"
		if entry.Kind == diff.Added {
			marker = "+"
		}
		_, _ = fmt.Fprintf(out, "%s %s\n", marker, entry.Line)
	}

	if len(outcome.Errors) > 0 {
		for _, err := range outcome.Errors {
			_, _ = fmt.Fprintln(out, errorLine(err))
		}
		return 2
	}

	return 0
}
