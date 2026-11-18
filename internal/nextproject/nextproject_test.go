package nextproject

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/replworks/repl-cli/internal/workspace"
)

func write(t *testing.T, dir, rel, text string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func ids(outcome Outcome) []string {
	var out []string
	for _, d := range outcome.Definitions {
		out = append(out, d.ID)
	}
	return out
}

func TestRunPrintsTaskAndReferences(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".replworks/TASKS.md", "- [x] T-001 One\n\n- [ ] T-002 Two\n  - Satisfies: FR-001, C3, AC-999, T-001, T-002\n")
	write(t, dir, ".replworks/PRODUCT_SPEC.md", "- FR-001: one\n")
	write(t, dir, ".replworks/ARCHITECTURE.md", "### C3. Three\n- body\n")

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run()

	if outcome.Task == nil || outcome.Task.ID != "T-002" || outcome.Task.Line != 3 || len(outcome.Task.Lines) != 2 {
		t.Fatalf("task = %+v", outcome.Task)
	}
	if !slices.Equal(ids(outcome), []string{"FR-001", "C3", "T-001"}) || !slices.Equal(outcome.Missing, []string{"AC-999"}) || len(outcome.Errors) != 0 {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestRunWithoutIncompleteTask(t *testing.T) {
	for _, text := range []string{"- [x] T-001 One\n", "# TASKS.md\n"} {
		dir := t.TempDir()
		write(t, dir, ".replworks/TASKS.md", text)

		outcome := Coordinator{Workspace: workspace.New(dir)}.Run()

		if outcome.Task != nil || len(outcome.Definitions) != 0 || len(outcome.Missing) != 0 || len(outcome.Errors) != 0 {
			t.Fatalf("outcome = %+v", outcome)
		}
	}
}

func TestRunWithoutTasksReadsNothingElse(t *testing.T) {
	dir := t.TempDir()
	// An unreadable PRODUCT_SPEC.md would add a second error if it were read.
	if err := os.MkdirAll(filepath.Join(dir, ".replworks", "PRODUCT_SPEC.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run()

	if outcome.Task != nil || len(outcome.Errors) != 1 {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestRunUnreadableDefiningDocument(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".replworks/TASKS.md", "- [ ] T-001 One\n  - C3 and FR-001\n")
	write(t, dir, ".replworks/PRODUCT_SPEC.md", "- FR-001: one\n")
	if err := os.MkdirAll(filepath.Join(dir, ".replworks", "ARCHITECTURE.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run()

	if !slices.Equal(ids(outcome), []string{"FR-001"}) || len(outcome.Missing) != 0 || len(outcome.Errors) != 1 {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestRunAbsentDefiningDocumentMakesReferencesMissing(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".replworks/TASKS.md", "- [ ] T-001 One\n  - C3 and FR-001\n")

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run()

	if len(outcome.Definitions) != 0 || !slices.Equal(outcome.Missing, []string{"C3", "FR-001"}) || len(outcome.Errors) != 0 {
		t.Fatalf("outcome = %+v", outcome)
	}
}
