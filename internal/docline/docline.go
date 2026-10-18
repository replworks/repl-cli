// Package docline splits a document into numbered lines and tags each line as
// inside or outside a fenced code block.
package docline

import "strings"

type Line struct {
	Number      int
	Text        string
	InsideFence bool
}

type fence struct {
	char  byte
	count int
}

func Classify(text string) []Line {
	parts := strings.Split(text, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	lines := make([]Line, 0, len(parts))
	var open *fence
	for i, part := range parts {
		line := strings.TrimSuffix(part, "\r")
		inside := open != nil
		if open == nil {
			if opened, ok := openingFence(line); ok {
				open = &opened
				inside = true
			}
		} else if closesFence(line, *open) {
			open = nil
		}

		lines = append(lines, Line{Number: i + 1, Text: line, InsideFence: inside})
	}

	return lines
}

func openingFence(line string) (fence, bool) {
	rest, ok := stripIndent(line)
	if !ok || len(rest) < 3 {
		return fence{}, false
	}

	char := rest[0]
	if char != '`' && char != '~' {
		return fence{}, false
	}

	count := runLength(rest, char)
	if count < 3 {
		return fence{}, false
	}

	return fence{char: char, count: count}, true
}

func closesFence(line string, open fence) bool {
	rest, ok := stripIndent(line)
	if !ok {
		return false
	}

	count := runLength(rest, open.char)
	if count < open.count {
		return false
	}

	return strings.TrimSpace(rest[count:]) == ""
}

func stripIndent(line string) (string, bool) {
	indent := 0
	for indent < len(line) && line[indent] == ' ' {
		indent++
	}
	if indent > 3 {
		return "", false
	}

	return line[indent:], true
}

func runLength(line string, char byte) int {
	count := 0
	for count < len(line) && line[count] == char {
		count++
	}

	return count
}
