package workspace

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/replworks/repl-cli/internal/fact"
)

type Workspace struct {
	root string
}

func New(root string) Workspace {
	return Workspace{root: root}
}

func (w Workspace) ReadText(rel string) fact.Fact[string] {
	abs, err := w.resolve(rel)
	if err != nil {
		return fact.Failed[string](err)
	}

	content, err := os.ReadFile(abs)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fact.Absent[string]()
		}
		return fact.Failed[string](fmt.Errorf("reading %s: %w", rel, err))
	}

	return fact.Present(string(content))
}

func (w Workspace) resolve(rel string) (string, error) {
	if rel == "" {
		return "", errors.New("path is empty")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("path %q is outside the project directory", rel)
	}

	root, err := filepath.Abs(w.root)
	if err != nil {
		return "", fmt.Errorf("resolving the project directory: %w", err)
	}

	abs := filepath.Join(root, rel)
	inside, err := filepath.Rel(root, abs)
	if err != nil {
		return "", fmt.Errorf("resolving %q: %w", rel, err)
	}
	if inside == ".." || strings.HasPrefix(inside, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q is outside the project directory", rel)
	}

	return abs, nil
}
