// Package ledger recognizes task entries in a TASKS.md text and derives
// progress from them.
package ledger

import (
	"slices"
	"strings"
	"unicode"

	"github.com/replworks/repl-cli/internal/docline"
)

type Entry struct {
	ID       string
	Complete bool
}

type Ledger struct {
	Total             int
	Complete          int
	FirstIncompleteID string
	IDs               []string
	Entries           []Entry
}

func New(text string) Ledger {
	var led Ledger

	for _, line := range docline.Classify(text) {
		id, complete, ok := parseEntry(line)
		if !ok {
			continue
		}

		led.Total++
		led.addID(id)
		led.Entries = append(led.Entries, Entry{ID: id, Complete: complete})
		if complete {
			led.Complete++
			continue
		}
		if led.FirstIncompleteID == "" {
			led.FirstIncompleteID = id
		}
	}

	return led
}

func (l *Ledger) addID(id string) {
	if slices.Contains(l.IDs, id) {
		return
	}

	l.IDs = append(l.IDs, id)
}

func parseEntry(line docline.Line) (string, bool, bool) {
	if line.InsideFence {
		return "", false, false
	}

	text := line.Text
	if len(text) < 6 {
		return "", false, false
	}
	if text[0] != '-' || text[1] != ' ' || text[2] != '[' {
		return "", false, false
	}
	box := text[3]
	if box != ' ' && box != 'x' && box != 'X' {
		return "", false, false
	}
	if text[4] != ']' || text[5] != ' ' {
		return "", false, false
	}

	rest := text[6:]
	if rest == "" {
		return "", false, false
	}
	end := strings.IndexFunc(rest, unicode.IsSpace)
	if end == 0 {
		return "", false, false
	}

	id := rest
	if end > 0 {
		id = rest[:end]
	}

	return id, box != ' ', true
}
