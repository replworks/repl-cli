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

func (f *fakeWriter) MakeExecutable(rel string) error {
	f.events = append(f.events, "make-executable "+rel)
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

func (f *fakeWriter) DeleteFile(rel string) error {
	f.events = append(f.events, "delete-file "+rel)
	delete(f.files, rel)
	f.present[rel] = false

	return nil
}

func (f *fakeWriter) ReplaceFile(rel, content string) error {
	f.events = append(f.events, "replace-file "+rel)
	f.files[rel] = content
	f.present[rel] = true

	return nil
}

func (f *fakeWriter) trace() string {
	return strings.Join(f.events, ", ")
}

func newSource() *fakeSource {
	return &fakeSource{canonical: fact.Present(canonicalText)}
}

func TestRunCreatesEveryEntry(t *testing.T) {
	source := newSource()
	writer := newFakeWriter()

	outcome := Coordinator{Source: source, Workspace: writer}.Run()

	if len(outcome.Errors) > 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	want := []Item{
		{Path: "AGENTS.md", Created: true},
		{Path: ".replworks", Created: true},
		{Path: ".replworks/PRODUCT_SPEC.md", Created: true},
		{Path: ".replworks/TECH_STACK.md", Created: true},
		{Path: ".replworks/ARCHITECTURE.md", Created: true},
		{Path: ".replworks/TASKS.md", Created: true},
	}
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
	for _, path := range []string{
		".replworks/PRODUCT_SPEC.md",
		".replworks/TECH_STACK.md",
		".replworks/ARCHITECTURE.md",
		".replworks/TASKS.md",
	} {
		content, ok := writer.files[path]
		if !ok {
			t.Fatalf("%s was not created", path)
		}
		if content != "" {
			t.Fatalf("%s content = %q, want empty", path, content)
		}
	}
	if !writer.dirs[".replworks"] {
		t.Fatal(".replworks was not created")
	}
	if source.calls != 1 {
		t.Fatalf("canonical source called %d times, want 1", source.calls)
	}
	wantTrace := "exists AGENTS.md, create-file AGENTS.md, exists .replworks, create-directory .replworks, " +
		"exists .replworks/PRODUCT_SPEC.md, create-file .replworks/PRODUCT_SPEC.md, " +
		"exists .replworks/TECH_STACK.md, create-file .replworks/TECH_STACK.md, " +
		"exists .replworks/ARCHITECTURE.md, create-file .replworks/ARCHITECTURE.md, " +
		"exists .replworks/TASKS.md, create-file .replworks/TASKS.md"
	if got := writer.trace(); got != wantTrace {
		t.Fatalf("trace = %q, want %q", got, wantTrace)
	}
}

func TestRunSkipsExistingEntries(t *testing.T) {
	writer := newFakeWriter(itemPaths...)

	outcome := Coordinator{Source: newSource(), Workspace: writer}.Run()

	if len(outcome.Errors) > 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	if len(outcome.Items) != len(itemPaths) {
		t.Fatalf("items = %+v, want %d skipped items", outcome.Items, len(itemPaths))
	}
	for i, item := range outcome.Items {
		want := Item{Path: itemPaths[i]}
		if item != want {
			t.Fatalf("items[%d] = %+v, want %+v", i, item, want)
		}
	}
	if len(writer.files) > 0 || len(writer.dirs) > 0 {
		t.Fatalf("the project directory was modified: files = %v, dirs = %v", writer.files, writer.dirs)
	}
	wantTrace := "exists " + strings.Join(itemPaths, ", exists ")
	if writer.trace() != wantTrace {
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
	if _, ok := writer.files["AGENTS.md"]; ok {
		t.Fatalf("AGENTS.md was written: %v", writer.files)
	}
	if !writer.dirs[".replworks"] {
		t.Fatal(".replworks was not created")
	}
}

func TestRunCreatesOnlyTheMissingDocuments(t *testing.T) {
	writer := newFakeWriter(
		"AGENTS.md",
		".replworks",
		".replworks/PRODUCT_SPEC.md",
		".replworks/ARCHITECTURE.md",
		".replworks/TASKS.md",
	)

	outcome := Coordinator{Source: newSource(), Workspace: writer}.Run()

	if len(outcome.Errors) > 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
	if len(outcome.Items) != len(itemPaths) {
		t.Fatalf("items = %+v, want %d items", outcome.Items, len(itemPaths))
	}
	for i, item := range outcome.Items {
		want := Item{Path: itemPaths[i]}
		if itemPaths[i] == ".replworks/TECH_STACK.md" {
			want.Created = true
		}
		if item != want {
			t.Fatalf("items[%d] = %+v, want %+v", i, item, want)
		}
	}
	if got := writer.files[".replworks/TECH_STACK.md"]; got != "" {
		t.Fatalf("TECH_STACK.md content = %q, want empty", got)
	}
	if len(writer.files) != 1 {
		t.Fatalf("files = %v, want only the created document", writer.files)
	}
	if len(writer.dirs) > 0 {
		t.Fatalf("dirs = %v, want none", writer.dirs)
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

func TestRunDoesNotWriteWhenTheAutomaticScriptCannotBeObtained(t *testing.T) {
	failure := errors.New("the helper script is unreachable")
	writer := newFakeWriter()

	outcome := Coordinator{
		Source:       newSource(),
		Script:       &fakeSource{canonical: fact.Failed[string](failure)},
		ManageScript: true,
		Workspace:    writer,
	}.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the helper-script failure", outcome.Errors)
	}
	if len(writer.events) != 0 {
		t.Fatalf("events = %q, want no workspace access before both sources are obtained", writer.events)
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

func TestRunReportsTheCreatedItemWhenTheSecondWriteFails(t *testing.T) {
	failure := errors.New("permission denied")
	writer := newFakeWriter()
	writer.createDirErr = failure

	outcome := Coordinator{Source: newSource(), Workspace: writer}.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the write failure", outcome.Errors)
	}
	if got := writer.files["AGENTS.md"]; got != canonicalText {
		t.Fatalf("AGENTS.md content = %q, want the canonical text", got)
	}
	if writer.dirs[".replworks"] {
		t.Fatal(".replworks was created although the write failed")
	}
	wantItems := []Item{{Path: "AGENTS.md", Created: true}}
	if len(outcome.Items) != len(wantItems) || outcome.Items[0] != wantItems[0] {
		t.Fatalf("items = %+v, want %+v", outcome.Items, wantItems)
	}
	wantTrace := "exists AGENTS.md, create-file AGENTS.md, exists .replworks, create-directory .replworks"
	if got := writer.trace(); got != wantTrace {
		t.Fatalf("trace = %q, want %q", got, wantTrace)
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
