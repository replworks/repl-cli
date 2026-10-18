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
