package basebranch

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=repl-cli test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=repl-cli test",
		"GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+os.DevNull,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.String()
}

func repository(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	git(t, dir, "init", "-b", "main")

	return dir
}

func write(t *testing.T, dir, path, content string) {
	t.Helper()

	abs := filepath.Join(dir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func commit(t *testing.T, dir string) {
	t.Helper()

	git(t, dir, "add", "-A")
	git(t, dir, "-c", "commit.gpgsign=false", "commit", "-m", "project")
}

func TestTasksReturnsTheTasksOfTheBranch(t *testing.T) {
	dir := repository(t)
	write(t, dir, tasksPath, "- [ ] T-001 first\n")
	commit(t, dir)

	git(t, dir, "checkout", "-q", "-b", "dev")
	write(t, dir, tasksPath, "- [x] T-001 first\n")
	commit(t, dir)
	git(t, dir, "checkout", "-q", "main")

	main := New(dir).Tasks("main")
	if !main.IsPresent() || main.Value() != "- [ ] T-001 first\n" {
		t.Fatalf("main tasks = %v, want the main text", main)
	}

	dev := New(dir).Tasks("dev")
	if !dev.IsPresent() || dev.Value() != "- [x] T-001 first\n" {
		t.Fatalf("dev tasks = %v, want the dev text", dev)
	}
}

func TestTasksIsAbsentWhenTheBranchHasNoTasks(t *testing.T) {
	dir := repository(t)
	write(t, dir, "AGENTS.md", "# AGENTS.md\n")
	commit(t, dir)

	got := New(dir).Tasks("main")

	if !got.IsAbsent() {
		t.Fatalf("tasks = %v, want absent", got)
	}
}

func TestTasksFailsOutsideARepository(t *testing.T) {
	got := New(t.TempDir()).Tasks("main")

	if !got.IsFailed() {
		t.Fatalf("tasks = %v, want a failure", got)
	}
	if !strings.Contains(got.Reason().Error(), "git rev-parse --show-toplevel failed") {
		t.Fatalf("reason = %q, want the failing git invocation", got.Reason())
	}
}

func TestTasksFailsWhenTheBranchDoesNotExist(t *testing.T) {
	dir := repository(t)
	write(t, dir, "AGENTS.md", "# AGENTS.md\n")
	commit(t, dir)

	got := New(dir).Tasks("dev")

	if !got.IsFailed() {
		t.Fatalf("tasks = %v, want a failure", got)
	}
	if !strings.Contains(got.Reason().Error(), "the base branch dev does not exist") {
		t.Fatalf("reason = %q, want the missing branch", got.Reason())
	}
}

func TestTasksTreatsAnOptionLikeBranchAsAMissingBranch(t *testing.T) {
	dir := repository(t)
	write(t, dir, "AGENTS.md", "# AGENTS.md\n")
	commit(t, dir)

	got := New(dir).Tasks("-upload-pack=/bin/false")

	if !got.IsFailed() {
		t.Fatalf("tasks = %v, want a failure", got)
	}
	if !strings.Contains(got.Reason().Error(), "does not exist") {
		t.Fatalf("reason = %q, want the missing branch", got.Reason())
	}
}

func TestTasksAcceptsASubdirectoryOfTheRepository(t *testing.T) {
	dir := repository(t)
	write(t, dir, "sub/"+tasksPath, "- [ ] T-001 first\n")
	write(t, dir, "AGENTS.md", "# AGENTS.md\n")
	commit(t, dir)

	got := New(filepath.Join(dir, "sub")).Tasks("main")

	if !got.IsPresent() || got.Value() != "- [ ] T-001 first\n" {
		t.Fatalf("tasks = %v, want the text at the project-relative location", got)
	}
}

func TestTasksIsAbsentWhenAnotherProjectHoldsTheTasks(t *testing.T) {
	dir := repository(t)
	write(t, dir, "sub/"+tasksPath, "- [ ] T-001 first\n")
	write(t, dir, "AGENTS.md", "# AGENTS.md\n")
	commit(t, dir)

	got := New(dir).Tasks("main")

	if !got.IsAbsent() {
		t.Fatalf("tasks = %v, want absent", got)
	}
}

func TestTasksDoesNotChangeRepositoryState(t *testing.T) {
	dir := repository(t)
	write(t, dir, tasksPath, "- [ ] T-001 first\n")
	commit(t, dir)

	before := state(t, dir)

	New(dir).Tasks("main")

	if after := state(t, dir); after != before {
		t.Fatalf("repository state changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func state(t *testing.T, dir string) string {
	t.Helper()

	return strings.Join([]string{
		git(t, dir, "rev-parse", "--abbrev-ref", "HEAD"),
		git(t, dir, "rev-parse", "HEAD"),
		git(t, dir, "show-ref"),
		git(t, dir, "status", "--porcelain"),
	}, "\n")
}
