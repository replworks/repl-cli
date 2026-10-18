// Package syncproject replaces AGENTS.md with the canonical AGENTS.md.
package syncproject

import (
	"errors"
	"fmt"

	"github.com/replworks/repl-cli/internal/diff"
	"github.com/replworks/repl-cli/internal/fact"
)

const agentsPath = "AGENTS.md"

// CanonicalSource provides the canonical AGENTS.md.
type CanonicalSource interface {
	Canonical() fact.Fact[string]
}

// Workspace is the project directory with the Read and Write capabilities.
type Workspace interface {
	ReadText(rel string) fact.Fact[string]
	ReplaceFile(rel, content string) error
}

// DifferenceCalculator computes the line-level difference between two texts.
type DifferenceCalculator interface {
	Compare(local, canonical string) []diff.Entry
}

// Outcome reports the result of one sync invocation. It is exactly one of:
// up to date (UpToDate), replaced (Difference), or replacement failed
// (Difference together with Errors).
type Outcome struct {
	UpToDate   bool
	Difference []diff.Entry
	Errors     []error
}

type Coordinator struct {
	Source     CanonicalSource
	Workspace  Workspace
	Difference DifferenceCalculator
}

func (c Coordinator) Run() Outcome {
	local := c.Workspace.ReadText(agentsPath)

	if local.IsAbsent() {
		return Outcome{Errors: []error{fmt.Errorf("%s does not exist", agentsPath)}}
	}
	if local.IsFailed() {
		return Outcome{Errors: []error{local.Reason()}}
	}

	canonical := c.Source.Canonical()
	if !canonical.IsPresent() {
		return Outcome{Errors: []error{canonicalReason(canonical)}}
	}

	if local.Value() == canonical.Value() {
		return Outcome{UpToDate: true}
	}

	difference := c.Difference.Compare(local.Value(), canonical.Value())
	if err := c.Workspace.ReplaceFile(agentsPath, canonical.Value()); err != nil {
		return Outcome{Difference: difference, Errors: []error{err}}
	}

	return Outcome{Difference: difference}
}

func canonicalReason(canonical fact.Fact[string]) error {
	if reason := canonical.Reason(); reason != nil {
		return reason
	}

	return errors.New("the canonical AGENTS.md could not be obtained")
}
