package status

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/replworks/repl-cli/internal/workspace"
)

func writeTasks(t *testing.T, dir, text string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".replworks", "TASKS.md"), []byte(text), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func TestRunReportsProgress(t *testing.T) {
	dir := t.TempDir()
	writeTasks(t, dir, ""+
		"- [x] T-001 one\n"+
		"- [X] T-002 two\n"+
		"- [ ] T-003 three\n"+
		"- [ ] T-004 four\n"+
		"- [ ] T-005 five\n")

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run()

	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	if outcome.Total != 5 || outcome.Complete != 2 {
		t.Fatalf("outcome = %+v, want 5 total and 2 complete", outcome)
	}
	if outcome.FirstIncompleteID != "T-003" {
		t.Fatalf("FirstIncompleteID = %q, want %q", outcome.FirstIncompleteID, "T-003")
	}
}

func TestRunReportsAllComplete(t *testing.T) {
	dir := t.TempDir()
	writeTasks(t, dir, "- [x] T-001 one\n- [x] T-002 two\n")

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run()

	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	if outcome.Total != 2 || outcome.Complete != 2 || outcome.FirstIncompleteID != "" {
		t.Fatalf("outcome = %+v, want all complete", outcome)
	}
}

func TestRunReportsEmptyLedger(t *testing.T) {
	dir := t.TempDir()
	writeTasks(t, dir, "")

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run()

	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	if outcome.Total != 0 || outcome.Complete != 0 || outcome.FirstIncompleteID != "" {
		t.Fatalf("outcome = %+v, want an empty ledger", outcome)
	}
}

func TestRunRecordsErrorWhenTasksIsAbsent(t *testing.T) {
	outcome := Coordinator{Workspace: workspace.New(t.TempDir())}.Run()

	if len(outcome.Errors) != 1 {
		t.Fatalf("errors = %v, want exactly one", outcome.Errors)
	}
}

func TestRunRecordsErrorWhenTasksCannotBeRead(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".replworks", "TASKS.md"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run()

	if len(outcome.Errors) != 1 {
		t.Fatalf("errors = %v, want exactly one", outcome.Errors)
	}
}
