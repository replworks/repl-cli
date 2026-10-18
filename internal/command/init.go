package command

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/canon"
	"github.com/replworks/repl-cli/internal/initproject"
	"github.com/replworks/repl-cli/internal/workspace"
)

func (r *runner) newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create AGENTS.md and the .replworks directory",
		Long:  "init writes the canonical AGENTS.md into the project directory and creates the .replworks directory. Existing entries are left unchanged.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			document, err := selectInitDocument(cmd)
			if err != nil {
				return err
			}

			dir, err := r.projectDir()
			if err != nil {
				return err
			}

			outcome := initproject.Coordinator{
				Source:    r.canonical(document),
				Workspace: workspace.GrantWrite(workspace.New(dir)),
			}.Run()
			r.status = renderInit(cmd, outcome)

			return nil
		},
	}
}

func selectInitDocument(cmd *cobra.Command) (canon.Document, error) {
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, "Select an AGENTS.md document:")
	_, _ = fmt.Fprintf(out, "1) %s\n", canon.AgentsDocument)
	_, _ = fmt.Fprintf(out, "2) %s\n", canon.AgentsAutoDocument)
	_, _ = fmt.Fprintln(out, "Choice [1-2]:")

	scanner := bufio.NewScanner(cmd.InOrStdin())
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("reading the init document selection: %w", err)
		}

		return "", errors.New("an init document selection is required")
	}

	switch strings.TrimSpace(scanner.Text()) {
	case "1":
		return canon.AgentsDocument, nil
	case "2":
		return canon.AgentsAutoDocument, nil
	default:
		return "", errors.New("the init document selection must be 1 or 2")
	}
}

func renderInit(cmd *cobra.Command, outcome initproject.Outcome) int {
	out := cmd.OutOrStdout()

	for _, item := range outcome.Items {
		state := "skipped"
		if item.Created {
			state = "created"
		}
		_, _ = fmt.Fprintf(out, "%s: %s\n", state, item.Path)
	}

	if len(outcome.Errors) > 0 {
		for _, err := range outcome.Errors {
			_, _ = fmt.Fprintln(out, errorLine(err))
		}
		return 2
	}

	return 0
}
