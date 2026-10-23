package syncproject

import (
	"errors"
	"strings"
	"testing"

	"github.com/replworks/repl-cli/internal/diff"
	"github.com/replworks/repl-cli/internal/fact"
)

type fakeSource struct {
	canonical fact.Fact[string]
	calls     int
}

func (f *fakeSource) Canonical() fact.Fact[string] {
	f.calls++

	return f.canonical
}

type fakeWorkspace struct {
	local      fact.Fact[string]
	replaceErr error

	events   []string
	replaced []string
}

func (f *fakeWorkspace) ReadText(rel string) fact.Fact[string] {
	f.events = append(f.events, "read "+rel)

	return f.local
}

func (f *fakeWorkspace) Exists(rel string) fact.Fact[bool] {
	f.events = append(f.events, "exists "+rel)
	return fact.Present(false)
}

func (f *fakeWorkspace) CreateDirectory(rel string) error {
	f.events = append(f.events, "create-directory "+rel)
	return nil
}

func (f *fakeWorkspace) CreateFile(rel, content string) error {
	f.events = append(f.events, "create-file "+rel)
	return nil
}

func (f *fakeWorkspace) MakeExecutable(rel string) error {
	f.events = append(f.events, "make-executable "+rel)
	return nil
}

func (f *fakeWorkspace) DeleteFile(rel string) error {
	f.events = append(f.events, "delete-file "+rel)
	return nil
}

func (f *fakeWorkspace) ReplaceFile(rel, content string) error {
	f.events = append(f.events, "replace "+rel)
	if f.replaceErr != nil {
		return f.replaceErr
	}
	f.replaced = append(f.replaced, content)

	return nil
}

func (f *fakeWorkspace) trace() string {
	return strings.Join(f.events, ", ")
}

func coordinator(local fact.Fact[string], canonical fact.Fact[string]) (Coordinator, *fakeSource, *fakeWorkspace) {
	source := &fakeSource{canonical: canonical}
	workspace := &fakeWorkspace{local: local}

	return Coordinator{Source: source, Workspace: workspace, Difference: diff.Calculator{}}, source, workspace
}

func TestRunReportsUpToDate(t *testing.T) {
	coord, source, workspace := coordinator(fact.Present("same\n"), fact.Present("same\n"))

	outcome := coord.Run()

	if !outcome.UpToDate {
		t.Fatalf("outcome = %+v, want up to date", outcome)
	}
	if len(outcome.Difference) > 0 || len(outcome.Errors) > 0 {
		t.Fatalf("outcome = %+v, want no difference and no errors", outcome)
	}
	if len(workspace.replaced) > 0 {
		t.Fatalf("replaced = %q, want no write", workspace.replaced)
	}
	if source.calls != 1 {
		t.Fatalf("canonical source called %d times, want 1", source.calls)
	}
	if workspace.trace() != "read AGENTS.md" {
		t.Fatalf("trace = %q, want only the local read", workspace.trace())
	}
}

func TestRunReplacesWhenTheTextsDiffer(t *testing.T) {
	coord, _, workspace := coordinator(fact.Present("old\n"), fact.Present("new\n"))

	outcome := coord.Run()

	if outcome.UpToDate {
		t.Fatal("outcome = up to date, want a replacement")
	}
	if len(outcome.Errors) > 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	want := []diff.Entry{{Kind: diff.Removed, Line: "old", Position: 1}, {Kind: diff.Added, Line: "new", Position: 1}}
	if len(outcome.Difference) != len(want) || outcome.Difference[0] != want[0] || outcome.Difference[1] != want[1] {
		t.Fatalf("difference = %+v, want %+v", outcome.Difference, want)
	}
	if len(workspace.replaced) != 1 || workspace.replaced[0] != "new\n" {
		t.Fatalf("replaced = %q, want the canonical text", workspace.replaced)
	}
}

func TestRunFailsWhenAGENTSIsAbsent(t *testing.T) {
	coord, source, workspace := coordinator(fact.Absent[string](), fact.Present("new\n"))

	outcome := coord.Run()

	if len(outcome.Errors) != 1 || !strings.Contains(outcome.Errors[0].Error(), agentsPath) {
		t.Fatalf("errors = %v, want one error naming %s", outcome.Errors, agentsPath)
	}
	if source.calls != 0 {
		t.Fatalf("canonical source called %d times, want 0", source.calls)
	}
	if len(workspace.replaced) > 0 {
		t.Fatalf("replaced = %q, want no write", workspace.replaced)
	}
}

func TestRunFailsWhenAGENTSCannotBeRead(t *testing.T) {
	failure := errors.New("permission denied")
	coord, source, workspace := coordinator(fact.Failed[string](failure), fact.Present("new\n"))

	outcome := coord.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the read failure", outcome.Errors)
	}
	if source.calls != 0 {
		t.Fatalf("canonical source called %d times, want 0", source.calls)
	}
	if len(workspace.replaced) > 0 {
		t.Fatalf("replaced = %q, want no write", workspace.replaced)
	}
}

func TestRunFailsWhenTheCanonicalTextCannotBeObtained(t *testing.T) {
	failure := errors.New("unreachable")
	coord, _, workspace := coordinator(fact.Present("old\n"), fact.Failed[string](failure))

	outcome := coord.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the canonical failure", outcome.Errors)
	}
	if len(workspace.replaced) > 0 {
		t.Fatalf("replaced = %q, want no write", workspace.replaced)
	}
	if workspace.trace() != "read AGENTS.md" {
		t.Fatalf("trace = %q, want the local read to happen first", workspace.trace())
	}
}

func TestRunDoesNotReplaceAGENTSWhenTheAutomaticScriptCannotBeObtained(t *testing.T) {
	failure := errors.New("the helper script is unreachable")
	coord, _, workspace := coordinator(fact.Present("old\n"), fact.Present("new\n"))
	coord.ManageScript = true
	coord.Script = &fakeSource{canonical: fact.Failed[string](failure)}

	outcome := coord.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the helper-script failure", outcome.Errors)
	}
	if len(workspace.replaced) > 0 {
		t.Fatalf("replaced = %q, want no replacement", workspace.replaced)
	}
}

func TestRunKeepsTheDifferenceWhenReplacementFails(t *testing.T) {
	failure := errors.New("permission denied")
	coord, _, workspace := coordinator(fact.Present("old\n"), fact.Present("new\n"))
	workspace.replaceErr = failure

	outcome := coord.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the write failure", outcome.Errors)
	}
	if len(outcome.Difference) != 2 {
		t.Fatalf("difference = %+v, want the computed difference", outcome.Difference)
	}
	if workspace.trace() != "read AGENTS.md, replace AGENTS.md" {
		t.Fatalf("trace = %q, want the read before the replacement", workspace.trace())
	}
}
