package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTasksFile(t *testing.T, dir, text string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".replworks", "TASKS.md"), []byte(text), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func TestStatusReportsProgress(t *testing.T) {
	dir := t.TempDir()
	writeTasksFile(t, dir, ""+
		"- [x] T-001 one\n"+
		"- [X] T-002 two\n"+
		"- [ ] T-003 three\n"+
		"- [ ] T-004 four\n"+
		"- [ ] T-005 five\n")
	t.Chdir(dir)

	code, stdout, stderr := runApp(t, "status")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	for _, want := range []string{"2", "5", "T-003"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestStatusReportsAllComplete(t *testing.T) {
	dir := t.TempDir()
	writeTasksFile(t, dir, "- [x] T-001 one\n- [X] T-002 two\n")
	t.Chdir(dir)

	code, stdout, stderr := runApp(t, "status")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if !strings.Contains(stdout, "no incomplete task") {
		t.Fatalf("stdout = %q, want the no-incomplete-task indication", stdout)
	}
}

func TestStatusReportsEmptyLedger(t *testing.T) {
	dir := t.TempDir()
	writeTasksFile(t, dir, "")
	t.Chdir(dir)

	code, stdout, stderr := runApp(t, "status")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	for _, want := range []string{"0", "no incomplete task"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestStatusModifiesNoFile(t *testing.T) {
	dir := t.TempDir()
	writeTasksFile(t, dir, "- [x] T-001 one\n- [ ] T-002 two\n")
	t.Chdir(dir)

	before := directoryState(t, dir)

	if code, _, stderr := runApp(t, "status"); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}

	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}
