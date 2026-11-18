// Package nextproject prints the next task and the definitions it references.
package nextproject

import (
	"fmt"

	"github.com/replworks/repl-cli/internal/definition"
	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/ledger"
	"github.com/replworks/repl-cli/internal/showproject"
	"github.com/replworks/repl-cli/internal/workspace"
)

type Outcome struct {
	Task        *definition.Excerpt
	Definitions []definition.Excerpt
	Missing     []string
	Errors      []error
}

type Coordinator struct {
	Workspace workspace.Workspace
}

func (c Coordinator) Run() Outcome {
	tasks := c.Workspace.ReadText(definition.TasksPath)
	if tasks.IsAbsent() {
		return Outcome{Errors: []error{fmt.Errorf("%s does not exist", definition.TasksPath)}}
	}
	if tasks.IsFailed() {
		return Outcome{Errors: []error{tasks.Reason()}}
	}

	id := ledger.New(tasks.Value()).FirstIncompleteID
	if id == "" {
		return Outcome{}
	}
	task, _ := definition.FindTask(id, tasks.Value())
	outcome := Outcome{Task: &task}

	needed := map[string]bool{}
	for _, ref := range definition.References(task.Text(), id, nil) {
		if path := definition.DocumentOf(ref); path != definition.TasksPath {
			needed[path] = true
		}
	}
	docs, errs := showproject.ReadDocuments(c.Workspace, needed)
	docs[definition.TasksPath] = tasks
	outcome.Errors = errs

	for _, ref := range definition.References(task.Text(), id, presentTexts(docs)) {
		excerpt, found, known := showproject.Lookup(ref, docs)
		switch {
		case found:
			outcome.Definitions = append(outcome.Definitions, excerpt)
		case known:
			outcome.Missing = append(outcome.Missing, ref)
		}
	}

	return outcome
}

func presentTexts(docs map[string]fact.Fact[string]) map[string]string {
	texts := map[string]string{}
	for path, text := range docs {
		if text.IsPresent() {
			texts[path] = text.Value()
		}
	}

	return texts
}
