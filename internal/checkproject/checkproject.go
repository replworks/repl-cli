// Package checkproject verifies a project directory against the check rules.
package checkproject

import (
	"errors"

	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/rule"
)

const agentsPath = "AGENTS.md"

// documents are the authoritative documents, in the order they are read.
var documents = []string{
	agentsPath,
	".replworks/PRODUCT_SPEC.md",
	".replworks/TECH_STACK.md",
	".replworks/ARCHITECTURE.md",
	".replworks/TASKS.md",
}

// Workspace is the project directory with the Read capability.
type Workspace interface {
	ReadText(rel string) fact.Fact[string]
}

// CanonicalSource provides the canonical AGENTS.md.
type CanonicalSource interface {
	Canonical() fact.Fact[string]
}

// BaseBranchReader provides the TASKS.md of the base branch.
type BaseBranchReader interface {
	Tasks(branch string) fact.Fact[string]
}

// Outcome reports the result of one check invocation.
type Outcome struct {
	Violations []rule.Violation
	Errors     []error
}

type Coordinator struct {
	Workspace Workspace
	Source    CanonicalSource
	Bases     BaseBranchReader
	Evaluator rule.Evaluator
}

// Run reads every authoritative document once, in order, obtains the canonical
// text, reads the base TASKS.md, and evaluates the rules over the gathered
// facts. The base branch is always read, even when TASKS.md is absent in the
// working tree.
func (c Coordinator) Run(branch string) Outcome {
	gathered := make([]rule.Document, 0, len(documents))
	var agents fact.Fact[string]
	var errors []error

	for _, path := range documents {
		text := c.Workspace.ReadText(path)
		if text.IsFailed() {
			errors = append(errors, text.Reason())
		}
		if path == agentsPath {
			agents = text
		}
		gathered = append(gathered, rule.Document{Path: path, Text: text})
	}

	canonical := c.Source.Canonical()
	if canonical.IsFailed() {
		errors = append(errors, canonicalReason(canonical))
	}

	base := c.Bases.Tasks(branch)
	if base.IsFailed() {
		errors = append(errors, base.Reason())
	}

	return Outcome{
		Violations: c.Evaluator.Evaluate(rule.Facts{
			Documents: gathered,
			Agents:    agents,
			Canonical: canonical,
			Base:      base,
		}),
		Errors: errors,
	}
}

func canonicalReason(canonical fact.Fact[string]) error {
	if reason := canonical.Reason(); reason != nil {
		return reason
	}

	return errors.New("the canonical AGENTS.md could not be obtained")
}
