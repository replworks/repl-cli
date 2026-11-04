// Package agentblock recognizes the preserved blocks of an AGENTS.md and
// composes the expected AGENTS.md from the canonical text and those blocks.
package agentblock

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/replworks/repl-cli/internal/docline"
)

var (
	tagOpening     = regexp.MustCompile(`^<([A-Za-z][A-Za-z0-9_-]*)>$`)
	commentOpening = regexp.MustCompile(`^<!-- BEGIN:([A-Za-z0-9][A-Za-z0-9_.-]*) -->$`)
)

// UnclosedError reports a block opening line that no block closing line for
// the same name follows.
type UnclosedError struct {
	Line    int
	Opening string
}

func (e *UnclosedError) Error() string {
	return fmt.Sprintf("AGENTS.md line %d opens %s without a matching closing line; add the closing line or remove the block", e.Line, e.Opening)
}

// Expected returns the expected AGENTS.md: the canonical text followed by each
// preserved block of local, each preceded by one empty line. It returns an
// *UnclosedError when local contains an opening line without a closing line.
func Expected(local, canonical string) (string, error) {
	excluded := map[string]bool{}
	for _, line := range docline.Classify(canonical) {
		if _, ok := closingFor(line); ok {
			excluded[line.Text] = true
		}
	}

	blocks, err := preservedBlocks(local, excluded)
	if err != nil {
		return "", err
	}
	if len(blocks) == 0 {
		return canonical, nil
	}

	var expected strings.Builder
	expected.WriteString(canonical)
	if !strings.HasSuffix(canonical, "\n") {
		expected.WriteString("\n")
	}
	for _, block := range blocks {
		expected.WriteString("\n")
		expected.WriteString(block)
		expected.WriteString("\n")
	}

	return expected.String(), nil
}

func preservedBlocks(local string, excluded map[string]bool) ([]string, error) {
	raw := strings.Split(local, "\n")

	var blocks []string
	var opening docline.Line
	closing := ""
	for _, line := range docline.Classify(local) {
		if line.InsideFence {
			continue
		}
		if closing == "" {
			if excluded[line.Text] {
				continue
			}
			if want, ok := closingFor(line); ok {
				opening, closing = line, want
			}
			continue
		}
		if line.Text == closing {
			blocks = append(blocks, strings.Join(raw[opening.Number-1:line.Number], "\n"))
			closing = ""
		}
	}
	if closing != "" {
		return nil, &UnclosedError{Line: opening.Number, Opening: opening.Text}
	}

	return blocks, nil
}

// closingFor returns the closing line for a block opening line.
func closingFor(line docline.Line) (string, bool) {
	if line.InsideFence {
		return "", false
	}
	if match := tagOpening.FindStringSubmatch(line.Text); match != nil {
		return "</" + match[1] + ">", true
	}
	if match := commentOpening.FindStringSubmatch(line.Text); match != nil {
		return "<!-- END:" + match[1] + " -->", true
	}

	return "", false
}
