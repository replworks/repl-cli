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

type fakeBaseReader struct {
	texts    map[string]fact.Fact[string]
	branches []string
}

func (r *fakeBaseReader) Tasks(branch string) fact.Fact[string] {
	r.branches = append(r.branches, branch)

	text, ok := r.texts[branch]
	if !ok {
		return fact.Absent[string]()
	}

	return text
}

func coordinator(texts map[string]fact.Fact[string]) (Coordinator, *fakeWorkspace) {
	workspace := &fakeWorkspace{texts: texts}

	return Coordinator{
		Workspace: workspace,
		Source:    &fakeSource{canonical: fact.Present("# AGENTS.md\n")},
		Bases:     &fakeBaseReader{},
		Evaluator: rule.Evaluator{},
	}, workspace
}

const baseBranch = "main"

func presentDocuments(agents, tasks string) map[string]fact.Fact[string] {
	texts := map[string]fact.Fact[string]{}
	for _, path := range documents {
		texts[path] = fact.Present("")
	}
	texts[agentsPath] = fact.Present(agents)
	texts[tasksPath] = fact.Present(tasks)

	return texts
}

func TestRunReadsEveryDocumentOnceInOrder(t *testing.T) {
	coord, workspace := coordinator(nil)

	coord.Run(baseBranch)

	want := documents
	if !reflect.DeepEqual(workspace.reads, want) {
		t.Fatalf("reads = %q, want %q", workspace.reads, want)
	}
}

func TestRunReportsMissingDocumentsInOrder(t *testing.T) {
	coord, _ := coordinator(map[string]fact.Fact[string]{
		"AGENTS.md": fact.Present("# AGENTS.md\n"),
	})

	outcome := coord.Run(baseBranch)

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
	coord, _ := coordinator(presentDocuments("# AGENTS.md\n", ""))

	outcome := coord.Run(baseBranch)

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

	outcome := coord.Run(baseBranch)

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

	first := coord.Run(baseBranch)
	second := coord.Run(baseBranch)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("outcomes differ across runs: %v and %v", first, second)
	}
}

func TestRunObtainsTheCanonicalTextOnce(t *testing.T) {
	coord, _ := coordinator(map[string]fact.Fact[string]{
		agentsPath: fact.Present("# AGENTS.md\n"),
	})
	source := coord.Source.(*fakeSource)

	coord.Run(baseBranch)

	if source.calls != 1 {
		t.Fatalf("canonical source called %d times, want 1", source.calls)
	}
}

func TestRunReportsADifferentAgentsMD(t *testing.T) {
	coord, _ := coordinator(presentDocuments("# local\n", ""))

	outcome := coord.Run(baseBranch)

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
	coord, _ := coordinator(presentDocuments("# AGENTS.md\n", ""))

	outcome := coord.Run(baseBranch)

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

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleDocumentsExist, Path: agentsPath, Message: "document does not exist"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
}

func TestRunRecordsAFailedCanonicalTextAsAnError(t *testing.T) {
	failure := errors.New("the canonical source is unreachable")
	coord, _ := coordinator(presentDocuments("# local\n", ""))
	coord.Source = &fakeSource{canonical: fact.Failed[string](failure)}

	outcome := coord.Run(baseBranch)

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the canonical failure", outcome.Errors)
	}

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
}

func TestRunReadsTheBaseBranchOnce(t *testing.T) {
	coord, _ := coordinator(map[string]fact.Fact[string]{
		agentsPath: fact.Present("# AGENTS.md\n"),
	})

	coord.Run(baseBranch)

	bases := coord.Bases.(*fakeBaseReader)
	want := []string{baseBranch}
	if !reflect.DeepEqual(bases.branches, want) {
		t.Fatalf("branches = %q, want %q", bases.branches, want)
	}
}

func TestRunRecordsAFailedBaseBranchAsAnError(t *testing.T) {
	failure := errors.New("the base branch main does not exist in the repository")
	coord, _ := coordinator(presentDocuments("# AGENTS.md\n", ""))
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Failed[string](failure),
	}}

	outcome := coord.Run(baseBranch)

	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the base branch failure", outcome.Errors)
	}
	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
}

