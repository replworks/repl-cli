// Package rule evaluates the check rules over gathered facts and owns the rule
// identifiers, their fixed evaluation order, and their violation messages.
package rule

import "github.com/replworks/repl-cli/internal/fact"

// RuleDocumentsExist is reported for each authoritative document that does not
// exist.
const RuleDocumentsExist = "documents-exist"

// RuleAgentsCanonical is reported for an AGENTS.md that differs from the
// canonical AGENTS.md.
const RuleAgentsCanonical = "agents-canonical"

const agentsPath = "AGENTS.md"

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
	Documents []Document
	Agents    fact.Fact[string]
	Canonical fact.Fact[string]
	Base      fact.Fact[string]
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
var order = []func(Facts) []Violation{documentsExist, agentsCanonical}

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
