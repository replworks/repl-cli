//go:build e2e

package command

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/replworks/repl-cli/internal/canon"
)

func TestSyncAgainstTheLiveAutomaticSource(t *testing.T) {
	want := liveDocument(t, canon.AgentsAutoDocument)
	wantScript := liveAskHumanScript(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# stale\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	code, stdout, stderr := runCommandWithInput(t, nil, "2\n", "sync")

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
		t.Fatal("AGENTS.md does not hold the selected live document")
	}
	script, err := os.ReadFile(filepath.Join(dir, ".replworks/scripts/ask-human.sh"))
	if err != nil {
		t.Fatalf("reading ask-human.sh: %v", err)
	}
	if string(script) != wantScript {
		t.Fatal("ask-human.sh does not hold the exact live script")
	}
}

func TestSyncAgainstTheLiveSourceReportsUpToDate(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(want), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, nil, "sync")

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
}

func TestSyncAgainstTheLiveSourceReplacesADifferentAgentsMD(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# stale\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "sync")

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
	if string(content) != want {
		t.Fatal("AGENTS.md does not hold the live canonical AGENTS.md")
	}
	if !strings.Contains(stdout, "- # stale") {
		t.Fatalf("stdout = %q, want the removed line of the difference", stdout)
	}

	code, stdout, stderr = runCommand(t, nil, "sync")
	if code != 0 {
		t.Fatalf("second exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if !strings.Contains(stdout, "up to date") {
		t.Fatalf("second stdout = %q, want an up-to-date indication", stdout)
	}
}

func TestSyncAgainstTheLiveSourceReplacesADifferentLineEnding(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(strings.ReplaceAll(want, "\n", "\r\n")), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if !strings.Contains(stdout, "- ") || !strings.Contains(stdout, "+ ") {
		t.Fatalf("stdout = %q, want the difference lines", stdout)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != want {
		t.Fatal("AGENTS.md does not hold the live canonical AGENTS.md")
	}
}

func TestSyncAgainstTheLiveSourceKeepsAGENTSMDWhenTheSourceIsUnobtainable(t *testing.T) {
	liveCanonical(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# local\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	before := directoryState(t, dir)
	client := &http.Client{Transport: &recordingTransport{target: "http://127.0.0.1:1"}}

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