func TestRunKeepsTheGatheringOrderOfErrors(t *testing.T) {
	readFailure := errors.New("reading .replworks/ARCHITECTURE.md: permission denied")
	canonicalFailure := errors.New("the canonical source is unreachable")
	baseFailure := errors.New("the base branch main does not exist in the repository")

	texts := presentDocuments("", "")
	texts[".replworks/ARCHITECTURE.md"] = fact.Failed[string](readFailure)
	coord, _ := coordinator(texts)
	coord.Source = &fakeSource{canonical: fact.Failed[string](canonicalFailure)}
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Failed[string](baseFailure),
	}}

	outcome := coord.Run(baseBranch)

	want := []error{readFailure, canonicalFailure, baseFailure}
	if !reflect.DeepEqual(outcome.Errors, want) {
		t.Fatalf("errors = %v, want %v", outcome.Errors, want)
	}
}

func TestRunReportsTaskIDsMissingFromTheWorkingTree(t *testing.T) {
	coord, _ := coordinator(presentDocuments("# AGENTS.md\n", "- [ ] T-001 first\n"))
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n- [ ] T-002 second\n- [ ] T-003 third\n"),
	}}

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleTaskIDsUnchanged, Path: tasksPath, Message: "missing task ID T-002"},
		{Rule: rule.RuleTaskIDsUnchanged, Path: tasksPath, Message: "missing task ID T-003"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
}

func TestRunDoesNotReportTaskIDsWhenTheBaseHasNoTasks(t *testing.T) {
	coord, _ := coordinator(presentDocuments("# AGENTS.md\n", "- [ ] T-001 first\n"))

	outcome := coord.Run(baseBranch)

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
}

func TestRunDoesNotReportTaskIDsWhenTheWorkingTreeHasNoTasks(t *testing.T) {
	texts := presentDocuments("# AGENTS.md\n", "")
	texts[tasksPath] = fact.Absent[string]()
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleDocumentsExist, Path: tasksPath, Message: "document does not exist"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
}

func TestRunDoesNotReportTaskIDsWhenTheBaseBranchFails(t *testing.T) {
	failure := errors.New("the base branch main does not exist in the repository")
	coord, _ := coordinator(presentDocuments("# AGENTS.md\n", "- [ ] T-001 first\n"))
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Failed[string](failure),
	}}

	outcome := coord.Run(baseBranch)

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the base branch failure", outcome.Errors)
	}
}

func TestRunReportsMoreThanOneNewlyCompleteTask(t *testing.T) {
	coord, _ := coordinator(presentDocuments("# AGENTS.md\n", "- [x] T-001 first\n- [x] T-002 second\n"))
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n- [ ] T-002 second\n"),
	}}

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleSingleTaskCompletion, Path: tasksPath, Message: "2 tasks are newly complete"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
}

func TestRunCountsEveryCompleteTaskWhenTheBaseHasNoTasks(t *testing.T) {
	coord, _ := coordinator(presentDocuments("# AGENTS.md\n", "- [x] T-001 first\n- [x] T-002 second\n"))

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleSingleTaskCompletion, Path: tasksPath, Message: "2 tasks are newly complete"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
}

func TestRunDoesNotReportNewlyCompleteTaskWhenTheBaseBranchFails(t *testing.T) {
	failure := errors.New("the base branch main does not exist in the repository")
	coord, _ := coordinator(presentDocuments("# AGENTS.md\n", "- [x] T-001 first\n- [x] T-002 second\n"))
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Failed[string](failure),
	}}

	outcome := coord.Run(baseBranch)

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], failure) {
		t.Fatalf("errors = %v, want the base branch failure", outcome.Errors)
	}
}

func TestRunReportsNoUnverifiedMarks(t *testing.T) {
	texts := presentDocuments("# AGENTS.md\n", "- [x] T-001 first\n")
	texts[productSpec] = fact.Present("# PRODUCT_SPEC.md\n\n## 1. Purpose [UNVERIFIED]\n")
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleNoUnverified, Path: productSpec, Message: "UNVERIFIED mark on line 3"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
}

func TestRunReportsMarksInDocumentOrder(t *testing.T) {
	texts := presentDocuments("# AGENTS.md\n", "- [x] T-001 first\n")
	texts[techStackPath] = fact.Present("## 6. CI [UNVERIFIED]\n")
	texts[architecture] = fact.Present("# ARCHITECTURE.md\n## 1. Purpose [UNVERIFIED]\n")
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleNoUnverified, Path: techStackPath, Message: "UNVERIFIED mark on line 1"},
		{Rule: rule.RuleNoUnverified, Path: architecture, Message: "UNVERIFIED mark on line 2"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
}

func TestRunDoesNotReportMarksFromAnAbsentDocument(t *testing.T) {
	texts := presentDocuments("# AGENTS.md\n", "- [x] T-001 first\n")
	texts[productSpec] = fact.Absent[string]()
	texts[architecture] = fact.Present("## 1. Purpose [UNVERIFIED]\n")
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleDocumentsExist, Path: productSpec, Message: "document does not exist"},
		{Rule: rule.RuleNoUnverified, Path: architecture, Message: "UNVERIFIED mark on line 1"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
}

