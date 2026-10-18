// Package checkproject verifies a project directory against the check rules.
package checkproject

import (
	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/rule"
)

// documents are the authoritative documents, in the order they are read.
var documents = []string{
	"AGENTS.md",
	".replworks/PRODUCT_SPEC.md",
	".replworks/TECH_STACK.md",
	".replworks/ARCHITECTURE.md",
	".replworks/TASKS.md",
}

// Workspace is the project directory with the Read capability.
type Workspace interface {
	ReadText(rel string) fact.Fact[string]
}

// Outcome reports the result of one check invocation.
type Outcome struct {
	Violations []rule.Violation
	Errors     []error
}

type Coordinator struct {
	Workspace Workspace
	Evaluator rule.Evaluator
}

// Run reads every authoritative document once, in order, and evaluates the
// rules over the gathered facts.
func (c Coordinator) Run() Outcome {
	gathered := make([]rule.Document, 0, len(documents))
	var errors []error

	for _, path := range documents {
		text := c.Workspace.ReadText(path)
		if text.IsFailed() {
			errors = append(errors, text.Reason())
		}
		gathered = append(gathered, rule.Document{Path: path, Text: text})
	}

	return Outcome{
		Violations: c.Evaluator.Evaluate(rule.Facts{Documents: gathered}),
		Errors:     errors,
	}
}
