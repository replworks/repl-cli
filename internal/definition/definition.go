// Package definition locates identifier definitions in the authoritative
// documents and extracts their definition excerpts.
package definition

import (
	"regexp"
	"strings"

	"github.com/replworks/repl-cli/internal/docline"
	"github.com/replworks/repl-cli/internal/ledger"
)

const (
	ProductSpecPath  = ".replworks/PRODUCT_SPEC.md"
	ArchitecturePath = ".replworks/ARCHITECTURE.md"
	TasksPath        = ".replworks/TASKS.md"
)

// DocumentOrder is the fixed order in which defining documents are read.
var DocumentOrder = []string{ProductSpecPath, ArchitecturePath, TasksPath}

var (
	identifierPattern = regexp.MustCompile(`^(?:T-[0-9]+|(?:FR|EL|ERR|AC)-[0-9]{3}|C[0-9]+|AR-[0-9]+)$`)
	sectionHeading    = regexp.MustCompile(`^#{1,3} `)
)

// Excerpt is the definition excerpt of one identifier.
type Excerpt struct {
	ID    string
	Path  string
	Line  int
	Lines []string
}

// Text joins the excerpt lines with a line feed, without a trailing line break.
func (e Excerpt) Text() string {
	return strings.Join(e.Lines, "\n")
}

// Valid reports whether id has one of the identifier forms.
func Valid(id string) bool {
	return identifierPattern.MatchString(id)
}

// DocumentOf returns the defining document of a valid identifier.
func DocumentOf(id string) string {
	switch {
	case strings.HasPrefix(id, "T-"):
		return TasksPath
	case strings.HasPrefix(id, "C"), strings.HasPrefix(id, "AR-"):
		return ArchitecturePath
	default:
		return ProductSpecPath
	}
}

// Find returns the excerpt of the first definition line of id in text, the
// text of its defining document.
func Find(id, text string) (Excerpt, bool) {
	return find(id, DocumentOf(id), text)
}

// FindTask returns the excerpt of the first Task Entry with the task ID id,
// whatever the form of that task ID.
func FindTask(id, tasksText string) (Excerpt, bool) {
	return find(id, TasksPath, tasksText)
}

func find(id, path, text string) (Excerpt, bool) {
	lines := docline.Classify(text)
	for i, line := range lines {
		if line.InsideFence || !isDefinition(id, path, line) {
			continue
		}

		excerpt := Excerpt{ID: id, Path: path, Line: line.Number, Lines: []string{line.Text}}
		if path == ArchitecturePath && strings.HasPrefix(id, "C") {
			excerpt.Lines = append(excerpt.Lines, sectionBody(lines[i+1:])...)
		} else {
			excerpt.Lines = append(excerpt.Lines, continuation(lines[i+1:])...)
		}

		return excerpt, true
	}

	return Excerpt{}, false
}

func isDefinition(id, path string, line docline.Line) bool {
	switch {
	case path == TasksPath:
		entryID, _, ok := ledger.ParseEntry(line)
		return ok && entryID == id
	case path == ArchitecturePath && strings.HasPrefix(id, "C"):
		return strings.HasPrefix(line.Text, "### "+id+". ")
	default:
		return strings.HasPrefix(line.Text, "- "+id+": ")
	}
}

func continuation(lines []docline.Line) []string {
	var body []string
	for _, line := range lines {
		if !strings.HasPrefix(line.Text, " ") && !strings.HasPrefix(line.Text, "\t") {
			break
		}
		body = append(body, line.Text)
	}

	return body
}

func sectionBody(lines []docline.Line) []string {
	var body []string
	for _, line := range lines {
		if !line.InsideFence && sectionHeading.MatchString(line.Text) {
			break
		}
		body = append(body, line.Text)
	}

	for len(body) > 0 && body[len(body)-1] == "" {
		body = body[:len(body)-1]
	}

	return body
}
