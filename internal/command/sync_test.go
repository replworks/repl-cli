package command

import (
	"bytes"
	"encoding/json"
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

func TestRenderSyncColorsDifferenceLinesForTerminalOutput(t *testing.T) {
	var buffer bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buffer)

	code := renderSyncOutput(cmd, syncproject.Outcome{
		Difference: []diff.Entry{
			{Kind: diff.Removed, Line: "old"},
			{Kind: diff.Added, Line: "new"},
		},
	}, true)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	want := "\x1b[31m- old\x1b[0m\n\x1b[32m+ new\x1b[0m\n"
	if buffer.String() != want {
		t.Fatalf("output = %q, want %q", buffer.String(), want)
	}
}

func TestRenderSyncLeavesDifferenceLinesPlainForNonTerminalOutput(t *testing.T) {
	var buffer bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buffer)

	renderSyncOutput(cmd, syncproject.Outcome{
		Difference: []diff.Entry{
			{Kind: diff.Removed, Line: "old"},
			{Kind: diff.Added, Line: "new"},
		},
	}, false)

	if got, want := buffer.String(), "- old\n+ new\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSyncUsesTheCanonicalDocumentAndInstallsTheScriptWithoutSelection(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# stale\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	want := "# AGENTS.md\n\n## AUTO_RUN\n"
	client, transport := servePage(t, want)

	code, stdout, stderr := runCommandWithInput(t, client, "ignored input\n", "sync")

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
	if len(requests) != 2 || requests[0] != string(canon.AgentsDocument) || requests[1] != string(canon.AskHumanScript) {
		t.Fatalf("requests = %q, want canonical document followed by %q", requests, canon.AskHumanScript)
	}
	if content, err := os.ReadFile(filepath.Join(dir, ".replworks/scripts/ask-human.sh")); err != nil || string(content) != want {
		t.Fatalf("ask-human.sh = %q, err = %v, want %q", content, err, want)
	}
	info, err := os.Stat(filepath.Join(dir, ".replworks/scripts/ask-human.sh"))
	if err != nil {
		t.Fatalf("stat ask-human.sh: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("ask-human.sh permissions = %o, want an executable file", info.Mode().Perm())
	}
}

