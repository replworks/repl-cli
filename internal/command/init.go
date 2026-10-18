package command

import (
	"fmt"

	"github.com/spf13/cobra"

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
			dir, err := r.projectDir()
			if err != nil {
				return err
			}

			outcome := initproject.Coordinator{
				Source:    r.canonical(),
				Workspace: workspace.GrantWrite(workspace.New(dir)),
			}.Run()
			r.status = renderInit(cmd, outcome)

			return nil
		},
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
