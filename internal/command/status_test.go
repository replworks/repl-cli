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

func TestStatusSendsNoNetworkRequest(t *testing.T) {
	dir := t.TempDir()
	writeTasksFile(t, dir, "- [ ] T-001 first\n")
	t.Chdir(dir)

	client, transport := servePage(t, string(recordedPage(t)))

	if code, _, stderr := runCommand(t, client, "status"); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}

	if requests := transport.requested(); len(requests) != 0 {
		t.Fatalf("requests = %q, want none", requests)
	}
}

func TestStatusReportsErrorWhenTasksIsAbsent(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	before := directoryState(t, dir)

	code, stdout, stderr := runApp(t, "status")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	lines := errorLines(t, stdout)
	if len(lines) != 1 {
		t.Fatalf("stdout has %d lines (%q), want 1", len(lines), stdout)
	}
	if !strings.Contains(lines[0], "repl:") || !strings.Contains(lines[0], "TASKS.md") {
		t.Fatalf("error line = %q, want it to identify TASKS.md", lines[0])
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestStatusReportsErrorWhenTasksCannotBeRead(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".replworks", "TASKS.md"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)
	before := directoryState(t, dir)

	code, stdout, stderr := runApp(t, "status")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	lines := errorLines(t, stdout)
	if len(lines) != 1 {
		t.Fatalf("stdout has %d lines (%q), want 1", len(lines), stdout)
	}
	if !strings.Contains(lines[0], "repl:") || !strings.Contains(lines[0], "TASKS.md") {
		t.Fatalf("error line = %q, want it to identify TASKS.md", lines[0])
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}
