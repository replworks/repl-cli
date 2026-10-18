package command

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/canon"
	"github.com/replworks/repl-cli/internal/diff"
	"github.com/replworks/repl-cli/internal/syncproject"
)

func renderSyncOutcome(t *testing.T, outcome syncproject.Outcome) (int, string) {
	t.Helper()

	var buffer bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buffer)

	return renderSync(cmd, outcome), buffer.String()
}

func TestSyncUsesTheAutomaticDocumentWhenSelected(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# stale\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	want := "# AGENTS.md\n\n## AUTO_RUN\n"
	client, transport := servePage(t, want)

	code, stdout, stderr := runCommandWithInput(t, client, "2\n", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q, stderr = %q)", code, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != want {
		t.Fatalf("AGENTS.md = %q, want %q", content, want)
	}
	requests := transport.requested()
	if len(requests) != 1 || requests[0] != string(canon.AgentsAutoDocument) {
		t.Fatalf("requests = %q, want one request to %q", requests, canon.AgentsAutoDocument)
	}
}

func TestSyncRepeatsAfterAnInvalidDocumentSelection(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# stale\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	want := "# AGENTS.md\n\n## AUTO_RUN\n"
	client, _ := servePage(t, want)

	code, stdout, stderr := runCommandWithInput(t, client, "3\n2\n", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q, stderr = %q)", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "Invalid selection. Please enter 1 or 2.") {
		t.Fatalf("stdout = %q, want the invalid-selection message", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestSyncFailsWithoutASelectionBeforeNetworkOrWrite(t *testing.T) {
	dir := t.TempDir()
	local := "# stale\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(local), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, transport := servePage(t, "# remote\n")

	code, _, stderr := runCommandWithInput(t, client, "", "sync")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stderr, "selection is required") {
		t.Fatalf("stderr = %q, want the missing-selection error", stderr)
	}
	if got := transport.requested(); len(got) != 0 {
		t.Fatalf("requests = %q, want none", got)
	}
	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != local {
		t.Fatalf("AGENTS.md = %q, want unchanged %q", content, local)
	}
}

func TestSyncReportsUpToDate(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(recordedText(t)), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, transport := servePage(t, string(recordedPage(t)))
	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "up to date") {
		t.Fatalf("stdout = %q, want an up-to-date indication", stdout)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
	if got := transport.requested(); len(got) != 1 {
		t.Fatalf("requests = %q, want 1", got)
	}
}

func TestSyncReplacesADifferentAgentsMd(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# local\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != recordedText(t) {
		t.Fatal("AGENTS.md does not hold the canonical AGENTS.md")
	}
	if !strings.Contains(stdout, "- # local") {
		t.Fatalf("stdout = %q, want the removed line of the difference", stdout)
	}
	if !strings.Contains(stdout, "+ # AGENTS.md") {
		t.Fatalf("stdout = %q, want the added line of the difference", stdout)
	}
}

