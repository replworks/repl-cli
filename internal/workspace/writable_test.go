package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	project := New(dir)

	if got := project.Exists("AGENTS.md"); !got.IsPresent() || !got.Value() {
		t.Fatalf("Exists(\"AGENTS.md\") = %+v, want present true", got)
	}
	if got := project.Exists(".replworks"); !got.IsPresent() || !got.Value() {
		t.Fatalf("Exists(\".replworks\") = %+v, want present true", got)
	}
	if got := project.Exists("missing.txt"); !got.IsPresent() || got.Value() {
		t.Fatalf("Exists(\"missing.txt\") = %+v, want present false", got)
	}
	if got := project.Exists("../outside"); !got.IsFailed() {
		t.Fatalf("Exists(\"../outside\") = %+v, want failed", got)
	}
}

func TestCreateFile(t *testing.T) {
	dir := t.TempDir()

	if err := GrantWrite(New(dir)).CreateFile("AGENTS.md", "content\n"); err != nil {
		t.Fatalf("CreateFile() error = %v, want nil", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading the created file: %v", err)
	}
	if string(content) != "content\n" {
		t.Fatalf("content = %q, want %q", string(content), "content\n")
	}
}

func TestCreateFileFailsWhenTheFileExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("keep\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := GrantWrite(New(dir)).CreateFile("AGENTS.md", "overwritten\n"); err == nil {
		t.Fatal("CreateFile() error = nil, want an error")
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading the existing file: %v", err)
	}
	if string(content) != "keep\n" {
		t.Fatalf("content = %q, want the existing content to be unchanged", string(content))
	}
}

func TestCreateDirectory(t *testing.T) {
	dir := t.TempDir()

	if err := GrantWrite(New(dir)).CreateDirectory(".replworks"); err != nil {
		t.Fatalf("CreateDirectory() error = %v, want nil", err)
	}

	info, err := os.Stat(filepath.Join(dir, ".replworks"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".replworks is not a directory")
	}
}

func TestCreateDirectoryFailsWhenTheDirectoryExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := GrantWrite(New(dir)).CreateDirectory(".replworks"); err == nil {
		t.Fatal("CreateDirectory() error = nil, want an error")
	}
}

func TestReplaceFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("old\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := GrantWrite(New(dir)).ReplaceFile("AGENTS.md", "new\n"); err != nil {
		t.Fatalf("ReplaceFile() error = %v, want nil", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading the replaced file: %v", err)
	}
	if string(content) != "new\n" {
		t.Fatalf("content = %q, want %q", string(content), "new\n")
	}
}

func TestWritesRejectPathsOutsideTheProjectDirectory(t *testing.T) {
	dir := t.TempDir()
	absolute := filepath.Join(dir, "AGENTS.md")
	writable := GrantWrite(New(dir))

	for _, path := range []string{"../escape", "../../escape", "", absolute} {
		t.Run(path, func(t *testing.T) {
			if err := writable.CreateFile(path, "x"); err == nil {
				t.Fatalf("CreateFile(%q) error = nil, want an error", path)
			}
			if err := writable.CreateDirectory(path); err == nil {
				t.Fatalf("CreateDirectory(%q) error = nil, want an error", path)
			}
			if err := writable.ReplaceFile(path, "x"); err == nil {
				t.Fatalf("ReplaceFile(%q) error = nil, want an error", path)
			}
		})
	}
}
