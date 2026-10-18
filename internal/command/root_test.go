package command

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runApp(t *testing.T, args ...string) (int, string, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	app := App{
		Version: "dev",
		Commit:  "none",
		Date:    "unknown",
		Stdout:  &stdout,
		Stderr:  &stderr,
	}

	return app.Execute(args), stdout.String(), stderr.String()
}

func errorLines(t *testing.T, stderr string) []string {
	t.Helper()

	if stderr == "" {
		return nil
	}
	if !strings.HasSuffix(stderr, "\n") {
		t.Fatalf("stderr is not newline terminated: %q", stderr)
	}

	return strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
}

func TestExecuteRejectsUnknownCommand(t *testing.T) {
	code, stdout, stderr := runApp(t, "bogus")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}

	lines := errorLines(t, stderr)
	if len(lines) != 1 {
		t.Fatalf("stderr has %d lines (%q), want 1", len(lines), stderr)
	}
	if !strings.Contains(lines[0], "unknown command") {
		t.Fatalf("error line = %q, want it to mention an unknown command", lines[0])
	}
}

func TestExecuteRejectsUnknownOption(t *testing.T) {
	code, stdout, stderr := runApp(t, "--bogus")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}

	lines := errorLines(t, stderr)
	if len(lines) != 1 {
		t.Fatalf("stderr has %d lines (%q), want 1", len(lines), stderr)
	}
	if !strings.Contains(lines[0], "unknown flag") {
		t.Fatalf("error line = %q, want it to mention an unknown flag", lines[0])
	}
}

func TestExecutePrintsHelpWithoutArguments(t *testing.T) {
	code, stdout, stderr := runApp(t)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Fatalf("stdout = %q, want usage output", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
}

func TestExecuteTouchesNoFileOnRejection(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("keep\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Chdir(dir)
	before := directoryState(t, dir)

	for _, args := range [][]string{{"bogus"}, {"--bogus"}} {
		if code, _, _ := runApp(t, args...); code != 2 {
			t.Fatalf("Execute(%q) exit code = %d, want 2", args, code)
		}
	}

	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name string
		app  App
		want string
	}{
		{
			name: "release build without commit",
			app:  App{Version: "v1.2.3", Commit: "none", Date: "unknown"},
			want: "v1.2.3",
		},
		{
			name: "release build with commit",
			app:  App{Version: "v1.2.3", Commit: "abc1234", Date: "2026-01-01T00:00:00Z"},
			want: "v1.2.3 (abc1234, 2026-01-01T00:00:00Z)",
		},
		{
			name: "development build",
			app:  App{Version: "dev", Commit: "", Date: ""},
			want: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.app.versionString(); got != tt.want {
				t.Fatalf("versionString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func directoryState(t *testing.T, dir string) string {
	t.Helper()

	var entries []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if entry.IsDir() {
			entries = append(entries, rel+"/")
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		entries = append(entries, rel+"="+string(content))
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}

	return strings.Join(entries, "\n")
}
