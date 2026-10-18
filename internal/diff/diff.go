// Package diff computes the line-level difference between two texts.
package diff

import (
	"strings"

	"github.com/aymanbagabas/go-udiff"
)

// Kind marks a line as removed from the local text or added by the canonical text.
type Kind int

const (
	Removed Kind = iota
	Added
)

// Entry is one line of the difference. Position is the 1-based line number of
// the line in the text it comes from.
type Entry struct {
	Kind     Kind
	Line     string
	Position int
}

// Calculator computes line-level differences. It is pure and deterministic.
type Calculator struct{}

// Compare returns the difference entries in document order. Removed entries
// precede added entries at the same position. The result is empty only when
// the two texts share the same lines.
func (Calculator) Compare(local, canonical string) []Entry {
	return Compare(local, canonical)
}

// Compare returns the difference entries between local and canonical.
func Compare(local, canonical string) []Entry {
	var entries []Entry

	localCursor, canonicalCursor := 0, 0
	for _, edit := range udiff.Lines(local, canonical) {
		canonicalCursor += edit.Start - localCursor

		for offset, line := range splitLines(local[edit.Start:edit.End]) {
			entries = append(entries, Entry{
				Kind:     Removed,
				Line:     line,
				Position: lineNumber(local, edit.Start) + offset,
			})
		}
		for offset, line := range splitLines(edit.New) {
			entries = append(entries, Entry{
				Kind:     Added,
				Line:     line,
				Position: lineNumber(canonical, canonicalCursor) + offset,
			})
		}

		localCursor = edit.End
		canonicalCursor += len(edit.New)
	}

	return entries
}

func splitLines(text string) []string {
	if text == "" {
		return nil
	}

	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

func lineNumber(text string, offset int) int {
	return 1 + strings.Count(text[:offset], "\n")
}
