// Package rule evaluates the check rules over gathered facts and owns the rule
// identifiers, their fixed evaluation order, and their violation messages.
package rule

import (
	"fmt"
	"slices"

	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/ledger"
)

// RuleDocumentsExist is reported for each authoritative document that does not
// exist.
const RuleDocumentsExist = "documents-exist"

// RuleTaskIDsUnchanged is reported for each task ID that exists on the base
// branch and does not exist in the working tree.
const RuleTaskIDsUnchanged = "task-ids-unchanged"

// RuleSingleTaskCompletion is reported when the working tree holds more than
// one newly complete task.
const RuleSingleTaskCompletion = "single-task-completion"

// RuleAgentsCanonical is reported for an AGENTS.md that differs from the
// canonical AGENTS.md.
const RuleAgentsCanonical = "agents-canonical"

const (
	agentsPath = "AGENTS.md"
	tasksPath  = ".replworks/TASKS.md"
)

// Violation is a failed rule.
type Violation struct {
	Rule    string
	Path    string
	Message string
}

// Document is the fact of one authoritative document.
type Document struct {
	Path string
	Text fact.Fact[string]
}

// Facts is the bundle of facts that rule evaluation reads.
type Facts struct {
	Documents     []Document
	Agents        fact.Fact[string]
	Canonical     fact.Fact[string]
	WorkingLedger fact.Fact[ledger.Ledger]
	BaseLedger    fact.Fact[ledger.Ledger]
}

// Evaluator owns the rule identifiers, their evaluation order, and their
// violation messages.
type Evaluator struct{}

// Evaluate returns the violations of every rule whose preconditions hold, in
// the fixed rule order.
func (Evaluator) Evaluate(facts Facts) []Violation {
	var violations []Violation

	for _, evaluate := range order {
		violations = append(violations, evaluate(facts)...)
	}

	return violations
}

// order is the fixed rule evaluation order.
var order = []func(Facts) []Violation{documentsExist, taskIDsUnchanged, singleTaskCompletion, agentsCanonical}

func documentsExist(facts Facts) []Violation {
	var violations []Violation

	for _, document := range facts.Documents {
		if !document.Text.IsAbsent() {
			continue
		}

		violations = append(violations, Violation{
			Rule:    RuleDocumentsExist,
			Path:    document.Path,
			Message: "document does not exist",
		})
	}

	return violations
}

func taskIDsUnchanged(facts Facts) []Violation {
	if !facts.WorkingLedger.IsPresent() || facts.BaseLedger.IsFailed() {
		return nil
	}

	base := ledger.Ledger{}
	if facts.BaseLedger.IsPresent() {
		base = facts.BaseLedger.Value()
	}

	var violations []Violation
	working := facts.WorkingLedger.Value().IDs

	for _, id := range base.IDs {
		if slices.Contains(working, id) {
			continue
		}

		violations = append(violations, Violation{
			Rule:    RuleTaskIDsUnchanged,
			Path:    tasksPath,
			Message: fmt.Sprintf("missing task ID %s", id),
		})
	}

	return violations
}

func singleTaskCompletion(facts Facts) []Violation {
	count, ok := newlyCompleteCount(facts)
	if !ok || count <= 1 {
		return nil
	}

	return []Violation{{
		Rule:    RuleSingleTaskCompletion,
		Path:    tasksPath,
		Message: fmt.Sprintf("%d tasks are newly complete", count),
	}}
}

// newlyCompleteCount counts the working entries that are complete and whose
// task ID is not complete in the base ledger. Entries are counted
// individually, so duplicate IDs are separate entries. The second result is
// false when the rule preconditions do not hold.
func newlyCompleteCount(facts Facts) (int, bool) {
	if !facts.WorkingLedger.IsPresent() || facts.BaseLedger.IsFailed() {
		return 0, false
	}

	complete := map[string]bool{}
	if facts.BaseLedger.IsPresent() {
		for _, entry := range facts.BaseLedger.Value().Entries {
			if entry.Complete {
				complete[entry.ID] = true
			}
		}
	}

	count := 0
	for _, entry := range facts.WorkingLedger.Value().Entries {
		if entry.Complete && !complete[entry.ID] {
			count++
		}
	}

	return count, true
}

func agentsCanonical(facts Facts) []Violation {
	if !facts.Agents.IsPresent() || !facts.Canonical.IsPresent() {
		return nil
	}
	if facts.Agents.Value() == facts.Canonical.Value() {
		return nil
	}

	return []Violation{{
		Rule:    RuleAgentsCanonical,
		Path:    agentsPath,
		Message: "AGENTS.md differs from the canonical AGENTS.md",
	}}
}
