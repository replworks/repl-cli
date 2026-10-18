package initproject

import (
	"errors"
	"strings"
	"testing"

	"github.com/replworks/repl-cli/internal/fact"
)

const canonicalText = "# AGENTS.md\n\n## DOCUMENT_ORDER\n"

type fakeSource struct {
	canonical fact.Fact[string]
	calls     int
}

func (f *fakeSource) Canonical() fact.Fact[string] {
	f.calls++

	return f.canonical
}

type fakeWriter struct {
	present       map[string]bool
	existsErr     error
	createFileErr error
	createDirErr  error

	events []string
	files  map[string]string
	dirs   map[string]bool
}

func newFakeWriter(present ...string) *fakeWriter {
	writer := &fakeWriter{
		present: map[string]bool{},
		files:   map[string]string{},
		dirs:    map[string]bool{},
	}
	for _, path := range present {
		writer.present[path] = true
	}

	return writer
}

func (f *fakeWriter) Exists(rel string) fact.Fact[bool] {
	f.events = append(f.events, "exists "+rel)
	if f.existsErr != nil {
		return fact.Failed[bool](f.existsErr)
	}

	return fact.Present(f.present[rel])
}

func (f *fakeWriter) CreateFile(rel, content string) error {
	f.events = append(f.events, "create-file "+rel)
	if f.createFileErr != nil {
		return f.createFileErr
	}
	f.files[rel] = content

	return nil
}

func (f *fakeWriter) CreateDirectory(rel string) error {
	f.events = append(f.events, "create-directory "+rel)
	if f.createDirErr != nil {
		return f.createDirErr
	}
	f.dirs[rel] = true

	return nil
}

func (f *fakeWriter) trace() string {
	return strings.Join(f.events, ", ")
}

func newSource() *fakeSource {
	return &fakeSource{canonical: fact.Present(canonicalText)}
}

func TestRunCreatesBothEntries(t *testing.T) {
	source := newSource()
	writer := newFakeWriter()

	outcome := Coordinator{Source: source, Workspace: writer}.Run()

	if len(outcome.Errors) > 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	want := []Item{{Path: "AGENTS.md", Created: true}, {Path: ".replworks", Created: true}}
	if len(outcome.Items) != len(want) {
		t.Fatalf("items = %+v, want %+v", outcome.Items, want)
	}
	for i, item := range outcome.Items {
		if item != want[i] {
			t.Fatalf("items[%d] = %+v, want %+v", i, item, want[i])
		}
	}
	if got := writer.files["AGENTS.md"]; got != canonicalText {
		t.Fatalf("AGENTS.md content = %q, want the canonical text", got)
	}
	if !writer.dirs[".replworks"] {
		t.Fatal(".replworks was not created")
	}
	if source.calls != 1 {
		t.Fatalf("canonical source called %d times, want 1", source.calls)
	}
	wantTrace := "exists AGENTS.md, create-file AGENTS.md, exists .replworks, create-directory .replworks"
	if got := writer.trace(); got != wantTrace {
		t.Fatalf("trace = %q, want %q", got, wantTrace)
	}
}

func TestRunSkipsExistingEntries(t *testing.T) {
	writer := newFakeWriter("AGENTS.md", ".replworks")

	outcome := Coordinator{Source: newSource(), Workspace: writer}.Run()

	if len(outcome.Errors) > 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	want := []Item{{Path: "AGENTS.md"}, {Path: ".replworks"}}
	for i, item := range outcome.Items {
		if item != want[i] {
			t.Fatalf("items[%d] = %+v, want %+v", i, item, want[i])
		}
	}
	if len(writer.files) > 0 || len(writer.dirs) > 0 {
		t.Fatalf("the project directory was modified: files = %v, dirs = %v", writer.files, writer.dirs)
	}
	if writer.trace() != "exists AGENTS.md, exists .replworks" {
		t.Fatalf("trace = %q, want only presence checks", writer.trace())
	}
}

func TestRunCreatesTheMissingEntryOnly(t *testing.T) {
	writer := newFakeWriter("AGENTS.md")

	outcome := Coordinator{Source: newSource(), Workspace: writer}.Run()

	if len(outcome.Errors) > 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	if outcome.Items[0].Created {
		t.Fatal("AGENTS.md was created although it exists")
	}
	if !outcome.Items[1].Created {
		t.Fatal(".replworks was not created")
	}
	if len(writer.files) > 0 {
		t.Fatalf("AGENTS.md was written: %v", writer.files)
	}
	if !writer.dirs[".replworks"] {
		t.Fatal(".replworks was not created")
	}
}

func TestRunFailsWhenTheCanonicalTextCannotBeObtained(t *testing.T) {
	failure := errors.New("the canonical source is unreachable")
	writer := newFakeWriter()

	outcome := Coordinator{
		Source:    &fakeSource{canonical: fact.Failed[string](failure)},
		Workspace: writer,
	}.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the canonical failure", outcome.Errors)
	}
	if len(writer.events) > 0 {
		t.Fatalf("the project directory was examined before the canonical text was obtained: %q", writer.trace())
	}
	if len(outcome.Items) > 0 {
		t.Fatalf("items = %+v, want none", outcome.Items)
	}
}

func TestRunFailsWhenTheCanonicalTextCannotBeObtainedEvenWhenAGENTSExists(t *testing.T) {
	writer := newFakeWriter("AGENTS.md", ".replworks")

	outcome := Coordinator{
		Source:    &fakeSource{canonical: fact.Failed[string](errors.New("unreachable"))},
		Workspace: writer,
	}.Run()

	if len(outcome.Errors) != 1 {
		t.Fatalf("errors = %v, want 1", outcome.Errors)
	}
	if len(writer.events) > 0 {
		t.Fatalf("the project directory was examined before the canonical text was obtained: %q", writer.trace())
	}
}

func TestRunStopsAtTheFirstWriteFailure(t *testing.T) {
	failure := errors.New("permission denied")
	writer := newFakeWriter()
	writer.createFileErr = failure

	outcome := Coordinator{Source: newSource(), Workspace: writer}.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the write failure", outcome.Errors)
	}
	if writer.trace() != "exists AGENTS.md, create-file AGENTS.md" {
		t.Fatalf("trace = %q, want it to stop after the failed write", writer.trace())
	}
	if len(outcome.Items) > 0 {
		t.Fatalf("items = %+v, want none", outcome.Items)
	}
}

func TestRunStopsWhenPresenceCannotBeDetermined(t *testing.T) {
	failure := errors.New("listing failed")
	writer := newFakeWriter()
	writer.existsErr = failure

	outcome := Coordinator{Source: newSource(), Workspace: writer}.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the presence failure", outcome.Errors)
	}
	if outcome.Items != nil {
		t.Fatalf("items = %+v, want none", outcome.Items)
	}
}
