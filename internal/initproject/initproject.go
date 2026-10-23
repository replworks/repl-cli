// Package initproject creates AGENTS.md and the .replworks directory with its authoritative documents.
package initproject

import (
	"errors"

	"github.com/replworks/repl-cli/internal/fact"
)

const (
	agentsPath       = "AGENTS.md"
	replworksPath    = ".replworks"
	productSpecPath  = ".replworks/PRODUCT_SPEC.md"
	techStackPath    = ".replworks/TECH_STACK.md"
	architecturePath = ".replworks/ARCHITECTURE.md"
	tasksPath        = ".replworks/TASKS.md"
	scriptsPath      = ".replworks/scripts"
	askHumanPath     = ".replworks/scripts/ask-human.sh"
)

// itemPaths is the order in which init processes entries.
var itemPaths = []string{
	agentsPath,
	replworksPath,
	productSpecPath,
	techStackPath,
	architecturePath,
	tasksPath,
}

// CanonicalSource provides the canonical AGENTS.md.
type CanonicalSource interface {
	Canonical() fact.Fact[string]
}

// Writer is the project directory with the Write capability.
type Writer interface {
	Exists(rel string) fact.Fact[bool]
	CreateFile(rel, content string) error
	MakeExecutable(rel string) error
	CreateDirectory(rel string) error
	DeleteFile(rel string) error
	ReplaceFile(rel, content string) error
}

// Item reports one managed entry.
type Item struct {
	Path     string
	Created  bool
	Deleted  bool
	Replaced bool
}

type Outcome struct {
	Items  []Item
	Errors []error
}

type Coordinator struct {
	Source       CanonicalSource
	Script       CanonicalSource
	ManageScript bool
	Workspace    Writer
}

func (c Coordinator) Run() Outcome {
	canonical := c.Source.Canonical()
	if !canonical.IsPresent() {
		return Outcome{Errors: []error{canonicalReason(canonical)}}
	}

	scriptContent := ""
	if c.ManageScript {
		if c.Script == nil {
			return Outcome{Errors: []error{errors.New("the ask-human.sh source is unavailable")}}
		}
		script := c.Script.Canonical()
		if !script.IsPresent() {
			return Outcome{Errors: []error{canonicalReason(script)}}
		}
		scriptContent = script.Value()
	}

	var outcome Outcome

	for _, path := range itemPaths {
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
		switch path {
		case replworksPath:
			err = c.Workspace.CreateDirectory(path)
		case agentsPath:
			err = c.Workspace.CreateFile(path, canonical.Value())
		default:
			err = c.Workspace.CreateFile(path, "")
		}
		if err != nil {
			outcome.Errors = append(outcome.Errors, err)
			return outcome
		}

		outcome.Items = append(outcome.Items, Item{Path: path, Created: true})
	}

	if !c.ManageScript {
		return outcome
	}

	scriptsExists := c.Workspace.Exists(scriptsPath)
	if scriptsExists.IsFailed() {
		outcome.Errors = append(outcome.Errors, scriptsExists.Reason())
		return outcome
	}
	if scriptsExists.Value() {
		outcome.Items = append(outcome.Items, Item{Path: scriptsPath})
	} else {
		if err := c.Workspace.CreateDirectory(scriptsPath); err != nil {
			outcome.Errors = append(outcome.Errors, err)
			return outcome
		}
		outcome.Items = append(outcome.Items, Item{Path: scriptsPath, Created: true})
	}

	scriptExists := c.Workspace.Exists(askHumanPath)
	if scriptExists.IsFailed() {
		outcome.Errors = append(outcome.Errors, scriptExists.Reason())
		return outcome
	}
	if scriptExists.Value() {
		if err := c.Workspace.ReplaceFile(askHumanPath, scriptContent); err != nil {
			outcome.Errors = append(outcome.Errors, err)
			return outcome
		}
		if err := c.Workspace.MakeExecutable(askHumanPath); err != nil {
			outcome.Errors = append(outcome.Errors, err)
			return outcome
		}
		outcome.Items = append(outcome.Items, Item{Path: askHumanPath, Replaced: true})
		return outcome
	}
	if err := c.Workspace.CreateFile(askHumanPath, scriptContent); err != nil {
		outcome.Errors = append(outcome.Errors, err)
		return outcome
	}
	if err := c.Workspace.MakeExecutable(askHumanPath); err != nil {
		outcome.Errors = append(outcome.Errors, err)
		return outcome
	}
	outcome.Items = append(outcome.Items, Item{Path: askHumanPath, Created: true})

	return outcome
}

func canonicalReason(canonical fact.Fact[string]) error {
	if reason := canonical.Reason(); reason != nil {
		return reason
	}

	return errors.New("the canonical AGENTS.md could not be obtained")
}
