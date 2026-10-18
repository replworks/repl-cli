package command

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/checkproject"
	"github.com/replworks/repl-cli/internal/rule"
)

func TestCheckReportsMissingDocuments(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(recordedText(t)), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	initRepository(t, dir)
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))
	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	want := []string{
		"documents-exist .replworks/PRODUCT_SPEC.md: document does not exist",
		"documents-exist .replworks/TECH_STACK.md: document does not exist",
		"documents-exist .replworks/ARCHITECTURE.md: document does not exist",
		"documents-exist .replworks/TASKS.md: document does not exist",
	}
	if len(lines) != len(want) {
		t.Fatalf("output = %q, want %d lines", stdout, len(want))
	}
	for i, line := range want {
		if lines[i] != line {
			t.Fatalf("line %d = %q, want %q", i+1, lines[i], line)
		}
	}

	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestCheckExitsZeroWithoutViolation(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, recordedText(t))
	initRepository(t, dir)

	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))
	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestCheckReportsAnExecutionErrorTogetherWithViolations(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, recordedText(t))
	if err := os.RemoveAll(filepath.Join(dir, ".replworks", "TECH_STACK.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, ".replworks", "ARCHITECTURE.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".replworks", "ARCHITECTURE.md"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	initRepository(t, dir)

	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))
	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "documents-exist .replworks/TECH_STACK.md: document does not exist") {
		t.Fatalf("stdout = %q, want the missing document as a violation", stdout)
	}
	if strings.Contains(stdout, "documents-exist .replworks/ARCHITECTURE.md") {
		t.Fatalf("stdout = %q, want no violation for the unreadable document", stdout)
	}
	if got := strings.Count(stdout, "repl: "); got != 1 {
		t.Fatalf("stdout = %q, want one error line, got %d", stdout, got)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestCheckRejectsArguments(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "check", "extra")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("stderr = %q, want it to reject the extra argument", stderr)
	}
}

func TestRenderCheck(t *testing.T) {
	tests := []struct {
		name       string
		outcome    checkproject.Outcome
		wantCode   int
		wantStdout string
	}{
		{
			name:       "no violation",
			outcome:    checkproject.Outcome{},
			wantCode:   0,
			wantStdout: "",
		},
		{
			name: "violations",
			outcome: checkproject.Outcome{Violations: []rule.Violation{
				{Rule: rule.RuleDocumentsExist, Path: ".replworks/TASKS.md", Message: "document does not exist"},
				{Rule: rule.RuleDocumentsExist, Path: ".replworks/PRODUCT_SPEC.md", Message: "document does not exist"},
			}},
			wantCode:   1,
			wantStdout: "documents-exist .replworks/TASKS.md: document does not exist\ndocuments-exist .replworks/PRODUCT_SPEC.md: document does not exist\n",
		},
		{
			name: "violations and errors",
			outcome: checkproject.Outcome{
				Violations: []rule.Violation{
					{Rule: rule.RuleDocumentsExist, Path: ".replworks/TASKS.md", Message: "document does not exist"},
				},
				Errors: []error{errors.New("reading .replworks/ARCHITECTURE.md: permission denied")},
			},
			wantCode:   2,
			wantStdout: "documents-exist .replworks/TASKS.md: document does not exist\nrepl: reading .replworks/ARCHITECTURE.md: permission denied\n",
		},
		{
			name:       "errors without violations",
			outcome:    checkproject.Outcome{Errors: []error{errors.New("reading AGENTS.md: permission denied")}},
			wantCode:   2,
			wantStdout: "repl: reading AGENTS.md: permission denied\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buffer bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buffer)

			code := renderCheck(cmd, tt.outcome)

			if code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d", code, tt.wantCode)
			}
			if buffer.String() != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", buffer.String(), tt.wantStdout)
			}
		})
	}
}

