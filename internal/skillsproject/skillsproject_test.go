package skillsproject

import (
	"errors"
	"strings"
	"testing"

	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/skill"
)

type fakeSource struct {
	bodies map[string]string
	fail   map[string]bool
}

func (s fakeSource) Fetch(url string) fact.Fact[string] {
	for name, fail := range s.fail {
		if fail && containsSkillName(url, name) {
			return fact.Failed[string](errors.New("unreachable"))
		}
	}

	for _, src := range skill.Sources {
		if src.URL == url {
			return fact.Present(s.bodies[src.Name])
		}
	}

	return fact.Failed[string](errors.New("unknown skill URL"))
}

func containsSkillName(url, name string) bool {
	return strings.HasSuffix(url, "/"+name+"/SKILL.md")
}

type memWorkspace struct {
	files     map[string]string
	dirs      map[string]bool
	failWrite string
	failRead  string
}

func newMemWorkspace() *memWorkspace {
	return &memWorkspace{files: map[string]string{}, dirs: map[string]bool{}}
}

func (w *memWorkspace) ReadText(rel string) fact.Fact[string] {
	if rel == w.failRead {
		return fact.Failed[string](errors.New("read failure"))
	}
	if content, ok := w.files[rel]; ok {
		return fact.Present(content)
	}

	return fact.Absent[string]()
}

func (w *memWorkspace) Exists(rel string) fact.Fact[bool] {
	if _, ok := w.dirs[rel]; ok {
		return fact.Present(true)
	}
	if _, ok := w.files[rel]; ok {
		return fact.Present(true)
	}

	return fact.Present(false)
}

func (w *memWorkspace) CreateDirectory(rel string) error {
	if _, ok := w.dirs[rel]; ok {
		return errors.New(rel + " already exists")
	}
	w.dirs[rel] = true

	return nil
}

func (w *memWorkspace) CreateFile(rel, content string) error {
	if rel == w.failWrite {
		return errors.New("write failure")
	}
	if _, ok := w.files[rel]; ok {
		return errors.New(rel + " already exists")
	}
	w.files[rel] = content

	return nil
}

func (w *memWorkspace) ReplaceFile(rel, content string) error {
	if rel == w.failWrite {
		return errors.New("write failure")
	}
	w.files[rel] = content

	return nil
}

func skillBodies() map[string]string {
	bodies := map[string]string{}
	for _, src := range skill.Sources {
		bodies[src.Name] = "# " + src.Name + "\n"
	}

	return bodies
}

func TestRunInstallsTheFiveSkills(t *testing.T) {
	workspace := newMemWorkspace()
	outcome := Coordinator{
		Source:    fakeSource{bodies: skillBodies()},
		Workspace: workspace,
	}.Run()

	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	if len(outcome.Items) != len(skill.Sources) {
		t.Fatalf("items = %d, want %d", len(outcome.Items), len(skill.Sources))
	}
	for i, src := range skill.Sources {
		content, ok := workspace.files[src.Path]
		if !ok {
			t.Fatalf("%s was not installed", src.Path)
		}
		if content != "# "+src.Name+"\n" {
			t.Errorf("%s = %q, want the skill body", src.Path, content)
		}
		if outcome.Items[i].State != Created {
			t.Errorf("%s state = %q, want created", src.Name, outcome.Items[i].State)
		}
		if outcome.Items[i].Name != src.Name || outcome.Items[i].Path != src.Path {
			t.Errorf("%s result = %#v, want name and path to match", src.Name, outcome.Items[i])
		}
	}
	if !workspace.dirs[".claude"] || !workspace.dirs[".claude/skills"] || !workspace.dirs[".claude/skills/repl-product-spec"] {
		t.Error("parent directories were not created")
	}
}

func TestRunReportsUpToDateOnRerun(t *testing.T) {
	workspace := newMemWorkspace()
	first := Coordinator{
		Source:    fakeSource{bodies: skillBodies()},
		Workspace: workspace,
	}.Run()
	if len(first.Errors) != 0 {
		t.Fatalf("setup: errors = %v", first.Errors)
	}
	before := len(workspace.files)

	second := Coordinator{
		Source:    fakeSource{bodies: skillBodies()},
		Workspace: workspace,
	}.Run()

	if len(second.Errors) != 0 {
		t.Fatalf("errors = %v, want none", second.Errors)
	}
	if len(workspace.files) != before {
		t.Fatal("a rerun created a file")
	}
	for _, item := range second.Items {
		if item.State != UpToDate {
			t.Errorf("%s state = %q, want up to date", item.Name, item.State)
		}
	}
}