func TestSyncDoesNotReadInputForDocumentSelection(t *testing.T) {
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
	if strings.Contains(stdout, "Select an AGENTS.md document") || strings.Contains(stdout, "Invalid selection") {
		t.Fatalf("stdout = %q, want no selection prompt", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestSyncMakesAnExistingAutomaticScriptExecutableWhenReplacingAgents(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".replworks/scripts"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# stale\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".replworks/scripts/ask-human.sh"), []byte("script\n"), 0o640); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, _ := servePage(t, "# AGENTS.md\n")
	code, _, stderr := runCommandWithInput(t, client, "2\n", "sync")

	if code != 0 || stderr != "" {
		t.Fatalf("code = %d, stderr = %q, want success", code, stderr)
	}
	info, err := os.Stat(filepath.Join(dir, ".replworks/scripts/ask-human.sh"))
	if err != nil {
		t.Fatalf("stat ask-human.sh: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("ask-human.sh permissions = %o, want an executable file", info.Mode().Perm())
	}
}

func TestSyncWithoutAGENTSDoesNotReadInputOrRequestNetwork(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, transport := servePage(t, "# remote\n")

	code, stdout, _ := runCommandWithInput(t, client, "", "sync")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(stdout, "AGENTS.md does not exist") {
		t.Fatalf("stdout = %q, want the missing-file error", stdout)
	}
	if got := transport.requested(); len(got) != 0 {
		t.Fatalf("requests = %q, want none", got)
	}
}

func TestSyncReportsUpToDate(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(recordedText(t)), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, transport := servePage(t, string(recordedPage(t)))
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
	if content, err := os.ReadFile(filepath.Join(dir, ".replworks/scripts/ask-human.sh")); err != nil || string(content) != recordedText(t) {
		t.Fatalf("ask-human.sh = %q, err = %v, want the canonical live script source", content, err)
	}
	if got := transport.requested(); len(got) != 2 {
		t.Fatalf("requests = %q, want canonical document and helper script", got)
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

func TestSyncReplacesTheScriptEvenWhenAgentsIsUpToDate(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".replworks/scripts"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	want := "script\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(want), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".replworks/scripts/ask-human.sh"), []byte("keep\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, transport := servePage(t, "script\n")
	code, stdout, stderr := runCommandWithInput(t, client, "", "sync")

	if code != 0 || stderr != "" || !strings.Contains(stdout, "up to date") {
		t.Fatalf("code = %d, stdout = %q, stderr = %q", code, stdout, stderr)
	}
	requests := transport.requested()
	if len(requests) != 2 || requests[0] != string(canon.AgentsDocument) || requests[1] != string(canon.AskHumanScript) {
		t.Fatalf("requests = %q, want canonical document and helper script", requests)
	}
	content, err := os.ReadFile(filepath.Join(dir, ".replworks/scripts/ask-human.sh"))
	if err != nil || string(content) != "script\n" {
		t.Fatalf("ask-human.sh = %q, err = %v, want published contents", content, err)
	}
}

func newJSONCommand(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	var buffer bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&buffer)
	cmd.Flags().Bool("json", true, "print one JSON object")

	return cmd, &buffer
}

func parseJSON(t *testing.T, text string) map[string]any {
	t.Helper()

	var value map[string]any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		t.Fatalf("output is not valid JSON: %v (output = %q)", err, text)
	}

	return value
}

func TestRenderSyncJSONUpToDate(t *testing.T) {
	cmd, buffer := newJSONCommand(t)

	code := renderSync(cmd, syncproject.Outcome{
		UpToDate:    true,
		ScriptPath:  ".replworks/scripts/ask-human.sh",
		ScriptState: "replaced",
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (output = %q)", code, buffer.String())
	}

	value := parseJSON(t, buffer.String())
	if value["up_to_date"] != true {
		t.Fatalf("up_to_date = %v, want true", value["up_to_date"])
	}
	if diff, ok := value["diff"].([]any); !ok || len(diff) != 0 {
		t.Fatalf("diff = %v, want empty array", value["diff"])
	}
	script, ok := value["script"].(map[string]any)
	if !ok {
		t.Fatalf("script = %v, want an object", value["script"])
	}
	if script["path"] != ".replworks/scripts/ask-human.sh" || script["state"] != "replaced" {
		t.Fatalf("script = %v, want the helper script", script)
	}
	if errors, ok := value["errors"].([]any); !ok || len(errors) != 0 {
		t.Fatalf("errors = %v, want empty array", value["errors"])
	}
}

func TestRenderSyncJSONWithDifference(t *testing.T) {
	cmd, buffer := newJSONCommand(t)

	code := renderSync(cmd, syncproject.Outcome{
		Difference: []diff.Entry{
			{Kind: diff.Removed, Line: "old"},
			{Kind: diff.Added, Line: "new"},
		},
		ScriptPath:  ".replworks/scripts/ask-human.sh",
		ScriptState: "created",
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (output = %q)", code, buffer.String())
	}

	value := parseJSON(t, buffer.String())
	if value["up_to_date"] != false {
		t.Fatalf("up_to_date = %v, want false", value["up_to_date"])
	}
	entries, ok := value["diff"].([]any)
	if !ok || len(entries) != 2 {
		t.Fatalf("diff = %v, want two entries", value["diff"])
	}
	removed, ok := entries[0].(map[string]any)
	if !ok || removed["kind"] != "removed" || removed["line"] != "old" {
		t.Fatalf("diff[0] = %v, want the removed line", entries[0])
	}
	added, ok := entries[1].(map[string]any)
	if !ok || added["kind"] != "added" || added["line"] != "new" {
		t.Fatalf("diff[1] = %v, want the added line", entries[1])
	}
	if script, ok := value["script"].(map[string]any); !ok || script["state"] != "created" {
		t.Fatalf("script = %v, want state created", value["script"])
	}
}

func TestRenderSyncJSONWithErrors(t *testing.T) {
	cmd, buffer := newJSONCommand(t)

	code := renderSync(cmd, syncproject.Outcome{
		Errors: []error{errors.New("boom")},
	})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (output = %q)", code, buffer.String())
	}

	value := parseJSON(t, buffer.String())
	if value["script"] != nil {
		t.Fatalf("script = %v, want null", value["script"])
	}
	errors, ok := value["errors"].([]any)
	if !ok || len(errors) != 1 {
		t.Fatalf("errors = %v, want one error", value["errors"])
	}
	if errors[0] != "boom" {
		t.Fatalf("errors[0] = %v, want boom", errors[0])
	}
}
