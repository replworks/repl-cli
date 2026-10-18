package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/checkproject"
	"github.com/replworks/repl-cli/internal/rule"
	"github.com/replworks/repl-cli/internal/workspace"
)

func (r *runner) newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Verify the project directory against the check rules",
		Long:  "check verifies the project directory against the check rules and prints one line per violation and one line per execution error.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := r.projectDir()
			if err != nil {
				return err
			}

			outcome := checkproject.Coordinator{
				Workspace: workspace.New(dir),
				Evaluator: rule.Evaluator{},
			}.Run()
			r.status = renderCheck(cmd, outcome)

			return nil
		},
	}
}

func renderCheck(cmd *cobra.Command, outcome checkproject.Outcome) int {
	out := cmd.OutOrStdout()

	for _, violation := range outcome.Violations {
		_, _ = fmt.Fprintf(out, "%s %s: %s\n", violation.Rule, violation.Path, violation.Message)
	}
	for _, err := range outcome.Errors {
		_, _ = fmt.Fprintln(out, errorLine(err))
	}

	if len(outcome.Errors) > 0 {
		return 2
	}
	if len(outcome.Violations) > 0 {
		return 1
	}

	return 0
}