func writeDocuments(t *testing.T, dir, agents string) {
	t.Helper()

	texts := map[string]string{
		filepath.Join("AGENTS.md"):                     agents,
		filepath.Join(".replworks", "PRODUCT_SPEC.md"): "# PRODUCT_SPEC.md\n",
		filepath.Join(".replworks", "TECH_STACK.md"):   "# TECH_STACK.md\n",
		filepath.Join(".replworks", "ARCHITECTURE.md"): "# ARCHITECTURE.md\n",
		filepath.Join(".replworks", "TASKS.md"):        "# TASKS.md\n",
	}

	if err := os.Mkdir(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	for path, text := range texts {
		if err := os.WriteFile(filepath.Join(dir, path), []byte(text), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
}

func TestCheckReportsADifferentAgentsMD(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, "# stale\n")
	initRepository(t, dir)
	t.Chdir(dir)

	client, transport := servePage(t, string(recordedPage(t)))

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

	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
	if got := transport.requested(); len(got) != 1 {
		t.Fatalf("requests = %q, want 1", got)
	}
}

func TestCheckReportsALineEndingDifference(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, strings.ReplaceAll(recordedText(t), "\n", "\r\n"))
	initRepository(t, dir)
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "check")

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

func TestCheckDoesNotReportAnIdenticalAgentsMD(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, recordedText(t))
	initRepository(t, dir)
	t.Chdir(dir)

	client, transport := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if got := transport.requested(); len(got) != 1 {
		t.Fatalf("requests = %q, want the canonical text retrieved exactly once", got)
	}
}

func TestCheckReportsAnAbsentAgentsMDOnlyOnce(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, "# stale\n")
	if err := os.Remove(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	initRepository(t, dir)
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, "agents-canonical") {
		t.Fatalf("stdout = %q, want no agents-canonical violation", stdout)
	}

	lines := errorLines(t, stdout)
	want := []string{"documents-exist AGENTS.md: document does not exist"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestCheckReportsAnUnobtainableCanonicalText(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, "# stale\n")
	if err := os.Remove(filepath.Join(dir, ".replworks", "TASKS.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	initRepository(t, dir)
	t.Chdir(dir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	client := &http.Client{Transport: &recordingTransport{target: server.URL}}

	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	if len(lines) != 2 {
		t.Fatalf("stdout = %q, want the violation and the error", stdout)
	}
	if lines[0] != "documents-exist .replworks/TASKS.md: document does not exist" {
		t.Fatalf("line 1 = %q, want the documents-exist violation first", lines[0])
	}
	if !strings.HasPrefix(lines[1], "repl: ") {
		t.Fatalf("line 2 = %q, want the execution error second", lines[1])
	}
	if strings.Contains(stdout, "agents-canonical") {
		t.Fatalf("stdout = %q, want no agents-canonical violation", stdout)
	}

	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestCheckUsesTheGivenBaseBranch(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, recordedText(t))
	initRepository(t, dir)
	gitCommand(t, dir, "checkout", "-q", "-b", "dev")
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))
	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "check", "--base", "dev")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestCheckReportsAMissingBaseBranch(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, recordedText(t))
	initRepository(t, dir)
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "check", "--base", "nope")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	want := []string{"repl: the base branch nope does not exist in the repository"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestCheckUsesMainWithoutTheBaseOption(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, recordedText(t))
	gitCommand(t, dir, "init", "-q", "-b", "trunk")
	gitCommand(t, dir, "add", "-A")
	gitCommand(t, dir, "-c", "commit.gpgsign=false", "commit", "-q", "-m", "project")
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	want := []string{"repl: the base branch main does not exist in the repository"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestCheckAcceptsASubdirectoryOfARepository(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	writeDocuments(t, sub, recordedText(t))
	initRepository(t, dir)
	t.Chdir(sub)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "check")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestCheckRejectsAnOptionLikeBaseBranch(t *testing.T) {
	dir := t.TempDir()
	writeDocuments(t, dir, recordedText(t))
	initRepository(t, dir)
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "check", "--base=-upload-pack=/bin/false")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (stdout = %q)", code, stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "repl: the base branch") {
		t.Fatalf("stdout = %q, want the base branch error", stdout)
	}
}

func initRepository(t *testing.T, dir string) {
	t.Helper()

	gitCommand(t, dir, "init", "-q", "-b", "main")
	gitCommand(t, dir, "add", "-A")
	gitCommand(t, dir, "-c", "commit.gpgsign=false", "commit", "-q", "-m", "project")
}

func gitCommand(t *testing.T, dir string, args ...string) {
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

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, stderr.String())
	}
}
