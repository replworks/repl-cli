package checkproject

import (
	"errors"
	"reflect"
	"testing"

	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/rule"
)

type fakeWorkspace struct {
	texts map[string]fact.Fact[string]
	reads []string
}

func (w *fakeWorkspace) ReadText(rel string) fact.Fact[string] {
	w.reads = append(w.reads, rel)

	text, ok := w.texts[rel]
	if !ok {
		return fact.Absent[string]()
	}

	return text
}

type fakeSource struct {
	canonical fact.Fact[string]
	calls     int
}

func (s *fakeSource) Canonical() fact.Fact[string] {
	s.calls++

	return s.canonical
}

func coordinator(texts map[string]fact.Fact[string]) (Coordinator, *fakeWorkspace) {
	workspace := &fakeWorkspace{texts: texts}

	return Coordinator{
		Workspace: workspace,
		Source:    &fakeSource{canonical: fact.Present("# AGENTS.md\n")},
		Evaluator: rule.Evaluator{},
	}, workspace
}

func TestRunReadsEveryDocumentOnceInOrder(t *testing.T) {
	coord, workspace := coordinator(nil)

	coord.Run()

	want := documents
	if !reflect.DeepEqual(workspace.reads, want) {
		t.Fatalf("reads = %q, want %q", workspace.reads, want)
	}
}

func TestRunReportsMissingDocumentsInOrder(t *testing.T) {
	coord, _ := coordinator(map[string]fact.Fact[string]{
		"AGENTS.md": fact.Present("# AGENTS.md\n"),
	})

	outcome := coord.Run()

	want := []rule.Violation{
		{Rule: rule.RuleDocumentsExist, Path: ".replworks/PRODUCT_SPEC.md", Message: "document does not exist"},
		{Rule: rule.RuleDocumentsExist, Path: ".replworks/TECH_STACK.md", Message: "document does not exist"},
		{Rule: rule.RuleDocumentsExist, Path: ".replworks/ARCHITECTURE.md", Message: "document does not exist"},
		{Rule: rule.RuleDocumentsExist, Path: ".replworks/TASKS.md", Message: "document does not exist"},
	}

	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
}

func TestRunReportsNoViolationWhenEveryDocumentExists(t *testing.T) {
	texts := map[string]fact.Fact[string]{}
	for _, path := range documents {
		texts[path] = fact.Present("")
	}
	texts[agentsPath] = fact.Present("# AGENTS.md\n")
	coord, _ := coordinator(texts)

	outcome := coord.Run()

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
}

func TestRunRecordsFailedReadsAsErrors(t *testing.T) {
	failure := errors.New("permission denied")
	texts := map[string]fact.Fact[string]{}
	for _, path := range documents {
		texts[path] = fact.Present("")
	}
	texts[agentsPath] = fact.Present("# AGENTS.md\n")
	texts[".replworks/PRODUCT_SPEC.md"] = fact.Failed[string](failure)
	texts[".replworks/TASKS.md"] = fact.Absent[string]()
	coord, _ := coordinator(texts)

	outcome := coord.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the read failure", outcome.Errors)
	}

	want := []rule.Violation{
		{Rule: rule.RuleDocumentsExist, Path: ".replworks/TASKS.md", Message: "document does not exist"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
}

func TestRunIsDeterministic(t *testing.T) {
	coord, _ := coordinator(map[string]fact.Fact[string]{
		"AGENTS.md": fact.Present("# AGENTS.md\n"),
	})

	first := coord.Run()
	second := coord.Run()

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("outcomes differ across runs: %v and %v", first, second)
	}
}

func TestRunObtainsTheCanonicalTextOnce(t *testing.T) {
	coord, _ := coordinator(map[string]fact.Fact[string]{
		agentsPath: fact.Present("# AGENTS.md\n"),
	})
	source := coord.Source.(*fakeSource)

	coord.Run()

	if source.calls != 1 {
		t.Fatalf("canonical source called %d times, want 1", source.calls)
	}
}

func TestRunReportsADifferentAgentsMD(t *testing.T) {
	texts := map[string]fact.Fact[string]{}
	for _, path := range documents {
		texts[path] = fact.Present("")
	}
	texts[agentsPath] = fact.Present("# local\n")
	coord, _ := coordinator(texts)

	outcome := coord.Run()

	want := []rule.Violation{
		{Rule: rule.RuleAgentsCanonical, Path: agentsPath, Message: "AGENTS.md differs from the canonical AGENTS.md"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
}

func TestRunDoesNotReportAnIdenticalAgentsMD(t *testing.T) {
	texts := map[string]fact.Fact[string]{}
	for _, path := range documents {
		texts[path] = fact.Present("")
	}
	texts[agentsPath] = fact.Present("# AGENTS.md\n")
	coord, _ := coordinator(texts)

	outcome := coord.Run()

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
}

func TestRunDoesNotReportAnAbsentAgentsMD(t *testing.T) {
	texts := map[string]fact.Fact[string]{}
	for _, path := range documents[1:] {
		texts[path] = fact.Present("")
	}
	coord, _ := coordinator(texts)

	outcome := coord.Run()

	want := []rule.Violation{
		{Rule: rule.RuleDocumentsExist, Path: agentsPath, Message: "document does not exist"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
}

func TestRunRecordsAFailedCanonicalTextAsAnError(t *testing.T) {
	failure := errors.New("the canonical source is unreachable")
	texts := map[string]fact.Fact[string]{}
	for _, path := range documents {
		texts[path] = fact.Present("")
	}
	texts[agentsPath] = fact.Present("# local\n")
	coord, _ := coordinator(texts)
	coord.Source = &fakeSource{canonical: fact.Failed[string](failure)}

	outcome := coord.Run()

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the canonical failure", outcome.Errors)
	}

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
}
