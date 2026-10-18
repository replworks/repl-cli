// Package status reports task progress read from .replworks/TASKS.md.
package status

import (
	"fmt"

	"github.com/replworks/repl-cli/internal/ledger"
	"github.com/replworks/repl-cli/internal/workspace"
)

const tasksPath = ".replworks/TASKS.md"

type Outcome struct {
	Total             int
	Complete          int
	FirstIncompleteID string
	Errors            []error
}

type Coordinator struct {
	Workspace workspace.Workspace
}

func (c Coordinator) Run() Outcome {
	text := c.Workspace.ReadText(tasksPath)

	if text.IsAbsent() {
		return Outcome{Errors: []error{fmt.Errorf("%s does not exist", tasksPath)}}
	}
	if text.IsFailed() {
		return Outcome{Errors: []error{text.Reason()}}
	}

	led := ledger.New(text.Value())

	return Outcome{
		Total:             led.Total,
		Complete:          led.Complete,
		FirstIncompleteID: led.FirstIncompleteID,
	}
}
