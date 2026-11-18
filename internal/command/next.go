package command

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/nextproject"
	"github.com/replworks/repl-cli/internal/workspace"
)

func (r *runner) newNextCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Show the next task and the definitions it references",
		Long:  "next prints the first incomplete task in .replworks/TASKS.md and the definition excerpt of each identifier it references.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := r.projectDir()
			if err != nil {
				return err
			}

			outcome := nextproject.Coordinator{Workspace: workspace.New(dir)}.Run()
			r.status = renderNext(cmd, outcome)

			return nil
		},
	}
}

type nextJSON struct {
	Task        *definitionJSON  `json:"task"`
	Definitions []definitionJSON `json:"definitions"`
	Missing     []string         `json:"missing"`
	Errors      []string         `json:"errors"`
}

func renderNext(cmd *cobra.Command, outcome nextproject.Outcome) int {
	code := 0
	if len(outcome.Errors) > 0 {
		code = 2
	}

	if jsonRequested(cmd) {
		var task *definitionJSON
		if outcome.Task != nil {
			value := toDefinitionJSON(*outcome.Task)
			task = &value
		}
		if !encodeJSON(cmd, nextJSON{
			Task:        task,
			Definitions: definitionsJSON(outcome.Definitions),
			Missing:     nonNil(outcome.Missing),
			Errors:      errorTexts(outcome.Errors),
		}) {
			return 2
		}
		return code
	}

	out := cmd.OutOrStdout()
	switch {
	case outcome.Task != nil:
		renderExcerpt(out, *outcome.Task)
	case len(outcome.Errors) == 0:
		_, _ = fmt.Fprintln(out, "no incomplete task")
	}
	renderExcerpts(out, outcome.Definitions)
	renderMissing(out, outcome.Missing)
	for _, err := range outcome.Errors {
		_, _ = fmt.Fprintln(out, errorLine(err))
	}

	return code
}
