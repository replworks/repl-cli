package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadTextPresent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got := New(dir).ReadText("AGENTS.md")
	if !got.IsPresent() {
		t.Fatalf("ReadText() kind = %v, want present", got.Kind())
	}
	if got.Value() != "hello\n" {
		t.Fatalf("ReadText() value = %q, want %q", got.Value(), "hello\n")
	}
}

func TestReadTextNested(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".replworks", "TASKS.md"), []byte("- [ ] T-001\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got := New(dir).ReadText(".replworks/TASKS.md")
	if !got.IsPresent() || got.Value() != "- [ ] T-001\n" {
		t.Fatalf("ReadText() = %+v, want the task ledger text", got)
	}
}

func TestReadTextAbsent(t *testing.T) {
	got := New(t.TempDir()).ReadText(".replworks/TASKS.md")
	if !got.IsAbsent() {
		t.Fatalf("ReadText() kind = %v, want absent", got.Kind())
	}
}

func TestReadTextFailedWhenNotReadable(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "TASKS.md"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	got := New(dir).ReadText("TASKS.md")
	if !got.IsFailed() {
		t.Fatalf("ReadText() kind = %v, want failed", got.Kind())
	}
	if got.Reason() == nil {
		t.Fatal("ReadText() reason = nil, want an error")
	}
}

func TestReadTextRejectsPathsOutsideTheProjectDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	absolute := filepath.Join(dir, "AGENTS.md")
	for _, path := range []string{"../escape", "../../escape", "sub/../../escape", "", absolute} {
		t.Run(path, func(t *testing.T) {
			got := New(dir).ReadText(path)
			if !got.IsFailed() {
				t.Fatalf("ReadText(%q) kind = %v, want failed", path, got.Kind())
			}
		})
	}
}

func TestReadTextConfinesToProjectDirectory(t *testing.T) {
	parent := t.TempDir()
	if err := os.WriteFile(filepath.Join(parent, "outside.txt"), []byte("outside\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	project := filepath.Join(parent, "project")
	if err := os.Mkdir(project, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if got := New(project).ReadText("../outside.txt"); !got.IsFailed() {
		t.Fatalf("ReadText(\"../outside.txt\") kind = %v, want failed", got.Kind())
	}
}
