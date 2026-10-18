package rule

import (
	"errors"
	"reflect"
	"testing"

	"github.com/replworks/repl-cli/internal/fact"
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

func TestEvaluateReportsDocumentsExistBeforeAgentsCanonical(t *testing.T) {
	facts := Facts{
		Documents: []Document{
			{Path: "AGENTS.md", Text: fact.Present("# local\n")},
			{Path: ".replworks/PRODUCT_SPEC.md", Text: fact.Absent[string]()},
		},
		Agents:    fact.Present("# local\n"),
		Canonical: fact.Present("# AGENTS.md\n"),
	}

	got := Evaluator{}.Evaluate(facts)

	want := []Violation{
		{Rule: RuleDocumentsExist, Path: ".replworks/PRODUCT_SPEC.md", Message: "document does not exist"},
		{Rule: RuleAgentsCanonical, Path: "AGENTS.md", Message: "AGENTS.md differs from the canonical AGENTS.md"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}