func TestRunReplacesOnlyDifferingFiles(t *testing.T) {
	workspace := newMemWorkspace()
	outcome := Coordinator{
		Source:    fakeSource{bodies: skillBodies()},
		Workspace: workspace,
	}.Run()
	if len(outcome.Errors) != 0 {
		t.Fatalf("setup: errors = %v", outcome.Errors)
	}
	workspace.files[".claude/skills/repl-tasks/SKILL.md"] = "custom"

	outcome = Coordinator{
		Source:    fakeSource{bodies: skillBodies()},
		Workspace: workspace,
	}.Run()

	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	for _, item := range outcome.Items {
		want := UpToDate
		if item.Name == "repl-tasks" {
			want = Replaced
		}
		if item.State != want {
			t.Errorf("%s state = %q, want %q", item.Name, item.State, want)
		}
	}
	if workspace.files[".claude/skills/repl-tasks/SKILL.md"] != "# repl-tasks\n" {
		t.Fatal("the differing skill file was not replaced")
	}
}

func TestRunFetchFailureWritesNothing(t *testing.T) {
	workspace := newMemWorkspace()

	outcome := Coordinator{
		Source:    fakeSource{bodies: skillBodies(), fail: map[string]bool{"repl-auto-run": true}},
		Workspace: workspace,
	}.Run()

	if len(outcome.Errors) != 1 {
		t.Fatalf("errors = %v, want one error", outcome.Errors)
	}
	if len(workspace.files) != 0 {
		t.Fatal("a fetch failure wrote a file")
	}
	if len(workspace.dirs) != 0 {
		t.Fatal("a fetch failure created a directory")
	}
}

func TestRunFetchFailureReportsEveryFailedSkill(t *testing.T) {
	workspace := newMemWorkspace()

	outcome := Coordinator{
		Source: fakeSource{
			bodies: skillBodies(),
			fail:   map[string]bool{"repl-product-spec": true, "repl-tasks": true},
		},
		Workspace: workspace,
	}.Run()

	if len(outcome.Errors) != 2 {
		t.Fatalf("errors = %v, want two errors", outcome.Errors)
	}
	if len(workspace.files) != 0 {
		t.Fatal("a fetch failure wrote a file")
	}
}

func TestRunStopsAtTheFirstWriteFailureAndReportsInstalledSkills(t *testing.T) {
	workspace := newMemWorkspace()
	workspace.failWrite = ".claude/skills/repl-architecture/SKILL.md"

	outcome := Coordinator{
		Source:    fakeSource{bodies: skillBodies()},
		Workspace: workspace,
	}.Run()

	if len(outcome.Errors) != 1 {
		t.Fatalf("errors = %v, want one error", outcome.Errors)
	}
	if len(outcome.Items) != 2 {
		t.Fatalf("items = %d, want the two skills installed before the failure", len(outcome.Items))
	}
	if _, ok := workspace.files[".claude/skills/repl-architecture/SKILL.md"]; ok {
		t.Fatal("the failing skill was installed")
	}
	if _, ok := workspace.files[".claude/skills/repl-tasks/SKILL.md"]; ok {
		t.Fatal("a skill after the failure was installed")
	}
}

func TestRunStopsOnAFailedLocalReadAndReportsInstalledSkills(t *testing.T) {
	workspace := newMemWorkspace()
	workspace.failRead = ".claude/skills/repl-architecture/SKILL.md"

	outcome := Coordinator{
		Source:    fakeSource{bodies: skillBodies()},
		Workspace: workspace,
	}.Run()

	if len(outcome.Errors) != 1 {
		t.Fatalf("errors = %v, want one error", outcome.Errors)
	}
	if len(outcome.Items) != 2 {
		t.Fatalf("items = %d, want the two skills installed before the failure", len(outcome.Items))
	}
}
