// Package initproject creates AGENTS.md and the .replworks directory.
package initproject

import (
	"errors"

	"github.com/replworks/repl-cli/internal/fact"
)

const (
	agentsPath    = "AGENTS.md"
	replworksPath = ".replworks"
)

// CanonicalSource provides the canonical AGENTS.md.
type CanonicalSource interface {
	Canonical() fact.Fact[string]
}

// Writer is the project directory with the Write capability.
type Writer interface {
	Exists(rel string) fact.Fact[bool]
	CreateFile(rel, content string) error
	CreateDirectory(rel string) error
}

// Item reports one managed entry.
type Item struct {
	Path    string
	Created bool
}

type Outcome struct {
	Items  []Item
	Errors []error
}

type Coordinator struct {
	Source    CanonicalSource
	Workspace Writer
}

func (c Coordinator) Run() Outcome {
	canonical := c.Source.Canonical()
	if !canonical.IsPresent() {
		return Outcome{Errors: []error{canonicalReason(canonical)}}
	}

	var outcome Outcome

	for _, path := range []string{agentsPath, replworksPath} {
		exists := c.Workspace.Exists(path)
		if exists.IsFailed() {
			outcome.Errors = append(outcome.Errors, exists.Reason())
			return outcome
		}

		if exists.Value() {
			outcome.Items = append(outcome.Items, Item{Path: path})
			continue
		}

		var err error
		if path == replworksPath {
			err = c.Workspace.CreateDirectory(path)
		} else {
			err = c.Workspace.CreateFile(path, canonical.Value())
		}
		if err != nil {
			outcome.Errors = append(outcome.Errors, err)
			return outcome
		}

		outcome.Items = append(outcome.Items, Item{Path: path, Created: true})
	}

	return outcome
}

func canonicalReason(canonical fact.Fact[string]) error {
	if reason := canonical.Reason(); reason != nil {
		return reason
	}

	return errors.New("the canonical AGENTS.md could not be obtained")
}
