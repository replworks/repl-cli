//go:build e2e

package command

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func liveRecordingClient() (*http.Client, *recordingTransport) {
	transport := &recordingTransport{target: "https://www.repl.net"}

	return &http.Client{Transport: transport}, transport
}

func TestCheckAgainstTheLiveSourceReportsADifferentAgentsMD(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, "# stale\n")
	initRepository(t, dir)
	t.Chdir(dir)

	client, transport := liveRecordingClient()
	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	want := []string{"agents-canonical AGENTS.md: AGENTS.md differs from the canonical AGENTS.md"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}

	if got := transport.requested(); len(got) != 1 {
		t.Fatalf("requests = %q, want the canonical text retrieved exactly once", got)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestCheckAgainstTheLiveSourceDoesNotReportAnIdenticalAgentsMD(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, want)
	initRepository(t, dir)
	t.Chdir(dir)

	client, transport := liveRecordingClient()

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want no violation", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if got := transport.requested(); len(got) != 1 {
		t.Fatalf("requests = %q, want the canonical text retrieved exactly once", got)
	}
}

func TestCheckAgainstTheLiveSourceReportsALineEndingDifference(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, strings.ReplaceAll(want, "\n", "\r\n"))
	initRepository(t, dir)
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "check")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "agents-canonical AGENTS.md") {
		t.Fatalf("stdout = %q, want an agents-canonical violation", stdout)
	}
}

func TestCheckAgainstTheLiveSourceReportsAnAbsentAgentsMD(t *testing.T) {
	liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, "# stale\n")
	if err := os.Remove(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	initRepository(t, dir)
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "check")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	want := []string{"documents-exist AGENTS.md: document does not exist"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func gitCommandOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
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

func repositoryState(t *testing.T, dir string) string {
	t.Helper()

	return strings.Join([]string{
		gitCommandOutput(t, dir, "rev-parse", "--abbrev-ref", "HEAD"),
		gitCommandOutput(t, dir, "rev-parse", "HEAD"),
		gitCommandOutput(t, dir, "show-ref"),
		gitCommandOutput(t, dir, "status", "--porcelain"),
	}, "\n")
}

func TestCheckAgainstTheLiveRepositoryUsesMain(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, want)
	initRepository(t, dir)
	t.Chdir(dir)

	client, transport := liveRecordingClient()
	before := repositoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want no line", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if got := transport.requested(); len(got) != 1 {
		t.Fatalf("requests = %q, want the canonical text retrieved exactly once", got)
	}
	if after := repositoryState(t, dir); after != before {
		t.Fatalf("repository state changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestCheckAgainstTheLiveRepositoryWithoutMain(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, want)
	gitCommand(t, dir, "init", "-q", "-b", "trunk")
	gitCommand(t, dir, "add", "-A")
	gitCommand(t, dir, "-c", "commit.gpgsign=false", "commit", "-q", "-m", "project")
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "check")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	want2 := []string{"repl: the base branch main does not exist in the repository"}
	if !reflect.DeepEqual(lines, want2) {
		t.Fatalf("stdout = %q, want %q", stdout, want2)
	}
}

func TestCheckAgainstTheLiveRepositoryUsesTheGivenBase(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, want)
	initRepository(t, dir)
	gitCommand(t, dir, "checkout", "-q", "-b", "dev")
	t.Chdir(dir)

	client, _ := liveRecordingClient()
	before := repositoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "check", "--base", "dev")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want no line", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if after := repositoryState(t, dir); after != before {
		t.Fatalf("repository state changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestCheckAgainstTheLiveRepositoryReportsAMissingBaseBranch(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, want)
	initRepository(t, dir)
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "check", "--base", "nope")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	want2 := []string{"repl: the base branch nope does not exist in the repository"}
	if !reflect.DeepEqual(lines, want2) {
		t.Fatalf("stdout = %q, want %q", stdout, want2)
	}

	for _, rule := range []string{"task-ids-unchanged", "single-task-completion", "no-unverified"} {
		if strings.Contains(stdout, rule) {
			t.Fatalf("stdout = %q, want no %s line", stdout, rule)
		}
	}
}

func TestCheckAgainstTheLiveRepositoryAcceptsASubdirectory(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writeDocuments(t, sub, want)
	initRepository(t, dir)
	t.Chdir(sub)

	client, _ := liveRecordingClient()

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want no line", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestCheckOutsideARepositoryReportsTheOtherRules(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, want)
	if err := os.Remove(filepath.Join(dir, ".replworks", "TASKS.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "check")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	if len(lines) != 2 {
		t.Fatalf("stdout = %q, want the violation and the error", stdout)
	}
	if lines[0] != "documents-exist .replworks/TASKS.md: document does not exist" {
		t.Fatalf("line 1 = %q, want the documents-exist violation", lines[0])
	}
	if !strings.HasPrefix(lines[1], "repl: ") || !strings.Contains(lines[1], "not a git repository") {
		t.Fatalf("line 2 = %q, want the repository error", lines[1])
	}
	if strings.Contains(stdout, "agents-canonical") {
		t.Fatalf("stdout = %q, want no agents-canonical violation", stdout)
	}
}

func TestCheckOutsideARepositoryReportsTheRepositoryErrorAlone(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, want)
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "check")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	if len(lines) != 1 {
		t.Fatalf("stdout = %q, want the error line alone", stdout)
	}
	if !strings.Contains(lines[0], "not a git repository") {
		t.Fatalf("line 1 = %q, want the repository error", lines[0])
	}
}

func TestCheckAgainstTheLiveRepositoryReportsRemovedTaskIDs(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writeDocuments(t, sub, want)
	writeTasks(t, sub, "- [ ] T-001 first\n- [ ] T-002 second\n- [ ] T-003 third\n")
	initRepository(t, dir)
	writeTasks(t, sub, "- [ ] T-001 first\n- [x] T-003 third\n")
	t.Chdir(sub)

	client, _ := liveRecordingClient()

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	want2 := []string{"task-ids-unchanged .replworks/TASKS.md: missing task ID T-002"}
	if !reflect.DeepEqual(lines, want2) {
		t.Fatalf("stdout = %q, want %q", stdout, want2)
	}
}

func TestCheckAgainstTheLiveRepositoryReportsNoRemovedTaskIDs(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	writeDocuments(t, dir, want)
	writeTasks(t, dir, "- [ ] T-001 first\n- [x] T-002 second\n")
	initRepository(t, dir)
	t.Chdir(dir)

	client, _ := liveRecordingClient()

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want no line", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}
