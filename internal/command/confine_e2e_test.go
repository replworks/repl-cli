//go:build e2e

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandsAreConfinedToTheProjectDirectory(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		agents     string
		documents  bool
		changes    bool
		wantCode   int
		wantStderr string
	}{
		{name: "init", args: []string{"init"}, changes: true},
		{name: "status", args: []string{"status"}, documents: true},
		{name: "check", args: []string{"check"}, documents: true},
		{name: "sync", args: []string{"sync"}, agents: "# stale\n", documents: true, changes: true},
		{name: "unknown command", args: []string{"release"}, documents: true, wantCode: 2, wantStderr: "unknown command"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			project := filepath.Join(root, "app")
			if err := os.MkdirAll(filepath.Join(root, "other"), 0o755); err != nil {
				t.Fatalf("setup: %v", err)
			}
			if err := os.Mkdir(project, 0o755); err != nil {
				t.Fatalf("setup: %v", err)
			}
			if err := os.WriteFile(filepath.Join(root, "parent.txt"), []byte("parent\n"), 0o644); err != nil {
				t.Fatalf("setup: %v", err)
			}
			if err := os.WriteFile(filepath.Join(root, "other", "sibling.txt"), []byte("sibling\n"), 0o644); err != nil {
				t.Fatalf("setup: %v", err)
			}
			if tt.documents {
				agents := tt.agents
				if agents == "" {
					agents = liveCanonical(t)
				}
				writeDocuments(t, project, agents)
			}
			initRepository(t, root)
			t.Chdir(project)

			outside := outsideState(t, root)
			before := directoryState(t, project)
			repository := repositoryState(t, root)

			client, _ := liveRecordingClient()
			code, stdout, stderr := runCommand(t, client, tt.args...)

			if code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d (stdout = %q, stderr = %q)", code, tt.wantCode, stdout, stderr)
			}
			if tt.wantStderr != "" && !strings.Contains(stderr, tt.wantStderr) {
				t.Fatalf("stderr = %q, want %q", stderr, tt.wantStderr)
			}
			if after := outsideState(t, root); after != outside {
				t.Fatalf("files outside the project directory changed:\nbefore: %s\nafter:  %s", outside, after)
			}
			if before2 := directoryState(t, project); (before2 != before) != tt.changes {
				t.Fatalf("project directory before: %s\nafter:  %s", before, before2)
			}
			if !tt.changes {
				if after := repositoryState(t, root); after != repository {
					t.Fatalf("repository state changed:\nbefore: %s\nafter:  %s", repository, after)
				}
			}
		})
	}
}

func TestInitInASubdirectoryCreatesFilesInTheProjectDirectoryOnly(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "app")
	if err := os.Mkdir(project, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("readme\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	initRepository(t, root)
	t.Chdir(project)

	client, transport := liveRecordingClient()

	code, stdout, stderr := runCommand(t, client, "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q, stderr = %q)", code, stdout, stderr)
	}
	if got := transport.requested(); len(got) != 1 {
		t.Fatalf("requests = %q, want the canonical text retrieved exactly once", got)
	}
	for _, rel := range []string{"AGENTS.md", ".replworks"} {
		if _, err := os.Stat(filepath.Join(project, rel)); err != nil {
			t.Fatalf("project entry %s: %v", rel, err)
		}
	}
	for _, rel := range []string{"AGENTS.md", ".replworks"} {
		if _, err := os.Stat(filepath.Join(root, rel)); !os.IsNotExist(err) {
			t.Fatalf("the repository root holds %s, want the project directory only (err = %v)", rel, err)
		}
	}
}

func outsideState(t *testing.T, root string) string {
	t.Helper()

	var contents []string

	for _, rel := range []string{"parent.txt", filepath.Join("other", "sibling.txt")} {
		content, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("reading %s: %v", rel, err)
		}
		contents = append(contents, rel+"="+string(content))
	}

	return strings.Join(contents, "\n")
}
