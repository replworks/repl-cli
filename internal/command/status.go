package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/status"
	"github.com/replworks/repl-cli/internal/workspace"
)

func (r *runner) newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Report task progress",
		Long:  "status reports how many tasks in .replworks/TASKS.md are complete and the ID of the first incomplete task.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := r.projectDir()
			if err != nil {
				return err
			}

			outcome := status.Coordinator{Workspace: workspace.New(dir)}.Run()
			r.status = renderStatus(cmd, outcome)

			return nil
		},
	}
}

func renderStatus(cmd *cobra.Command, outcome status.Outcome) int {
	out := cmd.OutOrStdout()

	if len(outcome.Errors) > 0 {
		for _, err := range outcome.Errors {
			_, _ = fmt.Fprintln(out, errorLine(err))
		}
		return 2
	}

	_, _ = fmt.Fprintf(out, "complete: %d of %d\n", outcome.Complete, outcome.Total)
	if outcome.FirstIncompleteID == "" {
		_, _ = fmt.Fprintln(out, "no incomplete task")
	} else {
		_, _ = fmt.Fprintf(out, "next: %s\n", outcome.FirstIncompleteID)
	}

	return 0
}
