// Package checkproject verifies a project directory against the check rules.
package checkproject

import (
	"errors"

	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/ledger"
	"github.com/replworks/repl-cli/internal/mark"
	"github.com/replworks/repl-cli/internal/rule"
)

const (
	agentsPath    = "AGENTS.md"
	tasksPath     = ".replworks/TASKS.md"
	productSpec   = ".replworks/PRODUCT_SPEC.md"
	techStackPath = ".replworks/TECH_STACK.md"
	architecture  = ".replworks/ARCHITECTURE.md"
)

// documents are the authoritative documents, in the order they are read.
var documents = []string{
	agentsPath,
	productSpec,
	techStackPath,
	architecture,
	tasksPath,
}

// markDocuments are the documents that may carry UNVERIFIED marks, in the
// order the Rule Evaluator reports them.
var markDocuments = []string{productSpec, techStackPath, architecture}

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
	texts := make(map[string]fact.Fact[string], len(documents))
	var errors []error

	for _, path := range documents {
		text := c.Workspace.ReadText(path)
		if text.IsFailed() {
			errors = append(errors, text.Reason())
		}
		texts[path] = text
		gathered = append(gathered, rule.Document{Path: path, Text: text})
	}

	markLists := make([]rule.MarkList, 0, len(markDocuments))
	for _, path := range markDocuments {
		markLists = append(markLists, rule.MarkList{Path: path, Marks: marksFact(path, texts[path])})
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
			Documents:     gathered,
			Agents:        texts[agentsPath],
			Canonical:     canonical,
			WorkingLedger: ledgerFact(texts[tasksPath]),
			BaseLedger:    ledgerFact(base),
			MarkLists:     markLists,
		}),
		Errors: errors,
	}
}

func ledgerFact(text fact.Fact[string]) fact.Fact[ledger.Ledger] {
	switch {
	case text.IsPresent():
		return fact.Present(ledger.New(text.Value()))
	case text.IsFailed():
		return fact.Failed[ledger.Ledger](text.Reason())
	default:
		return fact.Absent[ledger.Ledger]()
	}
}

func marksFact(path string, text fact.Fact[string]) fact.Fact[[]mark.Mark] {
	switch {
	case text.IsPresent():
		return fact.Present(mark.Scanner{}.Scan(path, text.Value()))
	case text.IsFailed():
		return fact.Failed[[]mark.Mark](text.Reason())
	default:
		return fact.Absent[[]mark.Mark]()
	}
}

func canonicalReason(canonical fact.Fact[string]) error {
	if reason := canonical.Reason(); reason != nil {
		return reason
	}

	return errors.New("the canonical AGENTS.md could not be obtained")
}
