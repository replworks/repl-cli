package showproject

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

func TestRunFoundAndMissingInGivenOrder(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, ".replworks/PRODUCT_SPEC.md", "- FR-002: two\n- FR-001: one\n")
	write(t, dir, ".replworks/ARCHITECTURE.md", "### C1. One\n- body\n")

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run([]string{"C1", "FR-001", "FR-999", "FR-002", "C9"})

	if !slices.Equal(ids(outcome), []string{"C1", "FR-001", "FR-002"}) {
		t.Fatalf("definitions = %q", ids(outcome))
	}
	if !slices.Equal(outcome.Missing, []string{"FR-999", "C9"}) {
		t.Fatalf("missing = %q", outcome.Missing)
	}
	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v", outcome.Errors)
	}
}

func TestRunAbsentDocumentMakesIdentifiersMissing(t *testing.T) {
	outcome := Coordinator{Workspace: workspace.New(t.TempDir())}.Run([]string{"T-001", "AR-01"})

	if len(outcome.Definitions) != 0 || !slices.Equal(outcome.Missing, []string{"T-001", "AR-01"}) || len(outcome.Errors) != 0 {
		t.Fatalf("outcome = %+v", outcome)
	}
}

func TestRunUnreadableDocumentIsAnErrorNotMissing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".replworks", "ARCHITECTURE.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, dir, ".replworks/PRODUCT_SPEC.md", "- FR-001: one\n")

	outcome := Coordinator{Workspace: workspace.New(dir)}.Run([]string{"C3", "AR-01", "FR-001"})

	if !slices.Equal(ids(outcome), []string{"FR-001"}) || len(outcome.Missing) != 0 || len(outcome.Errors) != 1 {
		t.Fatalf("outcome = %+v", outcome)
	}
}
