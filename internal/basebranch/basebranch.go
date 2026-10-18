// Package basebranch reads TASKS.md from a branch of the project's repository.
package basebranch

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/replworks/repl-cli/internal/fact"
)

const tasksPath = ".replworks/TASKS.md"

// Reader reads the base TASKS.md through the git executable.
type Reader struct {
	dir string
}

func New(dir string) Reader {
	return Reader{dir: dir}
}

// Tasks returns the TASKS.md of the given branch at the same project-relative
// location as in the working tree. It never changes repository state.
func (r Reader) Tasks(branch string) fact.Fact[string] {
	top, err := r.output("rev-parse", "--show-toplevel")
	if err != nil {
		return fact.Failed[string](err)
	}

	prefix, err := r.output("rev-parse", "--show-prefix")
	if err != nil {
		return fact.Failed[string](err)
	}

	root := Reader{dir: strings.TrimSpace(top)}

	ref := "refs/heads/" + branch
	if _, err := root.output("rev-parse", "--verify", "--quiet", ref); err != nil {
		return fact.Failed[string](fmt.Errorf("the base branch %s does not exist in the repository", branch))
	}

	path := strings.TrimSuffix(strings.TrimSuffix(prefix, "\n"), "\r") + tasksPath

	entries, err := root.output("ls-tree", ref, "--", path)
	if err != nil {
		return fact.Failed[string](err)
	}
	if strings.TrimSpace(entries) == "" {
		return fact.Absent[string]()
	}

	text, err := root.output("show", ref+":"+path)
	if err != nil {
		return fact.Failed[string](err)
	}

	return fact.Present(text)
}

func (r Reader) output(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}

		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), message)
	}

	return stdout.String(), nil
}
