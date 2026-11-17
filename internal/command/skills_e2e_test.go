//go:build e2e

package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/replworks/repl-cli/internal/skill"
)

func liveSkill(t *testing.T, name string) string {
	t.Helper()

	fact := skill.New(nil).Fetch("https://www.repl.net/skills/" + name + "/SKILL.md")
	if !fact.IsPresent() {
		t.Fatalf("the live skill source %s is unreachable: %v", name, fact.Reason())
	}

	return fact.Value()
}

func TestSkillsSyncAgainstTheLiveSkillURLs(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "skills", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q, stderr = %q)", code, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	for _, src := range skill.Sources {
		content, err := os.ReadFile(filepath.Join(dir, src.Path))
		if err != nil {
			t.Fatalf("reading %s: %v", src.Path, err)
		}
		want := liveSkill(t, src.Name)
		if string(content) != want {
			t.Errorf("%s does not hold the exact live skill file", src.Path)
		}
		if !strings.Contains(stdout, src.Name+": "+src.Path+": created\n") {
			t.Errorf("stdout = %q, want a created line for %s", stdout, src.Name)
		}
	}
}

func TestSkillsSyncAgainstTheLiveSkillURLsIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if code, _, stderr := runCommand(t, nil, "skills", "sync"); code != 0 {
		t.Fatalf("setup: exit code = %d (stderr = %q)", code, stderr)
	}

	code, stdout, stderr := runCommand(t, nil, "skills", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q, stderr = %q)", code, stdout, stderr)
	}
	for _, src := range skill.Sources {
		if !strings.Contains(stdout, src.Name+": "+src.Path+": up to date\n") {
			t.Errorf("stdout = %q, want an up to date line for %s", stdout, src.Name)
		}
	}
}
