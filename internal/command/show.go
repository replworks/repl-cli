package command

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/definition"
	"github.com/replworks/repl-cli/internal/showproject"
	"github.com/replworks/repl-cli/internal/workspace"
)

func (r *runner) newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <identifier>...",
		Short: "Show the definitions of identifiers",
		Long:  "show prints the definition excerpt of each given identifier (T, FR, EL, ERR, AC, C, or AR) from the authoritative documents.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := showIdentifiers(args)
			if err != nil {
				r.status = renderShow(cmd, showproject.Outcome{Errors: []error{err}})
				return nil
			}

			dir, err := r.projectDir()
			if err != nil {
				return err
			}

			outcome := showproject.Coordinator{Workspace: workspace.New(dir)}.Run(ids)
			r.status = renderShow(cmd, outcome)

			return nil
		},
	}
}

func showIdentifiers(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, errors.New("show requires at least one identifier")
	}

	var invalid, ids []string
	for _, arg := range args {
		switch {
		case !definition.Valid(arg):
			invalid = append(invalid, fmt.Sprintf("%q", arg))
		case !slices.Contains(ids, arg):
			ids = append(ids, arg)
		}
	}
	if len(invalid) > 0 {
		return nil, fmt.Errorf("not an identifier: %s", strings.Join(invalid, ", "))
	}

	return ids, nil
}

func renderShow(cmd *cobra.Command, outcome showproject.Outcome) int {
	code := showExitCode(outcome)
	if jsonRequested(cmd) {
		if !encodeJSON(cmd, showJSON{
			Definitions: definitionsJSON(outcome.Definitions),
			Missing:     nonNil(outcome.Missing),
			Errors:      errorTexts(outcome.Errors),
		}) {
			return 2
		}
		return code
	}

	out := cmd.OutOrStdout()
	renderExcerpts(out, outcome.Definitions)
	renderMissing(out, outcome.Missing)
	for _, err := range outcome.Errors {
		_, _ = fmt.Fprintln(out, errorLine(err))
	}

	return code
}

func showExitCode(outcome showproject.Outcome) int {
	switch {
	case len(outcome.Errors) > 0:
		return 2
	case len(outcome.Missing) > 0:
		return 1
	default:
		return 0
	}
}
