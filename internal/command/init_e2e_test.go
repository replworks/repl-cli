//go:build e2e

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/replworks/repl-cli/internal/canon"
)

func TestInitAgainstTheLiveSource(t *testing.T) {
	canonical := canon.New(nil).Canonical()
	if !canonical.IsPresent() {
		t.Fatalf("the live canonical source is unreachable: %v", canonical.Reason())
	}
	if !strings.HasPrefix(canonical.Value(), "# AGENTS.md") {
		t.Fatalf("the live canonical text starts with %q, want the AGENTS.md heading", firstLine(canonical.Value()))
	}

	dir := t.TempDir()
	t.Chdir(dir)

	code, stdout, stderr := runInit(t, nil, "init")

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
	if string(content) != canonical.Value() {
		t.Fatal("AGENTS.md does not hold the live canonical AGENTS.md")
	}

	info, err := os.Stat(filepath.Join(dir, ".replworks"))
	if err != nil {
		t.Fatalf("stat .replworks: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".replworks is not a directory")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the project directory: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("project directory holds %d entries, want 2", len(entries))
	}

	for _, want := range []string{"created: AGENTS.md", "created: .replworks"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func firstLine(text string) string {
	if index := strings.Index(text, "\n"); index >= 0 {
		return text[:index]
	}

	return text
}