func TestSyncFailsWhenTheCanonicalSourceIsUnavailable(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# local\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	client := &http.Client{Transport: &recordingTransport{target: server.URL}}

	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "sync")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if count := strings.Count(stdout, "repl:"); count != 1 {
		t.Fatalf("stdout has %d error lines (%q), want 1", count, stdout)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestSyncRejectsArguments(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "sync", "AGENTS.md")

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

func TestSyncTreatsALineEndingDifferenceAsDifferent(t *testing.T) {
	local := strings.ReplaceAll(recordedText(t), "\n", "\r\n")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(local), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if !strings.Contains(stdout, "- ") || !strings.Contains(stdout, "+ ") {
		t.Fatalf("stdout = %q, want the difference lines", stdout)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != recordedText(t) {
		t.Fatal("AGENTS.md does not hold the canonical AGENTS.md")
	}
}

func TestSyncTreatsATrailingWhitespaceDifferenceAsDifferent(t *testing.T) {
	local := strings.ReplaceAll(recordedText(t), "\n", " \n")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(local), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runCommand(t, client, "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if !strings.Contains(stdout, "- ") || !strings.Contains(stdout, "+ ") {
		t.Fatalf("stdout = %q, want the difference lines", stdout)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != recordedText(t) {
		t.Fatal("AGENTS.md does not hold the canonical AGENTS.md")
	}
}

func TestSyncOutputsTheSameDifferenceAcrossRuns(t *testing.T) {
	const local = "# local\n\nbody\n"

	run := func() string {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(local), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
		t.Chdir(dir)

		client, _ := servePage(t, string(recordedPage(t)))

		code, stdout, stderr := runCommand(t, client, "sync")
		if code != 0 {
			t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
		}

		return stdout
	}

	first, second := run(), run()

	if first != second {
		t.Fatalf("differences differ across runs:\nfirst:  %q\nsecond: %q", first, second)
	}
}

func TestSyncFailsWithoutAGENTSMD(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, transport := servePage(t, string(recordedPage(t)))
	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "sync")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if count := strings.Count(stdout, "repl:"); count != 1 {
		t.Fatalf("stdout has %d error lines (%q), want 1", count, stdout)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
	if got := transport.requested(); len(got) != 0 {
		t.Fatalf("requests = %q, want no request", got)
	}
}

func TestSyncFailsWhenAGENTSCannotBeRead(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "AGENTS.md"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, transport := servePage(t, string(recordedPage(t)))
	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, client, "sync")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if count := strings.Count(stdout, "repl:"); count != 1 {
		t.Fatalf("stdout has %d error lines (%q), want 1", count, stdout)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
	if got := transport.requested(); len(got) != 0 {
		t.Fatalf("requests = %q, want no request", got)
	}
}

func TestSyncReportsAReplacementFailureWithTheDifference(t *testing.T) {
	dir := t.TempDir()
	agents := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agents, []byte("# local\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	if err := os.Chmod(agents, 0o444); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(agents, 0o644)
	})

	code, stdout, stderr := runCommand(t, client, "sync")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "- # local") {
		t.Fatalf("stdout = %q, want the difference lines", stdout)
	}
	if got := strings.Count(stdout, "repl: "); got != 1 {
		t.Fatalf("stdout = %q, want one error line, got %d", stdout, got)
	}

	content, err := os.ReadFile(agents)
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != "# local\n" {
		t.Fatalf("AGENTS.md = %q, want the local content to be unchanged", content)
	}
}

func TestRenderSync(t *testing.T) {
	tests := []struct {
		name       string
		outcome    syncproject.Outcome
		wantCode   int
		wantStdout string
	}{
		{
			name:       "up to date",
			outcome:    syncproject.Outcome{UpToDate: true},
			wantCode:   0,
			wantStdout: "AGENTS.md is up to date\n",
		},
		{
			name: "replaced",
			outcome: syncproject.Outcome{Difference: []diff.Entry{
				{Kind: diff.Removed, Line: "old", Position: 1},
				{Kind: diff.Added, Line: "new", Position: 1},
			}},
			wantCode:   0,
			wantStdout: "- old\n+ new\n",
		},
		{
			name: "replacement failed",
			outcome: syncproject.Outcome{
				Difference: []diff.Entry{{Kind: diff.Added, Line: "new", Position: 1}},
				Errors:     []error{errors.New("permission denied")},
			},
			wantCode:   2,
			wantStdout: "+ new\nrepl: permission denied\n",
		},
		{
			name:       "execution error without a difference",
			outcome:    syncproject.Outcome{Errors: []error{errors.New("AGENTS.md does not exist")}},
			wantCode:   2,
			wantStdout: "repl: AGENTS.md does not exist\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout := renderSyncOutcome(t, tt.outcome)

			if code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d", code, tt.wantCode)
			}
			if stdout != tt.wantStdout {
				t.Fatalf("stdout = %q, want %q", stdout, tt.wantStdout)
			}
		})
	}
}
