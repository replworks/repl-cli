package command

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/canon"
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
				Source:       r.canonical(document),
				Script:       syncScriptSource(r, document),
				Auto:         document == canon.AgentsAutoDocument,
				ManageScript: true,
				Workspace:    workspace.GrantWrite(workspace.New(dir)),
				Difference:   diff.Calculator{},
			}.Run()
			r.status = renderSync(cmd, outcome)

			return nil
		},
	}
}

func syncScriptSource(r *runner, document canon.Document) syncproject.CanonicalSource {
	if document == canon.AgentsAutoDocument {
		return r.askHumanScript()
	}

	return nil
}

func renderSync(cmd *cobra.Command, outcome syncproject.Outcome) int {
	return renderSyncOutput(cmd, outcome, terminalOutput(cmd.OutOrStdout()))
}

func renderSyncOutput(cmd *cobra.Command, outcome syncproject.Outcome, color bool) int {
	out := cmd.OutOrStdout()

	if outcome.UpToDate {
		_, _ = fmt.Fprintln(out, "AGENTS.md is up to date")

		return 0
	}

	for _, entry := range outcome.Difference {
		marker := "-"
		colorCode := ""
		resetCode := ""
		if entry.Kind == diff.Added {
			marker = "+"
		}
		if color {
			colorCode = "\x1b[31m"
			resetCode = "\x1b[0m"
			if entry.Kind == diff.Added {
				colorCode = "\x1b[32m"
			}
		}
		_, _ = fmt.Fprintf(out, "%s%s %s%s\n", colorCode, marker, entry.Line, resetCode)
	}
	if outcome.ScriptPath != "" {
		_, _ = fmt.Fprintf(out, "%s: %s\n", outcome.ScriptState, outcome.ScriptPath)
	}

	if len(outcome.Errors) > 0 {
		for _, err := range outcome.Errors {
			_, _ = fmt.Fprintln(out, errorLine(err))
		}
		return 2
	}

	return 0
}

func terminalOutput(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
