// Package mark finds UNVERIFIED marks in the heading lines of documents.
package mark

import (
	"strings"

	"github.com/replworks/repl-cli/internal/docline"
)

// Text is the UNVERIFIED text.
const Text = "[UNVERIFIED]"

// Mark is one occurrence of the UNVERIFIED text within a heading line.
type Mark struct {
	Path    string
	Line    int
	Heading string
}

// Scanner owns the heading and mark definitions.
type Scanner struct{}

// Scan returns the marks of one document in document order.
func (Scanner) Scan(path, text string) []Mark {
	var marks []Mark

	for _, line := range docline.Classify(text) {
		if line.InsideFence || !isHeading(line.Text) {
			continue
		}

		for range strings.Count(line.Text, Text) {
			marks = append(marks, Mark{Path: path, Line: line.Number, Heading: line.Text})
		}
	}

	return marks
}

// isHeading reports whether a line starts with one to six '#' followed by a
// space.
func isHeading(line string) bool {
	hashes := 0
	for hashes < len(line) && line[hashes] == '#' {
		hashes++
	}

	return hashes >= 1 && hashes <= 6 && hashes < len(line) && line[hashes] == ' '
}
