package command

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/replworks/repl-cli/internal/checkproject"
	"github.com/replworks/repl-cli/internal/rule"
)

func TestCheckReportsMissingDocuments(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# AGENTS.md\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, nil, "check")

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
	writeDocuments(t, dir)

	t.Chdir(dir)

	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, nil, "check")

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
	writeDocuments(t, dir)
	if err := os.RemoveAll(filepath.Join(dir, ".replworks", "TECH_STACK.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, ".replworks", "ARCHITECTURE.md")); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".replworks", "ARCHITECTURE.md"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Chdir(dir)

	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, nil, "check")

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

func writeDocuments(t *testing.T, dir string) {
	t.Helper()

	paths := []string{
		"AGENTS.md",
		filepath.Join(".replworks", "PRODUCT_SPEC.md"),
		filepath.Join(".replworks", "TECH_STACK.md"),
		filepath.Join(".replworks", "ARCHITECTURE.md"),
		filepath.Join(".replworks", "TASKS.md"),
	}

	if err := os.Mkdir(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	for _, path := range paths {
		if err := os.WriteFile(filepath.Join(dir, path), []byte("# "+path+"\n"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
}