func TestRunDoesNotReportMarksWhenADocumentReadFails(t *testing.T) {
	readFailure := errors.New("permission denied")
	texts := presentDocuments("# AGENTS.md\n", "- [x] T-001 first\n")
	texts[techStackPath] = fact.Failed[string](readFailure)
	texts[architecture] = fact.Present("## 1. Purpose [UNVERIFIED]\n")
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], readFailure) {
		t.Fatalf("errors = %v, want the read failure", outcome.Errors)
	}
}

func TestRunDoesNotReportMarksWithoutATaskLedger(t *testing.T) {
	texts := presentDocuments("# AGENTS.md\n", "")
	texts[tasksPath] = fact.Absent[string]()
	texts[architecture] = fact.Present("## 1. Purpose [UNVERIFIED]\n")
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleDocumentsExist, Path: tasksPath, Message: "document does not exist"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v", outcome.Violations, want)
	}
}

func TestRunDoesNotReportMarksWithoutANewlyCompleteTask(t *testing.T) {
	texts := presentDocuments("# AGENTS.md\n", "- [x] T-001 first\n")
	texts[architecture] = fact.Present("## 1. Purpose [UNVERIFIED]\n")
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [x] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
}

func TestRunDoesNotReportMarksInAgentsOrTasks(t *testing.T) {
	agents := "# AGENTS.md\n\n## Notes [UNVERIFIED]\n"
	tasks := "- [x] T-001 first\n\n## Notes [UNVERIFIED]\n"
	texts := presentDocuments(agents, tasks)
	coord, _ := coordinator(texts)
	coord.Source = &fakeSource{canonical: fact.Present(agents)}
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none", outcome.Violations)
	}
	if len(outcome.Errors) != 0 {
		t.Fatalf("errors = %v, want none", outcome.Errors)
	}
}

func TestRunRecordsAnUnreadableTasksAsAnError(t *testing.T) {
	readFailure := errors.New("reading .replworks/TASKS.md: permission denied")
	texts := presentDocuments("# AGENTS.md\n", "")
	texts[tasksPath] = fact.Failed[string](readFailure)
	texts[architecture] = fact.Present("## 1. Purpose [UNVERIFIED]\n")
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n- [x] T-002 second\n"),
	}}

	outcome := coord.Run(baseBranch)

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none: an unreadable TASKS.md disables the task rules and is not missing", outcome.Violations)
	}
	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], readFailure) {
		t.Fatalf("errors = %v, want the read failure", outcome.Errors)
	}
}

func TestRunRecordsAnUnreadableAgentsAsAnError(t *testing.T) {
	readFailure := errors.New("reading AGENTS.md: permission denied")
	texts := presentDocuments("", "- [x] T-001 first\n")
	texts[agentsPath] = fact.Failed[string](readFailure)
	texts[architecture] = fact.Present("## 1. Purpose [UNVERIFIED]\n")
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	want := []rule.Violation{
		{Rule: rule.RuleNoUnverified, Path: architecture, Message: "UNVERIFIED mark on line 1"},
	}
	if !reflect.DeepEqual(outcome.Violations, want) {
		t.Fatalf("violations = %v, want %v: an unreadable AGENTS.md disables agents-canonical only", outcome.Violations, want)
	}
	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], readFailure) {
		t.Fatalf("errors = %v, want the read failure", outcome.Errors)
	}
}

func TestRunRecordsAnUnreadableMarkDocumentAsAnError(t *testing.T) {
	readFailure := errors.New("reading .replworks/TECH_STACK.md: permission denied")
	texts := presentDocuments("# AGENTS.md\n", "- [x] T-001 first\n")
	texts[techStackPath] = fact.Failed[string](readFailure)
	coord, _ := coordinator(texts)
	coord.Bases = &fakeBaseReader{texts: map[string]fact.Fact[string]{
		baseBranch: fact.Present("- [ ] T-001 first\n"),
	}}

	outcome := coord.Run(baseBranch)

	if len(outcome.Violations) != 0 {
		t.Fatalf("violations = %v, want none: a failed mark list disables no-unverified", outcome.Violations)
	}
	if len(outcome.Errors) != 1 || !errors.Is(outcome.Errors[0], readFailure) {
		t.Fatalf("errors = %v, want the read failure", outcome.Errors)
	}
}
