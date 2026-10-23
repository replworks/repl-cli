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

	upToDate := local.Value() == canonical.Value()
	var difference []diff.Entry
	if !upToDate {
		difference = c.Difference.Compare(local.Value(), canonical.Value())
	}

	scriptContent := ""
	if c.ManageScript {
		if c.Script == nil {
			return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{errors.New("the ask-human.sh source is unavailable")}}
		}
		script := c.Script.Canonical()
		if !script.IsPresent() {
			return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{canonicalReason(script)}}
		}
		scriptContent = script.Value()
	}
	if !upToDate {
		if err := c.Workspace.ReplaceFile(agentsPath, canonical.Value()); err != nil {
			return Outcome{Difference: difference, Errors: []error{err}}
		}
	}

	if c.ManageScript {
		return c.writeScript(upToDate, difference, scriptContent)
	}

	return Outcome{UpToDate: upToDate, Difference: difference}
}

func (c Coordinator) writeScript(upToDate bool, difference []diff.Entry, content string) Outcome {
	exists := c.Workspace.Exists(askHumanPath)
	if exists.IsFailed() {
		return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{exists.Reason()}}
	}

	if !exists.Value() {
		directory := c.Workspace.Exists(scriptsPath)
		if directory.IsFailed() {
			return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{directory.Reason()}}
		}
		if !directory.Value() {
			root := c.Workspace.Exists(replworksPath)
			if root.IsFailed() {
				return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{root.Reason()}}
			}
			if !root.Value() {
				if err := c.Workspace.CreateDirectory(replworksPath); err != nil {
					return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{err}}
				}
			}
			if err := c.Workspace.CreateDirectory(scriptsPath); err != nil {
				return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{err}}
			}
		}
	}
	state := "replaced"
	if exists.Value() {
		if err := c.Workspace.ReplaceFile(askHumanPath, content); err != nil {
			return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{err}}
		}
	} else {
		state = "created"
		if err := c.Workspace.CreateFile(askHumanPath, content); err != nil {
			return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{err}}
		}
	}
	if err := c.Workspace.MakeExecutable(askHumanPath); err != nil {
		return Outcome{UpToDate: upToDate, Difference: difference, Errors: []error{err}}
	}

	return Outcome{UpToDate: upToDate, Difference: difference, ScriptPath: askHumanPath, ScriptState: state}
}

func canonicalReason(canonical fact.Fact[string]) error {
	if reason := canonical.Reason(); reason != nil {
		return reason
	}

	return errors.New("the canonical AGENTS.md could not be obtained")
}
