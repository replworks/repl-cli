// Package skillsproject installs the five Agent Skills into the project
// directory.
package skillsproject

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/skill"
)

// SkillSource obtains the published skill file at url.
type SkillSource interface {
	Fetch(url string) fact.Fact[string]
}

// Workspace is the project directory with the Read and Write capabilities,
// restricted to `.claude/skills/` paths and their parent directories.
type Workspace interface {
	ReadText(rel string) fact.Fact[string]
	Exists(rel string) fact.Fact[bool]
	CreateDirectory(rel string) error
	CreateFile(rel, content string) error
	ReplaceFile(rel, content string) error
}

// SkillState is the outcome for one skill.
type SkillState string

const (
	// Created reports that the skill file did not exist and was created.
	Created SkillState = "created"
	// Replaced reports that the skill file existed with different content
	// and was replaced.
	Replaced SkillState = "replaced"
	// UpToDate reports that the skill file was left untouched.
	UpToDate SkillState = "up to date"
)

// SkillResult is the outcome for one skill.
type SkillResult struct {
	Name  string
	Path  string
	State SkillState
}

// Outcome reports the result of one invocation.
type Outcome struct {
	Items  []SkillResult
	Errors []error
}

type Coordinator struct {
	Source    SkillSource
	Workspace Workspace
}

// Run gathers the five skill files before writing any of them, then installs
// or refreshes the five skill files in the fixed order.
func (c Coordinator) Run() Outcome {
	facts := make([]fact.Fact[string], len(skill.Sources))
	var errors []error
	for i, src := range skill.Sources {
		facts[i] = c.Source.Fetch(src.URL)
		if facts[i].IsFailed() {
			errors = append(errors, skillError(src.Name, facts[i]))
		}
	}
	if len(errors) > 0 {
		return Outcome{Errors: errors}
	}

	var items []SkillResult
	for i, src := range skill.Sources {
		local := c.Workspace.ReadText(src.Path)
		switch {
		case local.IsFailed():
			return Outcome{Items: items, Errors: []error{local.Reason()}}
		case local.IsPresent() && local.Value() == facts[i].Value():
			items = append(items, SkillResult{Name: src.Name, Path: src.Path, State: UpToDate})
		case local.IsPresent():
			if err := c.Workspace.ReplaceFile(src.Path, facts[i].Value()); err != nil {
				return Outcome{Items: items, Errors: []error{err}}
			}
			items = append(items, SkillResult{Name: src.Name, Path: src.Path, State: Replaced})
		default:
			if err := c.install(src, facts[i].Value()); err != nil {
				return Outcome{Items: items, Errors: []error{err}}
			}
			items = append(items, SkillResult{Name: src.Name, Path: src.Path, State: Created})
		}
	}

	return Outcome{Items: items}
}

func (c Coordinator) install(src skill.Source, content string) error {
	for _, dir := range []string{".claude", ".claude/skills", filepath.Dir(src.Path)} {
		exists := c.Workspace.Exists(dir)
		switch {
		case exists.IsFailed():
			return exists.Reason()
		case exists.Value():
			continue
		default:
			if err := c.Workspace.CreateDirectory(dir); err != nil {
				return err
			}
		}
	}

	return c.Workspace.CreateFile(src.Path, content)
}

func skillError(name string, fact fact.Fact[string]) error {
	if reason := fact.Reason(); reason != nil {
		return fmt.Errorf("%s: %w", name, reason)
	}

	return errors.New(name + ": the skill file could not be obtained")
}
