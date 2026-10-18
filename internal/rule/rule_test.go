package rule

import (
	"errors"
	"reflect"
	"testing"

	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/ledger"
)

func TestEvaluateReportsEveryAbsentDocumentInOrder(t *testing.T) {
	facts := Facts{Documents: []Document{
		{Path: "AGENTS.md", Text: fact.Absent[string]()},
		{Path: ".replworks/PRODUCT_SPEC.md", Text: fact.Present("")},
		{Path: ".replworks/TECH_STACK.md", Text: fact.Absent[string]()},
		{Path: ".replworks/ARCHITECTURE.md", Text: fact.Present("# ARCHITECTURE.md")},
		{Path: ".replworks/TASKS.md", Text: fact.Absent[string]()},
	}}

	got := Evaluator{}.Evaluate(facts)

	want := []Violation{
		{Rule: RuleDocumentsExist, Path: "AGENTS.md", Message: "document does not exist"},
		{Rule: RuleDocumentsExist, Path: ".replworks/TECH_STACK.md", Message: "document does not exist"},
		{Rule: RuleDocumentsExist, Path: ".replworks/TASKS.md", Message: "document does not exist"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestEvaluateDoesNotReportPresentOrFailedDocuments(t *testing.T) {
	facts := Facts{Documents: []Document{
		{Path: "AGENTS.md", Text: fact.Present("")},
		{Path: ".replworks/PRODUCT_SPEC.md", Text: fact.Failed[string](errors.New("permission denied"))},
	}}

	if got := (Evaluator{}).Evaluate(facts); len(got) != 0 {
		t.Fatalf("violations = %v, want none", got)
	}
}

func TestEvaluateWithoutDocuments(t *testing.T) {
	if got := (Evaluator{}).Evaluate(Facts{}); len(got) != 0 {
		t.Fatalf("violations = %v, want none", got)
	}
}

func TestEvaluateIsDeterministic(t *testing.T) {
	facts := Facts{Documents: []Document{
		{Path: "AGENTS.md", Text: fact.Absent[string]()},
		{Path: ".replworks/PRODUCT_SPEC.md", Text: fact.Absent[string]()},
		{Path: ".replworks/TECH_STACK.md", Text: fact.Absent[string]()},
		{Path: ".replworks/ARCHITECTURE.md", Text: fact.Absent[string]()},
		{Path: ".replworks/TASKS.md", Text: fact.Absent[string]()},
	}}

	first := Evaluator{}.Evaluate(facts)
	second := Evaluator{}.Evaluate(facts)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("violations differ across runs: %v and %v", first, second)
	}
}

func TestEvaluateReportsADifferentAgentsMD(t *testing.T) {
	tests := []struct {
		name   string
		agents fact.Fact[string]
		canon  fact.Fact[string]
		want   []Violation
	}{
		{
			name:   "different text",
			agents: fact.Present("# local\n"),
			canon:  fact.Present("# AGENTS.md\n"),
			want: []Violation{{
				Rule:    RuleAgentsCanonical,
				Path:    "AGENTS.md",
				Message: "AGENTS.md differs from the canonical AGENTS.md",
			}},
		},
		{
			name:   "different line endings",
			agents: fact.Present("# AGENTS.md\r\n"),
			canon:  fact.Present("# AGENTS.md\n"),
			want: []Violation{{
				Rule:    RuleAgentsCanonical,
				Path:    "AGENTS.md",
				Message: "AGENTS.md differs from the canonical AGENTS.md",
			}},
		},
		{
			name:   "different trailing whitespace",
			agents: fact.Present("# AGENTS.md \n"),
			canon:  fact.Present("# AGENTS.md\n"),
			want: []Violation{{
				Rule:    RuleAgentsCanonical,
				Path:    "AGENTS.md",
				Message: "AGENTS.md differs from the canonical AGENTS.md",
			}},
		},
		{
			name:   "identical text",
			agents: fact.Present("# AGENTS.md\n"),
			canon:  fact.Present("# AGENTS.md\n"),
		},
		{
			name:   "absent AGENTS.md",
			agents: fact.Absent[string](),
			canon:  fact.Present("# AGENTS.md\n"),
		},
		{
			name:   "failed AGENTS.md",
			agents: fact.Failed[string](errors.New("permission denied")),
			canon:  fact.Present("# AGENTS.md\n"),
		},
		{
			name:   "failed canonical text",
			agents: fact.Present("# local\n"),
			canon:  fact.Failed[string](errors.New("unreachable")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluator{}.Evaluate(Facts{Agents: tt.agents, Canonical: tt.canon})

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("violations = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateReportsTaskIDsMissingFromTheWorkingTree(t *testing.T) {
	facts := Facts{
		WorkingLedger: fact.Present(ledger.New("- [ ] T-001 first\n- [ ] T-003 third\n")),
		BaseLedger:    fact.Present(ledger.New("- [ ] T-001 first\n- [ ] T-002 second\n- [ ] T-003 third\n")),
	}

	got := Evaluator{}.Evaluate(facts)

	want := []Violation{
		{Rule: RuleTaskIDsUnchanged, Path: ".replworks/TASKS.md", Message: "missing task ID T-002"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestEvaluateReportsDistinctBaseTaskIDsInBaseOrder(t *testing.T) {
	facts := Facts{
		WorkingLedger: fact.Present(ledger.New("")),
		BaseLedger:    fact.Present(ledger.New("- [ ] T-002 second\n- [ ] T-001 first\n- [ ] T-002 second again\n")),
	}

	got := Evaluator{}.Evaluate(facts)

	want := []Violation{
		{Rule: RuleTaskIDsUnchanged, Path: ".replworks/TASKS.md", Message: "missing task ID T-002"},
		{Rule: RuleTaskIDsUnchanged, Path: ".replworks/TASKS.md", Message: "missing task ID T-001"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestEvaluateDoesNotReportTaskIDsForOtherLedgerStates(t *testing.T) {
	base := fact.Present(ledger.New("- [ ] T-001 first\n"))
	working := fact.Present(ledger.New("- [ ] T-001 first\n"))

	tests := []struct {
		name    string
		working fact.Fact[ledger.Ledger]
		base    fact.Fact[ledger.Ledger]
	}{
		{name: "same IDs", working: working, base: base},
		{name: "empty base", base: fact.Present(ledger.New("")), working: working},
		{name: "absent base", working: working, base: fact.Absent[ledger.Ledger]()},
		{name: "failed base", working: working, base: fact.Failed[ledger.Ledger](errors.New("unreachable"))},
		{name: "absent working ledger", working: fact.Absent[ledger.Ledger](), base: base},
		{name: "failed working ledger", working: fact.Failed[ledger.Ledger](errors.New("permission denied")), base: base},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluator{}.Evaluate(Facts{WorkingLedger: tt.working, BaseLedger: tt.base})

			if len(got) != 0 {
				t.Fatalf("violations = %v, want none", got)
			}
		})
	}
}

func TestEvaluateReportsRulesInTheFixedOrder(t *testing.T) {
	facts := Facts{
		Documents: []Document{
			{Path: "AGENTS.md", Text: fact.Present("# local\n")},
			{Path: ".replworks/PRODUCT_SPEC.md", Text: fact.Absent[string]()},
		},
		Agents:        fact.Present("# local\n"),
		Canonical:     fact.Present("# AGENTS.md\n"),
		WorkingLedger: fact.Present(ledger.New("- [x] T-002 second\n- [x] T-003 third\n")),
		BaseLedger:    fact.Present(ledger.New("- [ ] T-001 first\n")),
	}

	got := Evaluator{}.Evaluate(facts)

	want := []Violation{
		{Rule: RuleDocumentsExist, Path: ".replworks/PRODUCT_SPEC.md", Message: "document does not exist"},
		{Rule: RuleTaskIDsUnchanged, Path: ".replworks/TASKS.md", Message: "missing task ID T-001"},
		{Rule: RuleSingleTaskCompletion, Path: ".replworks/TASKS.md", Message: "2 tasks are newly complete"},
		{Rule: RuleAgentsCanonical, Path: "AGENTS.md", Message: "AGENTS.md differs from the canonical AGENTS.md"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestEvaluateReportsMoreThanOneNewlyCompleteTask(t *testing.T) {
	tests := []struct {
		name    string
		working string
		base    string
		want    []Violation
	}{
		{
			name:    "two newly complete",
			working: "- [x] T-001 first\n- [x] T-002 second\n",
			base:    "- [ ] T-001 first\n- [ ] T-002 second\n",
			want: []Violation{{
				Rule:    RuleSingleTaskCompletion,
				Path:    ".replworks/TASKS.md",
				Message: "2 tasks are newly complete",
			}},
		},
		{
			name:    "one newly complete",
			working: "- [x] T-001 first\n- [ ] T-002 second\n",
			base:    "- [ ] T-001 first\n- [ ] T-002 second\n",
		},
		{
			name:    "none newly complete",
			working: "- [x] T-001 first\n",
			base:    "- [x] T-001 first\n",
		},
		{
			name:    "complete on the base branch",
			working: "- [x] T-001 first\n- [x] T-002 second\n",
			base:    "- [x] T-001 first\n- [x] T-002 second\n",
		},
		{
			name:    "absent on the base branch",
			working: "- [x] T-001 first\n- [x] T-002 second\n",
			base:    "- [ ] T-002 second\n",
			want: []Violation{{
				Rule:    RuleSingleTaskCompletion,
				Path:    ".replworks/TASKS.md",
				Message: "2 tasks are newly complete",
			}},
		},
		{
			name:    "duplicate IDs in the working tree",
			working: "- [x] T-001 first\n- [x] T-001 first again\n",
			base:    "- [ ] T-001 first\n",
			want: []Violation{{
				Rule:    RuleSingleTaskCompletion,
				Path:    ".replworks/TASKS.md",
				Message: "2 tasks are newly complete",
			}},
		},
		{
			name:    "present but incomplete on the base branch",
			working: "- [x] T-002 second\n",
			base:    "- [ ] T-002 second\n",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facts := Facts{
				WorkingLedger: fact.Present(ledger.New(tt.working)),
				BaseLedger:    fact.Present(ledger.New(tt.base)),
			}

			got := Evaluator{}.Evaluate(facts)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("violations = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateCountsEveryCompleteTaskWithoutBaseTasks(t *testing.T) {
	tests := []struct {
		name    string
		working string
		want    bool
	}{
		{name: "two complete", working: "- [x] T-001 first\n- [x] T-002 second\n", want: true},
		{name: "one complete", working: "- [x] T-001 first\n", want: false},
		{name: "none complete", working: "- [ ] T-001 first\n- [ ] T-002 second\n", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facts := Facts{
				WorkingLedger: fact.Present(ledger.New(tt.working)),
				BaseLedger:    fact.Absent[ledger.Ledger](),
			}

			got := Evaluator{}.Evaluate(facts)

			if (len(got) == 1) != tt.want {
				t.Fatalf("violations = %v, want a violation: %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateDoesNotReportNewlyCompleteForOtherLedgerStates(t *testing.T) {
	working := fact.Present(ledger.New("- [x] T-001 first\n- [x] T-002 second\n"))
	base := fact.Present(ledger.New("- [ ] T-001 first\n"))

	tests := []struct {
		name    string
		working fact.Fact[ledger.Ledger]
		base    fact.Fact[ledger.Ledger]
	}{
		{name: "failed base", working: working, base: fact.Failed[ledger.Ledger](errors.New("unreachable"))},
		{name: "absent working ledger", working: fact.Absent[ledger.Ledger](), base: base},
		{name: "failed working ledger", working: fact.Failed[ledger.Ledger](errors.New("permission denied")), base: base},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluator{}.Evaluate(Facts{WorkingLedger: tt.working, BaseLedger: tt.base})

			if len(got) != 0 {
				t.Fatalf("violations = %v, want none", got)
			}
		})
	}
}
