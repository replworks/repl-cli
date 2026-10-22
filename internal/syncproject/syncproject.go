// Package syncproject replaces AGENTS.md with the canonical AGENTS.md.
package syncproject

import (
	"errors"
	"fmt"

	"github.com/replworks/repl-cli/internal/diff"
	"github.com/replworks/repl-cli/internal/fact"
)

const (
	agentsPath    = "AGENTS.md"
	replworksPath = ".replworks"
	scriptsPath   = ".replworks/scripts"
	askHumanPath  = ".replworks/scripts/ask-human.sh"
)

// CanonicalSource provides the canonical AGENTS.md.
type CanonicalSource interface {
	Canonical() fact.Fact[string]
}

// Workspace is the project directory with the Read and Write capabilities.
type Workspace interface {
	ReadText(rel string) fact.Fact[string]
	Exists(rel string) fact.Fact[bool]
	CreateDirectory(rel string) error
	CreateFile(rel, content string) error
	MakeExecutable(rel string) error
	ReplaceFile(rel, content string) error
	DeleteFile(rel string) error
}

// DifferenceCalculator computes the line-level difference between two texts.
type DifferenceCalculator interface {
	Compare(local, canonical string) []diff.Entry
}

// Outcome reports the result of one sync invocation. It is exactly one of:
// up to date (UpToDate), replaced (Difference), or replacement failed
// (Difference together with Errors).
type Outcome struct {
	UpToDate    bool
	Difference  []diff.Entry
	ScriptPath  string
	ScriptState string
	Errors      []error
}

type Coordinator struct {
	Source       CanonicalSource
	Script       CanonicalSource
	Auto         bool
	ManageScript bool
	Workspace    Workspace
	Difference   DifferenceCalculator
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

	scriptExists := false
	scriptContent := ""
	if c.ManageScript {
		exists := c.Workspace.Exists(askHumanPath)
		if exists.IsFailed() {
			return Outcome{Difference: difference, Errors: []error{exists.Reason()}}
		}
		scriptExists = exists.Value()
	}
	if c.ManageScript && c.Auto && c.Script != nil && !scriptExists {
		script := c.Script.Canonical()
		if !script.IsPresent() {
			return Outcome{Difference: difference, Errors: []error{canonicalReason(script)}}
		}
		scriptContent = script.Value()
	}
	if err := c.Workspace.ReplaceFile(agentsPath, canonical.Value()); err != nil {
		return Outcome{Difference: difference, Errors: []error{err}}
	}

	if c.ManageScript && !c.Auto {
		if scriptExists {
			if err := c.Workspace.DeleteFile(askHumanPath); err != nil {
				return Outcome{Difference: difference, Errors: []error{err}}
			}
			return Outcome{Difference: difference, ScriptPath: askHumanPath, ScriptState: "deleted"}
		}
	}
	if c.ManageScript && c.Auto && !scriptExists {
		return c.ensureScript(difference, scriptContent)
	}
	if c.ManageScript && c.Auto {
		if err := c.Workspace.MakeExecutable(askHumanPath); err != nil {
			return Outcome{Difference: difference, Errors: []error{err}}
		}
	}

	return Outcome{Difference: difference}
}

func (c Coordinator) ensureScript(difference []diff.Entry, content string) Outcome {
	exists := c.Workspace.Exists(askHumanPath)
	if exists.IsFailed() {
		return Outcome{Difference: difference, Errors: []error{exists.Reason()}}
	}
	if exists.Value() {
		return Outcome{Difference: difference}
	}

	directory := c.Workspace.Exists(scriptsPath)
	if directory.IsFailed() {
		return Outcome{Difference: difference, Errors: []error{directory.Reason()}}
	}
	if !directory.Value() {
		root := c.Workspace.Exists(replworksPath)
		if root.IsFailed() {
			return Outcome{Difference: difference, Errors: []error{root.Reason()}}
		}
		if !root.Value() {
			if err := c.Workspace.CreateDirectory(replworksPath); err != nil {
				return Outcome{Difference: difference, Errors: []error{err}}
			}
		}
		if err := c.Workspace.CreateDirectory(scriptsPath); err != nil {
			return Outcome{Difference: difference, Errors: []error{err}}
		}
	}
	if err := c.Workspace.CreateFile(askHumanPath, content); err != nil {
		return Outcome{Difference: difference, Errors: []error{err}}
	}
	if err := c.Workspace.MakeExecutable(askHumanPath); err != nil {
		return Outcome{Difference: difference, Errors: []error{err}}
	}

	return Outcome{Difference: difference, ScriptPath: askHumanPath, ScriptState: "created"}
}

func canonicalReason(canonical fact.Fact[string]) error {
	if reason := canonical.Reason(); reason != nil {
		return reason
	}

	return errors.New("the canonical AGENTS.md could not be obtained")
}
