package workspace

import (
	"fmt"
	"os"
)

// Writable is a Project Workspace with the Write capability granted.
type Writable struct {
	Workspace
}

// GrantWrite returns a Writable over the same project directory.
func GrantWrite(w Workspace) Writable {
	return Writable{Workspace: w}
}

// CreateFile creates rel with content. It fails if rel already exists.
func (w Writable) CreateFile(rel, content string) error {
	abs, err := w.resolve(rel)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(abs, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("creating %s: %w", rel, err)
	}

	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		return fmt.Errorf("writing %s: %w", rel, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("writing %s: %w", rel, err)
	}

	return nil
}

// CreateDirectory creates the directory rel.
func (w Writable) CreateDirectory(rel string) error {
	abs, err := w.resolve(rel)
	if err != nil {
		return err
	}

	if err := os.Mkdir(abs, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", rel, err)
	}

	return nil
}

// ReplaceFile stores content as the full new content of rel.
func (w Writable) ReplaceFile(rel, content string) error {
	abs, err := w.resolve(rel)
	if err != nil {
		return err
	}

	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", rel, err)
	}

	return nil
}
