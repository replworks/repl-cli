//go:build e2e

package command

import (
	"net/http"
	"os"
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
