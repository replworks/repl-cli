package command

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/checkproject"
	"github.com/replworks/repl-cli/internal/diff"
	"github.com/replworks/repl-cli/internal/initproject"
	"github.com/replworks/repl-cli/internal/skillsproject"
	"github.com/replworks/repl-cli/internal/status"
	"github.com/replworks/repl-cli/internal/syncproject"
)

func jsonRequested(cmd *cobra.Command) bool {
	requested, err := cmd.Flags().GetBool("json")
	return err == nil && requested
}

func encodeJSON(cmd *cobra.Command, value any) bool {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(value); err != nil {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), errorLine(err))
		return false
	}

	return true
}

func errorText(err error) string {
	return strings.Join(strings.Fields(err.Error()), " ")
}

func errorTexts(errs []error) []string {
	texts := make([]string, 0, len(errs))
	for _, err := range errs {
		texts = append(texts, errorText(err))
	}

	return texts
}

type statusJSON struct {
	Complete        int      `json:"complete"`
	Total           int      `json:"total"`
	FirstIncomplete *string  `json:"first_incomplete"`
	Errors          []string `json:"errors"`
}

func renderStatusJSON(cmd *cobra.Command, outcome status.Outcome) int {
	var first *string
	if outcome.FirstIncompleteID != "" {
		id := outcome.FirstIncompleteID
		first = &id
	}

	value := statusJSON{
		Complete:        outcome.Complete,
		Total:           outcome.Total,
		FirstIncomplete: first,
		Errors:          errorTexts(outcome.Errors),
	}

	if !encodeJSON(cmd, value) {
		return 2
	}
	if len(outcome.Errors) > 0 {
		return 2
	}

	return 0
}

type violationJSON struct {
	Rule    string `json:"rule"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

type checkJSON struct {
	Violations []violationJSON `json:"violations"`
	Errors     []string        `json:"errors"`
}

func renderCheckJSON(cmd *cobra.Command, outcome checkproject.Outcome) int {
	violations := make([]violationJSON, 0, len(outcome.Violations))
	for _, violation := range outcome.Violations {
		violations = append(violations, violationJSON{
			Rule:    violation.Rule,
			Path:    violation.Path,
			Message: violation.Message,
		})
	}

	value := checkJSON{
		Violations: violations,
		Errors:     errorTexts(outcome.Errors),
	}

	if !encodeJSON(cmd, value) {
		return 2
	}
	if len(outcome.Errors) > 0 {
		return 2
	}
	if len(outcome.Violations) > 0 {
		return 1
	}

	return 0
}

type initItemJSON struct {
	Path  string `json:"path"`
	State string `json:"state"`
}

type initJSON struct {
	Items  []initItemJSON `json:"items"`
	Errors []string       `json:"errors"`
}

func renderInitJSON(cmd *cobra.Command, outcome initproject.Outcome) int {
	items := make([]initItemJSON, 0, len(outcome.Items))
	for _, item := range outcome.Items {
		state := "skipped"
		switch {
		case item.Created:
			state = "created"
		case item.Replaced:
			state = "replaced"
		case item.Deleted:
			state = "deleted"
		}

		items = append(items, initItemJSON{Path: item.Path, State: state})
	}

	value := initJSON{
		Items:  items,
		Errors: errorTexts(outcome.Errors),
	}

	if !encodeJSON(cmd, value) {
		return 2
	}
	if len(outcome.Errors) > 0 {
		return 2
	}

	return 0
}

type syncDiffJSON struct {
	Kind string `json:"kind"`
	Line string `json:"line"`
}

type syncScriptJSON struct {
	Path  string `json:"path"`
	State string `json:"state"`
}

type syncJSON struct {
	UpToDate bool            `json:"up_to_date"`
	Diff     []syncDiffJSON  `json:"diff"`
	Script   *syncScriptJSON `json:"script"`
	Errors   []string        `json:"errors"`
}

func renderSyncJSON(cmd *cobra.Command, outcome syncproject.Outcome) int {
	difference := make([]syncDiffJSON, 0, len(outcome.Difference))
	for _, entry := range outcome.Difference {
		kind := "removed"
		if entry.Kind == diff.Added {
			kind = "added"
		}

		difference = append(difference, syncDiffJSON{Kind: kind, Line: entry.Line})
	}

	var script *syncScriptJSON
	if outcome.ScriptPath != "" {
		script = &syncScriptJSON{Path: outcome.ScriptPath, State: outcome.ScriptState}
	}

	value := syncJSON{
		UpToDate: outcome.UpToDate,
		Diff:     difference,
		Script:   script,
		Errors:   errorTexts(outcome.Errors),
	}

	if !encodeJSON(cmd, value) {
		return 2
	}
	if len(outcome.Errors) > 0 {
		return 2
	}

	return 0
}

type skillsItemJSON struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	State string `json:"state"`
}

type skillsJSON struct {
	Items  []skillsItemJSON `json:"items"`
	Errors []string         `json:"errors"`
}

func renderSkillsSyncJSON(cmd *cobra.Command, outcome skillsproject.Outcome) int {
	items := make([]skillsItemJSON, 0, len(outcome.Items))
	for _, item := range outcome.Items {
		items = append(items, skillsItemJSON{
			Name:  item.Name,
			Path:  item.Path,
			State: string(item.State),
		})
	}

	value := skillsJSON{
		Items:  items,
		Errors: errorTexts(outcome.Errors),
	}

	if !encodeJSON(cmd, value) {
		return 2
	}
	if len(outcome.Errors) > 0 {
		return 2
	}

	return 0
}
